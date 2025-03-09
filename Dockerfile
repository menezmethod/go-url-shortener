# Build stage with extensive debugging
FROM golang:1.19 AS builder

# Set environment variable for verbose go command output
ENV GOFLAGS=-v
ENV GO111MODULE=on

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Print debug information
RUN pwd && ls -la && go env && go version

# Install git and other build essentials
RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Try to download dependencies with detailed errors
RUN go mod download -x || (echo "Go mod download failed with details:" && cat /root/.cache/go-build/log.txt && exit 1)

# Copy source code
COPY . .

# Build the application for ARM64 architecture (Raspberry Pi 5)
RUN go build -v -ldflags="-w -s" -o urlshortener ./cmd/server

# Simple runtime stage
FROM debian:bullseye-slim

# Install necessary packages
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates tzdata curl && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Create a non-root user to run the application
RUN useradd -m -s /bin/bash appuser

# Copy the binary from the builder stage
COPY --from=builder /app/urlshortener /app/urlshortener

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

# Environment variables
ENV PORT=8081 \
    ENVIRONMENT=production \
    READ_TIMEOUT=30s \
    WRITE_TIMEOUT=30s \
    IDLE_TIMEOUT=120s

# Run the application
ENTRYPOINT ["/app/urlshortener"]