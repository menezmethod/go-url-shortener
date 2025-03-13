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
  - [Current Status Summary](#current-status-summary)
    - [Well-Tested Components](#well-tested-components)
    - [Partially Tested Components](#partially-tested-components)
    - [Untested Components](#untested-components)
    - [Key Achievements](#key-achievements)
    - [Challenges](#challenges)
  - [Next Steps](#next-steps)
  - [Troubleshooting](#troubleshooting)
    - [Fixing Test Failures](#fixing-test-failures)
    - [Running Tests Properly](#running-tests-properly)
    - [Generating Detailed Coverage Reports](#generating-detailed-coverage-reports)

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
- [x] Test core packages:
  - [x] internal/config
  - [x] internal/logger
  - [ ] internal/models
  - [x] internal/domain

### Phase 2: Data Layer (Week 2)

- [x] Test data access packages:
  - [ ] internal/db
  - [ ] internal/redis
  - [ ] internal/cache
  - [x] internal/repository

### Phase 3: Business Logic (Week 3)

- [x] Test business logic packages:
  - [x] internal/service
  - [ ] internal/auth
  - [ ] internal/metrics

### Phase 4: API and Integration (Week 4)

- [x] Test API components:
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

We've created the following test helpers:

1. `internal/testutils/` package with utility functions for testing
2. `internal/testutils/mocks/` package containing:
   - `db_mock.go` - Mock implementations for database operations
   - `sql_mock.go` - Mock implementations for SQL result and rows
   - `repository_mock.go` - Mock implementations for repositories
3. `internal/common/interfaces.go` - Common interfaces for testing

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

| Component   | Status     | Coverage % | Notes                                         |
|-------------|------------|------------|-----------------------------------------------|
| config      | ✅ Complete | 83.8%      | Config loading tests with environment variables |
| logger      | ✅ Complete | 93.3%      | Logger initialization tests                   |
| models      | ⏳ Pending  | -          | -                                             |
| domain      | ✅ Complete | 100.0%     | Common domain errors and models               |
| db          | ⏳ Pending  | 0.0%       | -                                             |
| redis       | ⏳ Pending  | -          | -                                             |
| cache       | ✅ Complete | 100.0%     | In-memory cache implementation with concurrency tests |
| repository  | ✅ Complete | 67.2%      | Link repository with DB mocks                 |
| service     | ✅ Complete | 74.6%      | Link service, URL shortener, and cached service with mocks |
| auth        | ⏳ Pending  | 0.0%       | -                                             |
| metrics     | ⏳ Pending  | 0.0%       | -                                             |
| middleware  | ✅ Complete | 96.6%      | Comprehensive tests for all middleware components including integration tests |
| handlers    | 🔄 In Progress | 38.6%   | Link handler with service mocks, mock handler implementation at 100% |
| router      | ⏳ Pending  | 0.0%       | -                                             |
| integration | ⏳ Pending  | -          | End-to-end tests not started                  |

Legend:
- ✅ Complete
- 🔄 In Progress
- ⏳ Pending
- ❌ Blocked

## Current Status Summary

Overall, our test coverage has improved to approximately **43.6%** across the entire codebase. Here's a breakdown of our current testing status:

### Well-Tested Components
- **Cache**: 100% coverage - Complete implementation with concurrency tests
- **Domain**: 100% coverage - All domain models and errors tested
- **Middleware**: 96.6% coverage - Comprehensive tests for all middleware components
- **Logger**: 93.3% coverage - Core functionality tested
- **Config**: 83.8% coverage - Configuration loading and validation tested
- **Service**: 74.6% coverage - Core service implementations tested
- **Repository**: 67.2% coverage - Core database operations tested
- **Handlers**: 38.6% coverage - Mock handler implementation at 100%, main handler implementation needs work

### Partially Tested Components
- **Handlers**: 38.6% coverage
  - Mock handler implementation: 100% coverage
  - Main handler implementation: 0% coverage (needs work)
  - All test cases implemented but need to improve actual handler coverage

### Untested Components
- **Database**: 0% coverage
- **Redis**: Not started
- **Auth**: 0% coverage
- **Metrics**: 0% coverage
- **Router**: 0% coverage
- **Models**: Not started

### Key Achievements
- Successfully implemented testing for complex components using mocks
- Created a common interfaces package to improve testability
- Established a pattern for BDD-style tests with Ginkgo and Gomega
- Fixed type compatibility issues between production and test code
- Achieved 100% coverage for cache implementation with concurrency tests
- Consolidated service tests into a single file for better organization
- Added comprehensive tests for URL shortener and cached services
- **Added extensive middleware testing with 96.6% coverage**:
  - Implemented integration tests for all middleware components
  - Added stress tests for RateLimiter under high concurrency
  - Enhanced security header and CORS testing
  - Improved metrics collection testing

### Challenges
- Ensuring consistency between mock implementations and real components
- Achieving high coverage for complex database operations
- Testing error scenarios that are difficult to reproduce
- Managing test suite organization with multiple services
- Handling asynchronous operations in tests
- **Testing concurrent behavior in rate limiting and timeout middleware**

## Next Steps

1. **Focus on Handler Implementation Coverage**:
   - Implement the actual handler methods that currently have 0% coverage
   - Maintain the existing test cases while improving the implementation
   - Target achieving at least 80% coverage for the handler package

2. **Complete tests for remaining components**:
   - internal/models
   - internal/db package
   - internal/redis packages
   - internal/auth package
   - internal/metrics package
   - internal/api/router package

3. **Create integration tests**:
   - End-to-end API flow tests
   - Database integration tests
   - Redis integration tests

4. **Set up CI/CD integration**:
   - GitHub Actions workflow
   - Coverage reporting
   - Coverage thresholds

5. **Fix Common test issues**:
   - Update mock implementations as needed
   - Handle database connections in tests
   - Mock external dependencies
   - Improve async operation testing

6. **Run coverage reports**:
   - Generate coverage reports with `ginkgo -r -v --cover`
   - Identify uncovered code paths
   - Add tests to increase coverage

## Troubleshooting

### Fixing Test Failures

Initial test runs revealed some discrepancies between our test assumptions and the actual implementation. Here are some tips for fixing common test issues:

1. **Structure Mismatches**: Ensure your test is consistent with the actual field names and structure of the packages you're testing. For example, our initial config test had field name mismatches.

2. **Environment Setup**: Some tests need proper environment setup to work. For example, config tests may require certain environment variables to be set.

3. **Mock Completeness**: Make sure your mocks implement all required methods of the interface they're mocking.

4. **Interface Compatibility**: When using interfaces, ensure that your mock implementations are compatible with the actual interfaces used by the code.

5. **Ginkgo Command Not Found**: If the `ginkgo` command is not found, ensure that `$GOPATH/bin` is in your PATH or use the full path to the ginkgo executable.

### Running Tests Properly

- Run all Ginkgo tests with coverage:
  ```bash
  go install github.com/onsi/ginkgo/v2/ginkgo@latest
  $GOPATH/bin/ginkgo -r -v --cover ./...
  ```

- Run standard Go tests for specific packages:
  ```bash
  go test -v ./internal/config
  ```

- Generate and view coverage report:
  ```bash
  $GOPATH/bin/ginkgo -r -v --cover --coverprofile=coverage.out ./...
  go tool cover -html=coverage.out -o coverage.html
  open coverage.html
  ```

### Generating Detailed Coverage Reports

To get a detailed view of your test coverage, you can generate visual HTML reports:

1. Generate the coverage data:
   ```bash
   go test -coverprofile=coverage.out ./internal/...
   ```

2. Convert the coverage data to an HTML report:
   ```bash
   go tool cover -html=coverage.out -o coverage.html
   ```

3. Open the report in your browser:
   ```bash
   open coverage.html
   ```

The HTML report will show each file with coverage highlighting:
- Green: Covered lines
- Red: Uncovered lines
- Gray: Non-executable lines (comments, imports, etc.)

For a quick summary without visualization, use:
```bash
go tool cover -func=coverage.out
```

This command provides a function-by-function breakdown of coverage percentages.

This document will be updated regularly as we make progress on implementing test coverage. 