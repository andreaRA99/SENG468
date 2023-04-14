package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"reflect"
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
	Symbol   string `json:"symbol"`
	Quantity int    `json:"quantity"`
}

type balanceDif struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount"`
}

type users struct {
	user_id string
}

type accStatus struct {
	Cash_balance float64   `json:"cash_balance"`
	Stocks       []holding `json:"stocks"`
}

type quote struct {
	//ID    string
	Stock string
	Price float64
	CKey  string // Crytohraphic key
	// add timeout property
}

type LimitOrder struct {
	Stock  string  `json:"stock"`
	Price  float64 `json:"price"`
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
	User   string  `json:"ID"`
	Qty    float64 `json:"qty"`
}

type order struct {
	ID     string  `json:"id"`
	Stock  string  `json: "stock"`
	Amount float64 `json:"amount"`
	Price  float64
	Qty    int
	//Buy_id string  `json:"buy_id"`
	// figure out timeout feature
}

type displayCmdData struct {
	Transactions []logEntry   `json:"transactions"`
	Acc_Status   []accStatus  `json:"accStatus"`
	LimitOrders  []LimitOrder `json:"limitOrders"`
}

var quotes = []quote{}
var buys = []order{}
var sells = []order{}

var uncommited_limit_orders []LimitOrder

var logfile = []string{} //WILL BE MOVED TO DB
var transaction_counter int = 1
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

	// User Commands
	router.PUT("/users/addBal", addBalance)
	router.GET("/users/:id/quote/:stock", Quote)
	router.POST("/users/buy", buyStock)
	router.POST("/users/buy/commit", commitBuy)
	router.DELETE("/users/:id/buy/cancel", cancelBuy)
	router.POST("/users/sell", sellStock)
	router.POST("/users/sell/commit", commitSell)
	router.DELETE("/users/:id/sell/cancel", cancelSell)
	router.POST("/users/set/:type", setAmount)
	router.DELETE("/users/:id/set/:type/:stock/cancel", cancelSet)
	router.POST("/users/set/:type/trigger", setTrigger)
	router.POST("/dumplog", dumplog)
	router.GET("/displaysummary/:id", displaySummary)

	// Util routes
	router.GET("/users/:id", getAccount)
	router.GET("/health", healthcheck)

	// GET RID OF LATER, FOR DEBUGGING PURPOSES
	router.GET("/users", getAll) // Do we even need?? Not really
	router.GET("/log", logAll)
	router.GET("/orders", getOrders)
	router.GET("/quotes", getQuotes)

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

	connectToRedisCache()
	SetKeyWithExpirationInSecs("srock", "500", 0)
	GetKeyWithStringVal("srock")

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

func getQuotes(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, quotes)
}

func getOrders(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, buys)
}

func logAll(c *gin.Context) {
	r := readMany("logs", bson.D{})
	c.IndentedJSON(http.StatusOK, r)
}

func getAll(c *gin.Context) {
	// Bad on performance
	r := readMany("users", bson.D{})
	c.IndentedJSON(http.StatusOK, r)
}

func exists(ID string) bool {
	r := readOne("users", bson.D{{"user_id", ID}})
	n := bson.D{{"none", "none"}} // to compare and make sure not empty response

	if !reflect.DeepEqual(r, n) {
		return true
	} else {
		return false
	}
}

func createAcc(ID string) {
	// Else account not found
	err := insert("users", bson.D{{"user_id", ID}})
	if err != "ok" {
		panic(err)
	}
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

func addBalance(c *gin.Context) {
	var newBalDif balanceDif

	// Calling BindJSON to bind the recieved JSON to new BalDif
	if err := c.BindJSON(&newBalDif); err != nil {
		return
	}

	// Logging user command
	addCmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "ADD", Username: newBalDif.ID, Funds: newBalDif.Amount}
	logEvent(addCmdLog)

	// CREATING ACCOUNT IT DOES NOT EXIST
	if !exists(newBalDif.ID) {
		createAcc(newBalDif.ID)
	}

	if newBalDif.Amount >= 0 {
		u := updateOne("users", bson.D{{"user_id", newBalDif.ID}}, bson.D{{"cash_balance", newBalDif.Amount}}, "$inc")
		if u != "ok" {
			panic(u)
			c.IndentedJSON(http.StatusOK, u)
		}
	} else {
		c.IndentedJSON(http.StatusForbidden, "Enter valid amount")
	}

	// Logging account changes
	addDBLog := logEntry{LogType: ACC_TRANSACTION, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Action: "add", Username: newBalDif.ID, Funds: newBalDif.Amount}
	logEvent(addDBLog)

	transaction_counter += 1
}

func Quote(c *gin.Context) {
	//var newQuote quote

	id := c.Param("id")
	stock := c.Param("stock")

	// Logging user command
	quoteCmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "QUOTE", Username: id, StockSymbol: stock}
	logEvent(quoteCmdLog)

	theQuote := fetchQuote(id, stock)

	// newQuote should be sent to cache
	c.IndentedJSON(http.StatusOK, theQuote)

	transaction_counter += 1
}

func fetchQuote(id string, stock string) quote {
	var newQuote quote

	// check if quote for specified stock exists
	for _, o := range quotes {
		if o.Stock == stock {
			return o
		}
	}

	// else:  HITTING QUOTE SERVER
	// Currently: Read around type cache, do we want read through??
	var tmstmp int
	newQuote.Price, tmstmp, newQuote.CKey = mockQuoteServerHit(newQuote.Stock, id) //simulation of quote hit
	newQuote.Stock = stock

	quotes = append(quotes, newQuote)

	// Logging quote server hit
	QSHitLog := logEntry{LogType: QUOTESERVER, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Price: newQuote.Price, StockSymbol: stock, Username: id, QuoteServerTime: tmstmp, Cryptokey: newQuote.CKey}
	logEvent(QSHitLog)

	return newQuote
}

func mockQuoteServerHit(sym string, username string) (float64, int, string) {
	return rand.Float64() * 300, int(time.Now().Unix()), " thisISaCRYPTOkey "
}

func buyStock(c *gin.Context) {
	var newOrder order

	// Calling BindJSON to bind the recieved JSON to an order
	if err := c.BindJSON(&newOrder); err != nil {
		return
	}

	// Logging user command
	buyCmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "BUY", Username: newOrder.ID, StockSymbol: newOrder.Stock, Funds: newOrder.Amount}
	logEvent(buyCmdLog)
	transaction_counter += 1

	// CHECK IF USER HAS ENOUGH BALANCE
	r := rawreadField("users", bson.D{{"user_id", newOrder.ID}}, bson.D{{"cash_balance", 1}})

	// This would ideally go after checking if account has enough balance
	// Fetching most current price for that stock
	newOrder.Price = fetchQuote(newOrder.ID, newOrder.Stock).Price

	newOrder.Qty = int(math.Floor(newOrder.Amount))

	newOrder.Amount = newOrder.Price * float64(newOrder.Qty) // How much user will be charged based on  int Qty of stocks at surr price
	if newOrder.Amount == 0 {
		c.IndentedJSON(http.StatusForbidden, "Cannot afford stock with given amount")
		return
	}

	// Should refactor, very redundant code
	switch v := r[0][1].Value.(type) {
	case float64:
		//{
		if v > newOrder.Amount {
			buys = append(buys, newOrder)
			c.IndentedJSON(http.StatusOK, newOrder)
			return
		} else {
			c.IndentedJSON(http.StatusForbidden, "Not enough balance in your account")
		}
	case int64:
		a := float64(v)
		if a > newOrder.Amount {
			buys = append(buys, newOrder)
			c.IndentedJSON(http.StatusOK, newOrder)
			return
		} else {
			c.IndentedJSON(http.StatusForbidden, "Not enough balance in your account")
		}
	case int32:
		a := float64(v)
		if a > newOrder.Amount {
			buys = append(buys, newOrder)
			c.IndentedJSON(http.StatusOK, newOrder)
			return
		} else {
			c.IndentedJSON(http.StatusForbidden, "Not enough balance in your account")
		}
	}
}

func commitBuy(c *gin.Context) {
	var commitOrder order

	// Calling BindJSON to bind the recieved JSON to new BalDif
	if err := c.BindJSON(&commitOrder); err != nil {
		return
	}

	// Getting most recent order that took place within last 60 secs
	// Queue? Cache?
	match := false
	j := 0
	for _, o := range buys {

		if o.ID == commitOrder.ID {
			match = true

			// Logging user command
			commitBuyCmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "COMMIT_BUY", Username: commitOrder.ID, Funds: o.Amount}
			logEvent(commitBuyCmdLog)

			// change user balance
			to_match := bson.D{{"user_id", o.ID}}
			to_update := bson.D{{"cash_balance", -o.Amount}, {o.Stock, o.Qty}}
			r := updateOne("users", to_match, to_update, "$inc")

			if r != "ok" {
				panic(r)
			}

			// Logging account changes
			commitBuyDBLog := logEntry{LogType: ACC_TRANSACTION, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Action: "remove", Username: commitOrder.ID, Funds: o.Amount}
			logEvent(commitBuyDBLog)

			//remover order from orders
			c.IndentedJSON(http.StatusOK, r)

			//possible memory leak
			buys = append(buys[:j], buys[j+1:]...)
			break

		}
		j++
	}

	// Command did not execute but must be logged
	if !match {
		// Logging user command
		commitBuyCmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "COMMIT_BUY", Username: commitOrder.ID}
		logEvent(commitBuyCmdLog)

		// Logging command did not happen due to error
		errorLog := logEntry{LogType: ERR_EVENT, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "COMMIT_BUY", Username: commitOrder.ID}
		logEvent(errorLog)
	}
	transaction_counter += 1
}

func cancelBuy(c *gin.Context) {
	id := c.Param("id")
	match := false
	j := 0
	for _, o := range buys {
		if o.ID == id {
			match = true

			// Logging user command
			cancelBuyCmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "CANCEL_BUY", Username: id}
			logEvent(cancelBuyCmdLog)

			//remover order from orders
			//possible memory leak
			buys = append(buys[:j], buys[j+1:]...)
			break
			c.IndentedJSON(http.StatusOK, "ok")

		}
		j++
	}

	// Command did not execute but must be logged
	if !match {
		// Logging user command
		cancelBuyCmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "CANCEL_BUY", Username: id}
		logEvent(cancelBuyCmdLog)

		// Logging command did not happen due to error
		errorLog := logEntry{LogType: ERR_EVENT, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "CANCEL_BUY", Username: id}
		logEvent(errorLog)
	}
	transaction_counter += 1
}

func sellStock(c *gin.Context) {

	var newOrder order

	// Calling BindJSON to bind the recieved JSON to an order
	if err := c.BindJSON(&newOrder); err != nil {
		return
	}

	// Logging user command
	sellCmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "SELL", Username: newOrder.ID, StockSymbol: newOrder.Stock, Funds: newOrder.Amount}
	logEvent(sellCmdLog)
	transaction_counter += 1

	r := rawreadField("users", bson.D{{"user_id", newOrder.ID}}, bson.D{{newOrder.Stock, 1}})
	n := bson.D{{"none", "none"}}

	if reflect.DeepEqual(r, n) {
		panic("ERROR")
	}

	newOrder.Price = fetchQuote(newOrder.ID, newOrder.Stock).Price
	newOrder.Qty = int(math.Floor(newOrder.Amount / newOrder.Price))
	newOrder.Amount = newOrder.Price * float64(newOrder.Qty) // How much user will be charged based on  int Qty of stocks at surr price

	if (len(r)) < 1 {
		c.IndentedJSON(http.StatusForbidden, "Stock Not Owned!")
		return
	}
	if len(r[0]) < 1 {
		c.IndentedJSON(http.StatusForbidden, "Stock Not Owned!")
		return

	}

	switch v := r[0][1].Value.(type) {

	case int32:
		{
			q := v
			if int64(q) >= int64(newOrder.Qty) {
				sells = append(sells, newOrder)
				c.IndentedJSON(http.StatusOK, newOrder)
				return
			} else {
				c.IndentedJSON(http.StatusForbidden, "Not enough holdings")
				return
			}
		}

	default:
		c.IndentedJSON(http.StatusForbidden, "Server error")

		return
	}

	c.IndentedJSON(http.StatusForbidden, "Finished")

	// Check they have enough
	return
}

// User has enough balance, proceed creating order
//buy_id := len(orders) + 1
//newOrder.Buy_id = buy_id
//orders = append(orders, newOrder)
//return
//c.IndentedJSON(http.StatusOK, newOrder)

func commitSell(c *gin.Context) {
	var commitOrder order

	// Calling BindJSON to bind the recieved JSON to new BalDif
	if err := c.BindJSON(&commitOrder); err != nil {
		return
	}

	// Getting most recent order that took place within last 60 secs
	// Queue? Cache?
	match := false
	j := 0
	for _, o := range sells {
		if o.ID == commitOrder.ID {
			match = true

			// Logging user command
			commitSellCmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "COMMIT_SELL", Username: commitOrder.ID, Funds: commitOrder.Amount}
			logEvent(commitSellCmdLog)

			// change user balance
			to_match := bson.D{{"user_id", commitOrder.ID}}
			to_update := bson.D{{"cash_balance", +o.Amount}, {o.Stock, -o.Qty}}
			r := updateOne("users", to_match, to_update, "$inc")

			if r != "ok" {
				panic(r)
			}

			// Logging account changes
			commitSellDBLog := logEntry{LogType: ACC_TRANSACTION, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Action: "add", Username: commitOrder.ID, Funds: commitOrder.Amount}
			logEvent(commitSellDBLog)

			//remover order from orders

			sells = append(sells[:j], sells[j+1:]...)
			c.IndentedJSON(http.StatusOK, "ok")
			return
		}
		j++
	}

	// Command did not execute but must be logged
	if !match {
		// Logging user command
		commitSellCmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "COMMIT_SELL", Username: commitOrder.ID}
		logEvent(commitSellCmdLog)

		// Logging command did not happen due to error
		errorLog := logEntry{LogType: ERR_EVENT, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "COMMIT_SELL", Username: commitOrder.ID}
		logEvent(errorLog)
	}
	transaction_counter += 1

	c.IndentedJSON(http.StatusForbidden, "No previous sell order")
}

func cancelSell(c *gin.Context) {
	id := c.Param("id")
	match := false
	j := 0
	for _, o := range sells {
		if o.ID == id {
			match = true

			// Logging user command
			cancelSellCmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "CANCEL_SELL", Username: id}
			logEvent(cancelSellCmdLog)

			c.IndentedJSON(http.StatusOK, "ok")
			//remover order from orders
			//possible memory leak
			sells = append(sells[:j], sells[j+1:]...)
			return

		}
		j++
	}

	// Command did not execute but must be logged
	if !match {
		// Logging user command
		cancelSellCmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "CANCEL_SELL", Username: id}
		logEvent(cancelSellCmdLog)

		// Logging command did not happen due to error
		errorLog := logEntry{LogType: ERR_EVENT, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "CANCEL_SELL", Username: id}
		logEvent(errorLog)
	}
	transaction_counter += 1

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

func setAmount(c *gin.Context) {
	var limitorder LimitOrder
	limitorder.Type = c.Param("type")

	var cmd string
	if limitorder.Type == "buy" {
		cmd = "SET_BUY_AMOUNT"
	} else {
		cmd = "SET_SELL_AMOUNT"
	}

	// Logging user command
	cmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: cmd, Username: limitorder.User}
	logEvent(cmdLog)

	// Calling BindJSON to bind the recieved JSON
	if err := c.BindJSON(&limitorder); err != nil {
		return
	}

	match := false
	for _, o := range uncommited_limit_orders {
		if o.User == limitorder.User {
			if o.Type == limitorder.Type {
				match = true
				o.Amount = limitorder.Amount
			}
		}
	}

	uncommited_limit_orders = append(uncommited_limit_orders, limitorder)

	// TODO: not sure what constitutes this command not taking place
	if !match {
		// Logging error event
		errorLog := logEntry{LogType: ERR_EVENT, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: cmd, Username: limitorder.User}
		logEvent(errorLog)
	}
	transaction_counter += 1
}

func cancelSet(c *gin.Context) {
	var limitorder LimitOrder
	limitorder.Type = c.Param("type")

	var cmd string
	if limitorder.Type == "buy" {
		cmd = "CANCEL_SET_BUY"
	} else {
		cmd = "CANCEL_SET_SELL"
	}

	// Logging user command
	cmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: cmd, Username: limitorder.User}
	logEvent(cmdLog)

	// Calling BindJSON to bind the recieved JSON
	if err := c.BindJSON(&limitorder); err != nil {
		return
	}

	match := false
	j := 0
	for _, o := range uncommited_limit_orders {
		if o.User == limitorder.User {
			if o.Type == limitorder.Type {
				match = true
				uncommited_limit_orders = append(uncommited_limit_orders[:j], uncommited_limit_orders[j+1:]...)
			}
		}
		j++
	}

	// TODO: not sure what constitutes this command not taking place
	if !match {
		// Logging error event
		errorLog := logEntry{LogType: ERR_EVENT, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: cmd, Username: limitorder.User}
		logEvent(errorLog)
	}
	transaction_counter += 1
}

func setTrigger(c *gin.Context) {
	// Unresolved:
	//(a) a reserve account is created for the BUY transaction to hold the specified amount in reserve for when the transaction is triggered
	// (b)the user's cash account is decremented by the specified amount
	// Resolved:   (c) when the trigger point is reached the user's stock account is updated to reflect the BUY transaction.

	var limitorder LimitOrder
	limitorder.Type = c.Param("type")

	var cmd string
	if limitorder.Type == "buy" {
		cmd = "SET_BUY_TRIGGER"
	} else {
		cmd = "SET_SELL_TRIGGER"
	}

	// logging user command
	cmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: cmd, Username: limitorder.User, Funds: limitorder.Amount}
	logEvent(cmdLog)

	// Calling BindJSON to bind the recieved JSON
	if err := c.BindJSON(&limitorder); err != nil {
		return
	}

	j := 0
	for _, o := range uncommited_limit_orders {
		if o.User == limitorder.User {
			if o.Type == limitorder.Type {
				transaction_counter += 1

				o.Price = limitorder.Price
				parsedJson, err := json.Marshal(o)
				if err != nil {
					panic(err)
				}

				req, err := http.NewRequest(http.MethodPost, "http://polling_microservice:8081/new_limit", bytes.NewBuffer(parsedJson))
				if err != nil {
					panic(err)
				}
				_, err = http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}

				uncommited_limit_orders = append(uncommited_limit_orders[:j], uncommited_limit_orders[j+1:]...)
				return
			}

		}
		j++
	}

	// Logging error event
	errorLog := logEntry{LogType: ERR_EVENT, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: cmd, Username: limitorder.User, Funds: limitorder.Amount}
	logEvent(errorLog)
	transaction_counter += 1
}

func dumplog(c *gin.Context) {
	type dumplogParams struct {
		Filename string `json:"filename"`
		Id       string `json:"id"`
	}
	var dumpLog dumplogParams

	// Calling BindJSON to bind the recieved JSON
	if err := c.BindJSON(&dumpLog); err != nil {
		return
	}

	// Logging dumplog command
	dumplogCmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "DUMPLOG", Username: dumpLog.Id, Filename: dumpLog.Filename}
	logEvent(dumplogCmdLog)

	// Get logs from DB
	var logsd []bson.D
	var logs []logEntry
	if dumpLog.Id == "" {
		logsd = readMany("logs", bson.D{})
	} else {
		logsd = readMany("logs", bson.D{{"Username", dumpLog.Id}})
	}
	logs = mongo_read_logs(logsd)

	// Send logs as JSON response
	c.IndentedJSON(http.StatusOK, logs)

	transaction_counter += 1
}

// Provides a summary to the client of the given user's transaction history and the current
// status of their accounts as well as any set buy or sell triggers and their parameters
func displaySummary(c *gin.Context) {
	// Params: userid
	id := c.Param("id")

	// Logging displaySummary command
	cmdLog := logEntry{LogType: USERCOMMAND, Timestamp: time.Now().Unix(), Server: "own-server", TransactionNum: transaction_counter, Command: "DISPLAY_SUMMARY", Username: id}
	logEvent(cmdLog)

	// A summary of the given user's transaction history...
	var logsd []bson.D
	var logs []logEntry
	logsd = readMany("logs", bson.D{{"Username", id}})
	logs = mongo_read_logs(logsd)

	// ...and the current status of their accounts...
	var acc_status []accStatus
	r := readMany("users", bson.D{})
	n := bson.D{{"none", "none"}} // to compare and make sure not empty response

	if reflect.DeepEqual(r, n) {
		panic("Empty db read response")
	}

	acc_status = mongo_read_acc_status(r)

	// ...as well as any set buy or sell triggers and their parameters...
	var limitOrders []LimitOrder
	for idx := range uncommited_limit_orders {
		if uncommited_limit_orders[idx].User == id {
			limitOrders = append(limitOrders, uncommited_limit_orders[idx])
		}
	}

	// ...is displayed to the user.
	data := displayCmdData{Transactions: logs, Acc_Status: acc_status, LimitOrders: limitOrders}

	// Send data as JSON response
	c.IndentedJSON(http.StatusOK, data)

	transaction_counter += 1
}
