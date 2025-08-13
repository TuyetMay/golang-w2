package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "Size of HTTP requests in bytes",
			Buckets: prometheus.ExponentialBuckets(1, 2, 20),
		},
		[]string{"method", "endpoint"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Size of HTTP responses in bytes",
			Buckets: prometheus.ExponentialBuckets(1, 2, 20),
		},
		[]string{"method", "endpoint", "status"},
	)

	// Business metrics
	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Total number of active users",
		},
	)

	foldersCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "folders_created_total",
			Help: "Total number of folders created",
		},
		[]string{"user_role"},
	)

	notesCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notes_created_total",
			Help: "Total number of notes created",
		},
		[]string{"user_role"},
	)

	sharesCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shares_created_total",
			Help: "Total number of shares created",
		},
		[]string{"resource_type", "access_level"},
	)

	// Database metrics
	dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Error metrics
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "endpoint"},
	)

	// JWT metrics
	jwtTokensGenerated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "jwt_tokens_generated_total",
			Help: "Total number of JWT tokens generated",
		},
	)

	jwtTokensValidated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jwt_tokens_validated_total",
			Help: "Total number of JWT tokens validated",
		},
		[]string{"status"},
	)
)

// PrometheusMiddleware collects HTTP metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Get request size
		requestSize := float64(c.Request.ContentLength)
		if requestSize <= 0 && c.Request.Body != nil {
			// Estimate size if not provided
			requestSize = 0
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		
		// Get labels
		method := c.Request.Method
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}
		status := strconv.Itoa(c.Writer.Status())

		// Record metrics
		httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
		
		if requestSize > 0 {
			httpRequestSize.WithLabelValues(method, endpoint).Observe(requestSize)
		}
		
		responseSize := float64(c.Writer.Size())
		if responseSize > 0 {
			httpResponseSize.WithLabelValues(method, endpoint, status).Observe(responseSize)
		}

		// Record errors
		if c.Writer.Status() >= 400 {
			errorType := "client_error"
			if c.Writer.Status() >= 500 {
				errorType = "server_error"
			}
			errorsTotal.WithLabelValues(errorType, endpoint).Inc()
		}

		// Update active users (simple estimation based on unique requests)
		if userID, exists := c.Get("user_id"); exists && userID != nil {
			activeUsers.Set(1) // This is a simple implementation
		}
	}
}

// Business metrics functions
func RecordFolderCreated(userRole string) {
	foldersCreatedTotal.WithLabelValues(userRole).Inc()
}

func RecordNoteCreated(userRole string) {
	notesCreatedTotal.WithLabelValues(userRole).Inc()
}

func RecordShareCreated(resourceType, accessLevel string) {
	sharesCreatedTotal.WithLabelValues(resourceType, accessLevel).Inc()
}

// Database metrics functions
func RecordDBQuery(operation, table string, duration time.Duration) {
	dbQueriesTotal.WithLabelValues(operation, table).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

func SetActiveDBConnections(count int) {
	dbConnectionsActive.Set(float64(count))
}

// JWT metrics functions
func RecordJWTGenerated() {
	jwtTokensGenerated.Inc()
}

func RecordJWTValidated(success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	jwtTokensValidated.WithLabelValues(status).Inc()
}