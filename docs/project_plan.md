# Project Plan: URL Shortener in Go

## Project Overview

**Repository:** github.com/menezmethod/ref_go

This document outlines the plan for developing a URL shortener service using Go. The service will provide the ability to create shortened URLs, track their usage, and offer analytics on link clicks. We'll start with a simple foundation and incrementally add production features.

## Core Functionality

1. **URL Shortening**: Generate short codes for long URLs
2. **URL Redirection**: Redirect users from short URLs to original URLs
3. **Custom Aliases**: Allow custom short URLs
4. **Analytics**: Track link usage and provide basic analytics
5. **Token Authentication**: Secure the API with JWT tokens
6. **Rate Limiting**: Prevent abuse of the service

## Technical Stack

- **Language**: Go 1.22+ (to leverage new routing enhancements)
- **Database**: PostgreSQL 17.x (latest stable)
- **Logging**: Uber's zap logger for structured logging
- **HTTP Router**: Standard library net/http with ServeMux (Go 1.22+)
- **Database Access**: Standard library database/sql with lib/pq
- **Authentication**: JWT token-based auth with master password for token retrieval
- **Deployment**: Docker + Coolify

## Project Structure

```
ref_go/
├── cmd/
│   └── server/              # Main application entry point
├── internal/
│   ├── api/                 # HTTP API handlers and middleware
│   │   ├── handlers/        # Request handlers
│   │   ├── middleware/      # HTTP middleware (auth, logging, etc.)
│   │   └── router/          # HTTP router setup
│   ├── auth/                # Authentication and token management
│   ├── config/              # Configuration loading and validation
│   ├── db/                  # Database connection and migrations
│   │   ├── migrations/      # SQL migration files
│   │   └── models/          # Data models and DB operations
│   ├── domain/              # Business domain models
│   ├── repository/          # Data access layer
│   │   ├── postgres/        # PostgreSQL implementations
│   │   └── interfaces.go    # Repository interfaces
│   └── service/             # Business logic
├── scripts/                 # Helper scripts
├── configs/                 # Configuration files
├── docs/                    # Documentation
│   ├── api/                 # API documentation
│   ├── code_standard.md     # Coding standards
│   ├── code_tips.md         # Coding tips reference
│   └── project_milestones.md# Project tracking
├── .env.example             # Example environment variables
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
└── README.md                # Project documentation
```

## Development Phases

### Phase 1: Core Foundation (Week 1)

1. **Project Setup** (Day 1)
   - Initialize Go module with Go 1.22+
   - Set up basic project structure
   - Configure linting based on code_standard.md
   - Implement logging with zap

2. **Database Layer** (Day 2)
   - Design PostgreSQL schema with proper indexes
   - Create migrations using sql-migrate
   - Implement repository pattern with database/sql
   - Set up connection pooling with proper timeout configuration

3. **Core Service Layer** (Day 3)
   - Implement URL shortening algorithm (collision-resistant)
   - Create service for handling URL operations
   - Add proper error handling with wrapping

4. **Basic API & Authentication** (Day 4-5)
   - Set up HTTP server with graceful shutdown
   - Implement ServeMux routing with Go 1.22+ pattern matching
   - Create health check and readiness endpoints
   - Add basic request logging middleware
   - Implement JWT token generation endpoint with master password
   - Implement token validation middleware
   - Set up environment variable for master password

### Phase 2: Essential Features (Week 2)

5. **API Protection** (Day 1)
   - Implement token authentication middleware for protected routes
   - Add rate limiting per IP address
   - Create response headers for rate limit information
   - Add defensive security measures

6. **Redirection Service** (Day 2-3)
   - Implement URL redirection with proper status codes
   - Add tracking of redirects with goroutines for async logging
   - Handle validation, expired links, and error cases

7. **Custom Aliases** (Day 4)
   - Add support for custom URL aliases
   - Implement alias validation
   - Add conflict detection and resolution

8. **Analytics** (Day 5)
   - Implement click tracking with structured metadata
   - Add analytics storage with efficient queries
   - Create analytics dashboarding endpoints

### Phase 3: Production Readiness (Week 3)

9. **Rate Limiting** (Day 1)
   - Add rate limiting middleware with token bucket algorithm
   - Implement configurable limits per route
   - Add rate limit headers in responses

10. **Performance Optimization** (Day 2)
    - Implement request timeout middleware
    - Add context propagation for cancellation
    - Optimize database queries with proper indexes

11. **Observability** (Day 3)
    - Set up structured logging with correlation IDs
    - Add metrics collection with Prometheus format
    - Implement pprof endpoints for profiling

12. **Security Enhancements** (Day 4)
    - Add input validation and sanitization
    - Implement proper error handling that doesn't leak internals
    - Add security headers (CORS, CSP, etc.)

13. **Deployment Configuration** (Day 5)
    - Create production Dockerfile with multi-stage builds
    - Set up Coolify configuration
    - Document operational procedures

## API Endpoints

### Public Endpoints

- `GET /{code}` - Redirect to original URL
- `GET /api/health` - Health check
- `GET /api/ready` - Readiness check
- `POST /api/auth/token` - Get JWT token (requires master password)

### Protected Endpoints (require JWT token)

- `POST /api/links` - Create a new short link
- `GET /api/links` - List all links (paginated)
- `GET /api/links/{id}` - Get link details
- `PUT /api/links/{id}` - Update a link
- `DELETE /api/links/{id}` - Delete a link
- `GET /api/links/{id}/stats` - Get link statistics
- `GET /metrics` - Prometheus metrics (protected)

## Environment Variables

```
# Server
PORT=8081
ENVIRONMENT=development
LOG_LEVEL=debug

# Security
MASTER_PASSWORD=your_master_password  # Password for token generation
TOKEN_EXPIRY=24h                      # JWT token expiration time

# Database
POSTGRES_DB=url_shortener
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_db_password
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_MAX_CONNECTIONS=25
POSTGRES_MAX_IDLE_CONNECTIONS=5
POSTGRES_CONN_MAX_LIFETIME=15m

# Rate Limiting
RATE_LIMIT_REQUESTS=60
RATE_LIMIT_WINDOW=60s
```

## Authentication Flow

1. **Token Generation**:
   - Client sends master password to token endpoint
   - Server validates master password against environment variable
   - If valid, server generates a JWT token with expiration
   - Token is returned to client

2. **API Authentication**:
   - Client includes JWT token in Authorization header
   - Format: `Authorization: Bearer <jwt_token>`
   - Server validates token signature and expiration
   - If valid, request is allowed to proceed

3. **Token Implementation**:
   - Use standard library "github.com/golang-jwt/jwt/v5" for JWT handling
   - Include expiration time in token claims
   - Sign token with master password as secret
   - No database storage needed (stateless authentication)

4. **Security Considerations**:
   - Master password only transmitted during token retrieval
   - All communication over HTTPS
   - Tokens expire automatically
   - Token validation is computationally efficient

## Database Schema

The database will contain the following tables:

1. **urls**
   - `id` (UUID, primary key)
   - `original_url` (TEXT)
   - `hash` (TEXT, indexed)
   - `created_at` (TIMESTAMP)
   - `updated_at` (TIMESTAMP)

2. **short_links**
   - `id` (UUID, primary key)
   - `code` (TEXT, unique, indexed)
   - `custom_alias` (TEXT, unique, nullable, indexed)
   - `url_id` (UUID, foreign key to urls, indexed)
   - `expiration_date` (TIMESTAMP, nullable, indexed)
   - `is_active` (BOOLEAN, indexed)
   - `created_at` (TIMESTAMP)
   - `updated_at` (TIMESTAMP)

3. **link_clicks**
   - `id` (UUID, primary key)
   - `short_link_id` (UUID, foreign key to short_links, indexed)
   - `referrer` (TEXT, nullable)
   - `user_agent` (TEXT, nullable)
   - `ip_address` (TEXT, nullable)
   - `country` (TEXT, nullable)
   - `city` (TEXT, nullable)
   - `device` (TEXT, nullable)
   - `browser` (TEXT, nullable)
   - `os` (TEXT, nullable)
   - `created_at` (TIMESTAMP, indexed)

## Testing Strategy

1. **Unit Tests**
   - Test individual functions and methods
   - Use table-driven tests for comprehensive coverage
   - Mock external dependencies

2. **Integration Tests**
   - Test interactions between components
   - Use testcontainers for PostgreSQL integration tests
   - Test database migrations

3. **API Tests**
   - Test HTTP endpoints with httptest package
   - Validate response status codes, headers, and body
   - Test authentication and authorization

4. **Load Tests**
   - Test performance under load using k6 or similar
   - Benchmark critical paths

5. **Test Coverage**
   - Aim for >80% code coverage
   - Focus on business logic coverage

## Deployment Strategy

1. **Development**: Docker Compose with hot reloading
2. **Staging/Production**: Coolify with proper resource allocation

## Error Handling Strategy

1. **Error Types**:
   - Define domain-specific error types
   - Use error wrapping (Go 1.13+) for context
   - Apply sentinel errors for expected conditions

2. **User-Facing Errors**:
   - Sanitize error messages sent to clients
   - Use consistent error response format
   - Include request ID for correlation

3. **Internal Errors**:
   - Log detailed error context for debugging
   - Include stack traces for unexpected errors
   - Monitor error rates and types

## Performance Considerations

1. **Connection Pooling**:
   - Configure optimal PostgreSQL connection pool size
   - Implement proper connection lifetime management

2. **Request Processing**:
   - Use context for request cancellation
   - Implement timeouts for all external calls
   - Apply concurrency for independent operations

3. **Resource Management**:
   - Properly close all resources (db connections, file handles)
   - Implement graceful shutdown for clean termination
   - Monitor resource usage

## Coolify Deployment Configuration

1. **Container Configuration**:
   - Multi-stage Docker build for smaller images
   - Non-root user for security
   - Proper health check configuration

2. **Environment Variables**:
   - Secret management for sensitive data
   - Environment-specific configurations

3. **Backup Strategy**:
   - Regular PostgreSQL database backups
   - Backup verification process

## Go-Specific Best Practices

1. **Standard Library Usage**:
   - Prefer standard library over third-party when possible
   - Use context package for cancellation and timeouts
   - Apply io interfaces for flexibility

2. **Error Handling**:
   - Use error wrapping with fmt.Errorf("... %w", err)
   - Apply consistent error checking patterns
   - Create meaningful custom errors

3. **Concurrency**:
   - Use goroutines judiciously
   - Apply proper synchronization with channels/mutexes
   - Implement context cancellation

4. **Code Organization**:
   - Follow clean architecture principles
   - Use interfaces for dependency inversion
   - Apply domain-driven design where beneficial

## Next Steps

1. Initialize the Go module with Go 1.22+
2. Set up the basic project structure following the outlined plan
3. Create the database schema with proper indexes
4. Implement the core URL shortening logic
5. Set up JWT token authentication
6. Refer to code_standard.md before writing any code