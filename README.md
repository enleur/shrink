# Shrink - Service Showcase

This is a showcase project for a Go service implementing cloud-native practices, using a link shortener as the example application.

## Features and Technology Stack

- Go with Gin http router
- OpenAPI specification with oapi-codegen
- Redis for URL storage
- Zap for structured json logging
- OpenTelemetry/Jaeger for distributed tracing
- Testcontainers for integration testing
- Environment-based configuration
- ko for multi-platform Docker images
- GitHub Actions for CI 
- GitHub Packages publishing


## Getting Started

1. Clone the repository:
   ```
   gh repo clone enleur/shrink
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Set up environment variables (see `config.go` for required variables).

4. Generate API-related code:
   ```
   go generate ./...
   ```

5. Run the service:
   ```
   go run cmd/server/main.go
   ```

6. The service will be available at `http://localhost:8080` (or the configured port).

## API Endpoints

The API is defined using OpenAPI specification. The main endpoints are:

- `POST /shorten`: Shorten a URL
- `GET /{shortCode}`: Redirect to the original URL

API-related code is generated using `go generate` with oapi-codegen.

## Running Tests

To run the test suite:

```bash
go test -v ./...
```

Note: Docker is required for integration tests using test containers.