package main

import (
	"fmt"
	"log"

	"github.com/go-redis/redis"
)

func connectToRedcache() {
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
	fmt.Println("*************************")
	fmt.Println(pong)
	fmt.Println("*************************")
}
