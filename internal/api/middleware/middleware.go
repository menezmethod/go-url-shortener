package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"runtime"
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
		// Check if request ID is provided in header
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}

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

		// Get request body for POST/PUT/PATCH requests
		var body []byte
		if c.Request.Method != "GET" && c.Request.Body != nil {
			body, _ = c.GetRawData()
			// Restore the request body for later use
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		// Log request details
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}
		if len(body) > 0 {
			// Parse body as JSON to redact sensitive fields
			var jsonBody map[string]interface{}
			if err := json.Unmarshal(body, &jsonBody); err == nil {
				// Redact sensitive fields
				sensitiveFields := []string{"password", "token", "secret", "key", "auth"}
				for _, field := range sensitiveFields {
					if _, exists := jsonBody[field]; exists {
						jsonBody[field] = "[REDACTED]"
					}
				}
				// Convert back to JSON
				if redactedBody, err := json.Marshal(jsonBody); err == nil {
					fields = append(fields, zap.String("body", string(redactedBody)))
				}
			}
		}
		requestLogger.Info("Request started", fields...)

		// Process request
		c.Next()

		// Log any errors that occurred during request processing
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				fields := []zap.Field{
					zap.Error(err.Err),
					zap.Int("error_type", int(err.Type)),
				}
				if err.Meta != nil {
					fields = append(fields, zap.Any("meta", err.Meta))
				}
				requestLogger.Error("Error occurred during request", fields...)
			}
		}

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
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, false)]

				logger.Error("Recovered from panic",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("stack", string(stack)),
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
		// Create a done channel to signal completion
		done := make(chan bool, 1)
		panicChan := make(chan interface{}, 1)

		// Create timeout context
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()

			c.Next()
			done <- true
		}()

		select {
		case p := <-panicChan:
			panic(p) // Re-panic to let the Recovery middleware handle it
		case <-done:
			return
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				logger := GetLogger(c)
				logger.Warn("Request timed out",
					zap.String("path", c.Request.URL.Path),
					zap.Duration("timeout", timeout),
				)
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
					"error": "Request timeout",
				})
			}
		}
	}
}
