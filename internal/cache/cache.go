package cache

import (
	"sync/atomic"
	"time"

	goCache "github.com/patrickmn/go-cache"
)

// CacheInterface defines methods for caching
type CacheInterface interface {
	// Get retrieves a value from the cache
	Get(key string) (interface{}, bool)

	// Set adds a value to the cache
	Set(key string, value interface{}, ttl time.Duration)

	// Delete removes a value from the cache
	Delete(key string)

	// Clear empties the entire cache
	Clear()

	// GetStats returns statistics about cache usage
	GetStats() Stats
}

// Stats contains cache usage statistics
type Stats struct {
	Hits   uint64 `json:"hits"`
	Misses uint64 `json:"misses"`
	Items  int    `json:"items"`
}

// InMemoryCache implements the Cache interface using in-memory storage
type InMemoryCache struct {
	cache      *goCache.Cache
	hits       uint64
	misses     uint64
	defaultTTL time.Duration
}

// New creates a new in-memory cache
func New(defaultTTL, cleanupInterval time.Duration) *InMemoryCache {
	return &InMemoryCache{
		cache:      goCache.New(defaultTTL, cleanupInterval),
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	value, found := c.cache.Get(key)
	if found {
		atomic.AddUint64(&c.hits, 1)
	} else {
		atomic.AddUint64(&c.misses, 1)
	}
	return value, found
}

// Set adds a value to the cache
func (c *InMemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	if ttl == 0 {
		ttl = c.defaultTTL
	}
	c.cache.Set(key, value, ttl)
}

// Delete removes a value from the cache
func (c *InMemoryCache) Delete(key string) {
	c.cache.Delete(key)
}

// Clear empties the entire cache
func (c *InMemoryCache) Clear() {
	c.cache.Flush()
}

// GetStats returns statistics about cache usage
func (c *InMemoryCache) GetStats() Stats {
	items := c.cache.ItemCount()
	return Stats{
		Hits:   atomic.LoadUint64(&c.hits),
		Misses: atomic.LoadUint64(&c.misses),
		Items:  items,
	}
}
