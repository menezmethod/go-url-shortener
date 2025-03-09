package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics collects and provides API metrics
type Metrics struct {
	// Request counts
	requestCount         int64
	requestCountByPath   map[string]int64
	requestCountByPathMu sync.RWMutex

	// Error counts
	errorCount         int64
	errorCountByPath   map[string]int64
	errorCountByPathMu sync.RWMutex

	// Response time
	totalResponseTime         time.Duration
	totalResponseTimeByPath   map[string]time.Duration
	totalResponseTimeByPathMu sync.RWMutex

	// Request count by status code
	requestCountByStatus   map[int]int64
	requestCountByStatusMu sync.RWMutex

	// Active requests
	activeRequests int64

	// Link metrics
	shortLinkCount    int64
	totalRedirects    int64
	redirectsByLink   map[string]int64
	redirectsByLinkMu sync.RWMutex

	// Cache metrics
	cacheHits       int64
	cacheMisses     int64
	cacheTotalItems int64
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		requestCountByPath:      make(map[string]int64),
		errorCountByPath:        make(map[string]int64),
		totalResponseTimeByPath: make(map[string]time.Duration),
		requestCountByStatus:    make(map[int]int64),
		redirectsByLink:         make(map[string]int64),
	}
}

// RecordRequest records a request
func (m *Metrics) RecordRequest(path string) {
	atomic.AddInt64(&m.requestCount, 1)
	atomic.AddInt64(&m.activeRequests, 1)

	m.requestCountByPathMu.Lock()
	m.requestCountByPath[path]++
	m.requestCountByPathMu.Unlock()
}

// RecordResponse records a response
func (m *Metrics) RecordResponse(path string, statusCode int, duration time.Duration) {
	atomic.AddInt64(&m.activeRequests, -1)

	// Record response time
	atomic.AddInt64((*int64)(&m.totalResponseTime), int64(duration))

	m.totalResponseTimeByPathMu.Lock()
	m.totalResponseTimeByPath[path] += duration
	m.totalResponseTimeByPathMu.Unlock()

	// Record status code
	m.requestCountByStatusMu.Lock()
	m.requestCountByStatus[statusCode]++
	m.requestCountByStatusMu.Unlock()

	// Record error if status >= 400
	if statusCode >= 400 {
		atomic.AddInt64(&m.errorCount, 1)

		m.errorCountByPathMu.Lock()
		m.errorCountByPath[path]++
		m.errorCountByPathMu.Unlock()
	}
}

// RecordRedirect records a link redirect
func (m *Metrics) RecordRedirect(linkID string) {
	// Add a println for debugging to see if this method is being called
	fmt.Printf("[DEBUG] RecordRedirect called for link ID: %s\n", linkID)

	atomic.AddInt64(&m.totalRedirects, 1)

	m.redirectsByLinkMu.Lock()
	m.redirectsByLink[linkID]++
	m.redirectsByLinkMu.Unlock()
}

// SetShortLinkCount sets the current short link count
func (m *Metrics) SetShortLinkCount(count int64) {
	atomic.StoreInt64(&m.shortLinkCount, count)
}

// GetRequestCount returns the total request count
func (m *Metrics) GetRequestCount() int64 {
	return atomic.LoadInt64(&m.requestCount)
}

// GetErrorCount returns the total error count
func (m *Metrics) GetErrorCount() int64 {
	return atomic.LoadInt64(&m.errorCount)
}

// GetActiveRequests returns the current active request count
func (m *Metrics) GetActiveRequests() int64 {
	return atomic.LoadInt64(&m.activeRequests)
}

// GetAverageResponseTime returns the average response time
func (m *Metrics) GetAverageResponseTime() time.Duration {
	count := atomic.LoadInt64(&m.requestCount)
	if count == 0 {
		return 0
	}

	total := atomic.LoadInt64((*int64)(&m.totalResponseTime))
	return time.Duration(total) / time.Duration(count)
}

// GetTotalRedirects returns the total redirect count
func (m *Metrics) GetTotalRedirects() int64 {
	return atomic.LoadInt64(&m.totalRedirects)
}

// GetShortLinkCount returns the current short link count
func (m *Metrics) GetShortLinkCount() int64 {
	return atomic.LoadInt64(&m.shortLinkCount)
}

// GetRequestCountByPath returns request counts by path
func (m *Metrics) GetRequestCountByPath() map[string]int64 {
	result := make(map[string]int64)

	m.requestCountByPathMu.RLock()
	defer m.requestCountByPathMu.RUnlock()

	for path, count := range m.requestCountByPath {
		result[path] = count
	}

	return result
}

// GetErrorCountByPath returns error counts by path
func (m *Metrics) GetErrorCountByPath() map[string]int64 {
	result := make(map[string]int64)

	m.errorCountByPathMu.RLock()
	defer m.errorCountByPathMu.RUnlock()

	for path, count := range m.errorCountByPath {
		result[path] = count
	}

	return result
}

// GetRequestCountByStatus returns request counts by status code
func (m *Metrics) GetRequestCountByStatus() map[int]int64 {
	result := make(map[int]int64)

	m.requestCountByStatusMu.RLock()
	defer m.requestCountByStatusMu.RLock()

	for status, count := range m.requestCountByStatus {
		result[status] = count
	}

	return result
}

// GetRedirectsByLink returns redirect counts by link ID
func (m *Metrics) GetRedirectsByLink() map[string]int64 {
	result := make(map[string]int64)

	m.redirectsByLinkMu.RLock()
	defer m.redirectsByLinkMu.RUnlock()

	for linkID, count := range m.redirectsByLink {
		result[linkID] = count
	}

	return result
}

// GetCacheHits returns the cache hit count
func (m *Metrics) GetCacheHits() int64 {
	return atomic.LoadInt64(&m.cacheHits)
}

// SetCacheHits sets the cache hit count
func (m *Metrics) SetCacheHits(count int64) {
	atomic.StoreInt64(&m.cacheHits, count)
}

// GetCacheMisses returns the cache miss count
func (m *Metrics) GetCacheMisses() int64 {
	return atomic.LoadInt64(&m.cacheMisses)
}

// SetCacheMisses sets the cache miss count
func (m *Metrics) SetCacheMisses(count int64) {
	atomic.StoreInt64(&m.cacheMisses, count)
}

// GetCacheTotalItems returns the cache item count
func (m *Metrics) GetCacheTotalItems() int64 {
	return atomic.LoadInt64(&m.cacheTotalItems)
}

// SetCacheTotalItems sets the cache item count
func (m *Metrics) SetCacheTotalItems(count int64) {
	atomic.StoreInt64(&m.cacheTotalItems, count)
}

// ServeHTTP implements the http.Handler interface for metrics
func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Format metrics for Prometheus scraping or as JSON for manual review
	w.Header().Set("Content-Type", "text/plain")

	// Add a println for debugging to show the current redirect count
	totalRedirects := atomic.LoadInt64(&m.totalRedirects)
	fmt.Printf("[DEBUG] Current totalRedirects value: %d\n", totalRedirects)

	metrics := []struct {
		name  string
		value interface{}
		help  string
	}{
		{"url_shortener_requests_total", m.GetRequestCount(), "Total number of requests"},
		{"url_shortener_errors_total", m.GetErrorCount(), "Total number of errors"},
		{"url_shortener_active_requests", m.GetActiveRequests(), "Current number of active requests"},
		{"url_shortener_average_response_time_ms", m.GetAverageResponseTime().Milliseconds(), "Average response time in milliseconds"},
		{"url_shortener_redirects_total", m.GetTotalRedirects(), "Total number of redirects"},
		{"url_shortener_links_total", m.GetShortLinkCount(), "Total number of short links"},
		{"url_shortener_cache_hits_total", m.GetCacheHits(), "Total number of cache hits"},
		{"url_shortener_cache_misses_total", m.GetCacheMisses(), "Total number of cache misses"},
		{"url_shortener_cache_items_total", m.GetCacheTotalItems(), "Total number of items in cache"},
	}

	for _, metric := range metrics {
		w.Write([]byte(formatMetric(metric.name, metric.value, metric.help)))
	}
}

// formatMetric formats a Prometheus-style metric
func formatMetric(name string, value interface{}, help string) string {
	return "# HELP " + name + " " + help + "\n" +
		"# TYPE " + name + " gauge\n" +
		name + " " + formatValue(value) + "\n\n"
}

// formatValue formats a value for Prometheus output
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return "0"
	}
}
