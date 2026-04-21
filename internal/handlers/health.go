package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

var dbPool *pgxpool.Pool

// SetDBPool wires the primary database pool into handlers that need dependency checks.
func SetDBPool(pool *pgxpool.Pool) {
	dbPool = pool
}

// HealthHandler reports basic process health.
// @Summary      Health Check
// @Description  Check if the server is running
// @Tags         system
// @Produce      json
// @Success      200  {string}  string  "ok"
// @Router       /health [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HealthDepsHandler checks critical infrastructure dependencies.
// @Summary      Dependency Health Check
// @Description  Check database connectivity and other core dependencies
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      503  {object}  models.APIError
// @Router       /health/deps [get]
func HealthDepsHandler(w http.ResponseWriter, r *http.Request) {
	if dbPool == nil {
		writeJSON(w, http.StatusServiceUnavailable, models.APIError{Error: "Database pool not initialized"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := dbPool.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status":   "degraded",
			"database": "down",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"database": "up",
	})
}
