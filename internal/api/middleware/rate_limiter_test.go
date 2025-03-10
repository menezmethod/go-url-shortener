package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/menezmethod/ref_go/internal/api/middleware"
	"github.com/menezmethod/ref_go/internal/config"
)

var _ = Describe("RateLimiter", func() {
	var (
		router   *gin.Engine
		recorder *httptest.ResponseRecorder
		limiter  *middleware.RateLimiter
		cfg      *config.Config
		logger   *zap.Logger
	)

	BeforeEach(func() {
		// Set up gin in test mode
		gin.SetMode(gin.TestMode)
		router = gin.New()

		// Create test logger
		logger, _ = zap.NewDevelopment()

		// Create test config
		cfg = &config.Config{
			RateLimit: config.RateLimitConfig{
				Requests: 3,               // 3 requests
				Window:   2 * time.Second, // per 2 seconds
			},
		}

		// Create rate limiter with shorter cleanup period for testing
		limiter = middleware.NewRateLimiterWithCleanup(cfg, logger, 100*time.Millisecond)

		// Set up test recorder
		recorder = httptest.NewRecorder()

		// Set up test endpoint with rate limiting
		router.GET("/test", middleware.RateLimit(limiter), func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})
	})

	Describe("Rate Limiting", func() {
		Context("when requests are within the limit", func() {
			It("allows the requests and sets appropriate headers", func() {
				for i := 0; i < cfg.RateLimit.Requests; i++ {
					req, _ := http.NewRequest(http.MethodGet, "/test", nil)
					req.RemoteAddr = "192.168.1.1:12345" // Set client IP

					router.ServeHTTP(recorder, req)

					Expect(recorder.Code).To(Equal(http.StatusOK))
					Expect(recorder.Body.String()).To(Equal("success"))

					// Check rate limit headers
					limit := recorder.Header().Get("X-RateLimit-Limit")
					remaining := recorder.Header().Get("X-RateLimit-Remaining")

					Expect(limit).To(Equal(strconv.Itoa(cfg.RateLimit.Requests)))
					remainingInt, _ := strconv.Atoi(remaining)
					Expect(remainingInt).To(Equal(cfg.RateLimit.Requests - i - 1))

					recorder = httptest.NewRecorder() // Reset recorder for next request
				}
			})
		})

		Context("when requests exceed the limit", func() {
			It("blocks excess requests and sets retry-after header", func() {
				// Use up all available tokens
				for i := 0; i < cfg.RateLimit.Requests; i++ {
					req, _ := http.NewRequest(http.MethodGet, "/test", nil)
					req.RemoteAddr = "192.168.1.1:12345"
					router.ServeHTTP(recorder, req)
					recorder = httptest.NewRecorder()
				}

				// Make one more request that should be blocked
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.RemoteAddr = "192.168.1.1:12345"
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(429))
				Expect(recorder.Header().Get("Retry-After")).NotTo(BeEmpty())
			})
		})

		Context("when tokens are refilled", func() {
			It("allows requests after the refill period", func() {
				// Use up all tokens
				for i := 0; i < cfg.RateLimit.Requests; i++ {
					req, _ := http.NewRequest(http.MethodGet, "/test", nil)
					req.RemoteAddr = "192.168.1.1:12345"
					router.ServeHTTP(recorder, req)
					recorder = httptest.NewRecorder()
				}

				// Wait for tokens to be refilled
				time.Sleep(cfg.RateLimit.Window)

				// Try another request
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.RemoteAddr = "192.168.1.1:12345"
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal("success"))
			})
		})

		Context("when different clients make requests", func() {
			It("tracks rate limits separately for each client", func() {
				// First client uses all tokens
				for i := 0; i < cfg.RateLimit.Requests; i++ {
					req, _ := http.NewRequest(http.MethodGet, "/test", nil)
					req.RemoteAddr = "192.168.1.1:12345"
					router.ServeHTTP(recorder, req)
					recorder = httptest.NewRecorder()
				}

				// Second client should still have all tokens available
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.RemoteAddr = "192.168.1.2:12345"
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal("success"))
			})
		})

		Context("when cleanup runs", func() {
			It("removes old buckets", func() {
				// Make a request to create a bucket
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.RemoteAddr = "192.168.1.1:12345"
				router.ServeHTTP(recorder, req)

				// Wait for cleanup period
				time.Sleep(200 * time.Millisecond) // Wait for two cleanup cycles

				// The bucket should be cleaned up, so we should have full tokens again
				recorder = httptest.NewRecorder()
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				remaining := recorder.Header().Get("X-RateLimit-Remaining")
				remainingInt, _ := strconv.Atoi(remaining)
				Expect(remainingInt).To(Equal(cfg.RateLimit.Requests - 1))
			})
		})
	})

	Describe("NewRateLimiter", func() {
		var testLogger *zap.Logger

		BeforeEach(func() {
			testLogger, _ = zap.NewDevelopment()
		})

		It("creates a rate limiter with the correct configuration", func() {
			limiter := middleware.NewRateLimiter(cfg, testLogger)
			Expect(limiter).NotTo(BeNil())

			// Test that it behaves correctly
			allowed, remaining, _ := limiter.Allow("test-client")
			Expect(allowed).To(BeTrue())
			Expect(remaining).To(Equal(cfg.RateLimit.Requests - 1))
		})
	})

	Describe("min function", func() {
		var testLogger *zap.Logger

		BeforeEach(func() {
			testLogger, _ = zap.NewDevelopment()
		})

		Context("when comparing two integers", func() {
			It("returns the smaller value when first argument is smaller", func() {
				// Set up a config with specific limits
				testCfg := &config.Config{
					RateLimit: config.RateLimitConfig{
						Requests: 2, // Set a small capacity
						Window:   time.Second,
					},
				}

				limiter := middleware.NewRateLimiter(testCfg, testLogger)

				// Use up one token
				allowed1, remaining1, _ := limiter.Allow("test-client-min")
				Expect(allowed1).To(BeTrue())
				Expect(remaining1).To(Equal(1)) // Should be capacity - 1

				// Use up another token
				allowed2, remaining2, _ := limiter.Allow("test-client-min")
				Expect(allowed2).To(BeTrue())
				Expect(remaining2).To(Equal(0)) // Should be min(capacity, 0) = 0

				// Try to use another token
				allowed3, _, _ := limiter.Allow("test-client-min")
				Expect(allowed3).To(BeFalse()) // No more tokens
			})

			It("returns the smaller value when second argument is smaller", func() {
				// Set up a config with specific limits
				testCfg := &config.Config{
					RateLimit: config.RateLimitConfig{
						Requests: 10, // Set a larger capacity
						Window:   time.Second,
					},
				}

				limiter := middleware.NewRateLimiterWithCleanup(testCfg, testLogger, 10*time.Millisecond)

				// Add tokens incrementally (using the cleanup mechanism)
				// This will eventually refill partially instead of fully
				// Test this by exhausting tokens first
				for i := 0; i < 10; i++ {
					allowed, _, _ := limiter.Allow("test-client-min2")
					Expect(allowed).To(BeTrue())
				}

				// Now we're out of tokens
				allowed, _, _ := limiter.Allow("test-client-min2")
				Expect(allowed).To(BeFalse())

				// Wait for partial refill (about 20% of tokens)
				// This tests when tokensToAdd (e.g. 2) < capacity (10)
				time.Sleep(250 * time.Millisecond)

				// We should now have *some* tokens but not all
				// Should be able to get a couple but not all 10
				for i := 0; i < 3; i++ {
					allowed, _, _ := limiter.Allow("test-client-min2")
					if i < 2 {
						Expect(allowed).To(BeTrue(), "Should have refilled at least 2 tokens")
					}
				}
			})
		})
	})

	Describe("Min function", func() {
		It("returns the first argument when it is smaller", func() {
			result := middleware.Min(3, 5)
			Expect(result).To(Equal(3))
		})

		It("returns the second argument when it is smaller", func() {
			result := middleware.Min(8, 2)
			Expect(result).To(Equal(2))
		})

		It("returns the same value when both arguments are equal", func() {
			result := middleware.Min(4, 4)
			Expect(result).To(Equal(4))
		})
	})
})
