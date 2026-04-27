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
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/middleware"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/scheduler"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/pkg/steam"
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

	// Set auth logger
	middleware.SetAuthLogger(logger)

	// Initialize Sentry
	config.InitSentry(cfg.SentryDSN)
	defer config.FlushSentry()

	conn, err := config.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer conn.Close()
	handlers.SetDBPool(conn)

	redisClient, err := config.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()
	handlers.SetRedisClient(redisClient)
	handlers.SetSteamAPIKey(cfg.SteamAPIKey)

	catalogRepo := repositories.NewCatalogRepository(conn, redisClient)
	gamesService := services.NewGamesService(catalogRepo)
	handlers.SetGamesService(gamesService)

	dealsService := services.NewDealEvaluationService(catalogRepo)
	handlers.SetDealsService(dealsService)

	reviewRepo := repositories.NewReviewRepository(conn)
	reviewService := services.NewReviewService(reviewRepo, "", "") // API keys from env in production
	handlers.SetReviewService(reviewService)
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

	// Initialize library service
	libraryRepo := repositories.NewLibraryRepository(conn)
	steamClient := steam.NewClient(cfg.SteamAPIKey)
	libraryService := services.NewLibraryService(libraryRepo, catalogRepo, steamClient, logger)
	handlers.SetLibraryService(libraryService)

	// Initialize savings service
	savingsRepo := repositories.NewSavingsRepository(conn)
	savingsService := services.NewSavingsService(savingsRepo)
	handlers.SetSavingsService(savingsService)

	// Initialize clicks repository for analytics
	clicksRepo := repositories.NewClicksRepository(conn)
	handlers.SetClicksRepository(clicksRepo)

	// Initialize sales calendar repository
	salesCalendarRepo := repositories.NewSalesCalendarRepository(conn)

	// Initialize bundle service
	bundleService := services.NewBundleService(catalogRepo, logger)
	handlers.SetBundleService(bundleService)

	// Initialize buy timing service
	buyTimingService := services.NewBuyTimingService(salesCalendarRepo, logger)
	handlers.SetBuyTimingService(buyTimingService)

	log.Println("Database connected successfully")

	// Initialize and start scheduler
	sched := scheduler.New(logger)
	sched.AddJob(scheduler.Job{
		Name:     "price-refresh",
		Interval: 15 * time.Minute,
		Run:      scheduler.PriceRefreshJob(catalogRepo, logger),
	})
	sched.AddJob(scheduler.Job{
		Name:     "review-refresh",
		Interval: 24 * time.Hour,
		Run:      scheduler.ReviewRefreshJob(reviewService, logger),
	})

	sched.Start(context.Background())
	defer sched.Stop()

	wrappedHandler := newHTTPHandler(logger, cfg, redisClient)
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
