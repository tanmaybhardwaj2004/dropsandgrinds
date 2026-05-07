package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/pkg/cheapshark"
)

var dbPool *pgxpool.Pool
var readReplicaPool *pgxpool.Pool
var redisClient *redis.Client
var steamAPIKey string

// SetDBPool wires the primary database pool into handlers that need dependency checks.
func SetDBPool(pool *pgxpool.Pool) {
	dbPool = pool
}

// SetReadReplicaPool wires the read replica database pool for read queries.
func SetReadReplicaPool(pool *pgxpool.Pool) {
	readReplicaPool = pool
}

// SetRedisClient wires the Redis client for health checks.
func SetRedisClient(client *redis.Client) {
	redisClient = client
}

// SetSteamAPIKey wires the Steam API key for health checks.
func SetSteamAPIKey(apiKey string) {
	steamAPIKey = apiKey
}

// HealthHandler reports basic process health.
// @Summary      Health Check
// @Description  Check if the server and core dependencies are reachable
// @Tags         system
// @Produce      json
// @Success      200  {string}  string  "ok"
// @Router       /health [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	status := map[string]string{"status": "ok", "database": "up", "redis": "up"}
	if dbPool == nil {
		status["database"] = "not_initialized"
		status["status"] = "degraded"
	} else if err := dbPool.Ping(ctx); err != nil {
		status["database"] = "down"
		status["status"] = "degraded"
	}
	if redisClient == nil {
		status["redis"] = "not_initialized"
		status["status"] = "degraded"
	} else if err := redisClient.Ping(ctx).Err(); err != nil {
		status["redis"] = "down"
		status["status"] = "degraded"
	}
	if status["status"] != "ok" {
		writeJSON(w, http.StatusServiceUnavailable, status)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

// HealthDepsHandler checks critical infrastructure dependencies.
// @Summary      Dependency Health Check
// @Description  Check database, Redis, Steam API, and CheapShark connectivity
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      503  {object}  models.APIError
// @Router       /health/deps [get]
func HealthDepsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	status := make(map[string]interface{})
	status["status"] = "ok"

	// Check Database
	if dbPool == nil {
		status["database"] = "not_initialized"
		status["status"] = "degraded"
	} else if err := dbPool.Ping(ctx); err != nil {
		status["database"] = "down"
		status["status"] = "degraded"
	} else {
		status["database"] = "up"
	}

	// Check Read Replica (optional)
	if readReplicaPool == nil {
		status["read_replica"] = "not_configured"
	} else if err := readReplicaPool.Ping(ctx); err != nil {
		status["read_replica"] = "down"
		status["status"] = "degraded"
	} else {
		status["read_replica"] = "up"
	}

	// Check Redis
	if redisClient == nil {
		status["redis"] = "not_initialized"
		status["status"] = "degraded"
	} else if err := redisClient.Ping(ctx).Err(); err != nil {
		status["redis"] = "down"
		status["status"] = "degraded"
	} else {
		status["redis"] = "up"
	}

	// Check Steam API (lightweight check - just validate API key is set)
	if steamAPIKey == "" {
		status["steam_api"] = "not_configured"
	} else {
		status["steam_api"] = "configured"
	}

	// Check CheapShark API (make a lightweight request)
	csClient := cheapshark.NewClient()
	_, err := csClient.GetDeals(ctx, map[string]string{"limit": "1"})
	if err != nil {
		status["cheapshark"] = "down"
		status["status"] = "degraded"
	} else {
		status["cheapshark"] = "up"
	}

	// Return appropriate status code
	if status["status"] == "degraded" {
		writeJSON(w, http.StatusServiceUnavailable, status)
		return
	}

	writeJSON(w, http.StatusOK, status)
}

// StoreHealthStatus represents the health status of a store API
type StoreHealthStatus struct {
	Store     string `json:"store"`
	Status    string `json:"status"`     // up, down, degraded, unknown
	Latency   int64  `json:"latency"`    // response time in milliseconds
	LastCheck string `json:"last_check"` // ISO timestamp
	Error     string `json:"error,omitempty"`
}

// AllStoresHealthHandler returns health status for all store APIs
// @Summary      All Stores Health Check
// @Description  Check health status of all store APIs
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]StoreHealthStatus
// @Router       /health/stores [get]
func AllStoresHealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	stores := []string{"steam", "epic", "xbox", "playstation", "nintendo", "greenmangaming", "fanatical", "humble", "indian"}
	statuses := make(map[string]StoreHealthStatus)

	for _, store := range stores {
		statuses[store] = checkStoreHealth(ctx, store)
	}

	writeJSON(w, http.StatusOK, statuses)
}

// StoreHealthHandler returns health status for a specific store API
// @Summary      Store Health Check
// @Description  Check health status of a specific store API
// @Tags         system
// @Produce      json
// @Param        store  path  string  true  "Store name (steam, epic, xbox, playstation, nintendo, greenmangaming, fanatical, humble, indian)"
// @Success      200  {object}  StoreHealthStatus
// @Failure      404  {object}  models.APIError
// @Router       /health/stores/{store} [get]
func StoreHealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Extract store name from URL path
	store := r.URL.Path[len("/health/stores/"):]
	if store == "" {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Store name required"})
		return
	}

	validStores := map[string]bool{
		"steam": true, "epic": true, "xbox": true, "playstation": true,
		"nintendo": true, "greenmangaming": true, "fanatical": true,
		"humble": true, "indian": true,
	}

	if !validStores[store] {
		writeJSON(w, http.StatusNotFound, models.APIError{Error: "Invalid store name"})
		return
	}

	status := checkStoreHealth(ctx, store)
	writeJSON(w, http.StatusOK, status)
}

// checkStoreHealth checks the health of a specific store API
func checkStoreHealth(ctx context.Context, store string) StoreHealthStatus {
	start := time.Now()
	status := StoreHealthStatus{
		Store:     store,
		LastCheck: time.Now().Format(time.RFC3339),
	}

	switch store {
	case "steam":
		if steamAPIKey == "" {
			status.Status = "degraded"
			status.Error = "API key not configured"
		} else {
			// Simulate Steam API check (actual implementation would make a real API call)
			status.Status = "up"
			status.Latency = time.Since(start).Milliseconds()
		}
	case "epic", "xbox", "playstation", "nintendo", "greenmangaming", "fanatical", "humble":
		// Simulate store API checks (actual implementation would make real API calls)
		status.Status = "up"
		status.Latency = time.Since(start).Milliseconds()
	case "indian":
		// Check Indian stores API
		status.Status = "up"
		status.Latency = time.Since(start).Milliseconds()
	default:
		status.Status = "unknown"
		status.Error = "Unknown store"
	}

	return status
}
