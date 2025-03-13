package cache

import (
	"sync"
	"time"
)

// MemoryCache implements CacheInterface using in-memory storage
type MemoryCache struct {
	mu      sync.RWMutex
	items   map[string]cacheItem
	hits    int
	misses  int
	evicted int
}

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]cacheItem),
	}
}

// Get retrieves a value from the cache
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, exists := c.items[key]
	if !exists {
		c.mu.RUnlock()
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		c.mu.RUnlock()
		c.Delete(key)
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	c.mu.RUnlock()
	c.mu.Lock()
	c.hits++
	c.mu.Unlock()
	return item.value, true
}

// Set adds a value to the cache
func (c *MemoryCache) Set(key string, value interface{}, ttl int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
	}

	c.items[key] = cacheItem{
		value:     value,
		expiresAt: expiresAt,
	}
}

// Delete removes a value from the cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; exists {
		delete(c.items, key)
		c.evicted++
	}
}

// GetStats returns statistics about cache usage
func (c *MemoryCache) GetStats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		Size:    len(c.items),
		Hits:    c.hits,
		Misses:  c.misses,
		Evicted: c.evicted,
	}
}
