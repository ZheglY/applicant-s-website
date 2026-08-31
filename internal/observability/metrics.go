package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry        *prometheus.Registry
	HTTPRequests    *prometheus.CounterVec
	HTTPDuration    *prometheus.HistogramVec
	HTTPInFlight    prometheus.Gauge
	CacheOperations *prometheus.CounterVec
	RateLimits      *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	metrics := &Metrics{
		registry: prometheus.NewRegistry(),
		HTTPRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "unik", Subsystem: "http", Name: "requests_total",
			Help: "Total number of HTTP requests.",
		}, []string{"method", "route", "status"}),
		HTTPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "unik", Subsystem: "http", Name: "request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),
		HTTPInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "unik", Subsystem: "http", Name: "in_flight_requests",
			Help: "Current number of HTTP requests being processed.",
		}),
		CacheOperations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "unik", Subsystem: "cache", Name: "operations_total",
			Help: "Redis cache operations by key and result.",
		}, []string{"key", "result"}),
		RateLimits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "unik", Subsystem: "security", Name: "rate_limit_decisions_total",
			Help: "Rate limiting decisions.",
		}, []string{"scope", "decision"}),
	}
	metrics.registry.MustRegister(
		metrics.HTTPRequests,
		metrics.HTTPDuration,
		metrics.HTTPInFlight,
		metrics.CacheOperations,
		metrics.RateLimits,
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)
	return metrics
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{EnableOpenMetrics: true})
}
