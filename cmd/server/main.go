package main

import (
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/config"
	_ "github.com/tanmaybhardwaj2004/dropsandgrinds/docs"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/handlers"
)

// @Summary      Health Check
// @Description  Check if the server is running
// @Tags         system
// @Produce      json
// @Success      200  {string}  string  "ok"
// @Router       /health [get]
func healthHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok")) //cnvert string to byte then write response
}

// @title           DropsAndGrinds API
// @version         1.0
// @description     Smart Cross-Platform Game Deal Tracker (India-Focused)
// @host            localhost:8080
// @BasePath        /
func main() {
	godotenv.Load()

	http.HandleFunc("/health", healthHandler) //connects /health to healthHandler function
	
	// Auth Routes
	http.HandleFunc("/api/auth/register", handlers.RegisterHandler)
	http.HandleFunc("/api/auth/login", handlers.LoginHandler)
	http.HandleFunc("/api/auth/refresh", handlers.RefreshHandler)
	http.HandleFunc("/api/auth/logout", handlers.LogoutHandler)

	// Set up Swagger UI route
	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

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
