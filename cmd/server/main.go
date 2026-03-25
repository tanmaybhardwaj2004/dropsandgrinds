package main

import (
	"context"
	"log"
	"net/http"
	"os"

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

	dbURL := os.Getenv("DATABASE_URL")

	conn, err := config.ConnectDB(dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")

	defer conn.Close(context.Background())

	log.Println("Server listening on :8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
