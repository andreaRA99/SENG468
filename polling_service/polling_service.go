package main




import (
	"fmt"
	"github.com/gin-gonic/gin"
	"flag"
	"net/http"
	"math/rand"
	"time"
	"encoding/json"
	"bytes"
	"io/ioutil"
)

type LimitOrder struct {
	Stock  string
	Price  float64
	Type 	 string 
	Amount float64
	User   string `json:"ID"`
	Qty 	 float64

}

var reqUrlPrefix = "http://host.docker.internal:8080"

var active_orders []LimitOrder


func main(){

	router := gin.Default() // initializing Gin router
	router.SetTrustedProxies(nil)

	router.POST("/new_limit", new_limit)
	bind := flag.String("bind", "localhost:8081", "host:port to listen on")
	flag.Parse()


	err := router.Run(*bind)
	if err != nil {panic(err)} 


}

func mockQuoteServerHit(sym string, username string) (float64, int, string) {
	return rand.Float64() * 300, int(time.Now().Unix()), " thisISaCRYPTOkey "
}

func get_price(){
	j := 0
	for len(active_orders) > 0 {
		// do: update cache
		val, _, _ := mockQuoteServerHit(active_orders[j].Stock, active_orders[j].User)
		fmt.Println(val)
		fmt.Println(active_orders[j].Price)

		if val > active_orders[j].Price && active_orders[j].Type == "sell" {
			//"ID":active_orders[j].User, "Stock": active_orders[j].Stock, "Amount": active_orders[j].Amount, "Price": val
			active_orders[j].Qty = active_orders[j].Amount
			parsedJson, _ := json.Marshal(active_orders[j])
			req, err := http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/sell", bytes.NewBuffer(parsedJson))
			res, err:= http.DefaultClient.Do(req)
			resBody, err := ioutil.ReadAll(res.Body)
			fmt.Println(resBody)
			req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/sell/commit", bytes.NewBuffer(parsedJson))
			res, err = http.DefaultClient.Do(req)
			resBody, err = ioutil.ReadAll(res.Body)
			fmt.Println(resBody)

			//fmt.Println(res)
			if err != nil{
				fmt.Println("ERROR")
				fmt.Println(err)
			}
			active_orders = append(active_orders[:j], active_orders[j+1:]...)

		} else if val < active_orders[j].Price && active_orders[j].Type == "buy" {
			active_orders[j].Qty = active_orders[j].Amount

			parsedJson, _ := json.Marshal(active_orders[j])
			req, err := http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/buy", bytes.NewBuffer(parsedJson))
			res, err := http.DefaultClient.Do(req)

			
			resBody, err := ioutil.ReadAll(res.Body)
			fmt.Printf("RESBODY: %s\n", resBody)
			if err != nil{
				fmt.Println("ERROR")
				fmt.Println(err)
			}
			
			req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/buy/commit", bytes.NewBuffer(parsedJson))
			res, err = http.DefaultClient.Do(req)
			resBody, err = ioutil.ReadAll(res.Body)
			if err != nil{
				fmt.Println("ERROR")
				fmt.Println(err)
			}

			active_orders = append(active_orders[:j], active_orders[j+1:]...)
		}

		time.Sleep(1 * time.Second) // Math goes here 
		if len(active_orders) < 1 {
			return
		}
		j = (j + 1)%len(active_orders)
	}
	
	return
}
func new_limit(c *gin.Context){
	var limitorder LimitOrder
	if err := c.BindJSON(&limitorder); err != nil {
		c.IndentedJSON(http.StatusOK, err)
		return 
	}
	c.IndentedJSON(http.StatusOK, "ok")
	active_orders = append(active_orders, limitorder)
	if len(active_orders) == 1 {  go get_price()
	} else { return }
	
}