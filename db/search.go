package db

import (
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	configuration "main/config"
	"net/http"
	"strings"
	"sync"
)

func ConnectElasticsearch() (*elasticsearch.Client, error) {
	cfg := configuration.GetConfig()
	config := elasticsearch.Config{
		Addresses: []string{cfg.ElasticUrl},
	}
	client, err := elasticsearch.NewClient(config)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to Elasticsearch!")
	return client, nil
}

func CreateIndex(esClient *elasticsearch.Client, indexName string) error {
	existsRequest := esapi.IndicesExistsRequest{
		Index: []string{indexName},
	}

	response, err := existsRequest.Do(context.Background(), esClient)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		// Index already exists, no need to create it
		return nil
	}

	response, err = esClient.Indices.Create(indexName)

	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)

	if response.IsError() {
		return fmt.Errorf("error creating index: %s", response.String())
	}

	return nil
}

func ReindexData(client *mongo.Client, esClient *elasticsearch.Client, dataset string, indexName string, numWorkers int) error {
	cfg := configuration.GetConfig()
	collection := client.Database(cfg.DbName).Collection("poems")

	// Connect to Elasticsearch index
	err := CreateIndex(esClient, indexName)
	if err != nil {
		return err
	}

	// Retrieve data from MongoDB
	filter := bson.D{{
		Key:   "dataset",
		Value: dataset,
	}}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("collection err")
		return err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {

		}
	}(cursor, context.TODO())

	// Prepare bulk request
	var bulkRequest strings.Builder

	// Wait group to wait for all workers to finish
	var wg sync.WaitGroup

	// Channel to signal completion of each worker
	workerDone := make(chan struct{})

	// Channel to send bulk requests to workers
	bulkRequests := make(chan string, numWorkers)

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(esClient, indexName, bulkRequests, workerDone, &wg)
	}

	// Iterate over MongoDB documents and append to bulk request
	for cursor.Next(context.TODO()) {
		var document bson.M
		if err := cursor.Decode(&document); err != nil {
			fmt.Println("reading err")
			close(bulkRequests)
			wg.Wait()
			return err
		}

		delete(document, "_id")

		// Prepare the action line for the bulk request
		bulkRequest.WriteString(fmt.Sprintf(`{"index":{"_index":"%s"}}%s`, indexName, "\n"))

		// Convert document to JSON and append to bulk request
		documentString, err := bson.MarshalExtJSON(document, false, false)
		if err != nil {
			close(bulkRequests)
			wg.Wait()
			return err
		}
		bulkRequest.WriteString(fmt.Sprintf("%s%s", string(documentString), "\n"))

		// Send bulk request to workers when batch size is reached
		if bulkRequest.Len() > 10*1024 { // Adjust the threshold based on your requirements
			bulkRequests <- bulkRequest.String()
			bulkRequest.Reset()
		}
	}

	// Send any remaining documents as a final bulk request
	if bulkRequest.Len() > 0 {
		bulkRequests <- bulkRequest.String()
	}

	fmt.Println("Closing bulk request channel")

	// Close the bulk request channel to signal workers to finish
	close(bulkRequests)

	fmt.Println("Waiting for workers to finish")

	// Wait for all workers to finish
	wg.Wait()

	fmt.Println("Close workerDone channel")

	// Close the workerDone channel to release any waiting goroutines
	close(workerDone)

	return nil
}

func worker(esClient *elasticsearch.Client, indexName string, bulkRequests <-chan string, workerDone chan<- struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for bulkRequest := range bulkRequests {
		// Send bulk request to Elasticsearch
		response, err := esapi.BulkRequest{
			Index:   indexName,
			Body:    strings.NewReader(bulkRequest),
			Refresh: "false",
		}.Do(context.Background(), esClient)
		if err != nil {
			fmt.Println("error during bulk indexing:", err)
			return
		}

		// Check for errors in the response
		if response.IsError() {
			fmt.Printf("error during bulk indexing: %s\n", response.String())
			return
		}
	}

	// Signal completion to the main function
	workerDone <- struct{}{}
}

func SearchData(esClient *elasticsearch.Client, query string) error {
	searchRequest := esapi.SearchRequest{
		Index: []string{"poems"},
		Body:  strings.NewReader(fmt.Sprintf(`{"query":{"match":{"field":"%s"}}}`, query)),
	}

	response, err := searchRequest.Do(context.Background(), esClient)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)

	// TODO Process the search response

	return nil
}
