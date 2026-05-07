package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_WithoutDependenciesReturnsUnavailable(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	HealthHandler(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rr.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response json: %v", err)
	}
	if payload["status"] != "degraded" {
		t.Fatalf("expected status field 'degraded', got %q", payload["status"])
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
