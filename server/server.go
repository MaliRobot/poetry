package server

import (
	"encoding/json"
	"fmt"
	"log"
	db "poetry/db"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

var jobChan = make(chan []db.Poem, 10)

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

type AddPoemRequest struct {
	Dataset   string `json:"dataset"`
	Title     string `json:"title" binding:"required"`
	Poem      string `json:"poem" binding:"required"`
	Poet      string `json:"poet"`
	Tags      string `json:"tags"`
	Language  string `json:"language" binding:"required"`
	DatasetId string `json:"dataset_id"`
}

func addPoem(c *gin.Context, connection *db.MongoDBConnection) {
	var req AddPoemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tags := []string{}
	if req.Tags != "" {
		tags = strings.Split(req.Tags, ",")
	}

	poem := db.Poem{
		Dataset:   req.Dataset,
		DatasetId: req.DatasetId,
		Title:     req.Title,
		Poem:      req.Poem,
		Poet:      req.Poet,
		Tags:      tags,
		Language:  req.Language,
	}

	db.InsertOnePoemIntoDB(*connection, poem)
	c.JSON(200, gin.H{"message": "Poem added successfully"})
}

func startWorker(connection *db.MongoDBConnection) {
	go func() {
		for poems := range jobChan {
			var documents []interface{}
			for _, p := range poems {
				documents = append(documents, p)
			}
			collection, _ := db.GetCollection("poetry", "poems", connection)
			db.InsertManyIntoDB(*collection, documents)
		}
	}()
}

func addPoems(c *gin.Context, connection *db.MongoDBConnection) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "File is required"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to open file"})
		return
	}
	defer f.Close()
	var req []AddPoemRequest
	if err := json.NewDecoder(f).Decode(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON"})
		return
	}
	var poems []db.Poem
	for _, r := range req {
		tags := []string{}
		if r.Tags != "" {
			tags = strings.Split(r.Tags, ",")
		}
		poem := db.Poem{
			Dataset:   r.Dataset,
			DatasetId: r.DatasetId,
			Title:     r.Title,
			Poem:      r.Poem,
			Poet:      r.Poet,
			Tags:      tags,
			Language:  r.Language,
		}
		poems = append(poems, poem)
	}
	jobChan <- poems
	c.JSON(200, gin.H{"message": "Poems scheduled for processing"})
}

func Start() {
	mongoDBConnection, err := db.NewMongoDBConnection()

	if err != nil {
		log.Fatal(err)
	}
	defer mongoDBConnection.Disconnect()

	startWorker(mongoDBConnection)

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

		c.JSON(200, gin.H{"message": "Search successful"})
	})
	r.POST("/poem", func(c *gin.Context) {
		addPoem(c, mongoDBConnection)
	})
	r.POST("/poems", func(c *gin.Context) {
		addPoems(c, mongoDBConnection)
	})
	err = r.Run()
	if err != nil {
		fmt.Printf("Error running the server: %v\n", err)
		return
	}
}
