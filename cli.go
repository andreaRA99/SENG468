package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {

	requestURL := "http://localhost:8080/health"
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		os.Exit(1)
	}

	// res, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	fmt.Printf("client: error making http request: %s\n", err)
	// 	os.Exit(1)
	// }

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("client: response body: %s\n", resBody)
}
