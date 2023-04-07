package main

import (
	"context"

	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func rawreadField(collection_ string, filter bson.D, fields bson.D) []bson.D {
	databaseUri, found := os.LookupEnv("DATABASE_URI")
	if !found {
		log.Fatalln("No DATABASE_URI")
	}
	ctx := context.TODO()
	clientOptions := options.Client().ApplyURI(databaseUri)
	client, err := mongo.Connect(ctx, clientOptions)
	opts := options.Find().SetProjection(fields)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	collection := client.Database("daytrading").Collection(collection_)
	cursor, err := collection.Find(ctx, filter, opts)

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}
	if err != nil {
		panic(err)
	}

	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {

		panic(err)
	}

	return results
}


func readOne(collection_ string, filter bson.D) bson.D {
	databaseUri, found := os.LookupEnv("DATABASE_URI")
	if !found {
		log.Fatalln("No DATABASE_URI")
	}
	ctx := context.TODO()

	clientOptions := options.Client().ApplyURI(databaseUri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("daytrading").Collection(collection_)

	var results bson.D

	err = collection.FindOne(ctx, filter).Decode(&results)

	if err != nil {
		return bson.D{{"none", "none"}}
	}
	return results
}

func readMany(collection_ string, filter bson.D) []bson.D {
	databaseUri, found := os.LookupEnv("DATABASE_URI")
	if !found {
		log.Fatalln("No DATABASE_URI")
	}
	ctx := context.TODO()

	clientOptions := options.Client().ApplyURI(databaseUri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("daytrading").Collection(collection_)

	cursor, err := collection.Find(ctx, filter)

	var results []bson.D

	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	return results
}
