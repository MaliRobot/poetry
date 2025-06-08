package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	db "poetry/db"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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

func sendJobToWorker(poems []db.Poem) error {
	workerURL := getWorkerURL()

	jsonData, err := json.Marshal(poems)
	if err != nil {
		return fmt.Errorf("failed to marshal poems: %v", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Post(workerURL+"/jobs", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send job to worker: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("worker rejected job with status: %d", resp.StatusCode)
	}

	return nil
}

func getWorkerURL() string {
	workerHost := os.Getenv("WORKER_HOST")
	if workerHost == "" {
		workerHost = "localhost"
	}

	workerPort := os.Getenv("WORKER_PORT")
	if workerPort == "" {
		workerPort = "8081"
	}

	return fmt.Sprintf("http://%s:%s", workerHost, workerPort)
}

func addPoems(c *gin.Context) {
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

	// Send job to worker service
	err = sendJobToWorker(poems)
	if err != nil {
		log.Printf("Failed to send job to worker: %v", err)
		c.JSON(503, gin.H{"error": "Worker service unavailable"})
		return
	}

	c.JSON(200, gin.H{"message": "Poems scheduled for processing"})
}

func Start() {
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

		c.JSON(200, gin.H{"message": "Search successful"})
	})
	r.POST("/poem", func(c *gin.Context) {
		addPoem(c, mongoDBConnection)
	})
	r.POST("/poems", func(c *gin.Context) {
		addPoems(c)
	})
	err = r.Run()
	if err != nil {
		fmt.Printf("Error running the server: %v\n", err)
		return
	}
}
