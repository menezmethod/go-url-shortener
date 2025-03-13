package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/menezmethod/ref_go/internal/config"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu            sync.Mutex
	buckets       map[string]*tokenBucket
	capacity      int           // Maximum tokens per bucket
	refillRate    time.Duration // Rate at which tokens are refilled
	cleanupPeriod time.Duration // How often to clean up old buckets
	logger        *zap.Logger
}

// tokenBucket represents a token bucket for an individual client
type tokenBucket struct {
	tokens    int       // Current number of tokens
	lastSeen  time.Time // Last time this bucket was accessed
	lastRefil time.Time // Last time tokens were refilled
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cfg *config.Config, logger *zap.Logger) *RateLimiter {
	return NewRateLimiterWithCleanup(cfg, logger, 10*time.Minute)
}

// NewRateLimiterWithCleanup creates a new rate limiter with a custom cleanup period
func NewRateLimiterWithCleanup(cfg *config.Config, logger *zap.Logger, cleanupPeriod time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		buckets:       make(map[string]*tokenBucket),
		capacity:      cfg.RateLimit.Requests,
		refillRate:    cfg.RateLimit.Window,
		cleanupPeriod: cleanupPeriod,
		logger:        logger,
	}

	// Start a goroutine to periodically clean up old buckets
	go limiter.cleanupTask()

	return limiter
}

// cleanupTask removes buckets that haven't been seen in a while
func (rl *RateLimiter) cleanupTask() {
	ticker := time.NewTicker(rl.cleanupPeriod)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes old buckets
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	threshold := time.Now().Add(-rl.cleanupPeriod)
	for key, bucket := range rl.buckets {
		if bucket.lastSeen.Before(threshold) {
			delete(rl.buckets, key)
		}
	}
}

// Allow checks if a request is allowed based on the client's identifier
func (rl *RateLimiter) Allow(identifier string) (bool, int, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Get or create a bucket for this client
	bucket, exists := rl.buckets[identifier]
	if !exists {
		bucket = &tokenBucket{
			tokens:    rl.capacity,
			lastSeen:  now,
			lastRefil: now,
		}
		rl.buckets[identifier] = bucket
	}

	// Update last seen time
	bucket.lastSeen = now

	// Calculate tokens to add based on time elapsed since last refill
	elapsed := now.Sub(bucket.lastRefil)
	tokensToAdd := int(elapsed.Seconds() / rl.refillRate.Seconds() * float64(rl.capacity))

	if tokensToAdd > 0 {
		bucket.tokens = min(bucket.tokens+tokensToAdd, rl.capacity)
		bucket.lastRefil = now
	}

	// If the bucket has tokens, allow the request
	if bucket.tokens > 0 {
		bucket.tokens--
		return true, bucket.tokens, time.Time{}
	}

	// Calculate when the next token will be available
	nextRefill := bucket.lastRefil.Add(rl.refillRate)
	return false, 0, nextRefill
}

// RateLimit middleware limits the rate of requests
func RateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP address)
		clientIP := c.ClientIP()
		logger := GetLogger(c)

		// Check if the request is allowed
		allowed, remaining, retryAfter := limiter.Allow(clientIP)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.capacity))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if !allowed {
			// Set retry-after header
			c.Header("Retry-After", strconv.Itoa(int(time.Until(retryAfter).Seconds())))

			// Return 429 Too Many Requests
			logger.Info("Rate limit exceeded",
				zap.String("client_ip", clientIP),
				zap.Time("retry_after", retryAfter),
			)
			c.AbortWithStatusJSON(429, gin.H{"error": "Rate limit exceeded"})
			return
		}

		// Process the request
		c.Next()
	}
}

// Min returns the minimum of two integers (exported for testing)
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	return Min(a, b)
}

// GetBucketCount returns the current number of buckets (for testing)
func (rl *RateLimiter) GetBucketCount() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.buckets)
}
