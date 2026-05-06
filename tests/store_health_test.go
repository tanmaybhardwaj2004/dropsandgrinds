package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/handlers"
)

func TestStoreHealthEndpoints(t *testing.T) {
	// Test all stores health check endpoint
	t.Run("AllStoresHealthHandler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health/stores", nil)
		rr := httptest.NewRecorder()

		handlers.AllStoresHealthHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		var response map[string]handlers.StoreHealthStatus
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Check that all stores are present
		expectedStores := []string{"steam", "epic", "xbox", "playstation", "nintendo", "greenmangaming", "fanatical", "humble", "indian"}
		for _, store := range expectedStores {
			if _, exists := response[store]; !exists {
				t.Errorf("expected store %s in response", store)
			}
		}
	})

	// Test individual store health check endpoints
	testStores := []string{"steam", "epic", "xbox", "playstation", "nintendo", "greenmangaming", "fanatical", "humble", "indian"}

	for _, store := range testStores {
		t.Run("StoreHealthHandler_"+store, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health/stores/"+store, nil)
			rr := httptest.NewRecorder()

			handlers.StoreHealthHandler(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", rr.Code)
			}

			var response handlers.StoreHealthStatus
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response.Store != store {
				t.Errorf("expected store name %s, got %s", store, response.Store)
			}

			if response.Status != "up" && response.Status != "degraded" && response.Status != "unknown" {
				t.Errorf("unexpected status: %s", response.Status)
			}
		})
	}
}

func TestStoreHealthHandlerInvalidStore(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/stores/invalid_store", nil)
	rr := httptest.NewRecorder()

	handlers.StoreHealthHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

func TestHealthEndpointsIntegration(t *testing.T) {
	// Skip: This test requires the router to be initialised via cmd/server (a main package),
	// which cannot be imported. Move router construction into an internal package to enable this.
	t.Skip("Integration test requires router refactoring to be importable from tests")
}

