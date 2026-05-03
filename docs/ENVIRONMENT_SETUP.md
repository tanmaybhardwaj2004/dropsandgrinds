# Environment Setup Guide

Complete guide for setting up the DropsAndGrinds development environment.

## Prerequisites

- **Go** 1.21 or later
- **Node.js** 18+ (for frontend development)
- **PostgreSQL** 15+
- **Redis** 7+
- **Docker** & Docker Compose (optional, for containerized setup)
- **Git**

## Quick Start with Docker

The fastest way to get started:

```bash
# Clone the repository
git clone <repository-url>
cd dropsandgrinds

# Copy environment template
cp .env.example .env

# Edit .env with your configuration
# At minimum, set:
# - JWT_SECRET (generate a random string)
# - STEAM_API_KEY (from https://steamcommunity.com/dev/apikey)

# Start all services
docker-compose up -d

# Run database migrations
docker-compose exec api go run cmd/migrate/main.go

# Seed initial data
docker-compose exec api go run cmd/seed/main.go
```

Access the application:
- Frontend: http://localhost:8080
- API: http://localhost:8080/api
- Swagger UI: http://localhost:8080/api/swagger/index.html

## Manual Setup

### 1. Database Setup

```bash
# Create PostgreSQL database
createdb dropsandgrinds

# Create user with privileges
createuser -P dropsandgrinds_user
# Grant privileges
psql -c "GRANT ALL PRIVILEGES ON DATABASE dropsandgrinds TO dropsandgrinds_user;"
```

### 2. Redis Setup

```bash
# Start Redis server
redis-server

# Or with Docker
docker run -d -p 6379:6379 --name redis redis:7-alpine
```

### 3. Backend Setup

```bash
# Install Go dependencies
go mod download

# Copy and configure environment
cp .env.example .env

# Required environment variables:
# DATABASE_URL=postgres://user:pass@localhost/dropsandgrinds?sslmode=disable
# REDIS_URL=redis://localhost:6379/0
# JWT_SECRET=your-secret-key-min-32-chars
# STEAM_API_KEY=your-steam-api-key

# Run migrations
go run cmd/migrate/main.go

# Seed data
go run cmd/seed/main.go

# Start the API server
go run cmd/api/main.go
```

### 4. Frontend Setup

```bash
cd frontend

# Serve with any static file server
# Option 1: Python
python -m http.server 3000

# Option 2: Node.js npx
npx serve -p 3000

# Option 3: PHP
php -S localhost:3000
```

## Environment Variables

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@localhost/db?sslmode=disable` |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379/0` |
| `JWT_SECRET` | Secret for JWT signing (min 32 chars) | `your-super-secret-key-here` |
| `STEAM_API_KEY` | Steam Web API key | `ABCD1234...` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | API server port | `8080` |
| `GIN_MODE` | Gin framework mode | `release` |
| `LOG_LEVEL` | Logging level | `info` |
| `RATE_LIMIT_RPS` | Rate limit per second | `10` |
| `RATE_LIMIT_BURST` | Rate limit burst | `20` |

## Getting API Keys

### Steam API Key

1. Visit https://steamcommunity.com/dev/apikey
2. Log in with your Steam account
3. Enter any domain name (e.g., `localhost`)
4. Copy the generated key

## Verification

After setup, verify everything works:

```bash
# Test API health
curl http://localhost:8080/api/health

# Test database connection
curl http://localhost:8080/api/deals?page=1&limit=5

# View API documentation
curl http://localhost:8080/api/swagger/index.html
```

## Troubleshooting

### Database connection errors
- Verify PostgreSQL is running: `pg_isready`
- Check DATABASE_URL format
- Ensure user has correct privileges

### Redis connection errors
- Verify Redis is running: `redis-cli ping`
- Check REDIS_URL format

### JWT errors
- Ensure JWT_SECRET is at least 32 characters
- Check for proper escaping in .env file

### CORS errors in frontend
- Verify API is running on correct port
- Check that GIN_MODE is set correctly
- Ensure no conflicting services on port 8080
