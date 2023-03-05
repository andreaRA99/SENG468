package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
)



func mongo_read_bsonA(v bson.A) ([]holding){
	// Rather than Unmarshalling, which is best practice, this is a hack workaround since 
	// I couldnt get the Unmarshalling to work.


	// Hard coded to only work for reading for fields in account_holdings[{}]

	// bson.A is an array of bson.D, which contains bson.E, which is key value pairs.
	// loop through A to find each bson.D that corresponds to a document per holding, loop through that bson.D 
	// to extract the value

	var e []holding		


	for _, s := range v{
		fmt.Println(s)
		var temp holding

		switch c := s.(type){
			case bson.D:
			{
				for _, kv_pair := range c {
			
					switch d := kv_pair.Value.(type){
						case string:
						{
							temp.symbol = d
						}
						case int32:
						{
							fmt.Printf("Should never be this\n\n",d)
							fmt.Printf("%d\n\n",d)

						}			
						case float64:
						{
							if kv_pair.Key == "quantity" {
								temp.quantity = d
							} 	else // Value is pps
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