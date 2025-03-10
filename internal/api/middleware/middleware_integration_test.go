package middleware_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/menezmethod/ref_go/internal/api/middleware"
	"github.com/menezmethod/ref_go/internal/auth"
	"github.com/menezmethod/ref_go/internal/config"
)

// TokenValidator interface for the authentication middleware
type TokenValidator interface {
	ValidateToken(token string) (*auth.TokenClaims, error)
}

// MockTokenValidator for integration testing
type MockTokenValidator struct {
	validateFunc func(token string) (*auth.TokenClaims, error)
}

func (m *MockTokenValidator) ValidateToken(token string) (*auth.TokenClaims, error) {
	return m.validateFunc(token)
}

var _ = Describe("Middleware Integration", func() {
	var (
		router        *gin.Engine
		recorder      *httptest.ResponseRecorder
		observedLogs  *observer.ObservedLogs
		mockValidator *MockTokenValidator
		mockMetrics   *MockMetrics
		rateLimiter   *middleware.RateLimiter
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		router = gin.New()
		recorder = httptest.NewRecorder()

		// Set up logger with observer for verification
		core, obs := observer.New(zapcore.InfoLevel)
		observedLogs = obs
		logger := zap.New(core)

		// Create mock services
		mockValidator = &MockTokenValidator{
			validateFunc: func(token string) (*auth.TokenClaims, error) {
				if token == "valid-token" {
					return &auth.TokenClaims{}, nil
				}
				return nil, auth.ErrInvalidToken
			},
		}

		// Set up metrics collector
		mockMetrics = NewMockMetrics()

		// Set up rate limiter with shorter window for testing
		cfg := &config.Config{
			RateLimit: config.RateLimitConfig{
				Requests: 5,
				Window:   100 * time.Millisecond, // Shorter window for faster testing
			},
		}
		rateLimiter = middleware.NewRateLimiterWithCleanup(cfg, logger, 50*time.Millisecond)

		// Apply all middleware in the correct order
		router.Use(middleware.RequestID())                     // Add request ID first
		router.Use(middleware.Logging(logger))                 // Then logging to capture everything
		router.Use(middleware.Recovery())                      // Recovery should be early
		router.Use(middleware.SecurityHeaders())               // Security headers
		router.Use(middleware.Metrics(mockMetrics))            // Metrics collection
		router.Use(middleware.RateLimit(rateLimiter))          // Rate limiting
		router.Use(middleware.Timeout(200 * time.Millisecond)) // Short timeout for testing
		router.Use(middleware.CORS([]string{"*"}))             // CORS headers with all origins allowed

		// Define test endpoints
		router.GET("/public", func(c *gin.Context) {
			c.String(http.StatusOK, "Public content")
		})

		router.GET("/secure", middleware.Authentication(mockValidator), func(c *gin.Context) {
			claims := middleware.GetTokenClaims(c)
			c.JSON(http.StatusOK, gin.H{
				"authenticated": claims != nil,
				"request_id":    middleware.GetRequestID(c),
			})
		})

		router.GET("/slow", func(c *gin.Context) {
			time.Sleep(300 * time.Millisecond)
			c.String(http.StatusOK, "Slow response")
		})

		router.GET("/panic", func(c *gin.Context) {
			panic("Test panic")
		})

		router.POST("/data", func(c *gin.Context) {
			var data map[string]interface{}
			if err := c.BindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, data)
		})

		router.GET("/chain-error", func(c *gin.Context) {
			c.Error(fmt.Errorf("first error"))
			c.Error(fmt.Errorf("second error"))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Multiple errors occurred"})
		})
	})

	Context("with complete middleware stack", func() {
		It("successfully processes normal requests", func() {
			req := httptest.NewRequest(http.MethodGet, "/public", nil)
			router.ServeHTTP(recorder, req)

			// Check response
			Expect(recorder.Code).To(Equal(http.StatusOK))
			Expect(recorder.Body.String()).To(Equal("Public content"))

			// Verify request ID header was set
			Expect(recorder.Header().Get("X-Request-ID")).NotTo(BeEmpty())

			// Verify security headers were set
			Expect(recorder.Header().Get("X-Frame-Options")).To(Equal("DENY"))
			Expect(recorder.Header().Get("X-Content-Type-Options")).To(Equal("nosniff"))
			Expect(recorder.Header().Get("X-XSS-Protection")).To(Equal("1; mode=block"))
			Expect(recorder.Header().Get("Content-Security-Policy")).NotTo(BeEmpty())

			// Verify metrics were recorded
			Expect(mockMetrics.GetRequestCount()).To(Equal(int64(1)))
			Expect(mockMetrics.GetRequestCountByPath()["/public"]).To(Equal(int64(1)))

			// Verify logging happened
			logEntries := observedLogs.All()
			Expect(len(logEntries)).To(BeNumerically(">=", 2))

			// Find request start and end logs
			var foundStart, foundEnd bool
			for _, entry := range logEntries {
				if entry.Message == "Request started" {
					foundStart = true
				}
				if entry.Message == "Request completed" {
					foundEnd = true
				}
			}
			Expect(foundStart).To(BeTrue(), "Should have logged request start")
			Expect(foundEnd).To(BeTrue(), "Should have logged request end")
		})

		It("handles authentication correctly", func() {
			// Scenario 1: Valid token
			req := httptest.NewRequest(http.MethodGet, "/secure", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusOK))
			var response map[string]interface{}
			Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
			Expect(response["authenticated"]).To(BeTrue())
			Expect(response["request_id"]).NotTo(BeEmpty())

			// Reset recorder
			recorder = httptest.NewRecorder()

			// Scenario 2: Invalid token
			req = httptest.NewRequest(http.MethodGet, "/secure", nil)
			req.Header.Set("Authorization", "Bearer invalid-token")
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

			// Reset recorder
			recorder = httptest.NewRecorder()

			// Scenario 3: Malformed token
			req = httptest.NewRequest(http.MethodGet, "/secure", nil)
			req.Header.Set("Authorization", "InvalidFormat token")
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})

		It("handles rate limiting with bursts and recovery", func() {
			clientIP := "192.168.1.1"

			// Make initial burst of requests
			for i := 0; i < 5; i++ {
				req := httptest.NewRequest(http.MethodGet, "/public", nil)
				req.Header.Set("X-Forwarded-For", clientIP)
				recorder = httptest.NewRecorder()
				router.ServeHTTP(recorder, req)
				Expect(recorder.Code).To(Equal(http.StatusOK))
			}

			// Next request should be rate limited
			req := httptest.NewRequest(http.MethodGet, "/public", nil)
			req.Header.Set("X-Forwarded-For", clientIP)
			recorder = httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusTooManyRequests))

			// Wait for rate limit window to expire
			time.Sleep(150 * time.Millisecond)

			// Should be able to make requests again
			req = httptest.NewRequest(http.MethodGet, "/public", nil)
			req.Header.Set("X-Forwarded-For", clientIP)
			recorder = httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("recovers from panics with proper logging", func() {
			req := httptest.NewRequest(http.MethodGet, "/panic", nil)
			router.ServeHTTP(recorder, req)

			// Should recover and return 500
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

			// Verify error response format
			var response map[string]string
			Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
			Expect(response["error"]).To(Equal("Internal server error"))

			// Verify panic was logged with stack trace
			var foundPanicLog bool
			var hasStackTrace bool
			for _, entry := range observedLogs.All() {
				if entry.Message == "Recovered from panic" {
					foundPanicLog = true
					for _, field := range entry.Context {
						if field.Key == "stack" && field.String != "" {
							hasStackTrace = true
							break
						}
					}
					break
				}
			}
			Expect(foundPanicLog).To(BeTrue(), "Should have logged the panic recovery")
			Expect(hasStackTrace).To(BeTrue(), "Should have included stack trace")
		})

		It("handles timeouts appropriately", func() {
			req := httptest.NewRequest(http.MethodGet, "/slow", nil)
			router.ServeHTTP(recorder, req)

			// Should timeout and return 504
			Expect(recorder.Code).To(Equal(http.StatusGatewayTimeout))

			// Verify response format
			var response map[string]string
			Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
			Expect(response["error"]).To(Equal("Request timeout"))

			// Verify timeout was logged
			var foundTimeoutLog bool
			for _, entry := range observedLogs.All() {
				if entry.Message == "Request timed out" {
					foundTimeoutLog = true
					break
				}
			}
			Expect(foundTimeoutLog).To(BeTrue(), "Should have logged the timeout")
		})

		It("handles CORS requests correctly", func() {
			// Preflight request
			req := httptest.NewRequest(http.MethodOptions, "/public", nil)
			req.Header.Set("Origin", "http://example.com")
			req.Header.Set("Access-Control-Request-Method", "GET")
			req.Header.Set("Access-Control-Request-Headers", "Content-Type")

			router.ServeHTTP(recorder, req)

			// Check CORS headers
			Expect(recorder.Header().Get("Access-Control-Allow-Origin")).To(Equal("http://example.com"))
			Expect(recorder.Header().Get("Access-Control-Allow-Methods")).To(ContainSubstring("GET"))
			Expect(recorder.Header().Get("Access-Control-Allow-Headers")).To(ContainSubstring("Content-Type"))
			Expect(recorder.Code).To(Equal(http.StatusNoContent))

			// Actual request
			recorder = httptest.NewRecorder()
			req = httptest.NewRequest(http.MethodGet, "/public", nil)
			req.Header.Set("Origin", "http://example.com")

			router.ServeHTTP(recorder, req)

			Expect(recorder.Header().Get("Access-Control-Allow-Origin")).To(Equal("http://example.com"))
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("logs request body for POST requests with proper sanitization", func() {
			// Test with sensitive data
			body := map[string]interface{}{
				"username": "testuser",
				"password": "sensitive123",
				"email":    "test@example.com",
				"token":    "secret-token",
			}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/data", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusOK))

			// Verify body was logged but sensitive data was redacted
			var foundSensitiveLog bool
			for _, entry := range observedLogs.All() {
				for _, field := range entry.Context {
					if field.Key == "body" {
						foundSensitiveLog = true
						// Sensitive fields should be redacted
						Expect(field.String).NotTo(ContainSubstring("sensitive123"))
						Expect(field.String).NotTo(ContainSubstring("secret-token"))
						// Non-sensitive fields should be present
						Expect(field.String).To(ContainSubstring("testuser"))
						Expect(field.String).To(ContainSubstring("test@example.com"))
						break
					}
				}
			}
			Expect(foundSensitiveLog).To(BeTrue(), "Should have logged the sanitized request body")
		})

		It("handles multiple errors in the error chain", func() {
			req := httptest.NewRequest(http.MethodGet, "/chain-error", nil)
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

			// Verify all errors were logged
			var errorCount int
			for _, entry := range observedLogs.All() {
				if entry.Message == "Error occurred during request" {
					errorCount++
				}
			}
			Expect(errorCount).To(Equal(2), "Should have logged both errors")
		})

		It("handles concurrent requests with proper isolation", func() {
			var wg sync.WaitGroup
			concurrentRequests := 10
			results := make([]struct {
				statusCode int
				requestID  string
			}, concurrentRequests)

			// Make concurrent requests
			for i := 0; i < concurrentRequests; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()

					localRecorder := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodGet, "/public", nil)
					req.Header.Set("X-Forwarded-For", fmt.Sprintf("192.168.1.%d", idx))

					router.ServeHTTP(localRecorder, req)

					results[idx] = struct {
						statusCode int
						requestID  string
					}{
						statusCode: localRecorder.Code,
						requestID:  localRecorder.Header().Get("X-Request-ID"),
					}
				}(i)
			}

			wg.Wait()

			// Verify each request got a unique request ID and succeeded
			requestIDs := make(map[string]bool)
			for _, result := range results {
				Expect(result.statusCode).To(Equal(http.StatusOK))
				Expect(result.requestID).NotTo(BeEmpty())
				Expect(requestIDs[result.requestID]).To(BeFalse(), "Request IDs should be unique")
				requestIDs[result.requestID] = true
			}

			// Verify metrics
			Expect(mockMetrics.GetRequestCount()).To(BeNumerically(">=", int64(concurrentRequests)))
			Expect(mockMetrics.GetActiveRequests()).To(Equal(int64(0)))
		})
	})
})
