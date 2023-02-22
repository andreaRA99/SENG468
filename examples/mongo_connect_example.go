
package example



// DONT USE.  FOR EXAMPLE PURPOSES ONLY

import (
	"context"
	"fmt"
	"log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	ctx := context.TODO()
  uri := "mongodb+srv://daytrading.bpyesvi.mongodb.net/?authSource=%24external&authMechanism=MONGODB-X509&retryWrites=true&w=majority&tlsCertificateKeyFile=/Users/mateomoody/Desktop/X509-cert-3374770886339045150.cer"
  clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil { log.Fatal(err) }
	

	collection := client.Database("myFirstDatabase").Collection("TEST")

	fmt.Println("Connection Successful!")


	result, err := collection.InsertOne(ctx, bson.D{{"user_id","10"}, {"encrypted_key","10"}})


	docCount, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil { log.Fatal(err) }
}
