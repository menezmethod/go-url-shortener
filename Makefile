.PHONY: run build test lint clean deps docker-build docker-run migrate-up migrate-down migrate-create setup install-tools

# Build variables
BINARY_NAME=urlshortener
BUILD_DIR=./build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Docker parameters
DOCKER_IMAGE=urlshortener
DOCKER_TAG=latest

# Migration parameters
MIGRATE=migrate
MIGRATION_DIR=./migrations/postgres
DATABASE_URL=postgres://postgres:postgres@localhost:5432/url_shortener?sslmode=disable

# Default target
all: lint test build

# Setup development environment
setup:
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then cp .env.example .env && echo "Created .env file from .env.example"; else echo ".env file already exists"; fi
	@$(GOMOD) tidy

# Build the application
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

# Run the application
run:
	@$(GORUN) ./cmd/server

# Test the application
test:
	@echo "Testing..."
	@$(GOTEST) -v ./...

# Lint the code
lint:
	@echo "Linting..."
	@golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) tidy

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Docker run
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p 8081:8081 \
		-e POSTGRES_HOST=host.docker.internal \
		-e POSTGRES_PORT=5432 \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=url_shortener \
		-e MASTER_PASSWORD=development_master_password \
		-e BASE_URL=http://localhost:8081 \
		-e ENVIRONMENT=development \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker compose up
docker-compose-up:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d

# Docker compose down
docker-compose-down:
	@echo "Stopping services with Docker Compose..."
	@docker-compose down

# Run database migrations
migrate-up:
	@echo "Running migrations..."
	@$(MIGRATE) -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" up

# Rollback database migrations
migrate-down:
	@echo "Rolling back migrations..."
	@$(MIGRATE) -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" down

# Create a new migration
migrate-create:
	@echo "Creating migration..."
	@read -p "Enter migration name: " name; \
	$(MIGRATE) create -ext sql -dir $(MIGRATION_DIR) -seq $$name 

# Install required tools
install-tools:
	@echo "Installing required development tools..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest 