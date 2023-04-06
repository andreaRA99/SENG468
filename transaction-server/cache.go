package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

const MAX_QUOTE_VALIDITY_SECS = 60

func connectToRedisCache() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "rediscache:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		fmt.Println("*************************")
		log.Fatalln(err)
		fmt.Println("*************************")
	}
	//fmt.Println("*************************")
	//fmt.Println(pong)
	//fmt.Println("*************************")
	fmt.Println("Connnected")

	return client
}

func SetKeyWithExpirationInSecs(key string, val string, expSecs uint) error {
	secondsDelta := time.Duration(expSecs) * time.Second
	err := connectToRedisCache().Set(key, val, secondsDelta).Err()

	if err != nil {
		return errors.New("Could not set key " + key + "with expiration")
	}

	return nil
}

func GetKeyWithStringVal(key string) (string, error) {
	val, err := connectToRedisCache().Get(key).Result()
	//fmt.Println("*************************")
	//fmt.Println(val)
	return val, err
}

func writeQuoteToCache(symbol string, quote string) {
	err := SetKeyWithExpirationInSecs(symbol, quote, 60)
	if err != nil {
		fmt.Println("Error caching quote. Symbol: ", symbol, " Quote: ", quote, "error: ", err)
	}
}

func addQuoteToCaching(id string, stock string) {
	var quoteInCache string
	var newQuote quote
	var tmstmp string
	var price string

	newQuote.Price, tmstmp, newQuote.CKey = mockQuoteServerHit(newQuote.Stock, id)
	newQuote.Stock = stock

	quoteInCache, err := GetKeyWithStringVal(stock)
	price = strconv.FormatFloat(newQuote.Price, 'f', -1, 64)
	//ßfmt.Println(price + tmstmp)
	if err != nil {
		writeQuoteToCache(newQuote.Stock, price)
		//fmt.Println("*************************")
		//fmt.Println(" I am bad at coding  " + stock + "with expiration")
		//fmt.Println("*************************")
		log.Fatal(err)
	} else {

		fmt.Println("*************************")
		fmt.Println("*************************")
		fmt.Println(quoteInCache + tmstmp)
		fmt.Println("*************************")
		fmt.Println("*************************")

	}

	return quoteInCache
	//fmt.Println(newQuote)
	//return newQuote

}
