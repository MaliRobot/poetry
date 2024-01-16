package db

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

type MongoDBConnection struct {
	URI      string
	Database string
	Username string
	Password string
	Client   *mongo.Client
}

func loadEnv() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}
	return nil
}

func NewMongoDBConnection() (*MongoDBConnection, error) {
	if err := loadEnv(); err != nil {
		return nil, err
	}

	// Access the environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	database := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")

	uri := fmt.Sprintf("mongodb://%s:%s", dbHost, dbPort)

	clientOptions := options.Client().ApplyURI(uri)

	clientOptions.Auth = &options.Credential{
		Username: dbUser,
		Password: dbPass,
	}

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
		Database: database,
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
