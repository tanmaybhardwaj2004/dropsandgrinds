# DropsAndGrinds 🎮💰

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-4169E1?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D?style=flat&logo=redis&logoColor=white)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Required-2496ED?style=flat&logo=docker&logoColor=white)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**Smart Cross-Platform Game Deal Tracker for Indian Gamers**

DropsAndGrinds is an India-focused game deal aggregator that tracks prices across 15+ platforms, provides multi-source review scores, and offers intelligent bundle analysis. Built with Go, PostgreSQL, Redis, and deployed on AWS infrastructure.

---

## 📋 Table of Contents

- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [System Architecture](#-system-architecture)
- [Prerequisites](#-prerequisites)
- [Quick Start](#-quick-start)
- [Detailed Setup](#-detailed-setup)
- [Configuration](#-configuration)
- [API Documentation](#-api-documentation)
- [Development Guide](#-development-guide)
- [Deployment](#-deployment)
- [Troubleshooting](#-troubleshooting)
- [Contributing](#-contributing)

---

## ✨ Features

### 🎯 Core Features

| Feature | Description | Status |
|---------|-------------|--------|
| **Cross-Platform Deal Tracking** | Track game prices across Steam, Epic Games Store, GOG, and 12+ other platforms | ✅ Complete |
| **India-Focused Pricing** | Accurate Indian Rupee (₹) pricing with regional store support | ✅ Complete |
| **Price History & Analytics** | Historical price charts and all-time low tracking | ✅ Complete |
| **Smart Deal Evaluation** | AI-powered deal scoring based on price history and review scores | ✅ Complete |
| **Bundle Breaker** | Analyze if game bundles actually save you money | ✅ Complete |
| **Savings Dashboard** | Track gaming expenses, total savings, and spending patterns | ✅ Complete |

### 📚 Library & Collection Features

| Feature | Description | Status |
|---------|-------------|--------|
| **Steam Library Import** | Import your Steam library to track owned games | ✅ Complete |
| **Game Library Management** | Track game statuses: Want to buy, Purchased, Playing, Completed, Dropped | ✅ Complete |
| **Wishlist with Alerts** | Get notified when games hit your target price | ✅ Complete |
| **Buy Timing Predictions** | ML-based predictions for optimal purchase timing | ✅ Complete |

### 🔍 Search & Discovery

| Feature | Description | Status |
|---------|-------------|--------|
| **Meilisearch Integration** | Fast, typo-tolerant full-text search | ✅ Complete |
| **Multi-Source Reviews** | Aggregated scores from Metacritic, OpenCritic, Steam | ✅ Complete |
| **Sale Calendar** | Track ongoing and upcoming platform sales | ✅ Complete |
| **Deal Recommendations** | Personalized "Deals for You" based on library | ✅ Complete |

### 🔐 Authentication & User Features

| Feature | Description | Status |
|---------|-------------|--------|
| **JWT Authentication** | Secure token-based authentication | ✅ Complete |
| **OAuth Login** | Sign in with Google and Steam | ✅ Complete |
| **User Profiles** | Personalized dashboard and preferences | ✅ Complete |
| **PWA Support** | Install as Progressive Web App with offline support | ✅ Complete |

### 📊 Analytics & Monitoring

| Feature | Description | Status |
|---------|-------------|--------|
| **Analytics Dashboard** | Track user behavior and platform metrics | ✅ Complete |
| **File-Based Logging** | Comprehensive logging for debugging and monitoring | ✅ Complete |
| **Health Checks** | Service health monitoring endpoints | ✅ Complete |
| **Metrics Collection** | Prometheus-compatible metrics | ✅ Complete |

---

## 🛠 Tech Stack

### Backend
- **Go 1.21+** - Primary backend language
- **Gin Framework** - HTTP web framework
- **pgx** - PostgreSQL driver
- **go-redis** - Redis client
- **JWT-Go** - Authentication tokens
- **Swaggo** - API documentation

### Database & Cache
- **PostgreSQL 15+** - Primary database with full-text search
- **Redis 7+** - Session storage and caching
- **Meilisearch** - Advanced search engine

### Frontend
- **HTML5/CSS3** - Semantic markup and modern styling
- **Vanilla JavaScript** - No framework dependencies
- **Progressive Web App** - Service workers, offline support
- **Responsive Design** - Mobile-first approach

### Infrastructure & DevOps
- **Docker & Docker Compose** - Containerization
- **AWS** - EC2, RDS, ElastiCache, CloudFront, ALB
- **GitHub Actions** - CI/CD pipeline
- **Nginx** - Reverse proxy and static file serving

### External APIs
- **Steam API** - Library import and game data
- **RAWG API** - Game metadata and reviews
- **IGDB API** - Additional game information

---

## 🏗 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Client Layer                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │  Browser │  │  Mobile  │  │   PWA    │  │  Steam   │     │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘     │
└────────────────────┬────────────────────────────────────────┘
                     │ HTTPS
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                      CDN Layer (CloudFront)                 │
│                   Static Assets + Caching                   │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                    Load Balancer (ALB)                      │
│              SSL Termination + Routing                      │
└────────────────────┬────────────────────────────────────────┘
                     │
         ┌───────────┴───────────┐
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │    Backend      │
│  (Nginx:80)     │    │   (Go:8080)     │
│  Static Files   │    │   REST API      │
└─────────────────┘    └────────┬────────┘
                                │
              ┌─────────────────┼─────────────────┐
              │                 │                 │
              ▼                 ▼                 ▼
       ┌──────────┐      ┌─────────────┐      ┌───────────┐
       │PostgreSQL│      │  Redis      │      │Meilisearch│
       │  (RDS)   │      │(ElastiCache)|      │ (Search)  │
       └──────────┘      └─────────────┘      └───────────┘
```

---

## 📦 Prerequisites

Before you begin, ensure you have the following installed:

### Required Software

| Software | Version | Purpose | Download |
|----------|---------|---------|----------|
| **Docker** | 24.0+ | Container runtime | [Get Docker](https://docs.docker.com/get-docker/) |
| **Docker Compose** | 2.20+ | Multi-container orchestration | Included with Docker Desktop |
| **Git** | 2.40+ | Version control | [Download Git](https://git-scm.com/downloads) |

### Optional (for development)

| Software | Version | Purpose | Download |
|----------|---------|---------|----------|
| **Go** | 1.21+ | Backend development | [Download Go](https://golang.org/dl/) |
| **VS Code** | Latest | Recommended IDE | [Download VS Code](https://code.visualstudio.com/) |
| **Make** | - | Build automation | Included with Git Bash (Windows) |
| **curl** | - | API testing | Included with Git Bash |

### VS Code Extensions (Recommended)

- **Go** - Go language support by Google
- **Docker** - Docker management
- **PostgreSQL** - Database tools
- **REST Client** - API testing
- **Markdown All in One** - Documentation

---

## 🚀 Quick Start

### Option 1: Docker Compose (Recommended for Beginners)

The fastest way to get started with all services running:

```bash
# Step 1: Clone the repository 
git clone https://github.com/tanmaybhardwaj2004/dropsandgrinds.git

# Step 2: Navigate to project directory
cd dropsandgrinds

# Step 3: Create environment file
cp .env.example .env

# Step 4: Start all services with Docker Compose
docker-compose up -d

# Step 5: Verify services are running
docker-compose ps

# Step 6: Access the application
open http://localhost:80          # Frontend
open http://localhost:8080/health # Backend API Health Check
```

### Option 2: Local Development (Advanced)

For backend/frontend development with hot reloading:

```bash
# Step 1: Clone and navigate
git clone https://github.com/tanmaybhardwaj2004/dropsandgrinds.git
cd dropsandgrinds

# Step 2: Start infrastructure services only
docker-compose up -d postgres redis meilisearch

# Step 3: Set up Go environment
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Step 4: Install Go dependencies
go mod download
go mod tidy

# Step 5: Copy and configure environment
cp .env.example .env
# Edit .env with your local settings

# Step 6: Run database migrations
migrate -path migrations -database "$DATABASE_URL" up

# Step 7: Start the backend server
go run cmd/server/main.go

# Step 8: In a new terminal, serve frontend
# Using Python (built-in)
cd frontend && python -m http.server 8081

# OR using Node.js npx
npx serve frontend -p 8081

# Step 9: Access application
open http://localhost:8081  # Frontend
open http://localhost:8080  # Backend API
```

---

## 🔧 Detailed Setup

### 1. Environment Configuration

Create a `.env` file in the project root:

```bash
# Copy example environment file
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# ============================================
# Database Configuration
# ============================================
DATABASE_URL=postgres://postgres:your_password@localhost:5432/dropsandgrinds?sslmode=disable
DATABASE_READ_REPLICA_URL=postgres://postgres:your_password@localhost:5433/dropsandgrinds?sslmode=disable

POSTGRES_DB=dropsandgrinds
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_secure_password_here

# ============================================
# Redis Configuration
# ============================================
REDIS_URL=redis://localhost:6379

# ============================================
# JWT Configuration
# ============================================
JWT_SECRET=your_jwt_secret_here_change_in_production
JWT_EXPIRY=24h

# ============================================
# OAuth Configuration
# ============================================
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
STEAM_API_KEY=your_steam_api_key

# ============================================
# Meilisearch Configuration
# ============================================
MEILISEARCH_URL=http://localhost:7700
MEILISEARCH_API_KEY=your_meilisearch_key

# ============================================
# API Keys
# ============================================
RAWG_API_KEY=your_rawg_api_key
IGDB_CLIENT_ID=your_igdb_client_id
IGDB_CLIENT_SECRET=your_igdb_client_secret

# ============================================
# Logging Configuration
# ============================================
LOG_DIR=logs
LOG_FORMAT=json
LOG_LEVEL=info

# ============================================
# Application Configuration
# ============================================
PORT=8080
FRONTEND_URL=http://localhost:80
ENV=development
```

### 2. Docker Services

The project uses Docker Compose to run infrastructure services:

```yaml
# Services included in docker-compose.yml
- postgres:5432      # PostgreSQL database
- redis:6379        # Redis cache
- meilisearch:7700  # Search engine
- backend:8080      # Go backend API
- frontend:80       # Static file server
```

### 3. Database Setup

```bash
# Start database container
docker-compose up -d postgres

# Wait for PostgreSQL to be ready (30 seconds)
sleep 30

# Run migrations
migrate -path migrations -database "$DATABASE_URL" up

# Seed initial data (optional)
psql $DATABASE_URL -f seed/games.sql
```

### 4. Verify Installation

```bash
# Check all containers are running
docker-compose ps

# Test backend health
curl http://localhost:8080/health

# Test database connectivity
curl http://localhost:8080/health/deps

# Test frontend
curl -I http://localhost:80
```

Expected responses:
- `/health` - `{"status":"healthy"}`
- `/health/deps` - Database and Redis status

---

## ⚙ Configuration

### Backend Configuration

Configuration is loaded from environment variables via `config/config.go`:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Backend server port |
| `DATABASE_URL` | - | PostgreSQL connection string |
| `REDIS_URL` | - | Redis connection string |
| `JWT_SECRET` | - | Secret key for JWT signing |
| `LOG_LEVEL` | info | Logging level (debug/info/warn/error) |
| `LOG_FORMAT` | text | Log format (text/json) |

### Frontend Configuration

Frontend settings in `frontend/js/config.js`:

```javascript
const CONFIG = {
    API_BASE_URL: 'http://localhost:8080',
    WS_URL: 'ws://localhost:8080/ws',
    DEFAULT_THEME: 'dark',
    ITEMS_PER_PAGE: 20
};
```

### Docker Configuration

Key settings in `docker-compose.yml`:

| Service | Port | Volume |
|---------|------|--------|
| postgres | 5432 | ./data/postgres |
| redis | 6379 | ./data/redis |
| meilisearch | 7700 | ./data/meilisearch |
| backend | 8080 | ./logs/backend |
| frontend | 80 | ./frontend |

---

## 📚 API Documentation

### Interactive Documentation

- **Swagger UI**: http://localhost:8080/swagger/
- **OpenAPI Spec**: http://localhost:8080/swagger/doc.json

### API Endpoints

#### Health & Monitoring
```
GET  /health              # Service health check
GET  /health/deps         # Database/Redis health
GET  /metrics             # Prometheus metrics
```

#### Authentication
```
POST /api/auth/register   # User registration
POST /api/auth/login      # User login
POST /api/auth/refresh    # Refresh JWT token
POST /api/auth/logout     # User logout
GET  /auth/google         # Google OAuth login
GET  /auth/steam          # Steam OAuth login
```

#### Games & Catalog
```
GET  /api/games           # List all games
GET  /api/games/search    # Search games (Meilisearch)
GET  /api/games/:id       # Game details
GET  /api/games/:id/price # Price history
GET  /api/deals           # Current deals
GET  /api/deals/for-you   # Personalized deals
GET  /api/sales/active    # Active sales
GET  /api/sales/calendar  # Sale calendar
```

#### User Features
```
GET    /api/me              # Current user profile
GET    /api/wishlist        # User wishlist
POST   /api/wishlist        # Add to wishlist
DELETE /api/wishlist/:id    # Remove from wishlist
GET    /api/library         # User library
POST   /api/library/import  # Import Steam library
GET    /api/savings         # Savings dashboard
POST   /api/savings/purchase # Log purchase
```

#### Analytics & Tools
```
POST /api/bundles/analyze   # Analyze bundle value
GET  /api/games/:id/buy-timing # Best time to buy prediction
POST /api/analytics/events  # Track analytics events
```

---

## 💻 Development Guide

### Project Structure

```
dropsandgrinds/
├── cmd/
│   └── server/           # Application entry point
│       ├── main.go       # Server startup
│       └── router.go     # HTTP routing
├── config/               # Configuration management
│   ├── config.go         # Env/config loading
│   └── db.go             # Database connections
├── internal/
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # Auth, logging, rate limiting
│   └── models/           # Data models
├── pkg/
│   ├── logger/           # Logging utilities
│   └── utils/            # Helper functions
├── frontend/             # Static frontend files
│   ├── index.html        # Main page
│   ├── css/              # Stylesheets
│   ├── js/               # JavaScript modules
│   └── images/           # Static assets
├── migrations/           # Database migrations
├── docs/                 # Documentation
├── tests/                # Test files
├── docker-compose.yml    # Docker services
├── Dockerfile            # Backend container
└── README.md             # This file
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -v ./tests -run TestHealth

# Run smoke tests
./tests/smoke_test.sh
```

### Code Generation

```bash
# Generate Swagger documentation
swag init -g cmd/server/main.go

# Generate Go mocks (if needed)
go generate ./...
```

### Database Migrations

```bash
# Create new migration
migrate create -ext sql -dir migrations -seq add_user_preferences

# Run migrations up
migrate -path migrations -database "$DATABASE_URL" up

# Rollback migration
migrate -path migrations -database "$DATABASE_URL" down 1

# Force version (use with caution)
migrate -path migrations -database "$DATABASE_URL" force VERSION
```

---

## 🚀 Deployment

### Production Deployment (AWS)

1. **Infrastructure Setup**
   - Follow [docs/AWS_ALB_SETUP.md](docs/AWS_ALB_SETUP.md)
   - Configure RDS PostgreSQL per [docs/AWS_RDS_SETUP.md](docs/AWS_RDS_SETUP.md)
   - Set up ElastiCache Redis per [docs/AWS_ELASTICACHE_SETUP.md](docs/AWS_ELASTICACHE_SETUP.md)

2. **Environment Setup**
   ```bash
   # Set production environment variables
   export ENV=production
   export DATABASE_URL=postgres://...
   export REDIS_URL=redis://...
   ```

3. **Deploy with GitHub Actions**
   - Push to `main` branch triggers automatic deployment
   - Or manually trigger workflow from GitHub Actions tab

### Docker Production Build

```bash
# Build production image
docker build -t dropsandgrinds:latest .

# Run production container
docker run -d \
  -p 8080:8080 \
  -e DATABASE_URL=postgres://... \
  -e REDIS_URL=redis://... \
  -v $(pwd)/logs:/app/logs \
  dropsandgrinds:latest
```

---

## 🐛 Troubleshooting

### Common Issues

#### Docker Container Won't Start

```bash
# Check container logs
docker-compose logs backend
docker-compose logs postgres

# Restart services
docker-compose restart

# Rebuild containers
docker-compose down
docker-compose up -d --build
```

#### Database Connection Failed

```bash
# Verify PostgreSQL is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Reset database (WARNING: destroys data)
docker-compose down -v
docker-compose up -d postgres
migrate -path migrations -database "$DATABASE_URL" up
```

#### Redis Connection Failed

```bash
# Verify Redis is running
docker-compose ps redis
docker-compose logs redis

# Test Redis connection
docker-compose exec redis redis-cli ping
```

#### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>

# Or use different port in .env
PORT=8081
```

#### Frontend Not Loading

```bash
# Check Nginx configuration
docker-compose logs frontend

# Verify frontend files exist
ls -la frontend/index.html

# Rebuild frontend
docker-compose up -d --build frontend
```

### Getting Help

1. Check [docs/](docs/) directory for detailed guides
2. Review [GitHub Issues](https://github.com/tanmaybhardwaj2004/dropsandgrinds/issues)
3. Check logs: `docker-compose logs -f`

---

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/your-feature`
3. **Make your changes**
4. **Run tests**: `go test ./...`
5. **Commit**: `git commit -m "Add your feature"`
6. **Push**: `git push origin feature/your-feature`
7. **Open a Pull Request**

### Development Guidelines

- Follow Go conventions (gofmt, golint)
- Add tests for new features
- Update documentation
- Ensure Docker Compose works for all changes

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- **RAWG API** for game metadata
- **Steam API** for library import
- **Meilisearch** for search functionality
- **Contributors** who helped build this project

---

**Built with ❤️ for Indian gamers who love a good deal!**