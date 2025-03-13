# URL Shortener API

A RESTful API for shortening URLs with comprehensive analytics.

## Table of Contents

- [Features](#features)
- [API Documentation](#api-documentation)
- [Getting Started](#getting-started)
- [Testing](#testing)
  - [Unit and Integration Tests](#unit-and-integration-tests)
  - [Postman Tests](#postman-tests)
- [Authentication](#authentication)
- [Usage Examples](#usage-examples)

## Features

- Create short URLs with optional custom aliases
- Set expiration dates for links
- Track detailed analytics including browser, device, OS, and geographic information
- Secure API access with JWT authentication
- RESTful API with comprehensive documentation

## API Documentation

The API is documented using the OpenAPI/Swagger specification. When the server is running, you can access the Swagger UI at:

```
http://localhost:8081/swagger/index.html
```

## Getting Started

### Prerequisites

- Go (version 1.15 or later)
- PostgreSQL database

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/url-shortener.git
   cd url-shortener
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Configure environment variables:
   ```bash
   cp .env.example .env
   # Edit .env file with your configuration
   ```

4. Run the database migrations:
   ```bash
   go run cmd/migrate/main.go
   ```

5. Start the server:
   ```bash
   go run cmd/api/main.go
   ```

## Testing

### Unit and Integration Tests

Run the unit and integration tests with:

```bash
go test ./...
```

For more verbose output:

```bash
go test -v ./...
```

### Postman Tests

This project includes a comprehensive suite of Postman tests to validate API functionality. The tests cover all endpoints and include:

- Pre-request scripts for dynamic data generation
- Post-request tests with detailed assertions
- Environment variables for seamless test flow
- Visualizations for analytics data

#### Setting Up Postman Tests

1. Import the collections and environment from the `postman/` directory into Postman
2. Configure the environment variables, especially `masterPassword`
3. Run the collections in order:
   - First, run the auth collection to get a valid token
   - Then run the link operations collections

For detailed instructions, see the [Postman README](postman/README.md).

#### Running Automated Tests with Newman

```bash
# Install Newman
npm install -g newman

# Run the master collection
newman run ./postman/collections/URL_Shortener_API_Master.json -e ./postman/environments/URL_Shortener_API_Environment.json
```

## Authentication

The API uses JWT-based authentication:

1. Request a token with your master password:
   ```
   POST /api/auth/token
   ```

2. Use the token in subsequent requests as a Bearer token:
   ```
   Authorization: Bearer your_jwt_token
   ```

## Usage Examples

### Create a Short Link

```bash
curl -X POST "http://localhost:8081/api/links" \
  -H "Authorization: Bearer your_jwt_token" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/very-long-url-that-needs-shortening",
    "custom_alias": "my-link",
    "expiration_date": "2023-12-31T23:59:59Z"
  }'
```

### Get Link Statistics

```bash
curl -X GET "http://localhost:8081/api/links/my-link/stats" \
  -H "Authorization: Bearer your_jwt_token"
```

See the API documentation for more examples and details on all available endpoints.
