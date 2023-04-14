package main

import (
	"bufio"
	"bytes"
	"cache"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type LimitOrder struct {
	Stock  string
	Price  float64
	Type   string
	Amount float64
	User   string `json:"ID"`
	Qty    float64
}

type req struct {
	Sym      string `json:"Sym"`
	Username string `json:"Username"`
}

type quote_hit struct {
	Timestamp int     `json:"Timestamp"`
	Price     float64 `json:"Price"`
	Cryptokey string  `json:"Cryptokey"`
}

type logQSHit struct {
	Id        string  `json:"id"`
	Sym       string  `json:"sym"`
	Timestamp int     `json:"timestamp"`
	Price     float64 `json:"price"`
	Cryptokey string  `json:"cryptokey"`
}

var active_orders []LimitOrder

func main() {
	quoteServer, found := os.LookupEnv("QUOTE_SERVER")
	if !found {
		log.Fatalln("No QUOTE_SERVER")
	}

	transactionService, found := os.LookupEnv("TRANSACTION_SERVICE")
	if !found {
		log.Fatalln("No TRANSACTION_SERVICE")
	}
	router := gin.Default() // initializing Gin router
	router.SetTrustedProxies(nil)

	router.Use(func(ctx *gin.Context) {
		ctx.Set("quoteServer", quoteServer)
		ctx.Set("transactionService", transactionService)
		ctx.Next()
	})

	router.POST("/new_limit", new_limit)
	router.POST("/quote", get_price)
	bind := flag.String("bind", "localhost:8081", "host:port to listen on")
	flag.Parse()

	if err := router.Run(*bind); err != nil {
		panic(err)
	}

}

func quote_price(servAddr string, sym string, username string) (quote_hit, error) {
	strEcho := sym + " " + username + "\n"

	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		fmt.Println("\nResolveTCPAddr error: ", err)
		return quote_hit{}, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("\nDialTCP error: ", err)
		return quote_hit{}, err
	}

	defer conn.Close()

	//write to server SYM being requested and user
	_, err = conn.Write([]byte(strEcho))
	if err != nil {
		fmt.Println("\nWrite error: ", err)
		return quote_hit{}, err
	}

	reader := bufio.NewReader(conn)
	replyLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("\nRead error: ", err)
		return quote_hit{}, err
	}

	//parsing reply from server
	reply := strings.Split(strings.TrimRight(replyLine, "\n"), ",")
	quotePrice, err := strconv.ParseFloat(reply[0], 64)
	if err != nil {
		return quote_hit{}, err
	}
	timestamp, err := strconv.Atoi(reply[3])
	if err != nil {
		return quote_hit{}, err
	}
	cryptKey := reply[4]

	return quote_hit{
		Price: quotePrice,
		Timestamp: timestamp,
		Cryptokey: cryptKey,
	}, nil
}

func get_price(c *gin.Context) {
	var quote_req req
	if err := c.BindJSON(&quote_req); err != nil {
		c.IndentedJSON(http.StatusOK, err)
		return
	}

	q, err := quote_price(c.MustGet("quoteServer").(string), quote_req.Sym, quote_req.Username)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}

	cache.SetKeyWithExpirationInSecs(quote_req.Sym, q.Price, 0)

	c.IndentedJSON(http.StatusOK, q)
}

func do_limit_order(quoteServer string, transactionService string) {
	j := 0
	for len(active_orders) > 0 {
		// do: update cache
		val, err := quote_price(quoteServer, active_orders[j].Stock, active_orders[j].User)

		if err != nil {
			// Logging quote server hit
			logQSHit_ := logQSHit{Id: active_orders[j].User, Sym: active_orders[j].Stock, Timestamp: val.Timestamp, Price: val.Price, Cryptokey: val.Cryptokey}
			parsedJson, _ := json.Marshal(logQSHit_)
			_, err = http.NewRequest(http.MethodPost, transactionService+"/log_qs_hit", bytes.NewBuffer(parsedJson))
			if err != nil {
				fmt.Println("ERROR")
				log.Fatal(err)
			}

			if val.Price > active_orders[j].Price && active_orders[j].Type == "sell" {
				cache.SetKeyWithExpirationInSecs(active_orders[j].Stock, val.Price, 0)
				active_orders[j].Qty = active_orders[j].Amount

				parsedJson, _ := json.Marshal(active_orders[j])
				req, err := http.NewRequest(http.MethodPost, transactionService+"/users/sell", bytes.NewBuffer(parsedJson))
				_, err = http.DefaultClient.Do(req)

				req, err = http.NewRequest(http.MethodPost, transactionService+"/users/sell/commit", bytes.NewBuffer(parsedJson))
				_, err = http.DefaultClient.Do(req)

				if err != nil {
					fmt.Println("ERROR")
					fmt.Println(err)
				}

				active_orders = append(active_orders[:j], active_orders[j+1:]...)

			} else if val.Price < active_orders[j].Price && active_orders[j].Type == "buy" {
				cache.SetKeyWithExpirationInSecs(active_orders[j].Stock, val.Price, 0)
				active_orders[j].Qty = active_orders[j].Amount

				parsedJson, _ := json.Marshal(active_orders[j])
				req, err := http.NewRequest(http.MethodPost, transactionService+"/users/buy", bytes.NewBuffer(parsedJson))
				res, err := http.DefaultClient.Do(req)

				_, err = ioutil.ReadAll(res.Body)
				if err != nil {
					fmt.Println("ERROR")
					fmt.Println(err)
				}

				req, err = http.NewRequest(http.MethodPost, transactionService+"/users/buy/commit", bytes.NewBuffer(parsedJson))
				res, err = http.DefaultClient.Do(req)
				_, err = ioutil.ReadAll(res.Body)
				if err != nil {
					fmt.Println("ERROR")
					fmt.Println(err)
				}

				active_orders = append(active_orders[:j], active_orders[j+1:]...)
			}
		}

		time.Sleep(1 * time.Second) // Math goes here
		if len(active_orders) < 1 {
			return
		}
		j = (j + 1) % len(active_orders)
	}

	return
}

func new_limit(c *gin.Context) {
	quoteServer := c.MustGet("quoteServer").(string)
	transactionService := c.MustGet("transactionService").(string)

	var limitorder LimitOrder
	if err := c.BindJSON(&limitorder); err != nil {
		c.IndentedJSON(http.StatusOK, err)
		return
	}

	c.IndentedJSON(http.StatusOK, "ok")

	active_orders = append(active_orders, limitorder)
	if len(active_orders) == 1 {
		go do_limit_order(quoteServer, transactionService)
	} else {
		return
	}

}
