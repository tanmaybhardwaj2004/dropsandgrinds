package main

import (
	"log/slog"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/config"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/handlers"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/middleware"
)

func newHTTPHandler(logger *slog.Logger, cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/health/deps", handlers.HealthDepsHandler)

	// Auth Routes
	mux.HandleFunc("/api/auth/register", handlers.RegisterHandler)
	mux.HandleFunc("/api/auth/login", handlers.LoginHandler)
	mux.HandleFunc("/api/auth/refresh", handlers.RefreshHandler)
	mux.HandleFunc("/api/auth/logout", handlers.LogoutHandler)

	// Catalog and profile routes
	mux.HandleFunc("/api/games", handlers.GamesListHandler)
	mux.HandleFunc("/api/games/", handlers.GameDetailHandler)
	mux.HandleFunc("/api/deals", handlers.DealsListHandler)
	mux.HandleFunc("/api/prices/", handlers.PriceHistoryHandler)
	mux.Handle("/api/wishlist", middleware.RequireAuth([]byte(cfg.JWTSecret), http.HandlerFunc(handlers.WishlistCollectionHandler)))
	mux.Handle("/api/wishlist/", middleware.RequireAuth([]byte(cfg.JWTSecret), http.HandlerFunc(handlers.WishlistItemHandler)))
	mux.Handle("/api/me", middleware.RequireAuth([]byte(cfg.JWTSecret), http.HandlerFunc(handlers.MeHandler)))

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return middleware.RequestID(middleware.Logging(logger, middleware.RateLimit(mux, 60, time.Minute)))
}
