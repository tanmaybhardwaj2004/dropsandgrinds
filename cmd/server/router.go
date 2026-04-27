package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/config"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/handlers"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/middleware"
)

func newHTTPHandler(logger *slog.Logger, cfg config.Config, redisClient *redis.Client) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/health/deps", handlers.HealthDepsHandler)
	mux.HandleFunc("/metrics", handlers.MetricsHandler)

	// Auth Routes
	mux.HandleFunc("/api/auth/register", handlers.RegisterHandler)
	mux.HandleFunc("/api/auth/login", handlers.LoginHandler)
	mux.HandleFunc("/api/auth/refresh", handlers.RefreshHandler)
	mux.HandleFunc("/api/auth/logout", handlers.LogoutHandler)

	// Catalog and profile routes
	mux.HandleFunc("/api/games", handlers.GamesListHandler)
	mux.HandleFunc("/api/games/", handlers.GameDetailHandler)
	mux.HandleFunc("/api/games/", handlers.GameRedirectHandler)
	mux.HandleFunc("/api/games/", handlers.ReviewHandler)
	mux.HandleFunc("/api/deals", handlers.DealsListHandler)
	mux.HandleFunc("/api/deals/for-you", handlers.DealsForYouHandler)
	mux.HandleFunc("/api/prices/", handlers.PriceHistoryHandler)
	mux.HandleFunc("/api/prices/", handlers.IndiaArbitrageHandler)
	mux.Handle("/api/wishlist", middleware.RequireAuth([]byte(cfg.JWTSecret), http.HandlerFunc(handlers.WishlistCollectionHandler)))
	mux.Handle("/api/wishlist/", middleware.RequireAuth([]byte(cfg.JWTSecret), http.HandlerFunc(handlers.WishlistItemHandler)))
	mux.Handle("/api/me", middleware.RequireAuth([]byte(cfg.JWTSecret), http.HandlerFunc(handlers.MeHandler)))
	mux.Handle("/api/library", middleware.RequireAuth([]byte(cfg.JWTSecret), http.HandlerFunc(handlers.LibraryListHandler)))
	mux.Handle("/api/library/import", middleware.RequireAuth([]byte(cfg.JWTSecret), http.HandlerFunc(handlers.LibraryImportHandler)))
	mux.Handle("/api/savings", middleware.RequireAuth([]byte(cfg.JWTSecret), http.HandlerFunc(handlers.SavingsGetHandler)))
	mux.Handle("/api/savings/purchase", middleware.RequireAuth([]byte(cfg.JWTSecret), http.HandlerFunc(handlers.SavingsLogHandler)))
	mux.Handle("/api/savings/history", middleware.RequireAuth([]byte(cfg.JWTSecret), http.HandlerFunc(handlers.SavingsHistoryHandler)))
	mux.HandleFunc("/api/bundles/analyze", handlers.BundleAnalyzeHandler)
	mux.HandleFunc("/api/games/", handlers.BuyTimingHandler)
	mux.HandleFunc("/api/sales/active", handlers.ActiveSalesHandler)
	mux.HandleFunc("/api/sales/calendar", handlers.SalesCalendarHandler)

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return middleware.RequestID(middleware.Logging(logger, middleware.RateLimit(redisClient, 60, time.Minute)(mux)))
}
