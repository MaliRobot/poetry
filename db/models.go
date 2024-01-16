package db

type Poem struct {
	ID        string   `bson:"_id,omitempty"`
	Dataset   string   `bson:"dataset"`
	DatasetId string   `bson:"dataset_id"`
	Title     string   `bson:"title"`
	Poem      string   `bson:"poem"`
	Poet      string   `bson:"poet"`
	Tags      []string `bson:"tags"`
	Language  string   `bson:"language"`
}

type Song struct {
	Number            string `json:"#"`
	Country           string `json:"Country"`
	Artist            string `json:"Artist"`
	SongTitle         string `json:"Song"`
	Language          string `json:"Language"`
	EurovisionNumber  int    `json:"Eurovision_Number"`
	Year              string `json:"Year"`
	HostCountry       string `json:"Host_Country"`
	HostCity          string `json:"Host_City"`
	Lyrics            string `json:"Lyrics"`
	LyricsTranslation string `json:"Lyrics translation"`
}

type ChineseOneLinePoem struct {
	Line string   `json:"line"`
	Tags []string `json:"tags"`
}
