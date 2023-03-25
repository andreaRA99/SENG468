package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// mock db, actual requests will be sent to a Mongo DB
type account struct {
	ID      string  `json:"id"`
	Balance float64 `json:"balance"`
}

var accounts = []account{
	{ID: "1", Balance: 100},
	{ID: "2", Balance: 200},
	{ID: "3", Balance: 300},
}

type holding struct {
	symbol   string
	quantity float64
	pps      float64
}

type balanceDif struct {
	//ID     string
	Amount float64 `json:"amount"`
}

type users struct {
	user_id string
}

// Not used.  There is supposed to be a way to read mongo db stuff directly into struct, but I coldnt get it to work.
type c_bal struct {
	cash_balance int32
}

type quote struct {
	Stock string
	Price float64
	CKey  string // Crytohraphic key
	// add timeout property
}

type order struct {
	ID     string  `json:"id"`
	Stock  string  `json: "stock"`
	Amount float64 `json:"amount"`
	Buy_id string  `json:"buy_id"`
	// figure out timeout feature
}

var orders = []order{}

var logfile = []string{} //WILL BE MOVED TO DB
var transaction_counter = 1
var orders_counter = 1

func connectDb(databaseUri string) (*mongo.Client, error) {
	// adapted from https://github.com/mongodb/mongo-go-driver/blob/d957e67225a9ea82f1c7159020b4f9fd7c8d441a/README.md#usage
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return mongo.Connect(ctx, options.Client().ApplyURI(databaseUri))
}

// main
func main() {
	router := gin.Default() // initializing Gin router
	router.SetTrustedProxies(nil)

	var db *mongo.Database
	router.Use(func(ctx *gin.Context) {
		ctx.Set("db", db)
		ctx.Next()
	})

	router.GET("/users", getAll) // Do we even need?? Not really

	router.GET("/users/:id", getAccount)

	//router.POST("/newuser", addAccount) Migh be used if we do sign up

	router.PUT("/users/:id/add", addBalance)

	router.GET("/users/:id/quote/:stock", getQuote)

	router.POST("/users/:id/buy", buyStock)

	router.POST("/users/:id/buy/commit", commitBuy)

	router.POST("/users/:id/sell/:stock/amount/:quantity", sellStock)

	router.GET("/health", healthcheck)

	// using temp functions and http method to test cli
	// should be changed appropriately

	router.POST("/users/:id/buy/commit", commitBuy)
	router.POST("/users/:id/buy/cancel", cancelBuy)
	router.POST("/users/:id/sell/commit", commitSell)
	router.POST("/users/:id/sell/cancel", cancelSell)
	router.POST("/users/:id/set_buy/:stock/amount/:quantity", setBuyAmount)
	router.POST("/users/:id/set_buy/cancel/:stock", cancelSetBuy)
	router.POST("/users/:id/set_buy/trigger/:stock/amount/:quantity", setBuyTrigger)
	router.POST("/users/:id/set_sell/:stock/amount/:quantity", setSellAmount)
	router.POST("/users/:id/set_sell/trigger/:stock/amount/:quantity", setSellTrigger)
	router.POST("/users/:id/set_sell/cancel/:stock", cancelSetSell)
	router.POST("/users/:id/dumplog/:filename", dumplog)
	router.POST("/dumplog/:filename", dumplog)
	router.POST("/users/:id/display_summary", displaySummary)
	// GET RID OF LATER, FOR DEBUGGING PURPOSES
	router.GET("/log", logAll)
	router.GET("/orders", getOrders)

	bind := flag.String("bind", "localhost:8080", "host:port to listen on")
	flag.Parse()

	databaseUri, found := os.LookupEnv("DATABASE_URI")
	if !found {
		log.Fatalln("No DATABASE_URI")
	}

	mongoClient, err := connectDb(databaseUri)
	if err != nil {
		log.Fatalln(err)
	}

	db = mongoClient.Database("daytrading")

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	err = router.Run(*bind)
	log.Fatal(err)
}

func getOrders(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, orders)
}

func logAll(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, logfile)
}

func getAll(c *gin.Context) {
	// Bad on performance
	r := readMany("users", bson.D{})
	c.IndentedJSON(http.StatusOK, r)
}

func getAccount(c *gin.Context) {
	id := c.Param("id")

	r := readOne("users", bson.D{{"user_id", id}})
	n := bson.D{{"none", "none"}} // to compare and make sure not empty response

	if !reflect.DeepEqual(r, n) {
		c.IndentedJSON(http.StatusOK, r)
		return
	}
	// Else account not found
	err := insert("users", bson.D{{"user_id", id}})
	if err != "ok" {
		panic(err)
	}

	c.IndentedJSON(http.StatusOK, "success")
}

// func addAccount(id string) account {
// 	// NOT CURRENTLY USED
// 	var newAccount account
// 	newAccount.ID = id
// 	newAccount.Balance = 0

// 	accounts = append(accounts, newAccount)
// 	return newAccount
// }

// ADD ACCOUNT IS NEVER CALLED BY A REQUEST DIRECTLY
// func addAccount(c *gin.Context) {
// 	var newAccount account

// 	// Call BindJSON to bind the received JSON to newAccount.
// 	if err := c.BindJSON(&newAccount); err != nil {
// 		return
// 	}
// 	// Add the new account to the slice.
// 	accounts = append(accounts, newAccount)
// 	c.IndentedJSON(http.StatusCreated, newAccount)
// }

// THIS CODE MIGHT BE USEFUL IF WE DO SIGN UP FEATURE
// func addAccount(c *gin.Context) {
// 	var newAccount account

// 	// Call BindJSON to bind the received JSON to newAccount.
// 	if err := c.BindJSON(&newAccount); err != nil {
// 		return
// 	}
// 	// Add the new account to the slice.
// 	accounts = append(accounts, newAccount)
// 	c.IndentedJSON(http.StatusCreated, newAccount)
// }

// addBalance() USAGE:
// curl http://localhost:8080/users/01/add \
// --include --header \
// "Content-Type: application/json" \
// --request "PUT" --data \
// '{"amount": 1}'

func addBalance(c *gin.Context) {

	id := c.Param("id")
	var newBalDif balanceDif
	//var theAcc account

	// Calling BindJSON to bind the recieved JSON to new BalDif
	if err := c.BindJSON(&newBalDif); err != nil {
		return
	}

	// LOGGING USER COMMAND
	// timestamp-server-transaction-command-username-funds
	now := time.Now()
	t_num := strconv.Itoa(transaction_counter) //Is there way to make this global with pointers??
	transaction_counter += 1
	var log_entry = now.String() + " own_server " + t_num + " ADD " + id + fmt.Sprintf(" %f ", newBalDif.Amount)
	logfile = append(logfile, log_entry)
	//log.Println(log_entry)

	u := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"cash_balance", newBalDif.Amount}}, "$inc")

	if u != "ok" {
		panic(u)
	}

	// LOGGING ACCOUNT CHANGES
	//timestamp-server-transaction-action-username-funds
	// r := readOne("users", bson.D{{"user_id", id}})
	// n := bson.D{{"none", "none"}} // to compare and make sure not empty response

	// if !reflect.DeepEqual(r, n) {
	// 	c.IndentedJSON(http.StatusOK, r)
	// 	return
	// }

	// now = time.Now()
	// t_num = strconv.Itoa(transaction_counter) //Is there way to make this global with pointers??
	// transaction_counter += 1
	// log_entry = now.String() + " own_server" + t_num + " add " + id + strconv.Itoa(bal)
	// logfile = append(logfile, log_entry)
	// log.Println(log_entry)

	c.IndentedJSON(http.StatusOK, u)

}

func getQuote(c *gin.Context) {
	var newQuote quote

	// initializing stuff
	id := c.Param("id")
	newQuote.Stock = c.Param("stock")

	var tmstmp string

	newQuote.Price, tmstmp, newQuote.CKey = getQuoteLocal(newQuote.Stock, id)

	fmt.Println(tmstmp)

	// LOGGING FOR COMMAND: timestamp-server-transaction-command-username-stocksymbol
	now := time.Now()
	t_num := strconv.Itoa(transaction_counter) //Is there way to make this global with pointers??
	transaction_counter += 1
	var log_entry = now.String() + " own_server " + t_num + " QUOTE " + id + newQuote.Stock
	logfile = append(logfile, log_entry)
	log.Println(log_entry)

	c.IndentedJSON(http.StatusOK, "INCOMPLETE")

}

func getQuoteLocal(sym string, username string) (float64, string, string) {
	// WILL BE DELETED LATER
	// JUST SO THAT THERE IS A RETURN VALUE

	// Seems unecesary to send back and forth user and stock
	//return (rand.Float64() * 300), username, "thisISaCRYPTOkey"
	return rand.Float64() * 300, " thisisatimestamp ", " thisISaCRYPTOkey "
}

func getQuoteTEMP(sym string, username string) (float64, string, string) {

	//TEMPORARY NAME BECAUSE IT INTERFERS WITH GET QUOTE HTTP METHOD
	//make connection to server
	strEcho := sym + " " + username + "\n"
	servAddr := "quoteserve.seng.uvic.ca:4444"

	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		fmt.Println("\nResolveTCPAddr error: ", err)
		panic(err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("\nDialTCP error: ", err)
		panic(err)
	}

	//write to server SYM being requested and user
	_, err = conn.Write([]byte(strEcho))
	if err != nil {
		fmt.Println("\nWrite error: ", err)
		panic(err)
	}

	//reading from server
	_reply := make([]byte, 1024)

	_, err = conn.Read(_reply)
	if err != nil {
		fmt.Println("\nRead error: ", err)
		panic(err)
	}

	//parsing reply from server
	reply := strings.Split(strings.ReplaceAll(string(_reply), "\n", ""), ",")
	quotePrice, err := strconv.ParseFloat(reply[0], 64)
	if err != nil {
		panic(err)
	}
	timestamp := reply[3] // WHY TIME STAMP?? QUOTESERVER RETURNS QUOTE USER ID AND CK
	cryptKey := reply[4]

	conn.Close()

	return quotePrice, timestamp, cryptKey
}

//	curl http://localhost:8080/users/01/buy \
//	    --include --header \
//	    "Content-Type: application/json" \
//	    --request "POST" --data \
//	    '{"stock": "AMG", "amount":100}'
//
// SHOULD NOT CHANGE ACCOUNT
func buyStock(c *gin.Context) {
	var newOrder order

	// Calling BindJSON to bind the recieved JSON to new BalDif
	if err := c.BindJSON(&newOrder); err != nil {
		return
	}

	newOrder.ID = c.Param("id")

	// LOGGING USER COMMAND
	// timestamp-server-transaction-command-username
	now := time.Now()
	t_num := strconv.Itoa(transaction_counter) //Is there way to make this global with pointers??
	transaction_counter += 1
	var log_entry = now.String() + " own_server " + t_num + " BUY " + newOrder.ID
	logfile = append(logfile, log_entry)
	//log.Println(log_entry)

	// CHECK IF USER HAS ENOUGH BALANCE
	r := readField("users", bson.D{{"user_id", newOrder.ID}}, bson.D{{"cash_balance", 1}})

	n := bson.D{{"none", "none"}}

	if reflect.DeepEqual(r, n) {
		panic("ERROR")
	}

	// Check if user has enough balance
	//afford := false
	switch v := r[0][1].Value.(type) {
	case float64:
		{
			fmt.Println("FLOATING")
			if v > newOrder.Amount {
				// DO NOT WANT TO MAKE CHANGE TO ACCOUT BAL YET
				//r := updateOne("users", bson.D{{"user_id", newOrder.ID}}, bson.D{{"cash_balance", v - newOrder.Amount}}, "$set")
				//THIS SHOULD PRBLY GO IN AN ORDER CACHE??
				//afford = true

				newOrder.Buy_id = newOrder.ID + strconv.Itoa(orders_counter)
				orders = append(orders, newOrder)
				c.IndentedJSON(http.StatusOK, "AFFORD")

				// MAYBE STORE THIS IN CACHE??
				// i := updateOne("users", bson.D{{"user_id", newOrder.ID}}, bson.D{{"account_holdings", bson.D{{"symbol", newOrder.Stock}, {"quantity", quantity}, {"pps", pps}}}}, "$push")
				// if i != "ok" {
				// 	panic("PUSH ERROR")
				// }
				// if r != "ok" {
				// 	panic(r)
				// }
				//c.IndentedJSON(http.StatusBadRequest, accounts[index])
				return
			} else {
				c.IndentedJSON(http.StatusForbidden, "Not enough balance in your account")
			}
		}
	case int64:
		{
			a := float64(v)
			fmt.Println(a)
			fmt.Println(newOrder.Amount)
			if a > newOrder.Amount {
				// DO NOT WANT TO COMMIT CHANGES YET
				// fmt.Println("YES")
				// r := updateOne("users", bson.D{{"user_id", newOrder.ID}}, bson.D{{"cash_balance", a - newOrder.Amount}}, "$set")
				//THIS SHOULD PRBLY GO IN AN ORDER CACHE??
				//afford = true

				newOrder.Buy_id = newOrder.ID + strconv.Itoa(orders_counter)
				orders = append(orders, newOrder)
				c.IndentedJSON(http.StatusOK, "AFFORD")

				// STORE MAYBE IN CACHE
				// i := updateOne("users", bson.D{{"user_id", newOrder.ID}}, bson.D{{"account_holdings", bson.D{{"symbol", newOrder.Stock}, {"quantity", quantity}, {"pps", pps}}}}, "$push")
				// if i != "ok" {
				// 	panic("PUSH ERROR")
				// }
				// if r != "ok" {
				// 	panic(r)
				// }
				//c.IndentedJSON(http.StatusBadRequest, accounts[index])
				return
			} else {
				c.IndentedJSON(http.StatusForbidden, "Not enough balance in your account")
			}
		}
	case int32:
		{
			a := float64(v)
			fmt.Println(a)
			fmt.Println(newOrder.Amount)
			if a > newOrder.Amount {
				// fmt.Println("INT32")
				// r := updateOne("users", bson.D{{"user_id", newOrder.ID}}, bson.D{{"cash_balance", a - newOrder.Amount}}, "$set")
				//THIS SHOULD PRBLY GO IN AN ORDER CACHE??
				//afford = true

				newOrder.Buy_id = newOrder.ID + strconv.Itoa(orders_counter)
				orders = append(orders, newOrder)
				c.IndentedJSON(http.StatusOK, "AFFORD")

				// STORE MAYBE IN CACHE
				// i := updateOne("users", bson.D{{"user_id", newOrder.ID}}, bson.D{{"account_holdings", bson.D{{"symbol", newOrder.Stock}, {"quantity", quantity}, {"pps", pps}}}}, "$push")
				// if i != "ok" {
				// 	panic("PUSH ERROR")
				// }
				// if r != "ok" {
				// 	panic(r)
				// }
				//c.IndentedJSON(http.StatusBadRequest, accounts[index])
				return
			} else {
				c.IndentedJSON(http.StatusForbidden, "Not enough balance in your account")
			}
		}
	}

	// User has enough balance, proceed creating order
	//buy_id := len(orders) + 1
	//newOrder.Buy_id = buy_id
	//orders = append(orders, newOrder)
	//return
	//c.IndentedJSON(http.StatusOK, newOrder)
}

func commitBuy(c *gin.Context) {
	// 	`POST /users/:id/buy/:stock/:buyid`
	// **Arguments**
	// - `"id":string` User ID
	// - `"stock":string` Stock Symbol
	// - `"buy":float64` Dollar amount to buy
	// - `"buyid":int` Order identifier
	var commitOrder order

	// Calling BindJSON to bind the recieved JSON to new BalDif
	if err := c.BindJSON(&commitOrder); err != nil {
		return
	}

	for _, o := range orders {
		if o == commitOrder {
			// change user balance
			// add stock to user data
			//remover order from orders
			c.IndentedJSON(http.StatusOK, "Found order!")
		}
	}

}

func sellStock(c *gin.Context) {
	id := c.Param("id")
	stock := c.Param("stock")
	quantity, err := strconv.ParseFloat(c.Param("quantity"), 64)

	// SO PROGRAM WILL RUN
	//pps := getQuoteLocal(stock)
	pps := 100.00

	if err != nil {
		panic("ERR")
	}
	y := rawreadField("users", bson.D{{"user_id", id}}, bson.D{{"cash_balance", 1}})

	fmt.Println(y)

	r := rawreadField("users", bson.D{{"user_id", id}}, bson.D{{"account_holdings", 1}})
	n := bson.D{{"none", "none"}}

	if reflect.DeepEqual(r, n) {
		panic("ERROR")
	}
	//fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	//fmt.Println(r[0][1].Value)

	//fmt.Println()
	//fmt.Printf("TYPE = %T\n\n", r[0][1].Value)
	//fmt.Printf("TYPE = %T\n\n", r[0][0].Value)

	var these_holdings []holding

	//var temp holding

	switch v := r[0][1].Value.(type) {
	case bson.A:
		{
			// Only works with account holdings
			these_holdings = mongo_read_bsonA(v)
		}
	}

	//value of the trade
	value := pps * quantity

	// Check if user has the correct holdings
	for _, holding := range these_holdings {
		if holding.symbol == stock {
			//Will rewrite later
			fmt.Println("TRUE")
			r := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"cash_balance", value}}, "$inc")
			i := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"account_holdings", bson.D{{"symbol", holding.symbol}, {"quantity", holding.quantity}, {"pps", holding.pps}}}}, "$pull")
			if i != "ok" {
				panic("PUSH ERROR")
			}
			f := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"account_holdings", bson.D{{"symbol", stock}, {"quantity", holding.quantity - quantity}, {"pps", pps}}}}, "$push")

			if f != "ok" {
				panic("PUSH ERROR")
			}
			if r != "ok" {
				panic(r)
			}
			//c.IndentedJSON(http.StatusBadRequest, accounts[index])
			return
		}

	}

	// User has enough balance, proceed creating order
	//buy_id := len(orders) + 1
	//newOrder.Buy_id = buy_id
	//orders = append(orders, newOrder)
	//return
	//c.IndentedJSON(http.StatusOK, newOrder)
}

func healthcheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}

// temp functions to test cli
func commitBuy(c *gin.Context) {
	// health check code
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}

func cancelBuy(c *gin.Context) {
	// health check code
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}

func commitSell(c *gin.Context) {
	// health check code
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}

func cancelSell(c *gin.Context) {
	// health check code
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}

func setBuyAmount(c *gin.Context) {
	// health check code
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}

func cancelSetBuy(c *gin.Context) {
	// health check code
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}

func setBuyTrigger(c *gin.Context) {
	// health check code
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}

func setSellAmount(c *gin.Context) {
	// health check code
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}

func setSellTrigger(c *gin.Context) {
	// health check code
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}

func cancelSetSell(c *gin.Context) {
	// health check code
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}

func dumplog(c *gin.Context) {
	// health check code
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}

func displaySummary(c *gin.Context) {
	// health check code
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}
