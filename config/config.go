package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL            string
	DatabaseReadReplicaURL string
	RedisURL               string
	Port                   string
	JWTSecret              string
	AccessTokenTTLMinutes  int
	RefreshTokenTTLHours   int
	SentryDSN              string
	SteamAPIKey            string
	MeilisearchURL         string
	MeilisearchMasterKey   string
	GoogleClientID         string
	GoogleClientSecret     string
	GoogleRedirectURL      string
}

func LoadConfig() Config {
	dbURL := os.Getenv("DATABASE_URL")
	dbReadReplicaURL := os.Getenv("DATABASE_READ_REPLICA_URL")
	redisURL := os.Getenv("REDIS_URL")
	port := os.Getenv("PORT")
	jwtSecret := os.Getenv("JWT_SECRET")
	accessTokenTTLMinutes := parseEnvInt("ACCESS_TOKEN_TTL_MINUTES", 15)
	refreshTokenTTLHours := parseEnvInt("REFRESH_TOKEN_TTL_HOURS", 168)
	sentryDSN := os.Getenv("SENTRY_DSN")
	steamAPIKey := os.Getenv("STEAM_API_KEY")
	meilisearchURL := os.Getenv("MEILISEARCH_URL")
	meilisearchMasterKey := os.Getenv("MEILISEARCH_MASTER_KEY")
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	if port == "" {
		log.Fatal("PORT is not set")
	}
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}
	if redisURL == "" {
		redisURL = "localhost:6379"
		log.Printf("REDIS_URL not set, using default: %s", redisURL)
	}

	return Config{
		DatabaseURL:            dbURL,
		DatabaseReadReplicaURL: dbReadReplicaURL,
		RedisURL:               redisURL,
		Port:                   port,
		JWTSecret:              jwtSecret,
		AccessTokenTTLMinutes:  accessTokenTTLMinutes,
		RefreshTokenTTLHours:   refreshTokenTTLHours,
		SentryDSN:              sentryDSN,
		SteamAPIKey:            steamAPIKey,
		MeilisearchURL:         meilisearchURL,
		MeilisearchMasterKey:   meilisearchMasterKey,
		GoogleClientID:         googleClientID,
		GoogleClientSecret:     googleClientSecret,
		GoogleRedirectURL:      googleRedirectURL,
	}
}

func parseEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("invalid %s value %q, using fallback %d", key, value, fallback)
		return fallback
	}
	return parsed
}
