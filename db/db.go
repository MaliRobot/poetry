package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"poetry/config"
	"time"
)

type MongoDBConnection struct {
	URI      string
	Database string
	Username string
	Password string
	Client   *mongo.Client
}

func NewMongoDBConnection() (*MongoDBConnection, error) {
	cfg := config.GetConfig()

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", cfg.DbUser, cfg.DbPass, cfg.DbHost, cfg.DbPort)

	clientOptions := options.Client().ApplyURI(uri)

	// Set a timeout for the connection attempt.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	fmt.Println("Connected to MongoDB!")

	return &MongoDBConnection{
		URI:      uri,
		Database: cfg.DbName,
		Client:   client,
	}, nil
}

func (c *MongoDBConnection) Disconnect() {
	err := c.Client.Disconnect(context.Background())
	if err != nil {
		log.Printf("Error disconnecting from MongoDB: %v\n", err)
	} else {
		fmt.Println("Disconnected from MongoDB.")
	}
}

func InsertOnePoemIntoDB(mongoDBConnection MongoDBConnection, poem Poem) {
	collection := mongoDBConnection.Client.Database("poetry").Collection("poems")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, &poem)

	if err != nil {
		log.Fatal(err)
	}
	return
}

func InsertManyIntoDB(collection mongo.Collection, documents []interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertMany(ctx, documents)

	if err != nil {
		log.Fatal(err)
	}
}

func GetCollection(databaseName, collectionName string, mongoDBConnection *MongoDBConnection) (*mongo.Collection, error) {
	client := mongoDBConnection.Client
	collection := client.Database(databaseName).Collection(collectionName)
	return collection, nil
}
