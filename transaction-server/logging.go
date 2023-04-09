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
	LogType         string  `xml:"logType" json:"logType"`
	Timestamp       int64   `xml:"timestamp json:"timestamp"`
	Server          string  `xml:"server json:"server"`
	TransactionNum  int     `xml:"transactionNum" json:"transactionNum"`
	Command         string  `xml:"command" json:"command"`
	Username        string  `xml:"username" json:"username"`
	StockSymbol     string  `xml:"stockSymbol" json:"stockSymbol"`
	Filename        string  `xml:"filename" json:"filename"`
	Funds           float64 `xml:"funds" json:"funds"`
	Price           float64 `xml:"price" json:"price"`
	QuoteServerTime int     `xml:"quoteServerTime" json:"quoteServerTime"`
	Cryptokey       string  `xml:"cryptokey" json:"cryptokey"`
	Action          string  `xml:"action" json:"action"`
	ErrorMessage    string  `xml:"errorMessage" json:"errorMessage"`
	DebugMessage    string  `xml:"debugMessage" json:"debugMessage"`
}

func logEvent(logEntry logEntry) {
	switch logEntry.LogType {
	case USERCOMMAND, SYS_EVENT:
		resp := insert("logs", bson.D{{"LogType", logEntry.LogType}, {"Timestamp", logEntry.Timestamp}, {"Server", logEntry.Server},
			{"TransactionNum", logEntry.TransactionNum}, {"Command", logEntry.Command}, {"Username", logEntry.Username},
			{"StockSymbol", logEntry.StockSymbol}, {"Filename", logEntry.Filename}, {"Funds", logEntry.Funds}})
		if resp != "ok" {
			log.Fatal("Write to DB error")
		}
	case QUOTESERVER:
		resp := insert("logs", bson.D{{"LogType", logEntry.LogType}, {"Timestamp", logEntry.Timestamp}, {"Server", logEntry.Server},
			{"TransactionNum", logEntry.TransactionNum}, {"Price", logEntry.Price}, {"StockSymbol", logEntry.StockSymbol},
			{"Username", logEntry.Username}, {"QuoteServerTime", logEntry.QuoteServerTime}, {"Cryptokey", logEntry.Cryptokey}})
		if resp != "ok" {
			log.Fatal("Write to DB error")
		}
	case ACC_TRANSACTION:
		resp := insert("logs", bson.D{{"LogType", logEntry.LogType}, {"Timestamp", logEntry.Timestamp}, {"Server", logEntry.Server},
			{"TransactionNum", logEntry.TransactionNum}, {"Action", logEntry.Action}, {"Username", logEntry.Username}, {"Funds", logEntry.Funds}})
		if resp != "ok" {
			log.Fatal("Write to DB error")
		}
	case ERR_EVENT:
		resp := insert("logs", bson.D{{"LogType", logEntry.LogType}, {"Timestamp", logEntry.Timestamp}, {"Server", logEntry.Server},
			{"TransactionNum", logEntry.TransactionNum}, {"Command", logEntry.Command}, {"Username", logEntry.Username},
			{"StockSymbol", logEntry.StockSymbol}, {"Filename", logEntry.Filename}, {"Funds", logEntry.Funds}, {"ErrorMessage", logEntry.ErrorMessage}})
		if resp != "ok" {
			log.Fatal("Write to DB error")
		}
	case DEBUG_EVENT:
		resp := insert("logs", bson.D{{"LogType", logEntry.LogType}, {"Timestamp", logEntry.Timestamp}, {"Server", logEntry.Server},
			{"TransactionNum", logEntry.TransactionNum}, {"Command", logEntry.Command}, {"Username", logEntry.Username},
			{"StockSymbol", logEntry.StockSymbol}, {"Filename", logEntry.Filename}, {"Funds", logEntry.Funds}, {"DebugMessage", logEntry.DebugMessage}})
		if resp != "ok" {
			log.Fatal("Write to DB error")
		}
	}
}
