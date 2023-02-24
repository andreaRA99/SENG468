package main

import (
	"flag"
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

// main
func main() {
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
