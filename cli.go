package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli"
)

// Cmd struct is a representation of an isolated command executed by a user
type Cmd struct {
	Command  string `json:"Command"`
	Username string `json:"Username"`
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
		data := parseLine(scanner.Text())
		if false {
			fmt.Println(data)
		}
		// use parsed info to make get/post requests ?
	}

	if err := scanner.Err(); err != nil {
		return (err)
	}

	return nil
}

func parseLine(line string) *Cmd {
	line_arr := strings.Split(line, " ")
	cmd_arr := strings.Split(line_arr[1], ",")
	command := cmd_arr[0]

	if command == "ADD" {
		return &Cmd{Command: command, Username: cmd_arr[1], Amount: cmd_arr[2]}

	} else if command == "BUY" || command == "SELL" || command == "SET_BUY_AMOUNT" || command == "SET_BUY_TRIGGER" ||
		command == "SET_SELL_AMOUNT" || command == "SET_SELL_TRIGGER" {
		return &Cmd{Command: command, Username: cmd_arr[1], Stock: cmd_arr[2], Amount: cmd_arr[3]}

	} else if command == "QUOTE" || command == "CANCEL_SET_BUY" || command == "CANCEL_SET_SELL" {
		return &Cmd{Command: command, Username: cmd_arr[1], Stock: cmd_arr[2]}

	} else if command == "COMMIT_BUY" || command == "COMMIT_SELL" || command == "CANCEL_BUY" || command == "CANCEL_SELL" || command == "DISPLAY_SUMMARY" {
		return &Cmd{Command: command, Username: cmd_arr[1]}

	} else if command == "DUMPLOG" {
		if len(cmd_arr) == 2 {
			return &Cmd{Command: command, Filename: cmd_arr[1]}
		} else {
			return &Cmd{Command: command, Username: cmd_arr[1], Filename: cmd_arr[2]}
		}

	} else {
		fmt.Printf("Command received: %s, line: %s\n", command, line)
		panic("Unknown command received")

	}
}

// function that sends request  to server to execute command given
func executeCmd(cmd *Cmd) {

}
