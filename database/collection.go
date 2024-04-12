package database

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Collection {
	cilentOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	fmt.Println("Clientoptions TYPE:", reflect.TypeOf(cilentOptions), "\n")

	client, err := mongo.Connect(context.TODO(), cilentOptions)
	if err != nil {
		fmt.Println("Mongo.Connect() ERROR:", err)
		os.Exit(1)
	}
	// ctx, _ := context.WithTimeout(context.Background(), 25*time.Second)

	fmt.Println("Connected to MongoDB!")

	collection := client.Database("attendance").Collection("employees")

	fmt.Println("Collection type:", reflect.TypeOf(collection), "\n")

	return collection
}
