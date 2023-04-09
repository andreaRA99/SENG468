package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli"
)

type logEntry struct {
	LogType         string  `xml:"logType"`
	Timestamp       int64   `xml:"timestamp"`
	Server          string  `xml:"server"`
	TransactionNum  int     `xml:"transactionNum"`
	Command         string  `xml:"command"`
	Username        string  `xml:"username"`
	StockSymbol     string  `xml:"stockSymbol"`
	Filename        string  `xml:"filename"`
	Funds           float64 `xml:"funds"`
	Price           float64 `xml:"price"`
	QuoteServerTime int     `xml:"quoteServerTime"`
	Cryptokey       string  `xml:"cryptokey"`
	Action          string  `xml:"action"`
	ErrorMessage    string  `xml:"errorMessage"`
	DebugMessage    string  `xml:"debugMessage"`
}

// Cmd struct is a representation of an isolated command executed by a user
type Cmd struct {
	Command  string  `json:"cmd"`
	Id       string  `json:"id"`
	Stock    string  `json:"stock"`
	Amount   float64 `json:"amount"`
	Filename string  `json:"filename"`
	Price 	float64 `json:"Price"`
}

func main() {
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

				command := strings.ToUpper(c.String("cmd"))
				id := c.String("id")
				stock := strings.ToUpper(c.String("stock"))
				amount, err := strconv.ParseFloat(c.String("amount"), 64)
				if err != nil {
					panic(err)
				}
				filename := c.String("filename")

				cmd := Cmd{Command: command, Id: id, Stock: stock, Amount: amount, Filename: filename}

				executeCmd(cmd)
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
					Name:  "id",
					Usage: "username",
				},
				cli.StringFlag{
					Name:  "stock",
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

// Function reads file containing commands then parses and executes them line by line
func readFromFile(c *cli.Context) error {
	fileLocation := c.String("filelocation")
	file, err := os.Open(fileLocation)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		cmd := parseLine(scanner.Text())
		executeCmd(cmd)
	}

	if err := scanner.Err(); err != nil {
		return (err)
	}

	return nil
}

// Function parses single line from file and returns struct containing command params
func parseLine(line string) Cmd {
	line_arr := strings.Split(line, " ")
	cmd_arr := strings.Split(line_arr[1], ",")
	command := cmd_arr[0]

	switch command {
	case "ADD":
		amount, err := strconv.ParseFloat(cmd_arr[2], 64)
		if err != nil {
			panic(err)
		}
		return Cmd{Command: command, Id: cmd_arr[1], Amount: amount}
	case "BUY", "SELL", "SET_BUY_AMOUNT",  "SET_SELL_AMOUNT":
		amount, err := strconv.ParseFloat(cmd_arr[3], 64)
		if err != nil {
			panic(err)
		}
		return Cmd{Command: command, Id: cmd_arr[1], Stock: cmd_arr[2], Amount: amount}
	case "SET_BUY_TRIGGER", "SET_SELL_TRIGGER":
		price, err := strconv.ParseFloat(cmd_arr[3], 64)
		if err != nil {
			panic(err)
		}
		return Cmd{Command: command, Id: cmd_arr[1], Stock: cmd_arr[2], Price: price}
	case "QUOTE", "CANCEL_SET_BUY", "CANCEL_SET_SELL":
		return Cmd{Command: command, Id: cmd_arr[1], Stock: cmd_arr[2]}
	case "COMMIT_BUY", "COMMIT_SELL", "CANCEL_BUY", "CANCEL_SELL", "DISPLAY_SUMMARY":
		return Cmd{Command: command, Id: cmd_arr[1]}
	case "DUMPLOG":
		if len(cmd_arr) == 2 {
			return Cmd{Command: command, Filename: cmd_arr[1]}
		} else if len(cmd_arr) == 3 {
			return Cmd{Command: command, Id: cmd_arr[1], Filename: cmd_arr[2]}
		}
	}
	// fmt.Printf("Command received: %s, line: %s\n", command, line)
	panic("Unknown command received")
}

// Function sends request to server to execute command given
func executeCmd(cmd Cmd) {
	// fmt.Println(cmd)
	var req *http.Request
	var err error

	reqUrlPrefix := "http://localhost:8080"

	parsedJson, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
	}

	switch cmd.Command {
	case "ADD":
		req, err = http.NewRequest(http.MethodPut, reqUrlPrefix+"/users/addBal", bytes.NewBuffer(parsedJson))
	case "QUOTE":
		req, err = http.NewRequest(http.MethodGet, reqUrlPrefix+"/users/"+cmd.Id+"/quote/"+cmd.Stock, nil)
	case "BUY":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/buy", bytes.NewBuffer(parsedJson))
	case "COMMIT_BUY":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/buy/commit", bytes.NewBuffer(parsedJson))
	case "CANCEL_BUY":
		req, err = http.NewRequest(http.MethodDelete, reqUrlPrefix+"/users/"+cmd.Id+"/buy/cancel", nil)
	case "SELL":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/sell", bytes.NewBuffer(parsedJson))
	case "COMMIT_SELL":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/sell/commit", bytes.NewBuffer(parsedJson))
	case "CANCEL_SELL":
		req, err = http.NewRequest(http.MethodDelete, reqUrlPrefix+"/users/"+cmd.Id+"/sell/cancel", nil)
	case "SET_BUY_AMOUNT":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/set/buy", bytes.NewBuffer(parsedJson))
	case "CANCEL_SET_BUY":
		req, err = http.NewRequest(http.MethodDelete, reqUrlPrefix+"/users/"+cmd.Id+"/set/buy/"+cmd.Stock+"/cancel", nil)
	case "SET_BUY_TRIGGER":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/set/buy/trigger", bytes.NewBuffer(parsedJson))
	case "SET_SELL_AMOUNT":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/set/sell", bytes.NewBuffer(parsedJson))
	case "SET_SELL_TRIGGER":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/set/sell/trigger", bytes.NewBuffer(parsedJson))
	case "CANCEL_SET_SELL":
		req, err = http.NewRequest(http.MethodDelete, reqUrlPrefix+"/users/"+cmd.Id+"/set/sell/"+cmd.Stock+"/cancel", nil)
	case "DUMPLOG":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/dumplog", bytes.NewBuffer(parsedJson))
	case "DISPLAY_SUMMARY":
		req, err = http.NewRequest(http.MethodGet, reqUrlPrefix+"/displaysummary/"+cmd.Id, nil)
	}

	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Parse response body
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	// map each to logEntry struct
	var resStr []logEntry
	json.Unmarshal(resBody, &resStr)

	if cmd.Command == "DUMPLOG" {
		logsToFile(resStr)
	}

	if cmd.Command == "DISPLAY_SUMMARY" {
		displaySummary(resBody)
	}
}

func logsToFile(resp []logEntry) {
	// receiving in json, write in xml
	// fmt.Println(resp[len(resp)-1].Filename)
	filename := resp[len(resp)-1].Filename

	// Write to file
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer file.Close()

	_, err = file.WriteString("<?xml version='1.0'?>\n<log>\n")
	if err != nil {
		log.Fatal(err)
		return
	}

	// write logs from db
	for _, log_entry := range resp {
		// fmt.Println(log_entry)
		_, err = file.WriteString("\t<" + log_entry.LogType + ">\n")
		if err != nil {
			log.Fatal(err)
			return
		}
		switch log_entry.LogType {
		case "userCommand", "systemEvent":
			_, err = file.WriteString(
				"\t\t<timestamp>" + fmt.Sprintf("%d", log_entry.Timestamp) + "</timestamp>\n" +
					"\t\t<server>" + log_entry.Server + "</server>\n" +
					"\t\t<transactionNum>" + fmt.Sprintf("%d", log_entry.TransactionNum) + "</transactionNum>\n" +
					"\t\t<command>" + log_entry.Command + "</command>\n" +
					"\t\t<username>" + log_entry.Username + "</username>\n" +
					"\t\t<stockSymbol>" + log_entry.StockSymbol + "</stockSymbol>\n" +
					"\t\t<filename>" + log_entry.Filename + "</filename>\n" +
					"\t\t<funds>" + fmt.Sprintf("%f", log_entry.Funds) + "</funds>\n")
			if err != nil {
				log.Fatal(err)
				return
			}

		case "quoteServer":
			_, err = file.WriteString(
				"\t\t<timestamp>" + fmt.Sprintf("%d", log_entry.Timestamp) + "</timestamp>\n" +
					"\t\t<server>" + log_entry.Server + "</server>\n" +
					"\t\t<transactionNum>" + fmt.Sprintf("%d", log_entry.TransactionNum) + "</transactionNum>\n" +
					"\t\t<price>" + fmt.Sprintf("%f", log_entry.Price) + "</price>\n" +
					"\t\t<stockSymbol>" + log_entry.StockSymbol + "</stockSymbol>\n" +
					"\t\t<username>" + log_entry.Username + "</username>\n" +
					"\t\t<quoteServerTime>" + fmt.Sprintf("%d", log_entry.QuoteServerTime) + "</quoteServerTime>\n" +
					"\t\t<crytokey>" + log_entry.Cryptokey + "</cryptokey>\n")
			if err != nil {
				log.Fatal(err)
				return
			}

		case "accountTransaction":
			_, err = file.WriteString(
				"\t\t<timestamp>" + fmt.Sprintf("%d", log_entry.Timestamp) + "</timestamp>\n" +
					"\t\t<server>" + log_entry.Server + "</server>\n" +
					"\t\t<transactionNum>" + fmt.Sprintf("%d", log_entry.TransactionNum) + "</transactionNum>\n" +
					"\t\t<action>" + log_entry.Action + "</action>\n" +
					"\t\t<username>" + log_entry.Username + "</username>\n" +
					"\t\t<funds>" + fmt.Sprintf("%f", log_entry.Funds) + "</funds>\n")
			if err != nil {
				log.Fatal(err)
				return
			}
		case "errorEvent":
			_, err = file.WriteString(
				"\t\t<timestamp>" + fmt.Sprintf("%d", log_entry.Timestamp) + "</timestamp>\n" +
					"\t\t<server>" + log_entry.Server + "</server>\n" +
					"\t\t<transactionNum>" + fmt.Sprintf("%d", log_entry.TransactionNum) + "</transactionNum>\n" +
					"\t\t<command>" + log_entry.Command + "</command>\n" +
					"\t\t<username>" + log_entry.Username + "</username>\n" +
					"\t\t<stockSymbol>" + log_entry.StockSymbol + "</stockSymbol>\n" +
					"\t\t<filename>" + log_entry.Filename + "</filename>\n" +
					"\t\t<funds>" + fmt.Sprintf("%f", log_entry.Funds) + "</funds>\n" +
					"\t\t<errorMessage>" + log_entry.ErrorMessage + "</errorMessage>\n")
			if err != nil {
				log.Fatal(err)
				return
			}

		case "debugEvent":
			_, err = file.WriteString(
				"\t\t<timestamp>" + fmt.Sprintf("%d", log_entry.Timestamp) + "</timestamp>\n" +
					"\t\t<server>" + log_entry.Server + "</server>\n" +
					"\t\t<transactionNum>" + fmt.Sprintf("%d", log_entry.TransactionNum) + "</transactionNum>\n" +
					"\t\t<command>" + log_entry.Command + "</command>\n" +
					"\t\t<username>" + log_entry.Username + "</username>\n" +
					"\t\t<stockSymbol>" + log_entry.StockSymbol + "</stockSymbol>\n" +
					"\t\t<filename>" + log_entry.Filename + "</filename>\n" +
					"\t\t<funds>" + fmt.Sprintf("%f", log_entry.Funds) + "</funds>\n" +
					"\t\t<debugEvent>" + log_entry.ErrorMessage + "</debugEvent>\n")
			if err != nil {
				log.Fatal(err)
				return
			}

		}
		_, err = file.WriteString("\t</" + log_entry.LogType + ">\n")
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	_, err = file.WriteString("</log>")
	if err != nil {
		log.Fatal(err)
		return
	}
}

func displaySummary(resp []byte) {
	// print to stdout
}