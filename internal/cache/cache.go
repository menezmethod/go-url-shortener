package cache

// CacheInterface defines the interface for caching operations
type CacheInterface interface {
	// Get retrieves a value from the cache
	Get(key string) (interface{}, bool)

	// Set adds a value to the cache
	Set(key string, value interface{}, ttl int)

	// Delete removes a value from the cache
	Delete(key string)

	// GetStats returns statistics about cache usage
	GetStats() Stats
}

// Stats represents cache statistics
type Stats struct {
	Size    int `json:"size"`
	Hits    int `json:"hits"`
	Misses  int `json:"misses"`
	Evicted int `json:"evicted"`
}
