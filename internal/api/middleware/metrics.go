package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/menezmethod/ref_go/internal/metrics"
)

// Metrics middleware records metrics for each request
func Metrics(metrics *metrics.Metrics) gin.HandlerFunc {
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
