package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

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
	Adding float64
}

// main
func main() {
	router := gin.Default() // initializing Gin router
	router.SetTrustedProxies(nil)

	router.GET("/user", getAll)
	router.GET("/user/:id", getBalance)

	router.POST("/newuser", addAccount)
	router.POST("/user/:id/addBalance", addBalance)

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

func addAccount(c *gin.Context) {
	var newAccount account

	// Call BindJSON to bind the received JSON to newAccount.
	if err := c.BindJSON(&newAccount); err != nil {
		return
	}
	// Add the new account to the slice.
	accounts = append(accounts, newAccount)
	c.IndentedJSON(http.StatusCreated, newAccount)
}

func addBalance(c *gin.Context) {
	//id := c.Param("id")

	var addingAmount balanceDif
	//fmt.Println(addingAmount)

	// Call BindJSON to bind recieved json to newBalance type
	if err := c.BindJSON(&addingAmount); err != nil {
		return
	}

	fmt.Println(addingAmount.ID, addingAmount.Adding)

	for index, i := range accounts {
		if i.ID == addingAmount.ID {
			accounts[index].Balance = i.Balance + addingAmount.Adding

			fmt.Println(i.Balance)

			// Change this
			c.IndentedJSON(http.StatusCreated, accounts)
		}
	}
}
