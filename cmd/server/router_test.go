package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/config"
)

func TestNewHTTPHandler_HealthRoute(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil))
	h := newHTTPHandler(logger, config.Config{JWTSecret: "test-secret"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id header to be set")
	}
}

func TestNewHTTPHandler_ProtectedRouteRequiresAuth(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil))
	h := newHTTPHandler(logger, config.Config{JWTSecret: "test-secret"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
