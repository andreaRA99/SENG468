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
	//fmt.Println("Connnected")

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
	err := SetKeyWithExpirationInSecs(symbol, quote, 0)
	if err != nil {
		fmt.Println("Error caching quote. Symbol: ", symbol, " Quote: ", quote, "error: ", err)
	}
}

func addQuoteToCaching(stock string, price1 float64) string {
	var quoteInCache string
	//var newQuote quote
	var price string

	//newQuote.Price, _, newQuote.CKey = mockQuoteServerHit(newQuote.Stock, id)
	//newQuote.Stock = stock
	//price = strconv.FormatFloat(price1, 'f', -1, 64)
	//price = strconv.FormatFloat(price1, 'f', -1, 64)
	quoteInCache, err := GetKeyWithStringVal(stock)
	price = strconv.FormatFloat(price1, 'f', -1, 64)
	//ßfmt.Println(price + tmstmp)
	if err != nil {
		writeQuoteToCache(stock, price)
		quoteInCache = price
		//fmt.Println("*************************")
		//fmt.Println(" I am bad at coding  " + stock + "with expiration")
		//fmt.Println("*************************")
		return quoteInCache
		log.Fatal(err)
	} else {
		// fmt.Println("*************************")
		// fmt.Println("*************************")
		// fmt.Println(quoteInCache)
		// fmt.Println("*************************")
		// fmt.Println("*************************")
		quoteInCache, err := GetKeyWithStringVal(stock)
		return quoteInCache
		log.Fatal(err)

	}

	return quoteInCache
}
