package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HTTPRequestsTotal counts all HTTP requests processed by the service.
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests handled by the service.",
		},
		[]string{"path", "method", "status"},
	)

	// HTTPRequestDuration measures how long HTTP handlers take to respond.
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of latencies for HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	// ProviderOperations tracks operations performed by cache providers.
	ProviderOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_provider_operations_total",
			Help: "Count of cache provider operations.",
		},
		[]string{"provider", "operation", "status"},
	)

	// ProviderOperationDuration measures how long cache provider
	// operations take to complete.
	ProviderOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_provider_operation_duration_seconds",
			Help:    "Histogram of latencies for cache provider operations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider", "operation"},
	)

	// ExternalRequests counts calls to external HTTP APIs.
	ExternalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "external_requests_total",
			Help: "Count of requests to external API service.",
		},
		[]string{"cache", "status"},
	)

	// ExternalRequestDuration measures duration of calls to external HTTP APIs.
	ExternalRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "external_request_duration_seconds",
			Help:    "Histogram of external API request durations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"cache"},
	)

	// CacheLayerHits counts how many values were found on each cache layer.
	CacheLayerHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_layer_hits_total",
			Help: "Number of cache hits on each layer.",
		},
		[]string{"level"},
	)

	// CacheLayerMisses counts misses per cache layer.
	CacheLayerMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_layer_misses_total",
			Help: "Number of cache misses on each layer.",
		},
		[]string{"level"},
	)
)

// Register registers all metrics in the default registry.
func Register() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		ProviderOperations,
		ProviderOperationDuration,
		ExternalRequests,
		ExternalRequestDuration,
		CacheLayerHits,
		CacheLayerMisses,
	)
}

// RecordProviderOp increments ProviderOperations with result status.
func RecordProviderOp(provider, operation string, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	ProviderOperations.WithLabelValues(provider, operation, status).Inc()
}

// RecordProviderLatency records the duration of a provider operation.
func RecordProviderLatency(provider, operation string, durationSeconds float64) {
	ProviderOperationDuration.WithLabelValues(provider, operation).Observe(durationSeconds)
}

// RecordExternalRequest records metrics for an external API call.
func RecordExternalRequest(cacheName string, err error, durationSeconds float64) {
	status := "success"
	if err != nil {
		status = "error"
	}
	ExternalRequests.WithLabelValues(cacheName, status).Inc()
	ExternalRequestDuration.WithLabelValues(cacheName).Observe(durationSeconds)
}

// RecordCacheLayer records hits/misses for a cache layer.
func RecordCacheLayer(level int, hits, misses int) {
	CacheLayerHits.WithLabelValues(fmt.Sprintf("%d", level)).Add(float64(hits))
	CacheLayerMisses.WithLabelValues(fmt.Sprintf("%d", level)).Add(float64(misses))
}
