# Go Coding Standards

This document defines our project's coding standards, derived from code_tips.md and aligned with idiomatic Go practices. All code must follow these guidelines.

## Code Organization

### Project Structure

We follow the standard Go project layout:

```
ref_go/
├── cmd/                   # Main applications
├── internal/              # Private application code
│   ├── api/               # HTTP API handlers
│   ├── domain/            # Domain models
│   ├── service/           # Business logic 
│   └── repository/        # Data access
├── pkg/                   # Public libraries
└── configs/               # Configuration files
```

### Package Guidelines

- Package names should be short, lowercase, and descriptive
- Avoid package name collisions with standard libraries
- Group related functionality within packages
- Name packages after what they provide, not what they contain

```go
// Good
package validator

// Bad
package util
```

### Imports

- Group imports as standard library, external dependencies, and internal packages
- Use aliases when necessary to avoid naming collisions

```go
import (
    "context"
    "fmt"
    "net/http"
    
    "github.com/pkg/errors"
    "go.uber.org/zap"
    
    "github.com/menezmethod/ref_go/internal/domain"
)
```

## Coding Conventions

### Naming

- Use camelCase for private variables, functions, and methods
- Use PascalCase for exported (public) variables, functions, types and methods
- Be descriptive with names - clarity trumps brevity
- Use consistent variable names across related functions

```go
// Good
var userID string
type UserService struct {}
func (s *UserService) GetByID(id string) (*User, error) {}

// Bad
var uid string
type userservice struct {}
func (s *userservice) get_by_id(id string) (*User, error) {}
```

### Functions

- Keep functions small and focused on a single responsibility
- Return early for error conditions to avoid deep nesting
- Provide context in error messages
- Use named return values when they add clarity

```go
// Good
func GetUser(ctx context.Context, id string) (*User, error) {
    if id == "" {
        return nil, errors.New("user ID cannot be empty")
    }
    
    user, err := repository.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("finding user by ID: %w", err)
    }
    
    return user, nil
}
```

### Comments

- All exported functions, types, and variables must have comments
- Write comments in complete sentences with proper punctuation
- Use godoc-compatible comments for public API
- Comment complex logic and non-obvious behaviors

```go
// UserService provides methods for managing users.
type UserService struct {
    // ... fields
}

// GetByID retrieves a user by their unique identifier.
// Returns an error if the user cannot be found.
func (s *UserService) GetByID(id string) (*User, error) {
    // ...
}
```

## Error Handling

### Error Creation

- Use errors.New for simple errors
- Use fmt.Errorf with %w for wrapping errors
- Create custom error types for specific error conditions
- Use sentinel errors for expected error conditions

```go
// Sentinel errors
var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidInput = errors.New("invalid input")
)

// Custom error type
type ValidationError struct {
    Field string
    Error string
}

// Error wrapping
if err != nil {
    return fmt.Errorf("validating user data: %w", err)
}
```

### Error Handling

- Always check errors
- Handle each error at the appropriate level
- Don't use panic for normal error handling
- Log errors with context for debugging
- Only return sanitized errors to clients

```go
user, err := service.GetUser(ctx, id)
if err != nil {
    if errors.Is(err, ErrUserNotFound) {
        return nil, http.StatusNotFound, ErrUserNotFound
    }
    logger.Error("failed to get user", zap.String("id", id), zap.Error(err))
    return nil, http.StatusInternalServerError, errors.New("internal server error")
}
```

## Concurrency Patterns

### Goroutines

- Always ensure goroutines can exit properly
- Use context for cancellation
- Don't leak goroutines
- Pass variables explicitly to avoid closure-related issues

```go
func ProcessURLs(ctx context.Context, urls []string) error {
    errCh := make(chan error, len(urls))
    
    for _, url := range urls {
        go func(u string) {
            select {
            case errCh <- processURL(ctx, u):
            case <-ctx.Done():
                errCh <- ctx.Err()
            }
        }(url) // Pass url explicitly
    }
    
    // Collect results
    // ...
}
```

### Channels

- Use buffered channels when appropriate
- Close channels only from the sender, never from the receiver
- Check for closed channels with comma-ok syntax
- Consider using sync.WaitGroup for fan-out/fan-in patterns

```go
func ProcessItems(items []Item) error {
    results := make(chan Result, len(items))
    var wg sync.WaitGroup
    
    for _, item := range items {
        wg.Add(1)
        go func(i Item) {
            defer wg.Done()
            result, err := processItem(i)
            if err != nil {
                results <- Result{Error: err}
                return
            }
            results <- Result{Value: result}
        }(item)
    }
    
    // Close the channel when all goroutines are done
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Process results
    for result := range results {
        // Handle result
    }
    
    return nil
}
```

### Context Usage

- Pass context as the first parameter to functions
- Use context for cancellation and timeouts
- Don't store contexts in structs
- Propagate context through function calls

```go
func (s *Service) ProcessRequest(ctx context.Context, req Request) (Response, error) {
    // Create a timeout if needed
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // Use the context for downstream calls
    result, err := s.repository.Find(ctx, req.ID)
    if err != nil {
        return Response{}, err
    }
    
    // Check for cancellation
    select {
    case <-ctx.Done():
        return Response{}, ctx.Err()
    default:
        // Continue processing
    }
    
    // ...
}
```

## Type Definitions

### Struct Design

- Group related fields together
- Order fields for optimal memory alignment
- Use embedding judiciously
- Include field tags for JSON/DB when needed

```go
type User struct {
    ID        string    `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    Name      string    `json:"name" db:"name"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

### Interfaces

- Keep interfaces small and focused
- Define interfaces at the point of use
- Use composable interfaces

```go
// Repository defines the methods required for data access.
type Repository interface {
    Find(ctx context.Context, id string) (*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}
```

### Enums

- Use iota for creating enumerated constants
- Use typed constants for better type safety
- Use bitmasks for flags

```go
type Role int

const (
    RoleUser Role = iota + 1
    RoleAdmin
    RoleSystem
)

type Permission int

const (
    PermRead Permission = 1 << iota
    PermWrite
    PermDelete
    
    PermReadWrite = PermRead | PermWrite
    PermAll      = PermRead | PermWrite | PermDelete
)
```

## Testing

### Unit Tests

- Use table-driven tests
- Test both happy paths and error conditions
- Mock external dependencies
- Keep tests independent

```go
func TestUserService_GetByID(t *testing.T) {
    tests := []struct {
        name    string
        id      string
        want    *User
        wantErr bool
    }{
        {
            name: "valid ID",
            id:   "valid-id",
            want: &User{ID: "valid-id", Name: "John Doe"},
        },
        {
            name:    "empty ID",
            id:      "",
            wantErr: true,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks, dependencies...
            
            got, err := service.GetByID(tt.id)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("GetByID() got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

- Use docker-compose for integration tests
- Clean up resources after tests
- Use t.Parallel() for independent tests
- Set up test data in a repeatable way

## HTTP Handler Guidelines

### Request Handling

- Validate all input
- Use middleware for common concerns (logging, auth, etc.)
- Return appropriate status codes
- Provide consistent error responses

```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.renderError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    
    if err := validate(req); err != nil {
        h.renderError(w, http.StatusBadRequest, err.Error())
        return
    }
    
    user, err := h.service.CreateUser(r.Context(), req.ToModel())
    if err != nil {
        h.handleError(w, err)
        return
    }
    
    h.renderJSON(w, http.StatusCreated, user)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, ErrInvalidInput):
        h.renderError(w, http.StatusBadRequest, err.Error())
    case errors.Is(err, ErrUserNotFound):
        h.renderError(w, http.StatusNotFound, "user not found")
    default:
        h.logger.Error("internal error", zap.Error(err))
        h.renderError(w, http.StatusInternalServerError, "internal server error")
    }
}
```

### Response Handling

- Use consistent JSON structure
- Include helpful metadata (pagination, etc.)
- Set appropriate headers
- Provide informative error messages

```go
type Response struct {
    Data  interface{} `json:"data,omitempty"`
    Error string      `json:"error,omitempty"`
    Meta  *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    Total int `json:"total"`
    Page  int `json:"page"`
    Size  int `json:"size"`
}

func (h *Handler) renderJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    
    resp := Response{Data: data}
    if err := json.NewEncoder(w).Encode(resp); err != nil {
        h.logger.Error("failed to encode response", zap.Error(err))
    }
}

func (h *Handler) renderError(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    
    resp := Response{Error: message}
    if err := json.NewEncoder(w).Encode(resp); err != nil {
        h.logger.Error("failed to encode error response", zap.Error(err))
    }
}
```

## Database Access

### Connection Management

- Use connection pooling
- Set appropriate timeouts
- Configure max connections based on load
- Close resources properly

```go
func NewDB(cfg Config) (*sql.DB, error) {
    db, err := sql.Open("postgres", cfg.DSN)
    if err != nil {
        return nil, fmt.Errorf("opening database connection: %w", err)
    }
    
    db.SetMaxOpenConns(cfg.MaxConnections)
    db.SetMaxIdleConns(cfg.MaxIdleConnections)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("pinging database: %w", err)
    }
    
    return db, nil
}
```

### Query Execution

- Use prepared statements for repeated queries
- Use transactions for multi-step operations
- Use proper parameter binding to prevent SQL injection
- Set context with timeout for queries

```go
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
    query := `
        INSERT INTO users (id, email, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `
    
    _, err := r.db.ExecContext(
        ctx,
        query,
        user.ID,
        user.Email,
        user.Name,
        user.CreatedAt,
        user.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("inserting user: %w", err)
    }
    
    return nil
}

func (r *Repository) TransferFunds(ctx context.Context, fromID, toID string, amount float64) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("beginning transaction: %w", err)
    }
    defer tx.Rollback()
    
    // Execute multiple queries within the transaction
    
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("committing transaction: %w", err)
    }
    
    return nil
}
```

## Reference

For more detailed guidance, refer to:

1. [code_tips.md](./code_tips.md)
2. [Effective Go](https://go.dev/doc/effective_go)
3. [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
4. [Standard Go Project Layout](https://github.com/golang-standards/project-layout)