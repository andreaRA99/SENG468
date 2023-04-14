package cache

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

	return client
}

func SetKeyWithExpirationInSecs(key string, pricestck float64, expSecs uint) error {
	var val string
	secondsDelta := time.Duration(expSecs) * time.Second
	val = strconv.FormatFloat(pricestck, 'f', -1, 64)

	err := connectToRedisCache().Set(key, val, secondsDelta).Err()

	if err != nil {
		return errors.New("Could not set key " + key + "with expiration")
	}

	return nil
}

func GetKeyWithStringVal(key string) (string, error) {
	val, err := connectToRedisCache().Get(key).Result()
	return val, err
}

func writeQuoteToCache(symbol string, quote float64) {
	err := SetKeyWithExpirationInSecs(symbol, quote, 0)
	if err != nil {
		fmt.Println("Error caching quote. Symbol: ", symbol, " Quote: ", quote, "error: ", err)
	}
}
