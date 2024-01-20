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

func ReindexData(client *mongo.Client, esClient *elasticsearch.Client) error {
	cfg := configuration.GetConfig()
	collection := client.Database(cfg.DbName).Collection("poems")

	// Connect to Elasticsearch index
	indexName := "poetry"
	err := CreateIndex(esClient, indexName)
	if err != nil {
		return err
	}

	// Retrieve data from MongoDB
	filter := bson.D{{
		Key:   "dataset",
		Value: "kaggle-poetry-foundations-poems",
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

	// Iterate over MongoDB documents and index in Elasticsearch
	for cursor.Next(context.TODO()) {
		var document bson.M
		if err := cursor.Decode(&document); err != nil {
			fmt.Println("reading err")
			return err
		}

		delete(document, "_id")

		// Index document in Elasticsearch
		documentString, err := bson.MarshalExtJSON(document, false, false)
		if err != nil {
			return err
		}

		indexRequest := esapi.IndexRequest{
			Index:   indexName,
			Body:    strings.NewReader(string(documentString)),
			Refresh: "true",
		}

		// Send the index request
		response, err := indexRequest.Do(context.Background(), esClient)
		if err != nil {
			return err
		}

		// Check for errors in the response
		if response.IsError() {
			return fmt.Errorf("error indexing document: %s", response.String())
		}
	}

	return nil
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
