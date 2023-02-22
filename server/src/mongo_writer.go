
package main

import (
	"context"
	"fmt"
	"log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
)

func insert(collection_ string, data bson.D) {
	ctx := context.TODO()
  uri := "mongodb+srv://daytrading.bpyesvi.mongodb.net/?authSource=%24external&authMechanism=MONGODB-X509&retryWrites=true&w=majority&tlsCertificateKeyFile=/Users/mateomoody/Desktop/X509-cert-3374770886339045150.cer"
  clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil { log.Fatal(err) }
	defer client.Disconnect(ctx)

	collection := client.Database("myFirstDatabase").Collection("TEST")
	result, err := collection.InsertOne(ctx, bson.D{{"user_id","10"}, {"encrypted_key","10"}})
	fmt.Println(result)
	fmt.Println(err)
	docCount, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil { log.Fatal(err) }
	fmt.Println(docCount)
}
