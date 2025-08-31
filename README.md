# Crypto Price Service

A Go-based cryptocurrency price tracking service that fetches prices from multiple exchanges and provides REST API endpoints.

## Configuration

### Environment Variables

The application uses environment variables for configuration. Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

Required environment variables:

- `MONGO_HOST`: MongoDB connection URI
- `MONGO_USER`: MongoDB username (optional if using connection URI)
- `MONGO_PASSWORD`: MongoDB password (optional if using connection URI)
- `REDIS_HOST`: Redis server address
- `SENTRY_DSN`: Sentry DSN for error monitoring (optional)

### Configuration Priority

1. Environment variables (highest priority)
2. Configuration file (`pkg/config/env.yml`)
3. Default values (lowest priority)

## Running the Application

### Development
```bash
go run cmd/main.go
```

### Production
Set environment variables and run:
```bash
export MONGO_HOST="mongodb://your-host:27017"
export MONGO_USER="your-username"
export MONGO_PASSWORD="your-password"
export REDIS_HOST="your-redis-host:6379"
export SENTRY_DSN="your-sentry-dsn"

go run cmd/main.go
```

## Security Notes

- Never commit `.env` files to version control
- Use strong passwords for database connections
- Consider using Docker secrets or Kubernetes secrets in production
- The application supports both direct MongoDB connection URIs and separate user/password environment variables

## API Endpoints

- `GET /price`: Get cryptocurrency prices
- `GET /metrics`: Prometheus metrics
