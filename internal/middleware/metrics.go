package middleware

import (
	"net/http"
	"time"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/monitoring"
)

// Metrics records Prometheus request count and latency observations.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(recorder, r)

		monitoring.RecordHTTPRequest(r, recorder.status, time.Since(start))
	})
}
