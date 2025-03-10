package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/menezmethod/ref_go/internal/api/middleware"
)

var _ = Describe("Middleware", func() {
	var (
		router   *gin.Engine
		recorder *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		router = gin.New()
		recorder = httptest.NewRecorder()
	})

	Describe("RequestID", func() {
		BeforeEach(func() {
			router.Use(middleware.RequestID())
			router.GET("/test", func(c *gin.Context) {
				requestID := middleware.GetRequestID(c)
				c.String(http.StatusOK, requestID)
			})
		})

		Context("when handling a request", func() {
			It("should set a request ID header", func() {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				// Check X-Request-ID header
				requestID := recorder.Header().Get("X-Request-ID")
				Expect(requestID).NotTo(BeEmpty())

				// Check that the response body matches the request ID
				Expect(recorder.Body.String()).To(Equal(requestID))
			})

			It("should use provided request ID from header", func() {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				providedID := "test-request-id"
				req.Header.Set("X-Request-ID", providedID)

				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Header().Get("X-Request-ID")).To(Equal(providedID))
				Expect(recorder.Body.String()).To(Equal(providedID))
			})

			It("should generate different request IDs for different requests", func() {
				req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
				req2 := httptest.NewRequest(http.MethodGet, "/test", nil)

				recorder1 := httptest.NewRecorder()
				recorder2 := httptest.NewRecorder()

				router.ServeHTTP(recorder1, req1)
				router.ServeHTTP(recorder2, req2)

				id1 := recorder1.Header().Get("X-Request-ID")
				id2 := recorder2.Header().Get("X-Request-ID")

				Expect(id1).NotTo(BeEmpty())
				Expect(id2).NotTo(BeEmpty())
				Expect(id1).NotTo(Equal(id2))
			})
		})

		Context("when using GetRequestID", func() {
			It("should return empty string for context without request ID", func() {
				router := gin.New() // Router without RequestID middleware
				router.GET("/test", func(c *gin.Context) {
					requestID := middleware.GetRequestID(c)
					c.String(http.StatusOK, requestID)
				})

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, req)

				Expect(recorder.Body.String()).To(BeEmpty())
			})
		})
	})

	Describe("Logging", func() {
		var (
			observedLogs *observer.ObservedLogs
			core         zapcore.Core
			logger       *zap.Logger
		)

		BeforeEach(func() {
			core, observedLogs = observer.New(zapcore.InfoLevel)
			logger = zap.New(core)

			router.Use(middleware.Logging(logger))
			router.GET("/test", func(c *gin.Context) {
				time.Sleep(10 * time.Millisecond) // Add small delay to test latency logging
				c.String(http.StatusOK, "success")
			})

			router.POST("/test", func(c *gin.Context) {
				c.String(http.StatusCreated, "created")
			})

			router.GET("/error", func(c *gin.Context) {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "test error"})
			})
		})

		Context("when handling successful requests", func() {
			It("should log request and response details", func() {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("User-Agent", "test-agent")
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				// Wait for logging to complete
				Eventually(observedLogs.All).Should(HaveLen(2)) // Request start and end logs

				logs := observedLogs.All()
				startLog := logs[0]
				endLog := logs[1]

				// Check request start log
				Expect(startLog.Message).To(Equal("Request started"))
				Expect(startLog.Context).To(ContainElement(zap.String("method", "GET")))
				Expect(startLog.Context).To(ContainElement(zap.String("path", "/test")))
				Expect(startLog.Context).To(ContainElement(zap.String("user_agent", "test-agent")))

				// Check request end log
				Expect(endLog.Message).To(Equal("Request completed"))
				Expect(endLog.Context).To(ContainElement(zap.Int("status", http.StatusOK)))
				Expect(endLog.Context).To(ContainElement(BeAssignableToTypeOf(zap.Duration("latency", 0))))
			})

			It("should log request body for POST requests", func() {
				body := gin.H{"test_key": "test_value"}
				jsonBody, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusCreated))

				Eventually(observedLogs.All).Should(HaveLen(2))

				startLog := observedLogs.All()[0]
				Expect(startLog.Context).To(ContainElement(zap.String("body", string(jsonBody))))
			})
		})

		Context("when handling error responses", func() {
			It("should log error details", func() {
				req := httptest.NewRequest(http.MethodGet, "/error", nil)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				Eventually(observedLogs.All).Should(HaveLen(2))

				endLog := observedLogs.All()[1]
				Expect(endLog.Context).To(ContainElement(zap.Int("status", http.StatusInternalServerError)))
			})
		})

		Context("when request has correlation ID", func() {
			It("should include request ID in logs", func() {
				router.Use(middleware.RequestID()) // Add RequestID middleware

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				router.ServeHTTP(recorder, req)

				Eventually(observedLogs.All).Should(HaveLen(2))

				logs := observedLogs.All()
				for _, log := range logs {
					Expect(log.Context).To(ContainElement(WithTransform(func(f zapcore.Field) string {
						return f.Key
					}, Equal("request_id"))))
				}
			})
		})
	})

	Describe("Recovery", func() {
		var observedLogs *observer.ObservedLogs

		BeforeEach(func() {
			// Set up the logger with observer
			core, obs := observer.New(zapcore.ErrorLevel)
			observedLogs = obs
			logger := zap.New(core)

			// Add logging middleware first to set up the logger in context
			router.Use(func(c *gin.Context) {
				c.Set("logger", logger)
				c.Next()
			})

			// Add recovery middleware
			router.Use(middleware.Recovery())

			router.GET("/panic", func(c *gin.Context) {
				panic("test panic")
			})
			router.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "success")
			})
		})

		Context("when a panic occurs", func() {
			It("should recover and return 500 error", func() {
				req := httptest.NewRequest(http.MethodGet, "/panic", nil)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["error"]).To(Equal("Internal server error"))

				// Check that the panic was logged
				Eventually(observedLogs.All).Should(HaveLen(1))
				log := observedLogs.All()[0]
				Expect(log.Message).To(Equal("Recovered from panic"))
				Expect(log.Context).To(ContainElement(zap.Any("error", "test panic")))
				Expect(log.Context).To(ContainElement(zap.String("path", "/panic")))
			})
		})

		Context("when no panic occurs", func() {
			It("should process the request normally", func() {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal("success"))

				// No logs should be generated when there's no panic
				Expect(observedLogs.All()).To(BeEmpty())
			})
		})
	})

	Describe("Timeout", func() {
		BeforeEach(func() {
			router.Use(middleware.Timeout(100 * time.Millisecond))
		})

		Context("when request completes within timeout", func() {
			It("should process the request normally", func() {
				router.GET("/quick", func(c *gin.Context) {
					time.Sleep(50 * time.Millisecond)
					c.String(http.StatusOK, "success")
				})

				req := httptest.NewRequest(http.MethodGet, "/quick", nil)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal("success"))
			})
		})

		Context("when request exceeds timeout", func() {
			It("should return 504 Gateway Timeout", func() {
				router.GET("/slow", func(c *gin.Context) {
					time.Sleep(200 * time.Millisecond)
					c.String(http.StatusOK, "success")
				})

				req := httptest.NewRequest(http.MethodGet, "/slow", nil)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusGatewayTimeout))

				// Debug the response body
				body := recorder.Body.String()
				Expect(body).To(ContainSubstring("Request timeout"))
			})
		})

		Context("when request is canceled by client", func() {
			It("should handle cancellation gracefully", func() {
				router.GET("/cancel", func(c *gin.Context) {
					// Create a canceled context
					ctx, cancel := context.WithCancel(c.Request.Context())
					cancel()
					c.Request = c.Request.WithContext(ctx)

					// Wait for the cancellation to be processed
					time.Sleep(150 * time.Millisecond)
					c.String(http.StatusOK, "success")
				})

				req := httptest.NewRequest(http.MethodGet, "/cancel", nil)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusGatewayTimeout))

				// Debug the response body
				body := recorder.Body.String()
				Expect(body).To(ContainSubstring("Request timeout"))
			})
		})
	})
})
