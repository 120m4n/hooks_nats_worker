package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	// "go.mongodb.org/mongo-driver/bson"
	"github.com/120m4n/mongo_nats/model"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Connect to NATS server
	url := "nats://nats_server:4222"
	nc, err := nats.Connect(url)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://mongo_server:27017")
	mongoClient, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Get a handle for your collection
	collection := mongoClient.Database("test").Collection("coordinates")

	// Subscribe to "coordinates" topic
	_, err = nc.Subscribe("coordinates", func(m *nats.Msg) {
		var doc model.Document
		err := json.Unmarshal(m.Data, &doc)
		if err != nil {
			log.Printf("Error unmarshalling data: %v", err)
			return
		}

		// Now you can use the `doc` variable
		fmt.Printf("Received a document: %+v\n", doc)

		// Insert into MongoDB
		_, err = collection.InsertOne(context.Background(), doc)
		if err != nil {
			log.Printf("Error inserting into MongoDB: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("Error subscribing to topic: %v", err)
	}

	// Keep the connection alive
	select {}
}
