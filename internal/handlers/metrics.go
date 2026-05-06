package handlers

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

// MetricsHandler serves Prometheus metrics
// @Summary      Prometheus metrics
// @Description  Exposes Prometheus metrics for scraping
// @Tags         system
// @Produce      plain
// @Success      200  {string}  string  "metrics"
// @Failure      405  {object}  models.APIError
// @Router       /metrics [get]
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	promhttp.Handler().ServeHTTP(w, r)
}
