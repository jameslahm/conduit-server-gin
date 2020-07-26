package utils

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// GetConnection connect to db
func GetConnection() (*mongo.Client, context.Context, context.CancelFunc) {
	mongoURI := os.Getenv("MONGODBURI")
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Error:Generate Mongo Client!")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal("Error:Connect Error")
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("Error:Ping Error")
	}
	return client, ctx, cancel
}
