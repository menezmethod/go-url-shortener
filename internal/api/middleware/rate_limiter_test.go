package middleware_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
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

// Add this after the main RateLimiter tests
var _ = Describe("RateLimiter Stress Tests", func() {
	var (
		rateLimiter *middleware.RateLimiter
		cfg         *config.Config
		logger      *zap.Logger
	)

	BeforeEach(func() {
		cfg = &config.Config{
			RateLimit: config.RateLimitConfig{
				Requests: 50,                     // Allow 50 requests
				Window:   100 * time.Millisecond, // In a 100ms window
			},
		}
		logger, _ = zap.NewDevelopment()
		rateLimiter = middleware.NewRateLimiterWithCleanup(cfg, logger, 50*time.Millisecond)
	})

	Context("under high concurrency", func() {
		It("should properly limit concurrent requests from the same client", func() {
			// Number of concurrent goroutines
			concurrency := 100
			// Single client ID
			clientID := "192.168.1.1"

			// Channels to collect results
			allowed := make(chan bool, concurrency)
			finished := make(chan struct{})

			// Start a goroutine to collect results
			var allowedCount int32
			go func() {
				for a := range allowed {
					if a {
						atomic.AddInt32(&allowedCount, 1)
					}
				}
				close(finished)
			}()

			// Start concurrent goroutines to request rate limiting
			var wg sync.WaitGroup
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					isAllowed, _, _ := rateLimiter.Allow(clientID)
					allowed <- isAllowed
				}()
			}

			// Wait for all requests to finish
			wg.Wait()
			close(allowed)
			<-finished

			// Only 50 requests should be allowed
			Expect(int(allowedCount)).To(Equal(cfg.RateLimit.Requests))
		})

		It("should properly handle different clients concurrently", func() {
			// Number of concurrent goroutines per client
			requestsPerClient := 20
			// Number of different clients
			numClients := 10

			// Channels to collect results
			results := make(chan struct {
				clientID string
				allowed  bool
			}, requestsPerClient*numClients)

			finished := make(chan struct{})

			// Start a goroutine to collect results
			clientAllowed := make(map[string]int)
			go func() {
				for r := range results {
					if r.allowed {
						clientAllowed[r.clientID]++
					}
				}
				close(finished)
			}()

			// Start concurrent goroutines for multiple clients
			var wg sync.WaitGroup
			for c := 0; c < numClients; c++ {
				clientID := fmt.Sprintf("client-%d", c)

				for i := 0; i < requestsPerClient; i++ {
					wg.Add(1)
					go func(cid string) {
						defer wg.Done()
						isAllowed, _, _ := rateLimiter.Allow(cid)
						results <- struct {
							clientID string
							allowed  bool
						}{clientID: cid, allowed: isAllowed}
					}(clientID)
				}
			}

			// Wait for all requests to finish
			wg.Wait()
			close(results)
			<-finished

			// Each client should not exceed their quota
			for clientID, count := range clientAllowed {
				Expect(count).To(BeNumerically("<=", cfg.RateLimit.Requests),
					fmt.Sprintf("Client %s should not exceed rate limit", clientID))
			}
		})

		It("should enforce rate limits over time with bursts", func() {
			clientID := "burst-client"

			// First burst should allow up to the limit
			var initialAllowed int
			for i := 0; i < cfg.RateLimit.Requests+10; i++ {
				allowed, _, _ := rateLimiter.Allow(clientID)
				if allowed {
					initialAllowed++
				}
			}
			Expect(initialAllowed).To(BeNumerically("<=", cfg.RateLimit.Requests))

			// Wait for partial refill (50% of the window)
			time.Sleep(cfg.RateLimit.Window / 2)

			// Second burst should allow some requests
			var secondAllowed int
			for i := 0; i < cfg.RateLimit.Requests; i++ {
				allowed, _, _ := rateLimiter.Allow(clientID)
				if allowed {
					secondAllowed++
				}
			}

			// We expect some tokens to be refilled
			Expect(secondAllowed).To(BeNumerically(">", 0))

			// Wait for full window to refill all tokens
			time.Sleep(cfg.RateLimit.Window)

			// Third burst should allow requests again
			var thirdAllowed int
			for i := 0; i < cfg.RateLimit.Requests+10; i++ {
				allowed, _, _ := rateLimiter.Allow(clientID)
				if allowed {
					thirdAllowed++
				}
			}
			Expect(thirdAllowed).To(BeNumerically(">", 0))
		})

		It("should handle cleanup of inactive clients", func() {
			// Create a large number of clients and use them once
			numClients := 1000
			for i := 0; i < numClients; i++ {
				rateLimiter.Allow(fmt.Sprintf("one-time-client-%d", i))
			}

			// Get the current bucket count
			initialBucketCount := rateLimiter.GetBucketCount()
			Expect(initialBucketCount).To(Equal(numClients))

			// Wait for cleanup
			time.Sleep(200 * time.Millisecond)

			// Trigger cleanup by making one more request
			rateLimiter.Allow("trigger-cleanup")

			// Bucket count should be dramatically reduced
			Expect(rateLimiter.GetBucketCount()).To(BeNumerically("<", initialBucketCount))
		})

		It("should be resilient under extreme load", func() {
			// Simulate extremely high traffic with different traffic patterns

			// 1. Constant high-rate traffic from one client
			highRateClient := "high-rate-client"

			// 2. Bursty traffic from multiple clients
			burstyClients := 20

			// 3. Random low-rate traffic from many clients
			randomClients := 50

			var wg sync.WaitGroup

			// Start high-rate client
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 200; i++ {
					rateLimiter.Allow(highRateClient)
					time.Sleep(time.Millisecond) // Small delay between requests
				}
			}()

			// Start bursty clients
			for c := 0; c < burstyClients; c++ {
				wg.Add(1)
				go func(clientID int) {
					defer wg.Done()
					clientName := fmt.Sprintf("bursty-client-%d", clientID)

					// Make bursts of requests with pauses between
					for burst := 0; burst < 3; burst++ {
						// Burst of requests
						for i := 0; i < 20; i++ {
							rateLimiter.Allow(clientName)
						}
						// Pause
						time.Sleep(50 * time.Millisecond)
					}
				}(c)
			}

			// Start random clients
			for c := 0; c < randomClients; c++ {
				wg.Add(1)
				go func(clientID int) {
					defer wg.Done()
					clientName := fmt.Sprintf("random-client-%d", clientID)

					// Make random number of requests with random delays
					requests := 5 + rand.Intn(20)
					for i := 0; i < requests; i++ {
						rateLimiter.Allow(clientName)
						delay := time.Duration(1+rand.Intn(10)) * time.Millisecond
						time.Sleep(delay)
					}
				}(c)
			}

			// Wait for all traffic simulation to complete
			wg.Wait()

			// Verify the rate limiter is still functional
			// High-rate client should be rate limited now
			allowed, tokens, _ := rateLimiter.Allow(highRateClient)
			if allowed {
				// The client might have some tokens due to refill
				Expect(tokens).To(BeNumerically("<", cfg.RateLimit.Requests))
			} else {
				// Client is rate limited
				Expect(tokens).To(Equal(0))
			}

			// A new client should not be affected by the high traffic
			allowed, tokens, _ = rateLimiter.Allow("new-client")
			Expect(allowed).To(BeTrue())
			Expect(tokens).To(Equal(cfg.RateLimit.Requests - 1))
		})
	})
})
