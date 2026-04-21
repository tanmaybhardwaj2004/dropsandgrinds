package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/config"
	_ "github.com/tanmaybhardwaj2004/dropsandgrinds/docs"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/handlers"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

// @title           DropsAndGrinds API
// @version         1.0
// @description     Smart Cross-Platform Game Deal Tracker (India-Focused)
// @host            localhost:8080
// @BasePath        /
func main() {
	_ = godotenv.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := config.LoadConfig()

	conn, err := config.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer conn.Close()
	handlers.SetDBPool(conn)
	catalogRepo := repositories.NewCatalogRepository(conn)
	gamesService := services.NewGamesService(catalogRepo)
	handlers.SetGamesService(gamesService)
	wishlistRepo := repositories.NewWishlistRepository(conn)
	wishlistService := services.NewWishlistService(wishlistRepo)
	handlers.SetWishlistService(wishlistService)

	authService, err := services.NewAuthService(conn, services.AuthServiceConfig{
		JWTSecret:             cfg.JWTSecret,
		AccessTokenTTLMinutes: cfg.AccessTokenTTLMinutes,
		RefreshTokenTTLHours:  cfg.RefreshTokenTTLHours,
	})
	if err != nil {
		log.Fatal("Failed to initialize auth service:", err)
	}
	handlers.SetAuthService(authService)

	log.Println("Database connected successfully")

	wrappedHandler := newHTTPHandler(logger, cfg)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           wrappedHandler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			logger.Error("graceful shutdown failed", "error", shutdownErr)
		}
	}()

	logger.Info("server listening", "port", cfg.Port)
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
