package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"strconv"
	"time"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	activeSessions = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "simulation_active_sessions",
			Help: "Number of active simulation sessions",
		},
	)

	totalPrograms = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "simulation_programs_uploaded_total",
			Help: "Total number of programs uploaded",
		},
	)

	jobsQueued = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "simulation_jobs_queued",
			Help: "Number of jobs in queue",
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(activeSessions)
	prometheus.MustRegister(totalPrograms)
	prometheus.MustRegister(jobsQueued)
}

// PrometheusMiddleware collects metrics for HTTP requests
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}

// PrometheusHandler returns the Prometheus metrics handler
func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// UpdateSessionMetrics updates session-related metrics
func UpdateSessionMetrics(count int) {
	activeSessions.Set(float64(count))
}

// IncrementProgramMetrics increments program upload counter
func IncrementProgramMetrics() {
	totalPrograms.Inc()
}

// UpdateJobMetrics updates job queue metrics
func UpdateJobMetrics(jobType string, count int) {
	jobsQueued.WithLabelValues(jobType).Set(float64(count))
}
