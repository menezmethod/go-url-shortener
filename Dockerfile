# Build stage
FROM golang:1.23.4-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache ca-certificates git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with production optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o urlshortener ./cmd/server

# Install migrate tool for migrations
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Final stage
FROM alpine:latest

# Install certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata curl && \
    update-ca-certificates

# Create a non-root user to run the application
RUN adduser -D -g '' appuser

# Copy the binary from the builder stage
COPY --from=builder /app/urlshortener /app/urlshortener
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Copy migrations folder
COPY migrations /app/migrations

# Set the ownership of the application to appuser
RUN chown -R appuser:appuser /app

# Use the non-root user
USER appuser

# Set working directory
WORKDIR /app

# Expose the application port
EXPOSE 8081

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8081/api/health || exit 1

# Environment variables will be provided by docker-compose or Coolify
# Only set defaults for non-sensitive information
ENV PORT=8081 \
    ENVIRONMENT=production \
    READ_TIMEOUT=30s \
    WRITE_TIMEOUT=30s \
    IDLE_TIMEOUT=120s \
    POSTGRES_MAX_CONNECTIONS=25 \
    POSTGRES_MAX_IDLE_CONNECTIONS=5 \
    POSTGRES_CONN_MAX_LIFETIME=15m \
    SHORTLINK_DEFAULT_EXPIRY=30d

# Run the application
ENTRYPOINT ["/app/urlshortener"] 