package config

import (
	"log"
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func LoadConfig() Config {
	dbURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")

	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	if port == "" {
		log.Fatal("PORT is not set")
	}

	return Config{
		DatabaseURL: dbURL,
		Port:        port,
	}
}
