package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"main/db"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var DATAPATH = "cmd/parse_csv/data"

func customSplit(input string) []string {
	re := regexp.MustCompile(`[^,]+,[^,]+`)
	result := re.FindAllString(input, -1)
	return result
}

func createCSVReader(filePath string) (*csv.Reader, *os.File, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting current working directory:", err)
	}

	file, err := os.Open(filepath.Join(currentDir, DATAPATH, filePath))

	if err != nil {
		return nil, nil, err
	}

	reader := csv.NewReader(file)
	_, err = reader.Read()

	if err != nil {
		return nil, nil, err
	}

	return reader, file, nil
}

func importPoetryFoundation(mongoDBConnection db.MongoDBConnection) bool {
	path := "PoetryFoundationData.csv"
	reader, file, err := createCSVReader(path)

	if err != nil {
		log.Fatal(err)
	}

	for {
		record, err := reader.Read()

		if err != nil {
			break
		}

		tags := customSplit(record[4])

		poem := db.Poem{
			Dataset:   "kaggle-poetry-foundations-poems",
			DatasetId: record[0],
			Title:     strings.TrimSpace(record[1]),
			Poem:      strings.TrimSpace(record[2]),
			Poet:      record[3],
			Tags:      tags,
			Language:  "english",
		}

		db.InsertOnePoemIntoDB(mongoDBConnection, poem)
	}
	file.Close()
	return true
}

func importChineseOneLine(mongoDBConnection db.MongoDBConnection) bool {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting current working directory:", err)
	}

	filePath := filepath.Join(currentDir, DATAPATH, "/chinese_poetry_dataset_one_line_per_poem/poems_with_tags.json")
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return false
	}

	var poems []db.ChineseOneLinePoem
	if jsonErr := json.Unmarshal(jsonData, &poems); jsonErr != nil {
		log.Fatal("Error unmarshalling JSON:", jsonErr)
	}

	var documents []interface{}

	// Iterate over the songs
	for _, poem := range poems {
		poem := db.Poem{
			Dataset:  "chinese-poetry-one-line-kaggle",
			Poem:     poem.Line,
			Tags:     poem.Tags,
			Language: "chinese",
		}
		documents = append(documents, poem)
	}

	collection := mongoDBConnection.Client.Database("poetry").Collection("poems")
	db.InsertManyIntoDB(*collection, documents)
	return true
}

func importEurovision(mongoDBConnection db.MongoDBConnection) bool {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting current working directory:", err)
	}

	folderPath := filepath.Join(currentDir, DATAPATH, "eurovision")
	pattern := filepath.Join(folderPath, "*.json")

	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Println("Error finding files:", err)
		return false
	}

	var documents []interface{}

	for _, path := range files {
		fmt.Println("Processing file:", path)

		jsonData, readErr := os.ReadFile(path)
		if readErr != nil {
			fmt.Println("Error reading JSON file:", readErr)
			continue
		}

		var songs map[string]db.Song
		if jsonErr := json.Unmarshal(jsonData, &songs); jsonErr != nil {
			log.Fatal("Error unmarshalling JSON:", jsonErr)
		}

		// Iterate over the songs
		for key, song := range songs {
			poem := db.Poem{
				Dataset:   "eurovision-kaggle",
				DatasetId: key,
				Title:     song.SongTitle,
				Poem:      song.Lyrics,
				Poet:      song.Artist,
				Tags:      []string{song.Year},
				Language:  strings.ToLower(song.Language),
			}
			documents = append(documents, poem)

			// Some songs don't have translation, like UK songs
			if song.Language != song.LyricsTranslation {
				poem := db.Poem{
					Dataset:   "eurovision-kaggle",
					DatasetId: key,
					Title:     song.SongTitle,
					Poem:      song.LyricsTranslation,
					Poet:      song.Artist,
					Tags:      []string{song.Year},
					Language:  "english",
				}
				documents = append(documents, poem)
			}
		}
		collection := mongoDBConnection.Client.Database("poetry").Collection("poems")
		db.InsertManyIntoDB(*collection, documents)
	}

	return true
}

func importCollectionOfPoetry(mongoDBConnection db.MongoDBConnection) bool {
	path := "collection_of_poetry/poems.csv"
	reader, file, err := createCSVReader(path)

	if err != nil {
		log.Fatal(err)
	}

	for {
		record, err := reader.Read()

		if err != nil {
			break
		}

		poem := db.Poem{
			Dataset:   "kaggle-collection-of-poems",
			DatasetId: record[0],
			Title:     record[1],
			Poem:      record[7],
			Poet:      record[2],
			Language:  "english",
		}

		db.InsertOnePoemIntoDB(mongoDBConnection, poem)
	}
	file.Close()
	return true
}

func importPoemsData(mongoDBConnection db.MongoDBConnection) bool {
	path := "poems_data/gutenberg-poetry-dataset.csv"
	reader, file, err := createCSVReader(path)

	if err != nil {
		log.Fatal(err)
	}

	for {
		record, err := reader.Read()

		if err != nil {
			break
		}

		poem := db.Poem{
			Dataset:   "kaggle-poems-data-gutenberg",
			DatasetId: record[1],
			Title:     record[4],
			Poem:      record[2],
			Poet:      record[3],
			Language:  "english",
		}

		db.InsertOnePoemIntoDB(mongoDBConnection, poem)
	}
	file.Close()
	return true
}

func importRussianPoetryCorpus(mongoDBConnection db.MongoDBConnection) bool {
	path := "russian_poetry_corpus/russianPoetryWithTheme.csv"
	reader, file, err := createCSVReader(path)

	if err != nil {
		log.Fatal(err)
	}

	for {
		record, err := reader.Read()

		if err != nil {
			break
		}

		poem := db.Poem{
			Dataset:  "kaggle-russian-poetry-corpus",
			Title:    record[3],
			Poem:     record[2],
			Poet:     record[0],
			Language: "russian",
		}

		db.InsertOnePoemIntoDB(mongoDBConnection, poem)
	}
	err = file.Close()

	if err != nil {
		log.Fatal(err)
	}

	return true
}

func importArabicPoetryDataset(mongoDBConnection db.MongoDBConnection) bool {
	path := "Arabic_Poetry_Dataset.csv"
	reader, file, err := createCSVReader(path)

	if err != nil {
		log.Fatal(err)
	}

	for {
		record, err := reader.Read()

		if err != nil {
			break
		}

		poem := db.Poem{
			Dataset:   "kaggle-arabic-dataset",
			DatasetId: record[1],
			Title:     record[3],
			Poem:      record[4],
			Poet:      record[0],
			Language:  "arabic",
		}

		db.InsertOnePoemIntoDB(mongoDBConnection, poem)
	}
	file.Close()
	return true
}

var datasets = map[string]func(connection db.MongoDBConnection) bool{
	"poetry_foundation":     importPoetryFoundation,
	"chinese_one_line":      importChineseOneLine,
	"eurovision":            importEurovision,
	"collection_of_poetry":  importCollectionOfPoetry,
	"poems_data":            importPoemsData,
	"russian_poetry_corpus": importRussianPoetryCorpus,
	"arabic_poetry_dataset": importArabicPoetryDataset,
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("You must pass dataset argument")
		fmt.Println("Pass one of the following keys to import dataset:")
		for key, _ := range datasets {
			fmt.Printf("%s\n", key)
		}
		return
	}

	fmt.Println("Importing collection:", os.Args[1])
	dataset := os.Args[1]

	mongoDBConnection, err := db.NewMongoDBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer mongoDBConnection.Disconnect()

	result := datasets[dataset](*mongoDBConnection)
	if result {
		fmt.Printf("Dataset %s imported\n", dataset)
	}
}
