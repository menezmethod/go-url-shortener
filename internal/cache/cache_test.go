package cache

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Suite")
}

var _ = Describe("MemoryCache", func() {
	var (
		cache *MemoryCache
	)

	BeforeEach(func() {
		cache = NewMemoryCache()
	})

	Describe("Basic Operations", func() {
		It("should set and get values", func() {
			cache.Set("key1", "value1", 60) // 60 seconds TTL
			value, found := cache.Get("key1")
			Expect(found).To(BeTrue())
			Expect(value).To(Equal("value1"))
		})

		It("should handle non-existent keys", func() {
			value, found := cache.Get("non-existent")
			Expect(found).To(BeFalse())
			Expect(value).To(BeNil())
		})

		It("should delete values", func() {
			cache.Set("key1", "value1", 60)
			cache.Delete("key1")
			_, found := cache.Get("key1")
			Expect(found).To(BeFalse())
		})

		It("should handle TTL expiration", func() {
			cache.Set("key1", "value1", 1) // 1 second TTL
			time.Sleep(2 * time.Second)
			_, found := cache.Get("key1")
			Expect(found).To(BeFalse())
		})
	})

	Describe("Stats", func() {
		It("should track basic stats", func() {
			cache.Set("key1", "value1", 60)
			cache.Set("key2", "value2", 60)

			stats := cache.GetStats()
			Expect(stats.Size).To(Equal(2))
			Expect(stats.Hits).To(Equal(0))
			Expect(stats.Misses).To(Equal(0))
			Expect(stats.Evicted).To(Equal(0))

			// Generate some hits and misses
			cache.Get("key1")
			cache.Get("key1")
			cache.Get("non-existent")

			stats = cache.GetStats()
			Expect(stats.Hits).To(Equal(2))
			Expect(stats.Misses).To(Equal(1))
		})

		It("should track evictions", func() {
			cache.Set("key1", "value1", 1) // 1 second TTL
			time.Sleep(2 * time.Second)
			cache.Get("key1") // This will trigger eviction

			stats := cache.GetStats()
			Expect(stats.Evicted).To(Equal(1))
		})
	})

	Describe("Concurrent Operations", func() {
		It("should handle concurrent access safely", func() {
			const concurrentOps = 100
			done := make(chan bool)

			// Concurrent writes
			for i := 0; i < concurrentOps; i++ {
				go func(index int) {
					cache.Set("key", index, 60)
					done <- true
				}(i)
			}

			// Wait for all writes
			for i := 0; i < concurrentOps; i++ {
				<-done
			}

			// Concurrent reads
			for i := 0; i < concurrentOps; i++ {
				go func() {
					cache.Get("key")
					done <- true
				}()
			}

			// Wait for all reads
			for i := 0; i < concurrentOps; i++ {
				<-done
			}

			stats := cache.GetStats()
			Expect(stats.Hits + stats.Misses).To(Equal(concurrentOps))
		})
	})
})
