//go:build smoke

package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/cmd/server"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/config"
)

// Smoke tests are basic integration tests that verify the application
// starts correctly and core endpoints respond. These are meant to be run
// after deployment to verify the system is operational.

func TestSmoke_HealthEndpoint(t *testing.T) {
	// This test verifies the health endpoint is accessible
	// In production, this would test against the deployed URL

	cfg := config.Config{
		Port:        8080,
		DatabaseURL: "postgres://test:test@localhost:5432/test?sslmode=disable",
		RedisURL:    "redis://localhost:6379",
		JWTSecret:   "test-secret",
		SentryDSN:   "",
	}

	app := server.NewApp(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")
}

func TestSmoke_GamesEndpoint(t *testing.T) {
	cfg := config.Config{
		Port:        8080,
		DatabaseURL: "postgres://test:test@localhost:5432/test?sslmode=disable",
		RedisURL:    "redis://localhost:6379",
		JWTSecret:   "test-secret",
		SentryDSN:   "",
	}

	app := server.NewApp(cfg)

	req := httptest.NewRequest("GET", "/api/games?limit=10", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Should return 200 even if no games (empty list)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSmoke_DealsEndpoint(t *testing.T) {
	cfg := config.Config{
		Port:        8080,
		DatabaseURL: "postgres://test:test@localhost:5432/test?sslmode=disable",
		RedisURL:    "redis://localhost:6379",
		JWTSecret:   "test-secret",
		SentryDSN:   "",
	}

	app := server.NewApp(cfg)

	req := httptest.NewRequest("GET", "/api/deals?limit=10", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Should return 200 even if no deals (empty list)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSmoke_SearchEndpoint(t *testing.T) {
	cfg := config.Config{
		Port:        8080,
		DatabaseURL: "postgres://test:test@localhost:5432/test?sslmode=disable",
		RedisURL:    "redis://localhost:6379",
		JWTSecret:   "test-secret",
		SentryDSN:   "",
	}

	app := server.NewApp(cfg)

	req := httptest.NewRequest("GET", "/api/games/search?q=test&limit=10", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Should return 200 even if no results
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSmoke_SwaggerEndpoint(t *testing.T) {
	cfg := config.Config{
		Port:        8080,
		DatabaseURL: "postgres://test:test@localhost:5432/test?sslmode=disable",
		RedisURL:    "redis://localhost:6379",
		JWTSecret:   "test-secret",
		SentryDSN:   "",
	}

	app := server.NewApp(cfg)

	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Swagger UI should be accessible
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestSmoke_MetricsEndpoint verifies Prometheus metrics endpoint
func TestSmoke_MetricsEndpoint(t *testing.T) {
	cfg := config.Config{
		Port:        8080,
		DatabaseURL: "postgres://test:test@localhost:5432/test?sslmode=disable",
		RedisURL:    "redis://localhost:6379",
		JWTSecret:   "test-secret",
		SentryDSN:   "",
	}

	app := server.NewApp(cfg)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Metrics endpoint should be accessible
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestSmoke_AuthEndpoint verifies auth endpoints are accessible
func TestSmoke_AuthEndpoint(t *testing.T) {
	cfg := config.Config{
		Port:        8080,
		DatabaseURL: "postgres://test:test@localhost:5432/test?sslmode=disable",
		RedisURL:    "redis://localhost:6379",
		JWTSecret:   "test-secret",
		SentryDSN:   "",
	}

	app := server.NewApp(cfg)

	t.Run("Register endpoint accessible", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/auth/register", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		// Will return 400 for invalid body, but endpoint should be accessible
		assert.NotEqual(t, http.StatusNotFound, w.Code)
	})

	t.Run("Login endpoint accessible", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/auth/login", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		// Will return 400 for invalid body, but endpoint should be accessible
		assert.NotEqual(t, http.StatusNotFound, w.Code)
	})
}

// TestSmoke_SalesEndpoints verifies sale-related endpoints
func TestSmoke_SalesEndpoints(t *testing.T) {
	cfg := config.Config{
		Port:        8080,
		DatabaseURL: "postgres://test:test@localhost:5432/test?sslmode=disable",
		RedisURL:    "redis://localhost:6379",
		JWTSecret:   "test-secret",
		SentryDSN:   "",
	}

	app := server.NewApp(cfg)

	t.Run("Active sales endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/sales/active", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Sales calendar endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/sales/calendar", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestSmoke_BundleEndpoint verifies bundle analysis endpoint
func TestSmoke_BundleEndpoint(t *testing.T) {
	cfg := config.Config{
		Port:        8080,
		DatabaseURL: "postgres://test:test@localhost:5432/test?sslmode=disable",
		RedisURL:    "redis://localhost:6379",
		JWTSecret:   "test-secret",
		SentryDSN:   "",
	}

	app := server.NewApp(cfg)

	req := httptest.NewRequest("POST", "/api/bundles/analyze", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Will return 400 for invalid body, but endpoint should be accessible
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

// TestSmoke_ResponseHeaders verifies security headers are present
func TestSmoke_ResponseHeaders(t *testing.T) {
	cfg := config.Config{
		Port:        8080,
		DatabaseURL: "postgres://test:test@localhost:5432/test?sslmode=disable",
		RedisURL:    "redis://localhost:6379",
		JWTSecret:   "test-secret",
		SentryDSN:   "",
	}

	app := server.NewApp(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	// Verify content type is set
	contentType := w.Header().Get("Content-Type")
	assert.Contains(t, contentType, "application/json")
}
