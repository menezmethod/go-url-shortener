# URL Shortener Service

A high-performance, feature-rich URL shortening service built with Go, featuring caching, analytics, and a robust API.

![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)
![Postgres Version](https://img.shields.io/badge/PostgreSQL-17-336791?style=flat&logo=postgresql)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

## Features

- **URL Shortening**: Create short URLs with optional custom aliases
- **Analytics**: Track visits, referrers, user agents, and other metrics
- **Caching**: High-performance caching for frequently accessed URLs
- **Authentication**: Secure API access with JWT authentication
- **Swagger Documentation**: Interactive API documentation
- **Docker Support**: Ready to deploy with Docker and docker-compose
- **Database Migrations**: Automated database setup and migrations
- **Rate Limiting**: Configurable rate limiting to prevent abuse
- **Metrics**: Prometheus-compatible metrics endpoint
- **Configurable URL Expiration**: Set custom expiration dates for links

## Technology Stack

- **Backend**: Go (Gin framework)
- **Database**: PostgreSQL
- **Cache**: In-memory cache with configurable TTL
- **Documentation**: Swagger/OpenAPI
- **Containerization**: Docker & docker-compose
- **Logging**: Structured logging with Zap

## Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/url-shortener.git
   cd url-shortener
   ```

2. Configure environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your desired configuration
   ```

3. Start the service:
   ```bash
   docker-compose up -d
   ```

4. Access the service:
   - API: http://localhost:8081/api
   - Swagger UI: http://localhost:8081/swagger/index.html

### Manual Setup

1. Ensure Go 1.19+ and PostgreSQL 17 are installed

2. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/url-shortener.git
   cd url-shortener
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Configure environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your desired configuration
   ```

5. Start the PostgreSQL database

6. Run the application:
   ```bash
   go run cmd/server/main.go
   ```

## API Usage

### Authentication

Get an authentication token:

```bash
curl -X POST http://localhost:8081/api/auth/token \
  -H "Content-Type: application/json" \
  -d '{"master_password": "your_master_password"}'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Create a Short URL

```bash
curl -X POST http://localhost:8081/api/links \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "url": "https://example.com/very/long/url/that/needs/shortening",
    "custom_alias": "mylink",
    "expiration_date": "2023-12-31T23:59:59Z"
  }'
```

> **Note**: Use `custom_alias` to specify a custom code for your short URL. If not provided, the system will generate a random code.

Response:
```json
{
  "id": "5f3a4bc2-1234-5678-abcd-1234567890ab",
  "code": "mylink",
  "url_id": "15ebcec2-d0a9-4902-af30-e9b1f5645a2b",
  "expiration_date": "2023-12-31T23:59:59Z",
  "is_active": true,
  "created_at": "2023-01-15T14:30:15Z",
  "updated_at": "2023-01-15T14:30:15Z",
  "url": {
    "id": "15ebcec2-d0a9-4902-af30-e9b1f5645a2b",
    "original_url": "https://example.com/very/long/url/that/needs/shortening",
    "hash": "5a4519f4d7a77547b2adc9801e8e8241a1c3c2b2d53ef4709affee318ee4fdca",
    "created_at": "2023-01-15T14:30:15Z",
    "updated_at": "2023-01-15T14:30:15Z"
  }
}
```

### Retrieve Link Details

```bash
curl -X GET http://localhost:8081/api/links/mylink \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### List All Links

```bash
curl -X GET "http://localhost:8081/api/links?page=1&page_size=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Get Link Statistics

```bash
curl -X GET http://localhost:8081/api/links/mylink/stats \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Update a Link

```bash
curl -X PUT http://localhost:8081/api/links/mylink \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "is_active": false,
    "expiration_date": "2024-06-30T23:59:59Z"
  }'
```

### Delete a Link

```bash
curl -X DELETE http://localhost:8081/api/links/mylink \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Known Issues and Troubleshooting

### URL Redirection

The redirect functionality works best with custom aliases. System-generated codes may occasionally fail to redirect properly. If you encounter issues with a system-generated code, consider creating a new short URL with a custom alias.

### Swagger Documentation

If the Swagger documentation page appears blank, it may need to be regenerated using:

```bash
# Install swag if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/server/main.go
```

### Metrics

The metrics endpoint at `/metrics` provides insights into service operation. Note that the redirect counter (`url_shortener_redirects_total`) may not always accurately reflect the actual number of redirects performed.

## Configuration Options

The service can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `ENVIRONMENT` | Running environment (development/production) | `production` |
| `APP_PORT` | Port the application runs on | `8081` |
| `POSTGRES_HOST` | PostgreSQL host | `postgres` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | PostgreSQL username | `postgres` |
| `POSTGRES_PASSWORD` | PostgreSQL password | *(required)* |
| `POSTGRES_DB` | PostgreSQL database name | `url_shortener` |
| `MASTER_PASSWORD` | Password for admin authentication | *(required)* |
| `JWT_SECRET` | Secret for JWT token generation | *(required)* |
| `LOG_LEVEL` | Logging level (debug/info/notice/warn/error) | `notice` |
| `RATE_LIMIT_REQUESTS` | Number of requests allowed per window | `60` |
| `RATE_LIMIT_WINDOW` | Time window for rate limiting (seconds) | `60` |
| `READ_TIMEOUT` | HTTP server read timeout | `30s` |
| `WRITE_TIMEOUT` | HTTP server write timeout | `30s` |
| `IDLE_TIMEOUT` | HTTP server idle timeout | `120s` |
| `POSTGRES_MAX_CONNECTIONS` | Max DB connections | `25` |
| `POSTGRES_MAX_IDLE_CONNECTIONS` | Max idle DB connections | `5` |
| `POSTGRES_CONN_MAX_LIFETIME` | DB connection lifetime | `15m` |
| `SHORTLINK_DEFAULT_EXPIRY` | Default expiration for links | `30d` |

## Testing Progress

The project uses Ginkgo and Gomega for BDD-style testing. Current test coverage status:

### Overall Coverage: 43.6%

#### Well-Tested Components
- Cache System: 100% coverage
- Domain Logic: 100% coverage
- Middleware: 96.6% coverage
- Logger: 93.3% coverage
- Configuration: 83.8% coverage
- Service Layer: 74.6% coverage
- Repository Layer: 67.2% coverage
- Handlers: 38.6% coverage (Mock implementations at 100%)

#### Components in Progress
- API Handlers: Comprehensive test cases implemented, improving implementation coverage
- Repository Layer: Enhancing coverage for complex operations

#### Pending Components
- Database Layer
- Redis Integration
- Authentication System
- Metrics Collection

### Test Environment Setup

To run tests locally, you need to set up the test environment:

1. Copy the test environment template:
   ```bash
   cp .env.test.example .env.test
   ```

2. Edit `.env.test` with your secure test credentials:
   ```
   TEST_POSTGRES_PASSWORD=your_secure_password_here
   TEST_MASTER_PASSWORD=your_secure_master_password_here
   TEST_JWT_SECRET=your_secure_jwt_secret_here
   ```

3. Run the tests:
   ```bash
   make test
   # or for verbose output
   make test-v
   ```

> **Security Note**: Never commit `.env.test` to version control. It's already included in `.gitignore`.

## Integration with Your Applications

### Direct API Integration

Integrate with any programming language using standard HTTP requests:

**JavaScript Example**:
```javascript
// Creating a short URL
async function createShortUrl(longUrl) {
  const response = await fetch('http://localhost:8081/api/links', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer YOUR_TOKEN'
    },
    body: JSON.stringify({ url: longUrl })
  });
  return await response.json();
}

// Using the created short URL
createShortUrl('https://example.com/very/long/url')
  .then(data => {
    console.log('Short URL:', data.short_url);
  });
```

**Python Example**:
```python
import requests

def create_short_url(long_url, token):
    response = requests.post(
        'http://localhost:8081/api/links',
        headers={
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {token}'
        },
        json={'url': long_url}
    )
    return response.json()

# Get your token
auth_response = requests.post(
    'http://localhost:8081/api/auth/token',
    json={'master_password': 'your_master_password'}
)
token = auth_response.json()['token']

# Create short URL
result = create_short_url('https://example.com/very/long/url', token)
print(f"Short URL: {result['short_url']}")
```

### Embedding in Web Applications

To add a URL shortening feature to your website:

```html
<form id="url-shortener">
  <input type="url" id="long-url" placeholder="Enter a long URL" required>
  <button type="submit">Shorten</button>
  <div id="result"></div>
</form>

<script>
  document.getElementById('url-shortener').addEventListener('submit', async (e) => {
    e.preventDefault();
    const longUrl = document.getElementById('long-url').value;
    
    try {
      // Assuming you handle authentication separately
      const token = localStorage.getItem('auth_token');
      
      const response = await fetch('http://localhost:8081/api/links', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ url: longUrl })
      });
      
      const data = await response.json();
      
      if (response.ok) {
        document.getElementById('result').innerHTML = `
          <p>Short URL: <a href="${data.short_url}" target="_blank">${data.short_url}</a></p>
        `;
      } else {
        document.getElementById('result').innerHTML = `
          <p>Error: ${data.error || 'Failed to create short URL'}</p>
        `;
      }
    } catch (err) {
      document.getElementById('result').innerHTML = `
        <p>Error: ${err.message}</p>
      `;
    }
  });
</script>
```

## Monitoring and Metrics

Access Prometheus-compatible metrics:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8081/metrics
```

Available metrics include:
- Request counts and response times
- Error rates
- Cache hit/miss ratios
- Link usage statistics

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/cache

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./internal/cache
```

### Generating Swagger Documentation

```bash
# Install swag if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/server/main.go
```

## Deployment Options

### Docker Compose

The easiest way to deploy is using the provided docker-compose.yml:

```bash
docker-compose up -d
```

### Kubernetes

For Kubernetes deployment, sample configuration files are provided in the `k8s/` directory.

```bash
# Apply configurations
kubectl apply -f k8s/

# Verify deployment
kubectl get pods
```

### Cloud Providers

The service can be deployed to any cloud platform that supports Docker containers or Go applications, including:

- AWS ECS/EKS
- Google Cloud Run/GKE
- Azure App Service/AKS
- Digital Ocean App Platform

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
