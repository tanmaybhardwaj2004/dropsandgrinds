package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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
