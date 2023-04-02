package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
)

func connectToRedisCache() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "rediscache:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	if err != nil {
		fmt.Println("*************************")
		log.Fatal(err)
		fmt.Println("*************************")
	}
	//fmt.Println("*************************")
	fmt.Println(pong)
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
	fmt.Println("*************************")
	fmt.Println(val)
	return val, err
}
