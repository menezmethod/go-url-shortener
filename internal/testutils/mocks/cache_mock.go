package mocks

import (
	"github.com/menezmethod/ref_go/internal/cache"
)

// MockCache implements the CacheInterface for testing
type MockCache struct {
	GetFunc      func(key string) (interface{}, bool)
	SetFunc      func(key string, value interface{}, ttl int)
	DeleteFunc   func(key string)
	GetStatsFunc func() cache.Stats
}

// Get mocks the Get method
func (m *MockCache) Get(key string) (interface{}, bool) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	return nil, false
}

// Set mocks the Set method
func (m *MockCache) Set(key string, value interface{}, ttl int) {
	if m.SetFunc != nil {
		m.SetFunc(key, value, ttl)
	}
}

// Delete mocks the Delete method
func (m *MockCache) Delete(key string) {
	if m.DeleteFunc != nil {
		m.DeleteFunc(key)
	}
}

// GetStats mocks the GetStats method
func (m *MockCache) GetStats() cache.Stats {
	if m.GetStatsFunc != nil {
		return m.GetStatsFunc()
	}
	return cache.Stats{}
}
