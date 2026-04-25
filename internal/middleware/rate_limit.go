package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

// RateLimit applies a Redis-based per-IP request ceiling over a fixed window.
func RateLimit(redisClient *redis.Client, maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	if maxRequests <= 0 {
		maxRequests = 60
	}
	if window <= 0 {
		window = time.Minute
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := resolveClientIP(r)
			if clientIP == "" {
				clientIP = "unknown"
			}

			ctx := context.Background()
			key := fmt.Sprintf("ratelimit:%s", clientIP)

			// Get current count
			count, err := redisClient.Get(ctx, key).Int()
			if err != nil && err != redis.Nil {
				// If Redis fails, allow request (fail open)
				next.ServeHTTP(w, r)
				return
			}

			// Check if limit exceeded
			if count >= maxRequests {
				writeRateLimitError(w)
				return
			}

			// Increment counter with expiration
			if count == 0 {
				// First request in window, set with expiration
				redisClient.Set(ctx, key, 1, window)
			} else {
				// Increment existing
				redisClient.Incr(ctx, key)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func resolveClientIP(r *http.Request) string {
	forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func writeRateLimitError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(models.APIError{Error: "Rate limit exceeded"})
}
