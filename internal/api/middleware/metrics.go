package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsCollector defines the interface for collecting metrics
type MetricsCollector interface {
	RecordRequest(path string)
	RecordResponse(path string, statusCode int, duration time.Duration)
}

// Metrics middleware records metrics for each request
func Metrics(metrics MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		start := time.Now()

		// Record request
		path := c.Request.URL.Path
		metrics.RecordRequest(path)

		// Process request
		c.Next()

		// Record response
		duration := time.Since(start)
		metrics.RecordResponse(path, c.Writer.Status(), duration)
	}
}
