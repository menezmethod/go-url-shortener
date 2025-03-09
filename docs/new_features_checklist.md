# New Features Implementation Checklist

This document outlines the implementation plan for enhancing our URL shortener service with additional features.

## 1. Swagger API Documentation

### Implementation Steps
- [ ] Install required dependencies
  - [ ] github.com/swaggo/swag/cmd/swag (CLI tool)
  - [ ] github.com/swaggo/gin-swagger (Gin middleware)
  - [ ] github.com/swaggo/files (Swagger files)
- [ ] Add main API documentation annotations to main.go
- [ ] Add annotations to all handler methods
- [ ] Configure Swagger middleware in router.go
- [ ] Generate Swagger documentation using swag init
- [ ] Expose Swagger UI via /swagger/* endpoint
- [ ] Add version information to API documentation

### Acceptance Criteria
- Swagger UI is accessible via /swagger/index.html
- All API endpoints are properly documented with:
  - Request parameters
  - Request body schema
  - Response body schema
  - Response codes
  - Authentication requirements
- API documentation includes examples for each endpoint
- Documentation is automatically generated from code annotations

### Testing
```bash
# Generate Swagger docs
swag init -g cmd/server/main.go

# Start the server
go run cmd/server/main.go

# Access Swagger UI
curl http://localhost:8081/swagger/index.html
```

## 2. URL Caching Mechanism

### Implementation Steps
- [ ] Create a new cache package
  - [ ] Implement cache interface
  - [ ] Add in-memory cache implementation using github.com/patrickmn/go-cache
  - [ ] Add configurable TTL for cache entries
  - [ ] Add cache statistics (hits, misses)
- [ ] Integrate cache with URL service
  - [ ] Cache URLs by code for redirection
  - [ ] Update cache on URL creation/update/deletion
  - [ ] Implement LRU eviction strategy
- [ ] Add cache warming on startup (optional)
- [ ] Add metrics for cache performance

### Acceptance Criteria
- Frequently accessed URLs are served from cache
- Cache hit ratio is monitored and exposed in metrics
- Cache size is configurable via environment variables
- Cache entries expire based on configured TTL
- Cache is properly invalidated when URLs are updated/deleted

### Testing
```bash
# Test cache hit ratio
go test -v ./internal/cache

# Benchmark redirection performance
go test -bench=BenchmarkRedirect ./internal/service

# Verify metrics exposure
curl -H "Authorization: Bearer <token>" http://localhost:8081/metrics
```

## 3. URL Expiration Notifications

### Implementation Steps
- [ ] Create notification service interface
  - [ ] Add email notification implementation
  - [ ] Add webhook notification implementation
- [ ] Create expiration checker service
  - [ ] Implement scheduled job to check for expiring URLs
  - [ ] Configure notification thresholds (1 day, 3 days, 7 days)
- [ ] Update URL creation/update API to accept notification settings
  - [ ] Add notification preferences to URL model
  - [ ] Update database schema
- [ ] Create notification templates
- [ ] Add notification logging and metrics

### Acceptance Criteria
- Users can configure notification preferences when creating/updating URLs
- System sends notifications at configured thresholds before expiration
- Notifications can be delivered via multiple channels (email, webhook)
- Notification history is tracked and viewable via API
- Notification service is fault-tolerant (retries on failure)

### Testing
```bash
# Test notification service
go test -v ./internal/notification

# Test expiration checker
go test -v ./internal/service/expiration

# Test notification preferences API
curl -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"url":"https://example.com","expiration_date":"2023-12-31T23:59:59Z","notification_preferences":{"email":"user@example.com","thresholds":[1,7]}}' \
  http://localhost:8081/api/v1/links
```

## 4. Bulk Operations for Links

### Implementation Steps
- [ ] Update service layer to support bulk operations
  - [ ] Implement bulk create
  - [ ] Implement bulk update
  - [ ] Implement bulk delete
  - [ ] Add transaction support for atomic operations
- [ ] Add new API endpoints for bulk operations
  - [ ] POST /api/v1/links/bulk for creation
  - [ ] PUT /api/v1/links/bulk for updates
  - [ ] DELETE /api/v1/links/bulk for deletion
- [ ] Implement request validation for bulk operations
  - [ ] Add maximum batch size limit
  - [ ] Add validation for each item in batch
- [ ] Add proper error handling for partial failures
- [ ] Update Swagger documentation

### Acceptance Criteria
- API supports creating, updating, and deleting multiple links in a single request
- Bulk operations are atomic (all succeed or all fail)
- Proper validation is applied to each item in a batch
- System enforces maximum batch size to prevent abuse
- Detailed error reporting for failed items in a batch
- Performance is optimized for bulk operations

### Testing
```bash
# Test bulk creation
curl -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '[{"url":"https://example1.com"},{"url":"https://example2.com"}]' \
  http://localhost:8081/api/v1/links/bulk

# Test bulk update
curl -X PUT -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '[{"id":"id1","is_active":false},{"id":"id2","is_active":false}]' \
  http://localhost:8081/api/v1/links/bulk

# Test bulk deletion
curl -X DELETE -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '["id1","id2"]' \
  http://localhost:8081/api/v1/links/bulk
```

## 5. API Versioning Strategy

### Implementation Steps
- [ ] Design versioning strategy
  - [ ] Implement URL path versioning (e.g., /api/v1/links)
  - [ ] Update router to support versioned routes
  - [ ] Create version-specific handler packages
- [ ] Implement backward compatibility layer
  - [ ] Add request/response transformers for different versions
  - [ ] Document breaking vs. non-breaking changes
- [ ] Update route registration to use versioned handlers
- [ ] Add version deprecation mechanism
  - [ ] Add deprecation headers for old versions
  - [ ] Add sunset dates for deprecated versions
- [ ] Update documentation to reflect versioning

### Acceptance Criteria
- API routes include version information (e.g., /api/v1/links)
- Different versions can coexist in the same application
- Changes in newer versions don't break older versions
- Deprecated versions return appropriate warning headers
- Documentation clearly indicates version differences
- Version migration path is clearly documented

### Testing
```bash
# Test latest version
curl -v http://localhost:8081/api/v1/links

# Test with explicit version header
curl -v -H "Accept-Version: v1" http://localhost:8081/api/links
```

## Implementation Schedule

1. API Documentation (Swagger Integration): 2 days
2. URL Caching Mechanism: 2 days
3. API Versioning Strategy: 1 day
4. Bulk Operations: 2 days  
5. URL Expiration Notifications: 3 days

Total estimated time: 10 days 