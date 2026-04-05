package main

import (
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/config"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok")) //cnvert string to byte then write response
}

func main() {
	godotenv.Load()

	http.HandleFunc("/health", healthHandler) //connects /health to healthHandler function

	cfg := config.LoadConfig()

	conn, err := config.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")

	defer conn.Close(context.Background())

	log.Println("Server listening on :" + cfg.Port)

	err = http.ListenAndServe(":"+cfg.Port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
