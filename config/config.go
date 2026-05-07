package config

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
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
	GameSpotAPIKey         string
	MeilisearchURL         string
	MeilisearchMasterKey   string
	GoogleClientID         string
	GoogleClientSecret     string
	GoogleRedirectURL      string
}

func LoadConfig() Config {
	dbURL := readSecretEnv("DATABASE_URL")
	if dbURL == "" {
		dbURL = buildDatabaseURLFromEnv()
	}
	dbReadReplicaURL := os.Getenv("DATABASE_READ_REPLICA_URL")
	redisURL := os.Getenv("REDIS_URL")
	port := os.Getenv("PORT")
	jwtSecret := readSecretEnv("JWT_SECRET")
	accessTokenTTLMinutes := parseEnvInt("ACCESS_TOKEN_TTL_MINUTES", 15)
	refreshTokenTTLHours := parseEnvInt("REFRESH_TOKEN_TTL_HOURS", 168)
	sentryDSN := readSecretEnv("SENTRY_DSN")
	steamAPIKey := readSecretEnv("STEAM_API_KEY")
	gameSpotAPIKey := readSecretEnv("GAMESPOT_API_KEY")
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
		GameSpotAPIKey:         gameSpotAPIKey,
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

func readSecretEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	filePath := os.Getenv(key + "_FILE")
	if filePath == "" {
		return ""
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("failed to read %s from %s: %v", key, filePath, err)
		return ""
	}
	return strings.TrimSpace(string(data))
}

func buildDatabaseURLFromEnv() string {
	password := readSecretEnv("POSTGRES_PASSWORD")
	if password == "" {
		return ""
	}

	host := envOrDefault("POSTGRES_HOST", "postgres")
	port := envOrDefault("POSTGRES_PORT", "5432")
	user := envOrDefault("POSTGRES_USER", "postgres")
	dbName := envOrDefault("POSTGRES_DB", "dropsandgrinds")
	sslMode := envOrDefault("POSTGRES_SSLMODE", "disable")

	databaseURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     host + ":" + port,
		Path:     dbName,
		RawQuery: "sslmode=" + url.QueryEscape(sslMode),
	}
	return databaseURL.String()
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
