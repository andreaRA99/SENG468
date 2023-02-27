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

// main
func main() {
	sym, username := getParams()
	quotePrice, timestamp, cryptKey := connectToServer(sym, username)

	fmt.Println("qoute price = ", quotePrice, "\nsym = ", sym, "\nusername = ", username, "\ntimestamp = ", timestamp, "\ncrypt key = ", cryptKey)

	router := gin.Default() // initializing Gin router
	router.SetTrustedProxies(nil)

	router.GET("/user", getAll)
	router.GET("/user/:id", getBalance) // associate GET HTTP method and /user

	router.GET("/user/:id/add", addBalance)

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
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "User ID not found"})
}

func addBalance(c *gin.Context) {
	id := c.Param("id")
	//amount := c.Param("qty")
	c.IndentedJSON(http.StatusOK, id)
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

func connectToServer(sym string, username string) (string, string, string) {
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
