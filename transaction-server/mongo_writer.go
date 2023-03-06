package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func updateOne(collection_ string, who bson.D, with bson.D, _type string) string {

	update := bson.D{{_type, with}}

	ctx := context.TODO()
	uri := "mongodb+srv://daytrading.bpyesvi.mongodb.net/?authSource=%24external&authMechanism=MONGODB-X509&retryWrites=true&w=majority&tlsCertificateKeyFile=/Users/mateomoody/Desktop/X509-cert-3374770886339045150.cer"
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
		panic(err)
		return "Failed to Update Value"
	}
	defer client.Disconnect(ctx)

	collection := client.Database("daytrading").Collection(collection_)
	_, err = collection.UpdateOne(ctx, who, update)
	if err != nil {
		log.Fatal(err)
		panic(err)
		return "Failed to Update Value"
	}

	return "ok"

}
func insert(collection_ string, data bson.D) string {
	ctx := context.TODO()
	uri := "mongodb+srv://daytrading.bpyesvi.mongodb.net/?authSource=%24external&authMechanism=MONGODB-X509&retryWrites=true&w=majority&tlsCertificateKeyFile=/Users/mateomoody/Desktop/X509-cert-3374770886339045150.cer"
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("daytrading").Collection(collection_)
	_, err = collection.InsertOne(ctx, data)
	if err != nil {
		log.Fatal(err)
		return "Failed to Insert Value"
	}

	return "ok"
}
