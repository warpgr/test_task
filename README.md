# Rates Service

A Go-based microservice for fetching and providing Grinex exchange rates.

## Features
- **Fetch Rates**: Fetches USDT rates from Grinex.
- **Persistence**: Stores all fetched snapshots in PostgreSQL.
- **Observability**:
  - Structured logging with **Zap**.
  - Distributed tracing with **OpenTelemetry**.
  - Metrics collection with **Prometheus**.
- **Graceful Shutdown**: Handles OS signals for clean termination.

- **Monitoring & Tracing**:
  - `SERVICE_NAME`: Name of the service (default: `rates-service`).
  - `OTLP_ENDPOINT`: OpenTelemetry collector endpoint (e.g., `otel-collector:4317`). If empty, traces are exported to stdout.

## Getting Started

### 1. Build the application
```bash
make build
```

### 2. Configure environment
Copy the example environment file:
```bash
cp example.env config.env
```

### 3. Run with Docker Compose
```bash
docker-compose up -d
```

### 3. Run the application
```bash
./app
```

## Testing Client
A standalone gRPC client is included to simulate periodic requests (every 5 seconds) and verify tracing/metrics.

### Build the client
```bash
make client
```

### Run the client
```bash
./testing-client
```

## API
- **GRPC**: Default port `5001`.
- **Health Check**: `http://localhost:8080/healthcheck`
- **Prometheus Metrics**: `http://localhost:8080/metrics`

## Testing
Run unit and integration tests:
```bash
make test
```

## Linting
Run golangci-lint:
```bash
make lint
```
