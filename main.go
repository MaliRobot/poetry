package main

import (
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

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/collections", func(c *gin.Context) {
		getCollections(c, mongoDBConnection)
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
