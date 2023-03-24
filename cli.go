package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli"
)

// Cmd struct is a representation of an isolated command executed by a user
type Cmd struct {
	Command  string `json:"Command"`
	Userid   string `json:"Userid"`
	Stock    string `json:"Stock"`
	Amount   string `json:"Amount"`
	Filename string `json:"Filename"`
}

func main() {
	// requestURL := "http://localhost:8080/health"
	// res, err := http.Get(requestURL)
	// if err != nil {
	// 	fmt.Printf("client: could not create request: %s\n", err)
	// 	os.Exit(1)
	// }

	// fmt.Printf("client: got response!\n")
	// fmt.Printf("client: status code: %d\n", res.StatusCode)

	// resBody, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	fmt.Printf("client: could not read response body: %s\n", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("client: response body: %s\n", resBody)

	app := cli.NewApp()
	app.Name = "DayTrading Inc. CLI"
	app.Usage = "Lets you execute user commands from a file containing a list of commands as well as execute individual user commands from the command line"

	app.Commands = []cli.Command{
		{
			Name:     "read",
			Aliases:  []string{"r"},
			HelpName: "read",
			Action: func(c *cli.Context) error {
				readFromFile(c)
				return nil
			},
			Usage:       `Reads file from specified path`,
			Description: `Read and parse user commands' file`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "filelocation, fl",
					Usage: "full path of file containing commands",
				},
			},
		},
		{
			Name:     "execute",
			Aliases:  []string{"e"},
			HelpName: "execute",
			Action: func(c *cli.Context) error {

				// command := strings.ToUpper(c.String("cmd"))
				// userid := c.String("userid")
				// sym := strings.ToUpper(c.String("sym"))
				// amount := c.String("amount")
				// filename := c.String("filename")

				// cmd := Cmd{Command: command, Userid: userid, Stock: sym, Amount: amount, Filename: filename}

				// executeCmd(cmd)
				return nil
			},
			Usage:       `Executes specified user command`,
			Description: `Executes the given command`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "command, cmd",
					Usage: "user command",
				},
				cli.StringFlag{
					Name:  "userid, uid",
					Usage: "username",
				},
				cli.StringFlag{
					Name:  "stocksymbol, sym",
					Usage: "stock's symbol",
				},
				cli.Float64Flag{
					Name:  "amount, amt",
					Usage: "amount in dollars",
				},
				cli.StringFlag{
					Name:  "filename, fn",
					Usage: "file to print out to",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func readFromFile(c *cli.Context) error {
	fileLocation := c.String("filelocation")
	// file, err := os.Open("./" + fileLocation) //location relative to current dir
	file, err := os.Open(fileLocation)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		// parseLine(scanner.Text())
		data, cmd := parseLine(scanner.Text())
		executeCmd(data, cmd)
		// use parsed info to make get/post requests ?
	}

	if err := scanner.Err(); err != nil {
		return (err)
	}

	return nil
}

func parseLine(line string) (string, string) {
	line_arr := strings.Split(line, " ")
	cmd_arr := strings.Split(line_arr[1], ",")
	command := cmd_arr[0]

	// if command == "ADD" {
	// 	return Cmd{Command: command, Userid: cmd_arr[1], Amount: cmd_arr[2]}

	// } else if command == "BUY" || command == "SELL" || command == "SET_BUY_AMOUNT" || command == "SET_BUY_TRIGGER" ||
	// 	command == "SET_SELL_AMOUNT" || command == "SET_SELL_TRIGGER" {
	// 	return Cmd{Command: command, Userid: cmd_arr[1], Stock: cmd_arr[2], Amount: cmd_arr[3]}

	// } else if command == "QUOTE" || command == "CANCEL_SET_BUY" || command == "CANCEL_SET_SELL" {
	// 	return Cmd{Command: command, Userid: cmd_arr[1], Stock: cmd_arr[2]}

	// } else if command == "COMMIT_BUY" || command == "COMMIT_SELL" || command == "CANCEL_BUY" || command == "CANCEL_SELL" || command == "DISPLAY_SUMMARY" {
	// 	return Cmd{Command: command, Userid: cmd_arr[1]}

	// } else if command == "DUMPLOG" {
	// 	if len(cmd_arr) == 2 {
	// 		return Cmd{Command: command, Filename: cmd_arr[1]}
	// 	} else {
	// 		return Cmd{Command: command, Userid: cmd_arr[1], Filename: cmd_arr[2]}
	// 	}

	// } else {
	// 	fmt.Printf("Command received: %s, line: %s\n", command, line)
	// 	panic("Unknown command received")

	// }
	requestURL := "http://localhost:8080"
	switch command {
	case "ADD":
		requestURL += "/users/" + cmd_arr[1] + "/add/" + cmd_arr[2]
	case "QUOTE":
		requestURL += "/users/" + cmd_arr[1] + "/quote/" + cmd_arr[2]
	case "BUY":
		requestURL += "/users/" + cmd_arr[1] + "/buy/" + cmd_arr[2] + "/amount/" + cmd_arr[3]
	case "COMMIT_BUY":
		requestURL += "/users/" + cmd_arr[1] + "/buy/commit"
	case "CANCEL_BUY":
		requestURL += "/users/" + cmd_arr[1] + "/buy/cancel"
	case "SELL":
		requestURL += "/users/" + cmd_arr[1] + "/sell/" + cmd_arr[2] + "/amount/" + cmd_arr[3]
	case "COMMIT_SELL":
		requestURL += "/users/" + cmd_arr[1] + "/sell/commit"
	case "CANCEL_SELL":
		requestURL += "/users/" + cmd_arr[1] + "/sell/cancel"
	case "SET_BUY_AMOUNT":
		requestURL += "/users/" + cmd_arr[1] + "/set_buy/" + cmd_arr[2] + "/amount/" + cmd_arr[3]
	case "CANCEL_SET_BUY":
		requestURL += "/users/" + cmd_arr[1] + "/set_buy/cancel/" + cmd_arr[2]
	case "SET_BUY_TRIGGER":
		requestURL += "/users/" + cmd_arr[1] + "/set_buy/trigger/" + cmd_arr[2] + "/amount/" + cmd_arr[3]
	case "SET_SELL_AMOUNT":
		requestURL += "/users/" + cmd_arr[1] + "/set_sell/" + cmd_arr[1] + "/amount/" + cmd_arr[2]
	case "SET_SELL_TRIGGER":
		requestURL += "/users/" + cmd_arr[1] + "/set_sell/trigger/" + cmd_arr[2] + "/amount/" + cmd_arr[3]
	case "CANCEL_SET_SELL":
		requestURL += "/users/" + cmd_arr[1] + "/set_sell/cancel/" + cmd_arr[2]
	case "DUMPLOG":
		if len(cmd_arr) == 3 {
			requestURL += "/users/" + cmd_arr[1] + "/dumplog/" + cmd_arr[2]
		}
		if len(cmd_arr) == 2 {
			requestURL += "/dumplog/" + cmd_arr[1]
		}
	case "DISPLAY_SUMMARY":
		requestURL += "/users/" + cmd_arr[1] + "/display_summary"
	}

	if requestURL == "http://localhost:8080" {
		fmt.Printf("Command received: %s, line: %s\n", command, line)
		panic("Unkown command received")
	}

	return requestURL, command
}

// function that sends request to server to execute command given
func executeCmd(reqUrl string, cmd string) {
	fmt.Println(reqUrl)
	// requestURL := "http://localhost:8080/health"
	// res, err := http.MethodPut(requestURL)
	// if err != nil {
	// 	fmt.Printf("client: could not create request: %s\n", err)
	// 	panic(err)
	// }

	// fmt.Printf("client: got response!\n")
	// fmt.Printf("client: status code: %d\n", res.StatusCode)

	// resBody, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	fmt.Printf("client: could not read response body: %s\n", err)
	// 	panic(err)
	// }
	// fmt.Printf("client: response body: %s\n", resBody)

	// if cmd.Command == "ADD" {

	// }
	var res *http.Response
	var err error
	if cmd == "QUOTE" {
		res, err = http.Get(reqUrl)
	}
	if cmd == "ADD" {
		// res, err := http.Get(requestURL)
		fmt.Println("PUT request")
	} else {
		res, err = http.Post(reqUrl, "application/json", reader)
	}

	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("client: response body: %s\n", resBody)
}
