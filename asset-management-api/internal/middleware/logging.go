package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		logger.Error("Failed to create logs directory:", err)
	}

	// Set up file output
	logFile, err := os.OpenFile(filepath.Join("logs", "api.log"), 
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Error("Failed to open log file:", err)
	} else {
		logger.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}

	// Set JSON formatter for structured logging
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	logger.SetLevel(logrus.InfoLevel)
}

func StructuredLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logData := map[string]interface{}{
			"timestamp":    param.TimeStamp.Format(time.RFC3339),
			"status":       param.StatusCode,
			"latency":      param.Latency.String(),
			"client_ip":    param.ClientIP,
			"method":       param.Method,
			"path":         param.Path,
			"user_agent":   param.Request.UserAgent(),
			"error":        param.ErrorMessage,
			"body_size":    param.BodySize,
		}

		// Add request ID if available
		if requestID := param.Request.Header.Get("X-Request-ID"); requestID != "" {
			logData["request_id"] = requestID
		}

		// Log level based on status code
		var level string
		switch {
		case param.StatusCode >= 500:
			level = "ERROR"
		case param.StatusCode >= 400:
			level = "WARN"
		case param.StatusCode >= 300:
			level = "INFO"
		default:
			level = "INFO"
		}
		logData["level"] = level

		logJSON, _ := json.Marshal(logData)
		return string(logJSON) + "\n"
	})
}

func RequestResponseLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create response writer wrapper
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		// Process request
		c.Next()

		latency := time.Since(start)

		// Prepare log data
		logData := logrus.Fields{
			"timestamp":     start.Format(time.RFC3339),
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"query":         c.Request.URL.RawQuery,
			"status":        c.Writer.Status(),
			"latency_ms":    latency.Milliseconds(),
			"client_ip":     c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
			"request_size":  c.Request.ContentLength,
			"response_size": w.body.Len(),
		}

		// Add user context if available
		if userID, exists := c.Get("user_id"); exists {
			logData["user_id"] = userID
		}
		if userRole, exists := c.Get("user_role"); exists {
			logData["user_role"] = userRole
		}

		// Add request body for non-GET requests (excluding sensitive data)
		if c.Request.Method != "GET" && len(requestBody) > 0 && len(requestBody) < 1024 {
			// Don't log passwords or other sensitive fields
			if !containsSensitiveData(string(requestBody)) {
				logData["request_body"] = string(requestBody)
			}
		}

		// Add response body for errors
		if c.Writer.Status() >= 400 && w.body.Len() > 0 && w.body.Len() < 1024 {
			logData["response_body"] = w.body.String()
		}

		// Add error if present
		if len(c.Errors) > 0 {
			logData["errors"] = c.Errors.String()
		}

		// Log with appropriate level
		switch {
		case c.Writer.Status() >= 500:
			logger.WithFields(logData).Error("Server Error")
		case c.Writer.Status() >= 400:
			logger.WithFields(logData).Warn("Client Error")
		case c.Writer.Status() >= 300:
			logger.WithFields(logData).Info("Redirect")
		case latency > 1*time.Second:
			logger.WithFields(logData).Warn("Slow Request")
		default:
			logger.WithFields(logData).Info("Request Completed")
		}
	}
}

func containsSensitiveData(body string) bool {
	sensitiveFields := []string{
		"password", "token", "secret", "key", "auth",
		"Password", "Token", "Secret", "Key", "Auth",
	}
	
	for _, field := range sensitiveFields {
		if bytes.Contains([]byte(body), []byte(field)) {
			return true
		}
	}
	return false
}

// Error logging function
func LogError(err error, context map[string]interface{}) {
	fields := logrus.Fields{"error": err.Error()}
	for k, v := range context {
		fields[k] = v
	}
	logger.WithFields(fields).Error("Application Error")
}

// Info logging function
func LogInfo(message string, context map[string]interface{}) {
	fields := logrus.Fields{}
	for k, v := range context {
		fields[k] = v
	}
	logger.WithFields(fields).Info(message)
}

// Debug logging function
func LogDebug(message string, context map[string]interface{}) {
	fields := logrus.Fields{}
	for k, v := range context {
		fields[k] = v
	}
	logger.WithFields(fields).Debug(message)
}

// Business event logging
func LogBusinessEvent(event string, context map[string]interface{}) {
	fields := logrus.Fields{
		"event_type": "business",
		"event":      event,
	}
	for k, v := range context {
		fields[k] = v
	}
	logger.WithFields(fields).Info("Business Event")
}

// Security event logging
func LogSecurityEvent(event string, context map[string]interface{}) {
	fields := logrus.Fields{
		"event_type": "security",
		"event":      event,
	}
	for k, v := range context {
		fields[k] = v
	}
	logger.WithFields(fields).Warn("Security Event")
}

// Performance logging
func LogPerformance(operation string, duration time.Duration, context map[string]interface{}) {
	fields := logrus.Fields{
		"event_type":  "performance",
		"operation":   operation,
		"duration_ms": duration.Milliseconds(),
	}
	for k, v := range context {
		fields[k] = v
	}
	
	if duration > 1*time.Second {
		logger.WithFields(fields).Warn("Slow Operation")
	} else {
		logger.WithFields(fields).Info("Performance Metric")
	}
}