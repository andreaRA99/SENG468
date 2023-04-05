package main

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

func mongo_read_logs(v []bson.D) []logEntry {
	// fmt.Println("Entering mongo util")
	var temp []logEntry
	for _, s := range v {
		fmt.Println(s)
		var e logEntry
		for _, kv_pair := range s {
			tempk := kv_pair.Key
			tempv := kv_pair.Value
			switch d := tempv.(type) {

			case string:
				{
					if tempk == "LogType" {
						e.LogType = d
					}
					if tempk == "Server" {
						e.Server = d
					}
					if tempk == "Command" {
						e.Command = d
					}
					if tempk == "Username" {
						e.Username = d
					}
					if tempk == "StockSymbol" {
						e.StockSymbol = d
					}
					if tempk == "Filename" {
						e.Filename = d
					}
					if tempk == "Cryptokey" {
						e.Cryptokey = d
					}
					if tempk == "Action" {
						e.Action = d
					}
					if tempk == "ErrorMessage" {
						e.ErrorMessage = d
					}
					if tempk == "DebugMessage" {
						e.DebugMessage = d
					}
				}
			case int64:
				{
					e.Timestamp = d
				}
			case int:
				{
					if tempk == "TransactionNum" {
						e.TransactionNum = d
					}
					if tempk == "QuoteServerTime" {
						e.QuoteServerTime = d
					}
				}
			case float64:
				{
					if tempk == "Funds" {
						e.Funds = d
					}
					if tempk == "Price" {
						e.Price = d
					}
				}
			}
		}
		fmt.Println(e)
		temp = append(temp, e)
	}
	// fmt.Println("Leaving mongo utils")
	return temp
}

func mongo_read_bsonA(v bson.A) []holding {
	// Rather than Unmarshalling, which is best practice, this is a hack workaround since
	// I couldnt get the Unmarshalling to work.

	// Hard coded to only work for reading for fields in account_holdings[{}]

	// bson.A is an array of bson.D, which contains bson.E, which is key value pairs.
	// loop through A to find each bson.D that corresponds to a document per holding, loop through that bson.D
	// to extract the value

	var e []holding

	for _, s := range v {
		fmt.Println(s)
		var temp holding

		switch c := s.(type) {
		case bson.D:
			{
				for _, kv_pair := range c {

					switch d := kv_pair.Value.(type) {
					case string:
						{
							temp.symbol = d
						}
					case int32:
						{
							fmt.Printf("Should never be this\n\n", d)
							fmt.Printf("%d\n\n", d)

						}
					case float64:
						{
							if kv_pair.Key == "quantity" {
								temp.quantity = d
							} else // Value is pps
							{
								temp.pps = d
							}

						}
					}
				}
			}
		}
		e = append(e, temp)
	}

	return e
}
