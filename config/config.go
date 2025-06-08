package config

import (
	"os"
	"sync"
)

type Config struct {
	DbHost     string
	DbPort     string
	DbName     string
	DbUser     string
	DbPass     string
	ElasticUrl string
}

func NewConfig() *Config {
	//if err := godotenv.Load(); err != nil {
	//	log.Fatalf("error loading .env file: %v", err)
	//}

	return &Config{
		DbHost:     os.Getenv("DB_HOST"),
		DbPort:     os.Getenv("DB_PORT"),
		DbName:     os.Getenv("DB_NAME"),
		DbUser:     os.Getenv("DB_USER"),
		DbPass:     os.Getenv("DB_PASS"),
		ElasticUrl: os.Getenv("ELASTIC_URL"),
	}
}

func GetConfig() *Config {
	var (
		once     sync.Once
		instance *Config
	)
	once.Do(func() {
		instance = NewConfig()
	})
	return instance
}
