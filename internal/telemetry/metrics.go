package telemetry

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metrics of the service.
var (
	RateRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_requests_total",
			Help: "Total number of GetRates requests processed",
		},
		[]string{"method", "status"},
	)

	RateRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rate_request_duration_seconds",
			Help:    "Duration of GetRates requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	ExternalAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "external_api_duration_seconds",
			Help:    "Duration of external exchange API calls in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"exchange", "status"},
	)

	DBSaveDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_save_duration_seconds",
			Help:    "Duration of database write operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)
)

// MetricsHandler returns the HTTP handler for the /metrics endpoint.
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
