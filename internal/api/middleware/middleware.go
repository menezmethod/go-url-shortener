package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/menezmethod/ref_go/internal/logger"
)

type contextKey string

const (
	requestIDKey contextKey = "requestID"
	loggerKey    contextKey = "logger"
)

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.New().String()
		c.Set(string(requestIDKey), id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(string(requestIDKey)); exists {
		return id.(string)
	}
	return ""
}

// Logging logs requests with zap
func Logging(baseLogger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := GetRequestID(c)

		// Create a request-scoped logger
		requestLogger := logger.RequestLogger(baseLogger, requestID)

		// Add logger to context
		c.Set(string(loggerKey), requestLogger)

		// Log request details
		requestLogger.Info("Request started",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// Process request
		c.Next()

		// Log response time and status
		requestLogger.Info("Request completed",
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
		)
	}
}

// GetLogger retrieves the request-scoped logger from context
func GetLogger(c *gin.Context) *zap.Logger {
	if logger, exists := c.Get(string(loggerKey)); exists {
		return logger.(*zap.Logger)
	}
	// Fallback to global logger
	return zap.L()
}

// Recovery middleware handles panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger := GetLogger(c)
				logger.Error("Recovered from panic",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)

				c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
			}
		}()

		c.Next()
	}
}

// Timeout middleware adds request timeout
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		// Monitor context cancellation in a separate goroutine
		done := make(chan struct{})
		go func() {
			select {
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					c.AbortWithStatusJSON(408, gin.H{"error": "Request timeout"})
				}
			case <-done:
				// Request completed before timeout
			}
		}()

		c.Next()
		close(done)
	}
}
