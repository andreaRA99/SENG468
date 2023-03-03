package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
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

type balanceDif struct {
	ID     string
	Amount float64
}

type quote struct {
	Stock string
	Price float64
	CKey  string // Crytohraphic key
	// add timeout property
}

type order struct {
	ID     string
	Stock  string
	Buy    float64 // amount
	Buy_id int
	// figure out timeout feature
}

var orders = []order{}

// main
func main() {
	sym, username := getParams()
	quotePrice, timestamp, cryptKey := getQuote(sym, username)

	fmt.Println("qoute price = ", quotePrice, "\nsym = ", sym, "\nusername = ", username, "\ntimestamp = ", timestamp, "\ncrypt key = ", cryptKey)

	router := gin.Default() // initializing Gin router
	router.SetTrustedProxies(nil)

	router.GET("/users", getAll) // Do we even need?? Not really

	router.GET("/users/:id", getBalance)

	//router.POST("/newuser", addAccount) Migh be used if we do sign up

	router.PUT("/users/:id/addBal", addBalance)

	router.GET("/users/:id/quote/:stock", getQuote)

	router.POST("/users/:id/buy/:stock", buyQuote)

	bind := flag.String("bind", "localhost:8080", "host:port to listen on")
	flag.Parse()

	err := router.Run(*bind)
	log.Fatal(err)
}

func getAll(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, accounts)
}

func getBalance(c *gin.Context) {
	id := c.Param("id")

	//Loop over lost of accounts looking for id
	for _, a := range accounts {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	// If account not found
	var newAccount account = addAccount(id)
	c.IndentedJSON(http.StatusOK, newAccount)
}

func addAccount(id string) account {
	var newAccount account
	newAccount.ID = id
	newAccount.Balance = 0

	accounts = append(accounts, newAccount)
	return newAccount
}

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

func addBalance(c *gin.Context) {
	//id := c.Param("id")

	// creating a balanceDif to update account
	var addingAmount balanceDif
	//fmt.Println(addingAmount)
	// Call BindJSON to bind recieved json to newBalance type
	if err := c.BindJSON(&addingAmount); err != nil {
		return
	}

	fmt.Println(addingAmount.ID, addingAmount.Amount)

	for index, i := range accounts {
		if i.ID == addingAmount.ID {
			accounts[index].Balance = i.Balance + addingAmount.Amount

			fmt.Println(i.Balance)

			// Change this
			c.IndentedJSON(http.StatusOK, accounts[index])
		}
	}
}

func getQuote(c *gin.Context) {
	// TODO:
	// request quote from legacy server
	// update db
	//id := c.Param("id") not sure we need
	stock_sym := c.Param("stock")
	var newQuote quote
	newQuote.Stock = stock_sym
	newQuote.Price = 250.01
	newQuote.CKey = "n2378dnfq8"
	c.IndentedJSON(http.StatusOK, newQuote)
}

func buyQuote(c *gin.Context) {
	var newOrder order
	if err := c.BindJSON(&newOrder); err != nil {
		return
	}

	// Check if user has enough balance
	for index, i := range accounts {
		if i.ID == newOrder.ID {
			if accounts[index].Balance < newOrder.Buy {
				c.IndentedJSON(http.StatusBadRequest, accounts[index])
				return
			}
		}
	}

	// User has enough balance, proceed creating order
	buy_id := len(orders) + 1
	newOrder.Buy_id = buy_id
	orders = append(orders, newOrder)

	c.IndentedJSON(http.StatusOK, newOrder)
}

func getParams() (string, string) {
	// read SYM and username from stdin
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter SYM: ") // has to be max 3 letters
	text1, _ := reader.ReadString('\n')
	sym := strings.Trim(text1, "\n")

	fmt.Print("Enter username: ")
	text2, _ := reader.ReadString('\n')
	usrnme := strings.Trim(text2, "\n")

	return sym, usrnme
}

func getQuote(sym string, username string) (string, string, string) {
	//make connection to server
	strEcho := sym + " " + username + "\n"
	servAddr := "quoteserve.seng.uvic.ca:4444"

	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		fmt.Println("\nResolveTCPAddr error: ", err)
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("\nDialTCP error: ", err)
		os.Exit(1)
	}

	//write to server SYM being requested and user
	_, err = conn.Write([]byte(strEcho))
	if err != nil {
		fmt.Println("\nWrite error: ", err)
		os.Exit(1)
	}

	//reading from server
	_reply := make([]byte, 1024)

	_, err = conn.Read(_reply)
	if err != nil {
		fmt.Println("\nRead error: ", err)
		os.Exit(1)
	}

	//parsing reply from server
	reply := strings.Split(strings.ReplaceAll(string(_reply), "\n", ""), ",")
	quotePrice := reply[0]
	timestamp := reply[3]
	cryptKey := reply[4]

	conn.Close()

	return quotePrice, timestamp, cryptKey
}
