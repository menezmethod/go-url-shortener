# Ginkgo and Gomega Test Coverage Guide

## Overview

This document outlines the plan to achieve 100% test coverage for the ref_go project using Ginkgo and Gomega. This includes setting up the testing framework, writing tests for each component, and integrating test coverage reporting.

## Table of Contents

- [Ginkgo and Gomega Test Coverage Guide](#ginkgo-and-gomega-test-coverage-guide)
	- [Overview](#overview)
	- [Table of Contents](#table-of-contents)
	- [Introduction to Ginkgo and Gomega](#introduction-to-ginkgo-and-gomega)
	- [Test Coverage Strategy](#test-coverage-strategy)
	- [Implementation Plan](#implementation-plan)
		- [Phase 1: Setup and Core Components (Week 1)](#phase-1-setup-and-core-components-week-1)
		- [Phase 2: Data Layer (Week 2)](#phase-2-data-layer-week-2)
		- [Phase 3: Business Logic (Week 3)](#phase-3-business-logic-week-3)
		- [Phase 4: API and Integration (Week 4)](#phase-4-api-and-integration-week-4)
	- [Testing Tools and Setup](#testing-tools-and-setup)
		- [Initial Setup](#initial-setup)
		- [Makefile Updates](#makefile-updates)
		- [Test Helpers](#test-helpers)
	- [Writing Tests](#writing-tests)
		- [Basic Test Structure](#basic-test-structure)
		- [Test Categories](#test-categories)
		- [Testing with Mocks](#testing-with-mocks)
	- [Integration with CI/CD](#integration-with-cicd)
	- [Progress Tracking](#progress-tracking)
	- [Next Steps](#next-steps)

## Introduction to Ginkgo and Gomega

[Ginkgo](https://onsi.github.io/ginkgo/) is a BDD-style Go testing framework built to help you efficiently write expressive and comprehensive tests. [Gomega](https://onsi.github.io/gomega/) is a matcher/assertion library that works seamlessly with Ginkgo.

Key features:
- Structured, nested testing for easy organization
- Built-in support for asynchronous testing
- Comprehensive test reporting and output formatting
- Parallel test execution
- Focus and skip capabilities for targeted testing

## Test Coverage Strategy

We aim to achieve 100% test coverage by systematically addressing each component in our codebase. The strategy includes:

1. **Foundational Components First**: Start by testing the core infrastructure components like config, logger, etc.
2. **Building Upward**: Move to higher-level components like repositories and services
3. **External Interfaces Last**: Finally, test handlers and API integrations

## Implementation Plan

### Phase 1: Setup and Core Components (Week 1)

- [x] Set up Ginkgo and Gomega
- [x] Create test infrastructure (mocks, test helpers)
- [ ] Test core packages:
  - [x] internal/config
  - [x] internal/logger
  - [ ] internal/models
  - [ ] internal/domain

### Phase 2: Data Layer (Week 2)

- [ ] Test data access packages:
  - [ ] internal/db
  - [ ] internal/redis
  - [ ] internal/cache
  - [x] internal/repository

### Phase 3: Business Logic (Week 3)

- [ ] Test business logic packages:
  - [x] internal/service
  - [ ] internal/auth
  - [ ] internal/metrics

### Phase 4: API and Integration (Week 4)

- [ ] Test API components:
  - [x] internal/api/middleware
  - [x] internal/api/handlers
  - [ ] internal/api/router
- [ ] Create integration tests for end-to-end flows

## Testing Tools and Setup

### Initial Setup

```bash
# Add Ginkgo and Gomega to dependencies
go get github.com/onsi/ginkgo/v2
go get github.com/onsi/gomega

# Install the ginkgo CLI
go install github.com/onsi/ginkgo/v2/ginkgo

# Bootstrap a test suite in a package
cd internal/config
ginkgo bootstrap
ginkgo generate config
```

### Makefile Updates

The following targets have been added to the Makefile:

```makefile
# Run tests with Ginkgo
test-ginkgo:
	@echo "Running tests with Ginkgo..."
	@ginkgo -r -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@ginkgo -r -v --cover --coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out

# Run focused tests
test-focus:
	@echo "Running focused tests..."
	@ginkgo -r -v --focus="$(FOCUS)" ./...
```

### Test Helpers

We've created a test helpers package at `internal/testutils` that includes:
- Mock implementations for database interactions
- Mock implementations for repositories
- Test utility functions for environment setup

## Writing Tests

### Basic Test Structure

```go
package config_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	
	"github.com/menezmethod/ref_go/internal/config"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", func() {
	Describe("LoadConfig", func() {
		Context("with valid environment variables", func() {
			It("loads the configuration correctly", func() {
				// Test implementation
				cfg, err := config.LoadConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).NotTo(BeNil())
			})
		})
		
		Context("with missing environment variables", func() {
			It("returns an error", func() {
				// Test implementation
			})
		})
	})
})
```

### Test Categories

1. **Unit Tests**: Test individual functions and methods in isolation
   - Example: `internal/config/config_test.go`
   
2. **Integration Tests**: Test interactions between components
   - Example: `internal/repository/link_repository_test.go`
   
3. **End-to-End Tests**: Test complete user flows
   - Example: `internal/api/handlers/link_handler_test.go`

### Testing with Mocks

For dependencies that are difficult to test (database, external APIs), use mocks:

```go
// Define a mock repository
type MockRepository struct {
	GetUserFunc func(id string) (*models.User, error)
}

func (m *MockRepository) GetUser(id string) (*models.User, error) {
	return m.GetUserFunc(id)
}

// Use the mock in tests
var _ = Describe("UserService", func() {
	var (
		mockRepo *MockRepository
		service  *service.UserService
	)
	
	BeforeEach(func() {
		mockRepo = &MockRepository{}
		service = service.NewUserService(mockRepo)
	})
	
	Describe("GetUser", func() {
		It("returns a user when found", func() {
			mockRepo.GetUserFunc = func(id string) (*models.User, error) {
				return &models.User{ID: id, Name: "Test User"}, nil
			}
			
			user, err := service.GetUser("123")
			
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Name).To(Equal("Test User"))
		})
	})
})
```

## Integration with CI/CD

Add the following steps to your CI/CD pipeline:

1. Run tests with coverage reporting
2. Fail the build if coverage drops below a threshold
3. Archive test reports and coverage data

Example GitHub Actions workflow:

```yaml
name: Test Coverage

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.23
          
      - name: Install Ginkgo
        run: go install github.com/onsi/ginkgo/v2/ginkgo@latest
        
      - name: Run tests with coverage
        run: |
          ginkgo -r -v --cover --coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
          
      - name: Check coverage threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Test coverage is below 80%: $COVERAGE%"
            exit 1
          fi
          
      - name: Upload coverage report
        uses: actions/upload-artifact@v2
        with:
          name: coverage-report
          path: coverage.html
```

## Progress Tracking

| Component | Status | Coverage % | Notes |
|-----------|--------|------------|-------|
| config    | ✅ Complete | - | Basic config loading tests |
| logger    | ✅ Complete | - | Logger initialization tests |
| models    | ⏳ Pending | - | - |
| domain    | ⏳ Pending | - | - |
| db        | ⏳ Pending | - | - |
| redis     | ⏳ Pending | - | - |
| cache     | ⏳ Pending | - | - |
| repository| ✅ Complete | - | Link repository tests with mocks |
| service   | ✅ Complete | - | Link service tests with mocks |
| auth      | ⏳ Pending | - | - |
| metrics   | ⏳ Pending | - | - |
| middleware| ✅ Complete | - | Auth middleware tests |
| handlers  | ✅ Complete | - | Link handler tests |
| router    | ⏳ Pending | - | - |

Legend:
- ✅ Complete
- 🔄 In Progress
- ⏳ Pending
- ❌ Blocked

## Next Steps

1. Complete tests for remaining components
2. Set up CI/CD integration
3. Add integration tests for critical user flows
4. Implement coverage reporting
5. Address any uncovered code paths

This document will be updated regularly as we make progress on implementing test coverage. 