package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type rateEntry struct {
	windowStart time.Time
	count       int
}

type rateStore struct {
	mu      sync.Mutex
	entries map[string]*rateEntry
}

var globalRateStore = &rateStore{entries: make(map[string]*rateEntry)}

// RateLimit applies a simple per-IP request ceiling over a fixed window.
func RateLimit(next http.Handler, maxRequests int, window time.Duration) http.Handler {
	if maxRequests <= 0 {
		maxRequests = 60
	}
	if window <= 0 {
		window = time.Minute
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := resolveClientIP(r)
		if clientIP == "" {
			clientIP = "unknown"
		}

		if !allowRequest(clientIP, maxRequests, window) {
			writeRateLimitError(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func allowRequest(ip string, maxRequests int, window time.Duration) bool {
	now := time.Now().UTC()
	globalRateStore.mu.Lock()
	defer globalRateStore.mu.Unlock()

	entry, ok := globalRateStore.entries[ip]
	if !ok || now.Sub(entry.windowStart) >= window {
		globalRateStore.entries[ip] = &rateEntry{windowStart: now, count: 1}
		return true
	}

	if entry.count >= maxRequests {
		return false
	}

	entry.count++
	return true
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
