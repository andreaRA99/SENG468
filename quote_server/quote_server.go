package main

import (
	"crypto/md5"
	"encoding/binary"
	"flag"
	"github.com/gin-gonic/gin"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type req struct {
	sym      string
	username string
}

type quote_hit struct {
	Timestamp int     `json:"Timestamp"`
	Price     float64 `json:"Price"`
	Cryptokey string  `json:"Cryptokey"`
}

func main() {

	router := gin.Default() // initializing Gin router
	router.SetTrustedProxies(nil)

	router.POST("/", quote)
	bind := flag.String("bind", "localhost:8083", "host:port to listen on")
	flag.Parse()

	err := router.Run(*bind)
	if err != nil {
		panic(err)
	}
}

func RandStringBytesRmndr(n int, f string) string {
	h := md5.New()
	io.WriteString(h, f)
	var seed uint64 = binary.BigEndian.Uint64(h.Sum(nil))

	rand.Seed(int64(seed))

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func mockQuoteServerHit(sym string, username string) (float64, int, string) {
	t := int(time.Now().Unix())
	rand.Seed(int64(t))
	v := rand.Float64() * 300
	c := RandStringBytesRmndr(10, username)

	return v, t, c
}

func quote(c *gin.Context) {
	var request req
	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusOK, err)
		return
	}
	price, time, crypt := mockQuoteServerHit(request.sym, request.username)

	var send quote_hit

	send.Price = price
	send.Timestamp = time
	send.Cryptokey = crypt

	c.IndentedJSON(http.StatusOK, send)

}
