package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests handled by the API.",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "http_request_latency_seconds",
			Help:       "HTTP request latency distribution with p95 exposed as a quantile.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.005, 0.99: 0.001},
		},
		[]string{"method", "path", "status"},
	)

	CacheRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_requests_total",
			Help: "Total cache lookups by cache name and result.",
		},
		[]string{"cache", "result"},
	)
)

func init() {
	prometheus.MustRegister(HTTPRequestTotal, HTTPRequestLatency, CacheRequestsTotal)
}

// RecordHTTPRequest records the count and latency for one completed request.
func RecordHTTPRequest(r *http.Request, status int, duration time.Duration) {
	statusCode := strconv.Itoa(status)
	route := normalizePath(r.URL.Path)
	HTTPRequestTotal.WithLabelValues(r.Method, route, statusCode).Inc()
	HTTPRequestLatency.WithLabelValues(r.Method, route, statusCode).Observe(duration.Seconds())
}

// RecordCacheHit records a successful cache lookup.
func RecordCacheHit(cacheName string) {
	CacheRequestsTotal.WithLabelValues(cacheName, "hit").Inc()
}

// RecordCacheMiss records a cache miss or unreadable cached payload.
func RecordCacheMiss(cacheName string) {
	CacheRequestsTotal.WithLabelValues(cacheName, "miss").Inc()
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}
