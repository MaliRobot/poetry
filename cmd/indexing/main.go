package main

import (
	"fmt"
	"log"
	"os"
	"poetry/db"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("You must pass dataset and index name arguments")
		return
	}

	dataset := os.Args[1]
	indexName := os.Args[2]

	mongoDBConnection, err := db.NewMongoDBConnection()
	if err != nil {
		log.Fatalf("Mongo connection error while starting data indexing: %s", err)
	}

	esClient, err := db.ConnectElasticsearch()
	if err != nil {
		log.Fatalf("Elasticsearch error while attempting to index data: %s", err)
	}
	err = db.ReindexData(mongoDBConnection.Client, esClient, dataset, indexName, 4)
	if err != nil {
		log.Fatalf("Failed indexing data: %s", err)
	}
}
