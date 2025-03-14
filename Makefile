.PHONY: run build test lint clean deps docker-build docker-run migrate-up migrate-down migrate-create setup install-tools test-ginkgo test-coverage test-focus test-v test-postman

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
	@if [ -f .env.test ]; then \
		echo "Using .env.test for testing environment..."; \
		export $$(grep -v '^#' .env.test | xargs) && $(GOTEST) -v ./...; \
	else \
		echo "Warning: .env.test not found. Using default test environment."; \
		$(GOTEST) -v ./...; \
	fi

# Run tests with verbose output
test-v:
	@echo "Running tests with verbose output..."
	@if [ -f .env.test ]; then \
		echo "Using .env.test for testing environment..."; \
		export $$(grep -v '^#' .env.test | xargs) && $(GOTEST) -v -count=1 ./...; \
	else \
		echo "Warning: .env.test not found. Using default test environment."; \
		$(GOTEST) -v -count=1 ./...; \
	fi

# Run tests with Ginkgo
test-ginkgo:
	@echo "Running tests with Ginkgo..."
	@if [ -f .env.test ]; then \
		echo "Using .env.test for testing environment..."; \
		export $$(grep -v '^#' .env.test | xargs) && ~/go/bin/ginkgo -r -v ./...; \
	else \
		echo "Warning: .env.test not found. Using default test environment."; \
		~/go/bin/ginkgo -r -v ./...; \
	fi

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@if [ -f .env.test ]; then \
		echo "Using .env.test for testing environment..."; \
		export $$(grep -v '^#' .env.test | xargs) && ~/go/bin/ginkgo -r -v --cover --coverprofile=coverage.out ./...; \
	else \
		echo "Warning: .env.test not found. Using default test environment."; \
		~/go/bin/ginkgo -r -v --cover --coverprofile=coverage.out ./...; \
	fi
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out

# Run focused tests
test-focus:
	@echo "Running focused tests..."
	@if [ -f .env.test ]; then \
		echo "Using .env.test for testing environment..."; \
		export $$(grep -v '^#' .env.test | xargs) && ~/go/bin/ginkgo -r -v --focus="$(FOCUS)" ./...; \
	else \
		echo "Warning: .env.test not found. Using default test environment."; \
		~/go/bin/ginkgo -r -v --focus="$(FOCUS)" ./...; \
	fi

# Run Postman tests
test-postman:
	@echo "Running Postman tests..."
	@if ! command -v newman &> /dev/null; then \
		echo "Newman not found. Installing..."; \
		npm install -g newman newman-reporter-htmlextra; \
	fi
	@echo "Creating test environment file for Newman..."
	@if [ -f .env.test ]; then \
		echo "Using .env.test for test environment..."; \
		export $$(grep -v '^#' .env.test | xargs); \
	else \
		echo "Warning: .env.test not found. Using default environment variables."; \
	fi
	@echo '{' > newman-env.json
	@echo '  "name": "URL_Shortener_API_Local_Environment",' >> newman-env.json
	@echo '  "values": [' >> newman-env.json
	@echo '    {' >> newman-env.json
	@echo '      "key": "baseUrl",' >> newman-env.json
	@echo '      "value": "http://localhost:8081",' >> newman-env.json
	@echo '      "enabled": true' >> newman-env.json
	@echo '    },' >> newman-env.json
	@echo '    {' >> newman-env.json
	@echo '      "key": "apiPath",' >> newman-env.json
	@echo '      "value": "/api",' >> newman-env.json
	@echo '      "enabled": true' >> newman-env.json
	@echo '    },' >> newman-env.json
	@echo '    {' >> newman-env.json
	@echo '      "key": "masterPassword",' >> newman-env.json
	@echo '      "value": "$(MASTER_PASSWORD)",' >> newman-env.json
	@echo '      "enabled": true' >> newman-env.json
	@echo '    }' >> newman-env.json
	@echo '  ]' >> newman-env.json
	@echo '}' >> newman-env.json
	@echo "Running Postman collection with Newman..."
	@newman run ./postman/collections/master_collection.json -e newman-env.json --reporters cli,htmlextra --reporter-htmlextra-export postman-results.html
	@echo "Test results saved to postman-results.html"

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
	@go install github.com/onsi/ginkgo/v2/ginkgo@latest
	@npm install -g newman newman-reporter-htmlextra 