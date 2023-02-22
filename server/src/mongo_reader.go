
package main

import (
	"context"
	"fmt"
	"log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
)


func readOne(){


}

func readMany(){

}

func read(collection_ string, p string) {
	ctx := context.TODO()
  uri := "mongodb+srv://daytrading.bpyesvi.mongodb.net/?authSource=%24external&authMechanism=MONGODB-X509&retryWrites=true&w=majority&tlsCertificateKeyFile=/Users/mateomoody/Desktop/X509-cert-3374770886339045150.cer"
  clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil { log.Fatal(err) }
	defer client.Disconnect(ctx)

	pattern := bson.D{{"user_id", p}}
	fmt.Println(pattern)
	collection := client.Database("myFirstDatabase").Collection(collection_)

	cursor, err := collection.Find(ctx, pattern)
	
	if err != nil {
		panic(err)
  	}
  	var results []bson.D

  	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
  	}
	fmt.Println("RESULTS")
	fmt.Println(results)
	docCount, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil { log.Fatal(err) }
	fmt.Println(docCount)
}
