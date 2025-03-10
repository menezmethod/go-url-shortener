package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/menezmethod/ref_go/internal/api/middleware"
)

// MockMetrics implements the MetricsCollector interface for testing
type MockMetrics struct {
	mu                 sync.RWMutex
	requestCount       int64
	requestCountByPath map[string]int64
	errorCount         int64
	errorCountByPath   map[string]int64
	statusCount        map[int]int64
	activeRequests     int64
	totalDuration      time.Duration
}

func NewMockMetrics() *MockMetrics {
	return &MockMetrics{
		requestCountByPath: make(map[string]int64),
		errorCountByPath:   make(map[string]int64),
		statusCount:        make(map[int]int64),
	}
}

func (m *MockMetrics) RecordRequest(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestCount++
	m.requestCountByPath[path]++
	m.activeRequests++
}

func (m *MockMetrics) RecordResponse(path string, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeRequests--
	m.totalDuration += duration
	m.statusCount[statusCode]++
	if statusCode >= 400 {
		m.errorCount++
		m.errorCountByPath[path]++
	}
}

func (m *MockMetrics) GetRequestCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.requestCount
}

func (m *MockMetrics) GetErrorCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errorCount
}

func (m *MockMetrics) GetActiveRequests() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeRequests
}

func (m *MockMetrics) GetAverageResponseTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.requestCount == 0 {
		return 0
	}
	return m.totalDuration / time.Duration(m.requestCount)
}

func (m *MockMetrics) GetRequestCountByPath() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]int64)
	for k, v := range m.requestCountByPath {
		result[k] = v
	}
	return result
}

func (m *MockMetrics) GetErrorCountByPath() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]int64)
	for k, v := range m.errorCountByPath {
		result[k] = v
	}
	return result
}

func (m *MockMetrics) GetRequestCountByStatus() map[int]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[int]int64)
	for k, v := range m.statusCount {
		result[k] = v
	}
	return result
}

var _ = Describe("Metrics Middleware", func() {
	var (
		router   *gin.Engine
		recorder *httptest.ResponseRecorder
		metrics  *MockMetrics
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		router = gin.New()
		recorder = httptest.NewRecorder()
		metrics = NewMockMetrics()

		// Set up test endpoint with metrics middleware
		router.Use(middleware.Metrics(metrics))
		router.GET("/test", func(c *gin.Context) {
			time.Sleep(10 * time.Millisecond) // Add a small delay to test duration
			c.String(http.StatusOK, "success")
		})

		router.GET("/error", func(c *gin.Context) {
			c.String(http.StatusInternalServerError, "error")
		})
	})

	Context("when handling requests", func() {
		It("should record successful requests", func() {
			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusOK))
			Expect(metrics.GetRequestCount()).To(Equal(int64(1)))
			Expect(metrics.GetErrorCount()).To(Equal(int64(0)))
			Expect(metrics.GetRequestCountByPath()["/test"]).To(Equal(int64(1)))
		})

		It("should record error requests", func() {
			req := httptest.NewRequest("GET", "/error", nil)
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			Expect(metrics.GetRequestCount()).To(Equal(int64(1)))
			Expect(metrics.GetErrorCount()).To(Equal(int64(1)))
			Expect(metrics.GetRequestCountByPath()["/error"]).To(Equal(int64(1)))
			Expect(metrics.GetErrorCountByPath()["/error"]).To(Equal(int64(1)))
		})

		It("should handle concurrent requests correctly", func() {
			var wg sync.WaitGroup
			concurrentRequests := 10

			for i := 0; i < concurrentRequests; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					req := httptest.NewRequest("GET", "/test", nil)
					router.ServeHTTP(httptest.NewRecorder(), req)
				}()
			}

			wg.Wait()

			Expect(metrics.GetRequestCount()).To(Equal(int64(concurrentRequests)))
			Expect(metrics.GetRequestCountByPath()["/test"]).To(Equal(int64(concurrentRequests)))
			Expect(metrics.GetActiveRequests()).To(Equal(int64(0)))
		})

		It("should record response duration", func() {
			req := httptest.NewRequest("GET", "/test", nil)
			router.ServeHTTP(recorder, req)

			avgDuration := metrics.GetAverageResponseTime()
			Expect(avgDuration).To(BeNumerically(">=", 10*time.Millisecond))
		})

		It("should record status codes correctly", func() {
			req1 := httptest.NewRequest("GET", "/test", nil)
			req2 := httptest.NewRequest("GET", "/error", nil)

			router.ServeHTTP(recorder, req1)
			router.ServeHTTP(recorder, req2)

			statusCounts := metrics.GetRequestCountByStatus()
			Expect(statusCounts[http.StatusOK]).To(Equal(int64(1)))
			Expect(statusCounts[http.StatusInternalServerError]).To(Equal(int64(1)))
		})
	})
})
