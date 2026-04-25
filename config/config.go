package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL           string
	RedisURL              string
	Port                  string
	JWTSecret             string
	AccessTokenTTLMinutes int
	RefreshTokenTTLHours  int
}

func LoadConfig() Config {
	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")
	port := os.Getenv("PORT")
	jwtSecret := os.Getenv("JWT_SECRET")
	accessTokenTTLMinutes := parseEnvInt("ACCESS_TOKEN_TTL_MINUTES", 15)
	refreshTokenTTLHours := parseEnvInt("REFRESH_TOKEN_TTL_HOURS", 168)

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
		DatabaseURL:           dbURL,
		RedisURL:              redisURL,
		Port:                  port,
		JWTSecret:             jwtSecret,
		AccessTokenTTLMinutes: accessTokenTTLMinutes,
		RefreshTokenTTLHours:  refreshTokenTTLHours,
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
