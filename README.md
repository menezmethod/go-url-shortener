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

## Running Postman Tests

This project includes a comprehensive Postman collection for API testing. You can run these tests either locally or as part of the CI/CD pipeline.

### Running Tests Locally

1. Make sure you have the Docker Compose environment running:
   ```bash
   make run
   ```
   This will start the application and PostgreSQL in Docker containers.

2. Install Newman (the Postman CLI) if you don't have it already:
   ```bash
   npm install -g newman newman-reporter-htmlextra
   ```

3. Run the tests using the Makefile command:
   ```bash
   make test-postman
   ```
   This will:
   - Check if Docker Compose is running and start it if needed
   - Create a Newman environment file with the correct API credentials
   - Run the Postman collection with Newman
   - Generate an HTML report in the project root called `postman-results.html`

4. To stop the Docker Compose services when you're done:
   ```bash
   make docker-compose-down
   ```

### Troubleshooting

If you encounter issues with the Postman tests:

1. Check container status:
   ```bash
   make docker-compose-status
   ```

2. Inspect the container logs:
   ```bash
   docker compose logs
   ```

3. Ensure the master password in your `.env.dev` file matches what's being used in the Postman collection

### Running Tests in CI/CD Pipeline

The Postman tests are automatically run as part of the GitHub Actions workflow when you push to main or develop branches, or create a pull request to these branches.

1. Tests are defined in the `.github/workflows/postman-tests.yml` file
2. The workflow:
   - Sets up Docker Compose with the appropriate environment variables
   - Runs the Postman collection using Newman
   - Generates an HTML report and uploads it as an artifact

3. You can view the test results in the GitHub Actions tab of your repository.

4. To run the workflow manually, go to the Actions tab in your GitHub repository and select "Postman API Tests" from the workflows list, then click "Run workflow".

For more details on the Postman collections, see the [Postman README](postman/README.md).

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
