package httpserver

import (
	"net/http"
	"strconv"
	"time"

	"aur-cache-service/internal/metrics"
)

// statusRecorder wraps http.ResponseWriter to capture response status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware collects Prometheus metrics for each HTTP request.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, req)
		duration := time.Since(start).Seconds()
		path := req.URL.Path
		metrics.HTTPRequestsTotal.WithLabelValues(path, req.Method, strconv.Itoa(rec.status)).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(path, req.Method).Observe(duration)
	})
}
