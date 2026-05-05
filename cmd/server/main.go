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

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/config"
	_ "github.com/tanmaybhardwaj2004/dropsandgrinds/docs"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/handlers"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/middleware"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/scheduler"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/pkg/logger"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/pkg/steam"
)

// @title           DropsAndGrinds API
// @version         1.0
// @description     Smart Cross-Platform Game Deal Tracker (India-Focused)
// @host            localhost:8080
// @BasePath        /
func main() {
	_ = godotenv.Load()
	cfg := config.LoadConfig()

	// Initialize Sentry
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://ca3b71cc206fb5a094dca3953d3052bf@o4511301731811328.ingest.de.sentry.io/4511301742821456",
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)

	// Initialize file-based logger
	logFormat := "text"
	if os.Getenv("LOG_FORMAT") == "json" {
		logFormat = "json"
	}
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	if err := logger.Init(logger.Config{
		LogDir:      os.Getenv("LOG_DIR"),
		Level:       logLevel,
		Format:      logFormat,
		ServiceName: "backend",
	}); err != nil {
		log.Printf("Failed to initialize logger: %v", err)
		log.Println("Falling back to stdout logging")
		logger.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	}

	// Log application startup
	logger.LogStartup("1.0", cfg.Port)

	// Set auth logger
	middleware.SetAuthLogger(logger.Logger)

	// Initialize Sentry
	config.InitSentry(cfg.SentryDSN)
	defer config.FlushSentry()

	logger.LogComponentStartup("database", map[string]string{"url": cfg.DatabaseURL})
	conn, err := config.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		logger.LogError("database connection", err, map[string]string{"url": cfg.DatabaseURL})
		log.Fatal("Failed to connect to database:", err)
	}
	defer func() {
		conn.Close()
		logger.LogComponentShutdown("database", "normal shutdown")
	}()
	handlers.SetDBPool(conn)
	logger.LogInfo("database connected successfully", nil)

	// Connect to read replica if configured
	if cfg.DatabaseReadReplicaURL != "" {
		logger.LogComponentStartup("read_replica", map[string]string{"url": cfg.DatabaseReadReplicaURL})
		readReplicaConn, err := config.ConnectReadReplica(cfg.DatabaseReadReplicaURL)
		if err != nil {
			logger.LogError("read replica connection", err, map[string]string{"url": cfg.DatabaseReadReplicaURL})
			log.Printf("Failed to connect to read replica (continuing without): %v", err)
			readReplicaConn = nil
		}
		if readReplicaConn != nil {
			defer func() {
				readReplicaConn.Close()
				logger.LogComponentShutdown("read_replica", "normal shutdown")
			}()
			handlers.SetReadReplicaPool(readReplicaConn)
			logger.LogInfo("read replica connected successfully", nil)
		}
	}

	logger.LogComponentStartup("redis", map[string]string{"url": cfg.RedisURL})
	redisClient, err := config.NewRedisClient(cfg.RedisURL)
	if err != nil {
		logger.LogError("redis connection", err, map[string]string{"url": cfg.RedisURL})
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer func() {
		redisClient.Close()
		logger.LogComponentShutdown("redis", "normal shutdown")
	}()
	handlers.SetRedisClient(redisClient)
	handlers.SetSteamAPIKey(cfg.SteamAPIKey)
	logger.LogInfo("redis connected successfully", nil)

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
	libraryService := services.NewLibraryService(libraryRepo, catalogRepo, steamClient, logger.Logger)
	handlers.SetLibraryService(libraryService)

	// Initialize savings service
	savingsRepo := repositories.NewSavingsRepository(conn)
	savingsService := services.NewSavingsService(savingsRepo)
	handlers.SetSavingsService(savingsService)

	// Initialize clicks repository for analytics
	clicksRepo := repositories.NewClicksRepository(conn)
	handlers.SetClicksRepository(clicksRepo)

	// Initialize analytics repository
	analyticsRepo := repositories.NewAnalyticsRepository(conn)
	handlers.SetAnalyticsRepository(analyticsRepo)

	// Initialize sales calendar repository
	salesCalendarRepo := repositories.NewSalesCalendarRepository(conn)

	// Initialize bundle service
	bundleService := services.NewBundleService(catalogRepo, logger.Logger)
	handlers.SetBundleService(bundleService)

	// Initialize buy timing service
	buyTimingService := services.NewBuyTimingService(salesCalendarRepo, logger.Logger)
	handlers.SetBuyTimingService(buyTimingService)

	// Initialize arbitrage service
	arbitrageService := services.NewArbitrageService(catalogRepo, logger.Logger, 83.0, 0.18) // USD to INR exchange rate, 18% GST
	handlers.SetArbitrageService(arbitrageService)

	// Initialize OAuth service (optional)
	var oauthService *services.OAuthService
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		logger.LogComponentStartup("oauth", map[string]string{"provider": "google"})
		oauthService = services.NewOAuthService(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL, authService)
		handlers.SetOAuthService(oauthService)
		logger.LogInfo("OAuth service initialized", nil)
	}

	// Initialize Meilisearch service (optional)
	var meilisearchService *services.MeilisearchService
	if cfg.MeilisearchURL != "" && cfg.MeilisearchMasterKey != "" {
		logger.LogComponentStartup("meilisearch", map[string]string{"url": cfg.MeilisearchURL})
		meilisearchService = services.NewMeilisearchService(cfg.MeilisearchURL, cfg.MeilisearchMasterKey)
		if err := meilisearchService.ConfigureIndex(); err != nil {
			logger.LogError("meilisearch configuration", err, nil)
			logger.Logger.Warn("failed to configure Meilisearch index", "error", err)
		} else {
			handlers.SetMeilisearchService(meilisearchService)
			logger.LogInfo("Meilisearch service initialized", nil)
		}
	}

	// Initialize and start scheduler
	logger.LogComponentStartup("scheduler", nil)
	sched := scheduler.New(logger.Logger)
	sched.AddJob(scheduler.Job{
		Name:     "price-refresh",
		Interval: 15 * time.Minute,
		Run:      scheduler.PriceRefreshJob(catalogRepo, logger.Logger),
	})
	sched.AddJob(scheduler.Job{
		Name:     "review-refresh",
		Interval: 24 * time.Hour,
		Run:      scheduler.ReviewRefreshJob(reviewService, logger.Logger),
	})

	// Start Meilisearch sync if configured
	scheduler.StartMeilisearchSync(sched, catalogRepo, meilisearchService, logger.Logger)

	sched.Start(context.Background())
	defer func() {
		sched.Stop()
		logger.LogComponentShutdown("scheduler", "normal shutdown")
	}()

	wrappedHandler := newHTTPHandler(logger.Logger, cfg, redisClient)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           wrappedHandler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		logger.LogInfo("shutdown signal received", map[string]string{"signal": "interrupt"})
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			logger.LogError("graceful shutdown", shutdownErr, nil)
		} else {
			logger.LogInfo("graceful shutdown completed", nil)
		}
	}()

	logger.LogInfo("server listening", map[string]string{"port": cfg.Port, "address": ":" + cfg.Port})
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.LogError("server startup", err, nil)
		log.Fatal(err)
	}
	logger.LogShutdown("server closed")
}
