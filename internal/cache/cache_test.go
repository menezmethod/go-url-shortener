package cache

import (
	"sync"
	"testing"
	"time"
)

func TestInMemoryCache(t *testing.T) {
	// Create a new cache with a 1-second default TTL
	cache := New(1*time.Second, 5*time.Second)

	// Test Set and Get
	cache.Set("test-key", "test-value", 0)
	value, found := cache.Get("test-key")
	if !found {
		t.Error("Expected to find value in cache")
	}
	if value != "test-value" {
		t.Errorf("Expected value to be %s, got %s", "test-value", value)
	}

	// Test Delete
	cache.Delete("test-key")
	_, found = cache.Get("test-key")
	if found {
		t.Error("Expected key to be deleted")
	}

	// Test Clear
	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	cache.Clear()
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	if found1 || found2 {
		t.Error("Expected all keys to be cleared")
	}

	// Test expiration
	cache.Set("expiring-key", "expiring-value", 100*time.Millisecond)
	time.Sleep(200 * time.Millisecond)
	_, found = cache.Get("expiring-key")
	if found {
		t.Error("Expected key to be expired")
	}

	// Reset cache for stats test
	cache = New(1*time.Second, 5*time.Second)

	// Test stats
	cache.Set("key1", "value1", 0)
	cache.Get("key1") // Hit
	cache.Get("key2") // Miss
	stats := cache.GetStats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	if stats.Items != 1 {
		t.Errorf("Expected 1 item, got %d", stats.Items)
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := New(5*time.Minute, 10*time.Minute)
	var wg sync.WaitGroup
	concurrentOps := 100

	// Test concurrent writes
	wg.Add(concurrentOps)
	for i := 0; i < concurrentOps; i++ {
		go func(i int) {
			defer wg.Done()
			key := "key" + string(rune(i))
			cache.Set(key, i, 0)
		}(i)
	}
	wg.Wait()

	// Test concurrent reads
	hits := 0
	var mu sync.Mutex
	wg.Add(concurrentOps)
	for i := 0; i < concurrentOps; i++ {
		go func(i int) {
			defer wg.Done()
			key := "key" + string(rune(i))
			if val, found := cache.Get(key); found {
				mu.Lock()
				hits++
				mu.Unlock()
				if val.(int) != i {
					t.Errorf("Expected value %d for key %s, got %d", i, key, val)
				}
			}
		}(i)
	}
	wg.Wait()

	if hits != concurrentOps {
		t.Errorf("Expected %d hits, got %d", concurrentOps, hits)
	}

	// Test stats after concurrent operations
	stats := cache.GetStats()
	if stats.Hits != uint64(concurrentOps) {
		t.Errorf("Expected %d hits in stats, got %d", concurrentOps, stats.Hits)
	}
}

func BenchmarkCacheGetSet(b *testing.B) {
	cache := New(5*time.Minute, 10*time.Minute)
	cache.Set("bench-key", "bench-value", 0)

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cache.Get("bench-key")
		}
	})

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cache.Set("bench-key", "bench-value", 0)
		}
	})
}
