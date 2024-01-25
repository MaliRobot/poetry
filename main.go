package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	db "main/db"
)

func getCollections(c *gin.Context, connection *db.MongoDBConnection) {
	collection, _ := db.GetCollection("poetry", "poems", connection)

	results, err := collection.Distinct(c.Request.Context(), "dataset", bson.D{})

	if err != nil {
		log.Fatal(err)
	}

	c.JSON(200, gin.H{
		"datasets": results,
	})
}

func main() {
	mongoDBConnection, err := db.NewMongoDBConnection()

	if err != nil {
		log.Fatal(err)
	}
	defer mongoDBConnection.Disconnect()

	esClient, err := db.ConnectElasticsearch()
	if err != nil {
		fmt.Printf("Error connecting to Elasticsearch: %v\n", err)
		return
	}

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/collections", func(c *gin.Context) {
		getCollections(c, mongoDBConnection)
	})
	r.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(400, gin.H{"error": "Query parameter 'q' is required"})
			return
		}

		err := db.SearchData(esClient, query)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Process the search response and send it as JSON to the client
		// ...

		c.JSON(200, gin.H{"message": "Search successful"})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
