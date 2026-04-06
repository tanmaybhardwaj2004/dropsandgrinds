# DropsAndGrinds — Full Stack Project Document
### Smart Cross-Platform Game Deal Tracker (India-Focused)
**Last Updated: March 24, 2026**

---

## The Project

A production-grade game deal tracking platform — price comparison, deal alerts, Steam library scanning, bundle analysis, India-specific pricing, and multi-source review aggregation. Built right, deployable, CV-worthy.

Similar to: IsThereAnyDeal / GG.deals — but yours, with India-first features and a cleaner product vision.

**Key Differentiators:**
- Steam Library Scanner → hide owned games, flag missing DLCs
- India Arbitrage → Steam India vs Global + GST breakdown + UPI filter
- Bundle Breaker → is this bundle worth it vs buying separately?
- Multi-Source Review Score → weighted aggregate from Metacritic, OpenCritic, Steam user reviews, IGN, GameSpot (no single-platform bias)
- "Saved This Year" dashboard with savings graph
- "Best Time to Buy" calendar with historical sale predictions
- Platform redirect system with click analytics
- Personalised "Deals For You" feed based on wishlist + click history
- Beautiful cover grid with hover cards (price, % drop, historical low, review score)
- Live sale awareness (e.g., Steam Spring Sale 2026 — ending March 26)

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go (Golang) |
| Database | PostgreSQL (AWS RDS / local Docker) |
| Cache | Redis (AWS ElastiCache / local Docker) |
| Auth | JWT + refresh tokens → OAuth (Google/Steam) later |
| Frontend | HTML/CSS/JS → React/Next.js later |
| Infrastructure | AWS EC2 + ALB + S3 + CloudFront |
| DevOps | Docker + Docker Compose + GitHub Actions |
| API Docs | Swagger / OpenAPI 3.0 (swaggo/swag for Go) |
| Testing | Go testing package + testify + httptest |
| Security Scanning | Trivy (in CI/CD pipeline) |
| Monitoring | Prometheus metrics + Sentry (production errors) |
| Price Data | CheapShark API, GG.deals, IsThereAnyDeal feeds |
| Game Metadata | RAWG, IGDB, Steam Store API |
| Review Scores | Metacritic, OpenCritic, Steam reviews, IGN, GameSpot |
| User Library | Steam Web API (GetOwnedGames) |
| Bundle Data | Humble Bundle, Fanatical (light scraping) |

---

## System Architecture

```
Internet
    ↓
CloudFront (CDN)
    ↓
AWS ALB (Load Balancer)  ←── /health checks
    ↓
EC2 — Go App (Docker)  ←→  Redis (rate limiting + cache)
    ↓
RDS PostgreSQL (price history, users, deals, wishlists, review scores)
    ↓  (backend only — never frontend)
External APIs: CheapShark, GG.deals, Steam, RAWG, IGDB,
               Metacritic, OpenCritic, IGN, GameSpot
```

**Container Network:**
```
docker-compose.yml
├── frontend   (nginx, static files)  → port 80  [PUBLIC]
├── backend    (Go app)               → port 8080 [INTERNAL ONLY]
├── postgres   (PostgreSQL)           → port 5432 [INTERNAL ONLY]
└── redis      (Redis)                → port 6379 [INTERNAL ONLY]
All on: dropsandgrinds_network
```

---

## Separation of Logic

**Rules — never break these:**
- Frontend NEVER touches the DB directly. Ever.
- Backend NEVER renders HTML. Ever.
- DB is NOT exposed outside Docker network. Ever.
- External APIs (Steam, CheapShark, review sites, etc.) called ONLY from backend.
- Frontend and backend only talk through defined API contracts (JSON).

| Layer | Responsibility |
|---|---|
| Frontend | Rendering deal grid, wishlist, library view, hover cards, filters, redirect |
| Backend | Business logic, auth, DB queries, external API calls, analytics logging, review aggregation |
| Database | Storing and retrieving data — nothing else |

---

## Multi-Source Review Score System

To avoid bias from any single platform, DropsAndGrinds aggregates review scores from multiple sources with configurable weights. Each score is normalised to a 0–100 scale before weighting.

**Sources and default weights:**

| Source | Type | Weight | Notes |
|---|---|---|---|
| Metacritic | Critic aggregate | 25% | Industry standard, wide coverage |
| OpenCritic | Critic aggregate | 25% | Broader critic pool, no paid influence |
| Steam User Reviews | Player aggregate | 30% | Largest player base, most recent sentiment |
| IGN | Single critic outlet | 10% | Popular reference, offset by others |
| GameSpot | Single critic outlet | 10% | Long-standing, good genre coverage |

**Weighted Score Formula:**
```
ReviewScore = (Metacritic × 0.25) + (OpenCritic × 0.25) +
              (SteamUserScore × 0.30) + (IGN × 0.10) + (GameSpot × 0.10)
```

**Normalisation rules:**
- Metacritic: already 0–100
- OpenCritic: already 0–100
- Steam: convert "% positive" directly (85% positive → 85)
- IGN/GameSpot: convert from 10-point scale (8.5/10 → 85)
- If a source is unavailable for a game, redistribute its weight proportionally among available sources
- Minimum 2 sources required to display a score — otherwise show "Not enough reviews"

**Where it appears:**
- Deal grid hover card (badge: score + source count)
- Game detail page (full breakdown per source with links)
- "Best Time to Buy" recommendation (higher-scored games weighted slightly more in predictions)

**Caching:** Review scores cached in PostgreSQL with a 24-hour TTL. Re-fetched via background cron job to avoid stale scores during major review updates.

---

## Platform Redirect System

```
User clicks "View on Steam" button
    ↓
GET /api/games/{id}/redirect?platform=steam
    ↓
Backend:
  - Looks up store URL for game + platform
  - Logs: user_id, game_id, platform, timestamp (analytics)
  - Returns: { "url": "https://store.steampowered.com/app/..." }
    ↓
Frontend: window.location.href = url
    ↓
User lands on platform store page
```

**Why through the backend:**
- Track which platforms users click most
- Validate the URL is still live
- Rate limit abuse
- Change URLs without redeploying frontend
- Foundation for "best time to buy" and click-pattern features

---

## Core Features — MVP Scope

### 1. Deal Cover Grid (Default Home View)
- Responsive grid (Pinterest-style)
- Each card: game cover, current price (₹), % drop, historical low, review score badge
- Hover: Add to Wishlist, Mark as Bought, quick review summary
- Filter bar: Platform (PC/PS/Xbox/Switch), Payment (UPI, Card, Wallet), Price range, Discount %
- "Deals For You" personalised section at top (wishlist + click history)
- Live sale banner: e.g., "Steam Spring Sale ends March 26 — 47 games on your wishlist are discounted"

### 2. Steam Library Scanner
- User enters SteamID (no login needed — public API)
- Backend calls Steam GetOwnedGames from server side
- Hide owned games from deal grid
- Flag missing DLCs for owned base games
- No Steam credentials stored — SteamID only, with GDPR-like consent prompt

### 3. India Arbitrage View
- Side-by-side: Steam India (₹) vs Steam Global (USD→₹)
- GST-inclusive breakdown
- Cheapest region flag (which store/region is genuinely cheapest right now)
- UPI / PayPal / Wallet filter (not all stores support UPI)

### 4. Bundle Breaker
- User pastes a bundle URL (Humble, Fanatical, Steam)
- Backend fetches bundle contents + individual current prices
- Verdict: "Bundle saves ₹X — worth it" or "Wait, you can buy separately for ₹Y less"
- Per-game price breakdown table

### 5. Best Time to Buy Calendar
- Calendar showing historical sale patterns (Steam Summer, Winter, regional)
- Rule-based recommendation: Buy Now / Wait for Summer Sale / Wait for Winter Sale
- Current date awareness: e.g., "Steam Spring Sale is live right now (ends March 26, 2026) — this is the best price in 6 months"
- Data from SteamDB public history + local cached DB

### 6. Saved This Year Dashboard
- User logs purchases (manual form: "I bought this at ₹X, was ₹Y")
- Total savings counter
- Monthly savings breakdown (Plotly / chart)
- Fun message: "You've saved enough for 3 free games this year!"

### 7. Wishlist & Alerts
- Add games from deal grid
- Set price alert threshold (e.g., "Alert me when below ₹499")
- Email notification when price drops to threshold or hits all-time low
- Price history graph per game
- GDPR-like consent checkbox before enabling email alerts

### 8. Search & Filters
- Full-text search by title, genre, platform
- Filter: payment method, % off, ₹ range, review score range
- Paginated results — never load full table in one query

---

## Security & Privacy

| Area | Implementation |
|---|---|
| Password hashing | bcrypt with salt (cost factor 12) |
| JWT | Short expiry (15 min) + refresh token rotation |
| Rate limiting | Redis-based on all public endpoints |
| Input validation | Sanitise all inputs — no SQL injection, no XSS |
| Steam data | Only public SteamID stored — no Steam credentials ever |
| Email alerts | GDPR-like opt-in consent before enabling |
| Click analytics | GDPR-like consent with opt-out option |
| HTTPS | Enforced via ALB + CloudFront (TLS termination) |
| Secrets | .env + Docker secrets — never committed to repo |
| Dependency scanning | Trivy on Docker images in CI/CD pipeline |
| SQL safety | Parameterised queries only — no string interpolation in SQL |
| CORS | Strict allow-list in backend middleware |

---

## Error Handling & Observability

| Area | Implementation |
|---|---|
| Logging | Structured logging with `slog` — every request gets a unique request ID |
| Custom errors | Custom error types with user-friendly messages vs internal detail |
| Graceful degradation | Show cached data if external API (Steam, CheapShark, review sources) fails |
| Health checks | `/health` endpoint checks DB, Redis, and external API reachability |
| Dependency health | `/health/deps` returns status of each dependency individually |
| Metrics | Prometheus `/metrics` endpoint (request count, latency, cache hit rate) |
| Error reporting | Sentry integration for production error alerting |
| API docs | Swagger UI at `/swagger/index.html` — auto-generated from Go annotations |

---

## Performance & Scalability

| Area | Implementation |
|---|---|
| Pagination | Every list endpoint uses limit + offset (default 20, max 100) |
| Redis caching | Hot paths: deal grid, "Deals For You", search results |
| DB indexes | On: game_id, user_id, fetched_at, price_inr, platform |
| Connection pooling | pgxpool for PostgreSQL |
| Stateless backend | JWT-only auth — no server-side session → horizontal scaling ready |
| Cache TTLs | Deal list: 5 min / Review scores: 24 hrs / Price history: 1 hr |
| Future scaling | Read replicas for heavy read traffic during Steam sales |

---

## Testing Strategy

### Unit Tests (Go testing + testify)
- Service layer: business logic (arbitrage calc, bundle verdict, score weighting)
- Repository layer: DB query correctness (using test DB or sqlmock)
- Middleware: auth, rate limiting, input validation
- Utility functions: currency conversion, score normalisation

### Integration Tests
- Handler tests using `net/http/httptest`
- Full request/response cycle per endpoint
- Redis behaviour under rate limiting

### CI/CD Gate
- `go test ./...` must pass before any build step
- `go vet ./...` for static analysis
- Minimum coverage target: 70% on service layer

### Running Tests
```bash
# All tests
go test ./...

# With coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Specific package
go test ./internal/services/...

# Verbose
go test -v ./internal/handlers/...
```

---

## API Documentation (Swagger)

Using `swaggo/swag` to auto-generate OpenAPI 3.0 docs from Go annotations.

```bash
# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs (run from project root)
swag init -g cmd/server/main.go

# Docs available at:
# http://localhost:8080/swagger/index.html
```

**Annotation example:**
```go
// GetDeals returns paginated active deals
// @Summary      Get current deals
// @Tags         deals
// @Produce      json
// @Param        platform  query  string  false  "Filter by platform"
// @Param        limit     query  int     false  "Results per page" default(20)
// @Success      200  {object}  models.DealListResponse
// @Failure      400  {object}  models.ErrorResponse
// @Router       /api/deals [get]
func (h *DealHandler) GetDeals(w http.ResponseWriter, r *http.Request) {
```

Update docs whenever an endpoint changes. Swagger is the source of truth for the API contract between the two developers.

---

## Database Schema

```sql
-- Users
users (id, email, password_hash, steam_id, consent_analytics, consent_alerts,
       created_at, updated_at)

-- Games
games (id, title, slug, platform, cover_url, metadata_source, created_at)

-- Prices
prices (id, game_id, store, price_inr, price_usd, region,
        is_historical_low, fetched_at)

-- Deals
deals (id, game_id, store, discount_pct, deal_url, expires_at, cached_at)

-- Review scores (per source, aggregated separately)
review_scores (id, game_id, source, score_normalised, raw_score, fetched_at)

-- Wishlists
wishlists (id, user_id, game_id, alert_threshold_inr, created_at)

-- Purchases (Saved This Year)
purchases (id, user_id, game_id, paid_inr, original_inr, bought_at)

-- Click analytics
clicks (id, user_id, game_id, platform, clicked_at)

-- Bundles
bundles (id, name, bundle_url, store, total_price_inr, expires_at, cached_at)
bundle_games (id, bundle_id, game_id)
```

**Indexes (critical for performance):**
```sql
CREATE INDEX idx_prices_game_id ON prices(game_id);
CREATE INDEX idx_prices_fetched_at ON prices(fetched_at);
CREATE INDEX idx_deals_expires_at ON deals(expires_at);
CREATE INDEX idx_clicks_user_game ON clicks(user_id, game_id);
CREATE INDEX idx_review_scores_game ON review_scores(game_id, source);
```

---

## Backend API Endpoints

### Health & Docs
```
GET  /health                              → overall health check
GET  /health/deps                         → per-dependency health (DB, Redis, APIs)
GET  /metrics                             → Prometheus metrics
GET  /swagger/index.html                  → Swagger UI
```

### Auth
```
POST /api/auth/register
POST /api/auth/login
POST /api/auth/logout
POST /api/auth/refresh                    → refresh token rotation
```

### Games
```
GET  /api/games                           → paginated deal grid
GET  /api/games/{id}                      → game detail + price history + review scores
GET  /api/games/{id}/redirect?platform=X  → click log + redirect URL
GET  /api/games/search?q=...              → full-text search + filters
```

### Deals
```
GET  /api/deals                           → current hot deals (paginated)
GET  /api/deals/for-you                   → personalised deals
```

### Prices & Arbitrage
```
GET  /api/prices/{game_id}/history        → price history data (for chart)
GET  /api/prices/{game_id}/india          → India arbitrage (IN vs Global + GST)
GET  /api/games/{id}/buy-timing           → Best Time to Buy recommendation
```

### Reviews
```
GET  /api/games/{id}/reviews              → aggregated score + per-source breakdown
```

### Wishlist
```
GET    /api/wishlist
POST   /api/wishlist
DELETE /api/wishlist/{id}
PATCH  /api/wishlist/{id}/threshold       → update alert price
```

### Steam Library
```
POST /api/library/import                  → import via SteamID
GET  /api/library                         → list owned game IDs
```

### Bundles
```
POST /api/bundles/analyze                 → Bundle Breaker
```

### Savings
```
GET  /api/savings                         → Saved This Year dashboard
POST /api/savings/purchase                → log a purchase
```

---

## Folder Structure

```
/cmd
    /server              ← main.go
/internal
    /handlers            ← HTTP handlers (one file per domain)
    /services            ← business logic (arbitrage, bundle, reviews, alerts)
    /models              ← data structs + request/response types
    /repositories        ← all DB queries
    /middleware          ← auth, rate limiting, logging, CORS, request ID
    /scheduler           ← cron jobs (price refresh, review score refresh)
/pkg
    /utils               ← currency conversion, score normalisation, helpers
/config                  ← env config loading
/docker                  ← Dockerfiles
/docs                    ← Swagger auto-generated output (swag init)
/migrations              ← SQL migration files (ordered, named)
/.github
    /workflows           ← GitHub Actions CI/CD
/tests
    /integration         ← integration tests (httptest)
```

---

## CI/CD Pipeline

```
git push to main
    ↓
GitHub Actions triggers
    ↓
Job 1: TEST (all branches)
    - go test ./...
    - go vet ./...
    - go build
    ↓ (only if tests pass)
Job 2: SECURITY SCAN
    - Trivy scans Docker image for CVEs
    - Fails build on HIGH/CRITICAL vulnerabilities
    ↓
Job 3: BUILD (main only)
    - docker build
    - tag image with git SHA
    - push to AWS ECR
    ↓
Job 4: DEPLOY (zero downtime)
    - Old container still serves traffic
    - New container starts
    - /health check passes
    - ALB switches traffic
    - Old container drains and stops
```

**Zero downtime achieved by:**
- Graceful shutdown in Go (SIGTERM → finish in-flight requests → exit)
- `/health` endpoint checked by ALB before routing traffic
- Watchtower on EC2 pulls new ECR image automatically

**File:** `.github/workflows/deploy.yml`

---

## AWS Infrastructure

| Service | Role | Free Tier |
|---|---|---|
| EC2 t2.micro | Runs Docker containers | 750 hrs/month free |
| RDS PostgreSQL | Managed DB | 750 hrs/month free |
| ElastiCache Redis | Managed cache | Not free — use local first |
| S3 | Game cover images, static assets | 5GB free |
| CloudFront | CDN for static assets | 1TB transfer free |
| ECR | Docker image registry | 500MB free |
| ALB | Load balancer + health checks | ~$16/month — skip early |

**Free alternative stack (zero budget):**

| AWS Service | Free Alternative |
|---|---|
| EC2 | Fly.io |
| RDS PostgreSQL | Supabase |
| ElastiCache Redis | Upstash |
| S3 | Cloudflare R2 |
| CloudFront | Cloudflare (free forever) |
| ECR | Docker Hub |

---

## AWS IAM — Least Privilege

```
Root AWS Account
    ↓ (used once to create IAM users — never again)

IAM User: "dev-user"
    For: both developers (local CLI, development)
    Permissions: development-scoped only

IAM Role: "ec2-app-role"
    Attached to: EC2 instance
    Permissions: pull from ECR only

IAM Role: "github-actions-role"
    Used by: CI/CD pipeline
    Permissions: push to ECR + SSH to EC2 only
```

If EC2 is compromised → attacker can only pull images.
If GitHub Actions is compromised → push + SSH only.
Principle of Least Privilege: minimum permissions per identity, always.

---

## Team Split — Two Developer Workflow

**Fixed ownership (non-negotiable):**
- Dev B owns 100% of frontend (HTML/CSS/JS → React) and 100% of Swagger annotations
- Dev A owns 100% of CI/CD pipeline and AWS infrastructure

**Everything else** — Go backend code, DB migrations, repositories, services, middleware, tests — is split roughly 50/50 by domain. Neither developer is "the backend person" or "the frontend person" for the Go work. Both write Go. Both write SQL. Both write tests.

---

### At a Glance

| Domain | Dev A | Dev B |
|---|---|---|
| Project setup + Docker Compose | ✓ | |
| Go module, folder structure, shared error types | ✓ | |
| Docker Compose local dev config | ✓ | |
| PostgreSQL + Redis connection layer | ✓ | |
| DB migration tooling | ✓ | |
| **Auth system** (register, login, JWT, refresh tokens) | ✓ | |
| Auth middleware | ✓ | |
| Rate limiting middleware (Redis) | ✓ | |
| Input validation + CORS middleware | ✓ | |
| Request ID + structured logging (slog) | ✓ | |
| Graceful shutdown (SIGTERM) | ✓ | |
| `/health` + `/health/deps` endpoints | ✓ | |
| Prometheus `/metrics` endpoint | ✓ | |
| Sentry integration | ✓ | |
| **Game listings** (model, repo, handler, pagination) | ✓ | |
| **Deals engine** (price history, historical low, deal eval) | ✓ | |
| **Platform redirect** + click analytics logging | ✓ | |
| **"Deals For You"** personalisation endpoint | ✓ | |
| Background cron scheduler (framework) | ✓ | |
| Cron: deal price refresh (every 15 min) | ✓ | |
| DB indexes + query optimisation | ✓ | |
| CI/CD pipeline (GitHub Actions) | ✓ | |
| Trivy security scanning in CI | ✓ | |
| AWS setup (EC2, ECR, IAM, Watchtower) | ✓ | |
| Docker production config (multi-stage, secrets) | ✓ | |
| AWS managed services (RDS, ElastiCache, ALB, CloudFront) | ✓ | |
| **Auth** — Swagger annotations | | ✓ |
| **All Swagger annotations** (every endpoint, every model) | | ✓ |
| Swagger UI route + `swag init` setup | | ✓ |
| **India Arbitrage engine** (Steam IN vs Global, GST) | | ✓ |
| **Multi-source Review Score** (fetch, normalise, weight, cache) | | ✓ |
| Cron: review score refresh (every 24 hrs) | | ✓ |
| **Steam Library Scanner** (SteamID → GetOwnedGames → store) | | ✓ |
| **Bundle Breaker** (fetch bundle, individual prices, verdict) | | ✓ |
| **Wishlist** (CRUD endpoints, alert threshold) | | ✓ |
| Email alert trigger (price ≤ threshold or all-time low) | | ✓ |
| GDPR consent logic (alerts + analytics gates) | | ✓ |
| **Savings dashboard** ("Saved This Year" endpoints) | | ✓ |
| **Best Time to Buy** (calendar logic, live sale detection) | | ✓ |
| **Search** (full-text, filters, paginated) | | ✓ |
| 100% of frontend (deal grid, all pages, dark mode, mobile) | | ✓ |
| Unit + integration tests for all B-owned backend code | | ✓ |
| Unit + integration tests for all A-owned backend code | ✓ | |

---

### Developer A — Backend Foundation + Infrastructure

Dev A builds the skeleton everything else attaches to. If Dev A's layers are solid, Dev B can build features without being blocked. Dev A also owns all DevOps and cloud.

**Backend (Go):**
- Go module init, folder structure, Docker Compose local config
- PostgreSQL connection (pgxpool), Redis connection, `.env` loading
- DB migration tooling + initial schema migrations (all tables)
- Shared error types and error handling conventions doc (Dev B follows this)
- Auth system end-to-end: `POST /api/auth/register`, `POST /api/auth/login`, `POST /api/auth/logout`, `POST /api/auth/refresh`
- bcrypt password hashing (cost 12), JWT (15 min expiry), refresh token rotation
- Auth middleware (validates JWT on protected routes)
- Rate limiting middleware (Redis-based, per-IP, all public endpoints)
- Input validation + sanitisation middleware
- CORS middleware (strict allow-list)
- Request ID middleware
- Structured logging with `slog` (every request logs method, path, status, duration, request ID)
- Graceful shutdown (SIGTERM → finish in-flight requests → exit cleanly)
- `/health` — checks DB + Redis reachability
- `/health/deps` — per-dependency status (DB, Redis, each external API)
- Prometheus `/metrics` endpoint (request count, latency p95, cache hit rate)
- Sentry integration wired into error middleware
- Game listings domain: game model, game repository, `GET /api/games` (paginated + filtered), `GET /api/games/{id}` (detail)
- Deals engine: price history repository, historical low tracking, deal evaluation logic, `GET /api/deals`, `GET /api/deals/for-you`
- Platform redirect: `GET /api/games/{id}/redirect?platform=X` — URL lookup, click log, return redirect URL
- Background cron scheduler framework (shared — Dev B plugs jobs into it)
- Cron job: deal price refresh every 15 minutes
- DB indexes on all critical fields
- `EXPLAIN ANALYZE` query review before Phase 10
- Unit tests for all auth, deals, and game listing logic

**DevOps + AWS (100% Dev A):**
- Multi-stage Dockerfile for Go backend
- nginx Dockerfile for frontend static serving
- GitHub Actions workflow (`.github/workflows/deploy.yml`)
  - Job 1: `go test ./...` + `go vet ./...` (all branches)
  - Job 2: Trivy Docker image scan (fail on HIGH/CRITICAL CVEs)
  - Job 3: Docker build + push to ECR (main only)
  - Job 4: Watchtower zero-downtime restart on EC2
- AWS EC2 setup, ECR registry, Watchtower config
- IAM roles: `ec2-app-role` (pull ECR only), `github-actions-role` (push ECR + SSH), `dev-user` (both devs)
- Manual deploy walkthrough first (Phase 11), then automate (Phase 12)
- Move to managed services: RDS PostgreSQL, ElastiCache Redis (or Upstash), ALB, CloudFront

---

### Developer B — Features + External Integrations + Frontend + Swagger

Dev B builds everything the user actually sees and interacts with — both the backend feature logic and the frontend that renders it. Every endpoint Dev B builds gets Swagger annotations immediately, not after.

**Backend (Go) — Feature Domains:**
- India Arbitrage engine: Steam India (₹) vs Steam Global (USD→₹), GST (18%) calculation, cheapest region logic, `GET /api/prices/{game_id}/india`, `GET /api/prices/{game_id}/history`
- Multi-source Review Score system:
  - HTTP clients for Metacritic, OpenCritic, Steam reviews, IGN, GameSpot
  - Score normalisation (all sources → 0–100)
  - Weighted average (Metacritic 25%, OpenCritic 25%, Steam 30%, IGN 10%, GameSpot 10%)
  - Proportional redistribution if source unavailable; minimum 2 sources to show score
  - `review_scores` model + migration + repository
  - `GET /api/games/{id}/reviews`
  - Cron job: review score refresh every 24 hours (plugged into Dev A's scheduler)
- Steam Library Scanner: accept SteamID + GDPR consent, call GetOwnedGames (server-side), store per user, `POST /api/library/import`, `GET /api/library`
- Bundle Breaker: accept bundle URL, scrape contents (respectful delay, robots.txt), fetch individual prices, verdict logic, `POST /api/bundles/analyze`
- Wishlist: full CRUD, price alert threshold per entry, `GET/POST/DELETE /api/wishlist`, `PATCH /api/wishlist/{id}/threshold`
- Email alert trigger: check wishlist thresholds on every price refresh, send alert when price ≤ threshold or new all-time low
- GDPR consent gates: opt-in before email alerts, opt-in before click analytics; stored on user record
- Savings dashboard: `POST /api/savings/purchase`, `GET /api/savings` (totals + monthly breakdown)
- Best Time to Buy: historical sale calendar data (Steam Summer, Winter, regional), rule-based recommendation, live sale detection (current date vs known sale windows), `GET /api/games/{id}/buy-timing`, live sale banner endpoint
- Search: full-text game search, genre/platform/score/price filters, `GET /api/games/search?q=...` (paginated)
- Unit + integration tests for every endpoint and service listed above

**Swagger (100% Dev B — all endpoints, not just B's):**
- Install `swaggo/swag`, run `swag init`, wire Swagger UI route at `/swagger/index.html`
- Write `@Summary`, `@Tags`, `@Param`, `@Success`, `@Failure`, `@Router` annotations on every single endpoint — including Dev A's auth, health, deals, and game listing endpoints
- Keep Swagger as the source of truth for the API contract; update it before or alongside any endpoint change
- Swagger is how Dev B knows what Dev A's endpoints return, and vice versa — agree on shapes in Swagger first, then build

**Frontend (100% Dev B):**
- Deal cover grid: responsive Pinterest-style layout, hover cards (cover, price ₹, % drop, historical low, review score badge), Add to Wishlist / Mark as Bought actions
- Live sale banner: "Steam Spring Sale ends March 26 — X games on your wishlist are discounted"
- Filter bar: Platform, Payment method (UPI/Card/Wallet), Price range, Discount %, Review score range
- "Deals For You" personalised section at top
- Game detail page: full price history chart, per-source review score breakdown, Best Time to Buy recommendation
- India Arbitrage view: side-by-side ₹ comparison table + GST breakdown + cheapest region flag
- Bundle Breaker UI: URL input + results table + buy/wait verdict
- Best Time to Buy calendar: visual sale calendar + current recommendation
- Saved This Year dashboard: purchase log form + savings graph + fun message
- Wishlist management page: list, threshold edit, remove
- Steam Library import form: SteamID input + GDPR consent checkbox + result summary
- Search page: search bar + filter panel + paginated results
- Dark mode (default), mobile-friendly layout, loading skeleton cards, error/empty states
- All frontend calls go through backend API — no direct external API calls from browser

---

### Shared Responsibilities

| Responsibility | How it works |
|---|---|
| Code review | Every PR requires approval from the other developer before merge |
| Tests | Each developer writes unit + integration tests for their own code |
| `.env.example` | Both keep it updated — never commit real secrets |
| DB schema changes | Talk first, Dev A writes migration, Dev B reviews and approves |
| API contract | Agree on request/response shape in Swagger before either side builds against it |
| Error format | Dev A defines the standard error response struct; Dev B uses it everywhere |
| PR size | Keep PRs focused — one feature or one fix per PR, not five things at once |

### Git Workflow
```
main branch      → always deployable, CI/CD deploys from here
feature branches → feature/bundle-breaker, feature/review-scores, etc.

1. Pick a task from the issue board
2. Create branch: git checkout -b feature/your-task
3. Write code + tests + Swagger annotations (for B) in the same branch
4. Open PR → other developer reviews + approves
5. Merge to main → GitHub Actions runs → zero-downtime deploy
6. Delete feature branch
```

---

## Development Roadmap

### Phase 0 — Environment Setup
**Both developers**
- Install Go, VS Code, Git, Docker Desktop
- Clone repo, `docker compose up`, confirm everything starts
- Verify: `go version`, `docker --version`, `git --version`
- OS: Windows now → CachyOS (Arch Linux) when pendrive arrives

---

### Phase 1 — Project Skeleton
**Dev A**
- Go module init, folder structure (`/cmd`, `/internal`, `/pkg`, `/config`, `/migrations`, `/docs`, `/tests`)
- Docker Compose for local dev (postgres, redis, backend, frontend containers)
- pgxpool PostgreSQL connection + Redis client wired from `.env`
- DB migration tooling (golang-migrate or goose)
- Initial schema migrations: all tables defined upfront (users, games, prices, deals, review_scores, wishlists, purchases, clicks, bundles, bundle_games)
- Shared error types and response structs — document them in `/internal/models/errors.go`
- `/health` endpoint (checks DB + Redis)
- Basic `net/http` router wired

**Dev B** (parallel, once repo is created)
- Install `swaggo/swag`: `go install github.com/swaggo/swag/cmd/swag@latest`
- Wire Swagger UI route: `GET /swagger/index.html`
- Run `swag init -g cmd/server/main.go` — confirm docs generate
- Annotate `/health` as first practice endpoint
- Set up frontend folder: `index.html` shell, CSS reset, basic layout grid

---

### Phase 2 — Authentication
**Dev A**
- `POST /api/auth/register` — bcrypt hash (cost 12), store user
- `POST /api/auth/login` — verify hash, issue JWT (15 min) + refresh token
- `POST /api/auth/refresh` — validate refresh token, rotate and issue new pair
- `POST /api/auth/logout` — invalidate refresh token
- Auth middleware — parse + validate JWT, attach user ID to context
- Rate limiting middleware — Redis token bucket, per-IP on all public endpoints
- Input validation middleware — sanitise all request bodies
- CORS middleware — strict allow-list
- Request ID middleware — UUID per request, propagated through logs
- Structured logging — `slog`, log every request (method, path, status, latency, request ID)
- Graceful shutdown — listen for SIGTERM, drain in-flight requests

**Dev B** (parallel)
- Swagger annotations for all 4 auth endpoints (request body, success, error responses)
- Frontend: login page + register page (forms, validation, JWT stored in memory — never localStorage)
- Frontend: auth state management (logged in / out, redirect to login if JWT expired)

---

### Phase 3 — Game Listings + Price Fetching
**Dev A**
- Game model + repository (`/internal/repositories/game_repo.go`)
- `GET /api/games` — paginated deal grid (limit + offset, default 20, max 100), supports platform filter
- `GET /api/games/{id}` — game detail + latest cached prices
- CheapShark API client — fetch current deals, store in `prices` table
- Cache deal list in Redis (5 min TTL)
- DB indexes: `game_id`, `fetched_at`, `price_inr`, `platform`
- Unit tests: game repo, CheapShark client

**Dev B** (parallel)
- Swagger annotations for game endpoints
- Frontend: deal cover grid — responsive card layout, game cover image, price ₹, % drop badge
- Frontend: filter bar UI (platform, price range) — wired to `GET /api/games` query params
- Frontend: loading skeleton cards (show while API responds)
- Frontend: empty state + error state

---

### Phase 4 — Deals Engine
**Dev A**
- Price history repository — store every price fetch, mark `is_historical_low` flag
- Deal evaluation service — % discount vs historical low, deal quality tier (hot / good / meh)
- `GET /api/deals` — current best deals, paginated, sorted by discount
- `GET /api/deals/for-you` — filter deals to games on user's wishlist + past clicks
- `GET /api/games/{id}/redirect?platform=X` — look up store URL, log click to `clicks` table, return URL
- Unit tests: price history logic, deal evaluation, click logging

**Dev B** (parallel)
- Swagger annotations for deals + redirect endpoints
- Frontend: "Deals For You" section at top of home page (calls `/api/deals/for-you`)
- Frontend: live sale banner component (static for now — wired to real data in Phase 9)
- Frontend: hover card overlay on deal grid (price, % drop, historical low, buttons)
- Frontend: click redirect handler (`window.location.href = url` from backend response)

---

### Phase 5 — India Arbitrage + Price History
**Dev B**
- India Arbitrage service:
  - Fetch Steam India price (₹) and Steam Global price (USD)
  - Convert USD → ₹ using stored exchange rate (refreshed daily)
  - Add GST (18%) to both where applicable
  - Determine cheapest region / store
- `GET /api/prices/{game_id}/india` — returns IN price, Global price, GST breakdown, cheapest flag
- `GET /api/prices/{game_id}/history` — returns price time series for chart
- Unit tests: GST calculation, currency conversion, arbitrage logic

**Dev A** (parallel)
- Background cron scheduler framework (`/internal/scheduler/`) — Dev B plugs price + review jobs in
- Cron: CheapShark price refresh every 15 minutes
- Prometheus `/metrics` endpoint — request count, latency p95, cache hit rate
- Sentry error reporting wired into error middleware

**Dev B** (also)
- Swagger annotations for price + arbitrage endpoints
- Frontend: India Arbitrage view — side-by-side table (Steam IN vs Global), GST row, cheapest region badge
- Frontend: Price history chart on game detail page (line chart, ₹ over time)

---

### Phase 6 — Multi-Source Review Scores
**Dev B**
- HTTP clients for each review source:
  - Metacritic (scrape or unofficial API)
  - OpenCritic (public API)
  - Steam reviews (`appreviews` endpoint — % positive)
  - IGN (scrape review page)
  - GameSpot (scrape or API)
- Score normalisation service — all sources → 0–100
- Weighted average calculator:
  - Metacritic 25%, OpenCritic 25%, Steam 30%, IGN 10%, GameSpot 10%
  - Proportional redistribution if source unavailable
  - Require ≥ 2 sources to display score; otherwise return `"score": null, "reason": "not enough reviews"`
- `review_scores` repository — store normalised scores + source + fetched_at
- `GET /api/games/{id}/reviews` — aggregated score + per-source breakdown
- Cron job: review score refresh every 24 hours (registered in Dev A's scheduler)
- Unit tests: normalisation, weighting, redistribution, edge cases

**Dev A** (parallel)
- Review score DB migration (already in schema from Phase 1, just verify indexes)
- Integration test helpers shared with Dev B

**Dev B** (also)
- Swagger annotations for review endpoint
- Frontend: review score badge on deal grid card (score + source count, colour-coded green/amber/red)
- Frontend: game detail page — full per-source breakdown (score, label, link to source)

---

### Phase 7 — Steam Library Scanner
**Dev B**
- Accept SteamID from user + show GDPR consent checkbox (consent stored on user record)
- Backend calls Steam `GetOwnedGames` API (server-side only — SteamID never leaves backend)
- Store owned `app_id` list per user in DB
- `POST /api/library/import` — trigger import for authenticated user
- `GET /api/library` — list owned game IDs for current user
- Filtering: `GET /api/games` respects `?exclude_owned=true` param — joins against library table
- DLC flagging: for each owned base game, check if DLC exists in deal grid and flag it
- Unit tests: library import service, DLC flagging logic

**Dev A** (parallel)
- `/health/deps` endpoint — per-dependency status (DB, Redis, Steam API, CheapShark)

**Dev B** (also)
- Swagger annotations for library endpoints
- Frontend: Steam Library import page — SteamID input field, GDPR consent checkbox, import button, "X games imported" result summary
- Frontend: deal grid — toggle "Hide owned games" (sends `exclude_owned=true` param), owned game badge on cards still shown if toggle is off

---

### Phase 8 — Wishlist, Alerts + Savings Dashboard
**Dev B**
- Wishlist repository + service
- `GET /api/wishlist` — user's wishlist with current prices
- `POST /api/wishlist` — add game
- `DELETE /api/wishlist/{id}` — remove game
- `PATCH /api/wishlist/{id}/threshold` — update alert price threshold
- Email alert trigger: on every price refresh cron run, check all wishlist thresholds; send email if price ≤ threshold or new all-time low; respect user's `consent_alerts` flag
- GDPR consent gate: prompt user to opt-in before first alert; store on user record
- Savings dashboard:
  - `POST /api/savings/purchase` — log "I bought this at ₹X, was ₹Y"
  - `GET /api/savings` — total saved ₹, monthly breakdown, "equivalent free games" message
- Click analytics: GDPR gate on `clicks` table logging (already wired in Phase 4 — add consent check here)
- Integration tests: wishlist CRUD, alert trigger logic, savings calculations

**Dev A** (parallel)
- Review and harden auth middleware edge cases (token expiry, refresh race conditions)
- Ensure cron scheduler handles panics gracefully (recover + log, don't crash the app)

**Dev B** (also)
- Swagger annotations for wishlist + savings endpoints
- Frontend: Wishlist page — game list with current price, threshold input, remove button, alert status
- Frontend: Saved This Year dashboard — purchase log form (game, paid ₹, original ₹), savings total counter, monthly bar chart, fun message ("You saved enough for X free games!")
- Frontend: GDPR consent modals (email alerts, click tracking)

---

### Phase 9 — Bundle Breaker + Best Time to Buy
**Dev B**
- Bundle Breaker:
  - Accept bundle URL (Humble, Fanatical, Steam bundle)
  - Scrape bundle page for game list (respectful: 1s delay, check robots.txt)
  - Fetch current individual prices for each game in bundle from price DB
  - Verdict: bundle total vs sum of individual current prices → buy bundle / buy separately / mixed
  - `POST /api/bundles/analyze`
  - Unit tests: verdict logic, edge cases (game not in DB, price unavailable)
- Best Time to Buy:
  - Historical sale calendar: Steam Summer (June), Winter (Dec), Spring (Mar), regional sales — stored as date ranges in DB
  - Current date awareness: check today's date against known sale windows (e.g., Spring Sale March 13–26, 2026 → "ON SALE NOW — ends in X days")
  - Rule-based recommendation: if sale is live → "Buy Now", if sale < 30 days away → "Wait — sale soon", otherwise → "Wait for [next sale name]"
  - `GET /api/games/{id}/buy-timing` — recommendation + calendar data
  - Live sale banner endpoint: `GET /api/sales/active` — returns currently active sales
  - Cron: refresh sale calendar data from SteamDB public data (weekly)
  - Unit tests: recommendation logic, date range checks

**Dev A** (parallel)
- Performance pass: `EXPLAIN ANALYZE` on slow queries, add any missing indexes
- Redis cache audit: confirm all hot paths are cached
- Pagination audit: every list endpoint

**Dev B** (also)
- Swagger annotations for bundle + buy-timing + active sales endpoints
- Frontend: Bundle Breaker page — URL input, results table (game | individual price | bundle share), verdict banner
- Frontend: Best Time to Buy calendar — visual sale calendar, current recommendation card, countdown if sale is active
- Frontend: live sale banner on home page now wired to `/api/sales/active`

---

### Phase 10 — Search
**Dev B**
- Full-text game search (PostgreSQL `tsvector` / `ILIKE` for MVP, Meilisearch later)
- `GET /api/games/search?q=...` — search by title, genre, platform
- Filters: payment method, % off, ₹ range, review score range
- Paginated results (limit + offset)
- Cache popular search queries in Redis (1 min TTL)
- Unit + integration tests for search + filters

**Dev A** (parallel)
- GIN index on `games.title` for full-text search
- Search result caching strategy review

**Dev B** (also)
- Swagger annotations for search endpoint
- Frontend: search bar (sticky in nav), search results page with filter panel, paginated result cards

---

### Phase 11 — Deployment (Manual First)
**Dev A** (lead, Dev B assists and follows along)
- Multi-stage Dockerfile for Go backend (build stage → minimal runtime image)
- nginx Dockerfile for frontend static files
- `docker compose -f docker-compose.prod.yml up` — production-equivalent local test
- Manual SSH deploy to AWS EC2 — both developers run through this once
- `.env` management on EC2, Docker secrets for DB password
- Both developers understand every step before CI/CD automates it

**Dev B** (parallel)
- Frontend production build — minify JS/CSS, optimise images
- Confirm all API base URLs are configurable via environment variable (not hardcoded)
- Full frontend smoke test against manually deployed backend

---

### Phase 12 — CI/CD Pipeline
**Dev A**
- `.github/workflows/deploy.yml` — four-job pipeline:
  - Job 1: `go test ./...` + `go vet ./...` (all branches, blocks merge if failing)
  - Job 2: Trivy Docker image scan (fail on HIGH/CRITICAL CVEs)
  - Job 3: Docker build + tag with git SHA + push to ECR (main only)
  - Job 4: Watchtower on EC2 pulls new image, zero-downtime restart
- Branch protection: main requires passing CI before merge
- Notify both developers on pipeline failure (GitHub Actions notification)

**Dev B** (parallel)
- Ensure `swag init` runs as a pre-build step in CI (docs must be generated before `go build`)
- Confirm frontend build is included in nginx Docker image build step
- End-to-end smoke test suite (basic happy-path curl checks after deploy)

---

### Phase 13 — Production AWS Infrastructure
**Dev A** (lead)
- AWS ALB: configure listeners, target groups, health check on `/health`
- Move DB to RDS PostgreSQL (migrate data, update connection strings)
- Move Redis to ElastiCache or Upstash
- CloudFront distribution for static assets (S3 origin)
- IAM roles finalised: `ec2-app-role`, `github-actions-role`, `dev-user` (shared between both devs)
- HTTPS: TLS certificate via ACM, enforced at ALB + CloudFront

**Dev B** (parallel)
- Update all frontend asset URLs to use CloudFront domain
- Load test key endpoints (`/api/games`, `/api/deals`) to confirm Redis cache holds under traffic
- Fix any CORS issues that surface with production domain

---

### Phase 14 — Polish + Scale
**Both**
- Full-text search upgrade: Meilisearch (replaces PostgreSQL ILIKE)
- OAuth login: Google and/or Steam (Dev A does backend OAuth flow, Dev B does frontend OAuth UI)
- Analytics dashboard (internal): click trends, platform popularity, most-wishlisted games (Dev A backend, Dev B frontend)
- PWA manifest + service worker for "install to phone" (Dev B)
- Read replicas for PostgreSQL under heavy sale traffic (Dev A)
- Final CV review: make sure every bullet is earned and demonstrable

---

## CV Bullets This Project Earns

### Developer A
- Designed and built a fault-tolerant REST API in Go with PostgreSQL and Redis using the clean repository/service/handler pattern
- Implemented JWT authentication with bcrypt password hashing and refresh token rotation for stateless, horizontally scalable auth
- Built Redis-based rate limiting middleware and DB connection pooling (pgxpool) for concurrent traffic handling
- Implemented structured request logging (slog) with unique request IDs, graceful SIGTERM shutdown, and custom error types
- Built `/health`, `/health/deps`, and Prometheus `/metrics` endpoints for full production observability
- Integrated Sentry error reporting for real-time production alerting
- Built game listings and deals engine with price history tracking, historical low detection, and personalised "Deals For You" feed
- Implemented platform redirect system with click analytics logging and consent gating
- Built zero-downtime CI/CD pipeline using GitHub Actions, Docker, Trivy, and Watchtower — push to main, live in minutes
- Integrated Trivy vulnerability scanning into CI/CD — blocks deploys on HIGH/CRITICAL CVEs
- Deployed on AWS EC2 with ALB, RDS PostgreSQL, ElastiCache Redis, S3, and CloudFront CDN
- Implemented AWS IAM least-privilege roles for EC2, CI/CD pipeline, and developer access
- Containerised all services with Docker Compose using isolated networks and zero public DB/backend exposure

### Developer B
- Designed and built backend feature services in Go: India Arbitrage engine, Bundle Breaker, Steam Library Scanner, Best Time to Buy, Wishlist & Alerts, Savings Dashboard, and full-text Search
- Engineered India-specific price arbitrage engine comparing Steam India vs Global pricing with live GST (18%) calculation and cheapest region detection
- Built multi-source review score aggregation engine (Metacritic, OpenCritic, Steam, IGN, GameSpot) with normalisation, configurable weighting, and proportional redistribution for missing sources
- Built Steam Library Scanner using public Steam Web API (GetOwnedGames) — no credential storage, server-side only
- Designed Bundle Breaker verdict logic comparing bundle price vs sum of individual current prices
- Built "Best Time to Buy" calendar with current-date-aware live sale detection (e.g., Steam Spring Sale March 2026)
- Implemented GDPR-like consent gates for email alerts and click analytics with opt-in stored per user
- Auto-generated Swagger/OpenAPI 3.0 documentation with swaggo/swag covering 100% of API endpoints — used as the team's API contract
- Built responsive frontend from scratch: deal cover grid, hover cards, price history charts, India arbitrage view, dark mode, mobile layout, skeleton loading states
- Wrote unit tests (testify) and integration tests (httptest) for all owned backend services with 70%+ coverage
- Integrated external APIs: CheapShark, Steam Web API, Metacritic, OpenCritic, IGN, GameSpot — all cached and paginated

---

## Legal & Ethical Notes

- Use public APIs; add delays (minimum 1s) between scraping requests
- Steam library import uses only public GetOwnedGames — SteamID only, no credentials
- Respect robots.txt on all scraped sources (Humble, Fanatical)
- Cache aggressively to reduce external API load
- Personal/educational use; label clearly: "not affiliated with Steam, Epic, GOG, etc."
- GDPR-like consent prompts before enabling email alerts or click analytics
- No personal data sold or shared with third parties

---

## Your Rule

No copy-pasting.
Figure it out yourself.
Use Claude for direction and debugging only.

---

## Current Status

```
OS:     Windows now → switching to CachyOS (Arch Linux) when pendrive arrives
Editor: VS Code
Start:  Phase 0 — install tools, verify everything works

NEXT STEPS:
1. Install Go              → go.dev/dl
2. Install VS Code         → code.visualstudio.com
3. Install Git             → git-scm.com
4. Install Docker Desktop  → docker.com
5. Both developers clone the repo, run docker compose up, confirm it works
6. Start Phase 1 (Dev A) + Swagger setup (Dev B) in parallel
```

---

*Start small. Build right.*
*login + deal grid + wishlist → steam library → India arbitrage → review scores*
