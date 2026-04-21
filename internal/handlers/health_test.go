package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_ReturnsOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	HealthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response json: %v", err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("expected status field 'ok', got %q", payload["status"])
	}
}

func TestHealthDepsHandler_WithoutDBPool(t *testing.T) {
	SetDBPool(nil)

	req := httptest.NewRequest(http.MethodGet, "/health/deps", nil)
	rr := httptest.NewRecorder()

	HealthDepsHandler(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rr.Code)
	}
}
