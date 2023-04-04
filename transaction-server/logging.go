package main

import (
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

var USERCOMMAND = "userCommand"
var QUOTESERVER = "quoteServer"
var ACC_TRANSACTION = "accountTransaction"
var SYS_EVENT = "systemEvent"
var ERR_EVENT = "errorEvent"
var DEBUG_EVENT = "debugEvent"

type logEntry struct {
	LogType      string  `xml:"logType"`
	Timestamp    int64   `xml:"timestamp"`
	Server       string  `xml:"server"`
	Tnum         int     `xml:"transactionNum"`
	Command      string  `xml:"command"`
	Username     string  `xml:"username"`
	Stock        string  `xml:"stockSymbol"`
	Filename     string  `xml:"filename"`
	Funds        float64 `xml:"funds"`
	Price        float64 `xml:"price"`
	QSTime       int     `xml:"quoteServerTime"`
	Cryptokey    string  `xml:"cryptokey"`
	Action       string  `xml:"action"`
	ErrorMessage string  `xml:"errorMessage"`
	DebugMessage string  `xml:"debugMessage"`
}

func logEvent(logEntry logEntry) {
	switch logEntry.LogType {
	case USERCOMMAND, SYS_EVENT:
		resp := insert("logs", bson.D{{"logType", logEntry.LogType}, {"timestamp", logEntry.Timestamp}, {"server", logEntry.Server},
			{"transactionNum", logEntry.Tnum}, {"command", logEntry.Command}, {"username", logEntry.Username},
			{"stockSymbol", logEntry.Stock}, {"filename", logEntry.Filename}, {"funds", logEntry.Funds}})
		if resp != "ok" {
			log.Fatal("Write to DB error")
		}
	case QUOTESERVER:
		resp := insert("logs", bson.D{{"logType", logEntry.LogType}, {"timestamp", logEntry.Timestamp}, {"server", logEntry.Server},
			{"transactionNum", logEntry.Tnum}, {"price", logEntry.Price}, {"stockSymbol", logEntry.Stock},
			{"username", logEntry.Username}, {"quoteServerTime", logEntry.QSTime}, {"cryptoKey", logEntry.Cryptokey}})
		if resp != "ok" {
			log.Fatal("Write to DB error")
		}
	case ACC_TRANSACTION:
		resp := insert("logs", bson.D{{"logType", logEntry.LogType}, {"timestamp", logEntry.Timestamp}, {"server", logEntry.Server},
			{"transactionNum", logEntry.Tnum}, {"action", logEntry.Action}, {"username", logEntry.Username}, {"funds", logEntry.Funds}})
		if resp != "ok" {
			log.Fatal("Write to DB error")
		}
	case ERR_EVENT:
		resp := insert("logs", bson.D{{"logType", logEntry.LogType}, {"timestamp", logEntry.Timestamp}, {"server", logEntry.Server},
			{"transactionNum", logEntry.Tnum}, {"command", logEntry.Command}, {"username", logEntry.Username},
			{"stockSymbol", logEntry.Stock}, {"filename", logEntry.Filename}, {"funds", logEntry.Funds}, {"errorMessage", logEntry.ErrorMessage}})
		if resp != "ok" {
			log.Fatal("Write to DB error")
		}
	case DEBUG_EVENT:
		resp := insert("logs", bson.D{{"logType", logEntry.LogType}, {"timestamp", logEntry.Timestamp}, {"server", logEntry.Server},
			{"transactionNum", logEntry.Tnum}, {"command", logEntry.Command}, {"username", logEntry.Username},
			{"stockSymbol", logEntry.Stock}, {"filename", logEntry.Filename}, {"funds", logEntry.Funds}, {"debugMessage", logEntry.DebugMessage}})
		if resp != "ok" {
			log.Fatal("Write to DB error")
		}
	}
	transaction_counter += 1
	// return logEntry
}

// func main() {
// 	// uc := userCommand{Timestamp: time.Now(), Server: "own-server"}
// 	// thing := reflect.TypeOf(uc)
// 	// fmt.Println(reflect.TypeOf(thing))
// 	log(userCommand{Timestamp: time.Now(), Server: "own-server"})
// }
