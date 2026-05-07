# Developer Guide

## Section 1 - Starting the App from VSCodium

### Prerequisites

- Docker
- Go
- Git

### Clone the repo

```bash
git clone https://github.com/tanmaybhardwaj2004/dropsandgrinds.git
cd dropsandgrinds
```

### Configure environment

```bash
cp .env.example .env
```

Open `.env` in VSCodium and fill in the required values.

### Start the app

```bash
docker compose up -d
```

Open http://localhost in your browser.

### View backend logs

```bash
docker compose logs -f backend
```

### Stop the app

```bash
docker compose down
```

### Rebuild after code changes

```bash
docker compose up -d --build
```

## Section 2 - Making Code Changes

### Backend changes

Edit Go files, then rebuild the backend:

```bash
docker compose up -d --build backend
```

### Frontend changes

Edit HTML, CSS, or JS files, then rebuild the frontend:

```bash
docker compose up -d --build frontend
```

### Database migrations

Add a new `.sql` file in `/migrations/`, then restart the backend. Migrations auto-apply on backend startup.

### Running tests

```bash
go test ./...
```

### Running vet

```bash
go vet ./...
```

## Section 3 - Deploying to AWS EC2 (replacing the existing deployment)

### Build and push new Docker image to ECR

```bash
docker build -t dropsandgrinds .
docker tag dropsandgrinds:latest 727646495620.dkr.ecr.ap-south-1.amazonaws.com/dropsandgrinds:latest
aws ecr get-login-password --region ap-south-1 | docker login --username AWS --password-stdin 727646495620.dkr.ecr.ap-south-1.amazonaws.com
docker push 727646495620.dkr.ecr.ap-south-1.amazonaws.com/dropsandgrinds:latest
```

### SSH into EC2

```bash
ssh -i /home/tan/Documents/dropsandgrinds-key.pem ubuntu@3.110.58.88
```

### Pull latest code and restart

```bash
cd ~/dropsandgrinds
git pull origin feat/devops-infra
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

### Verify deployment

```bash
curl http://3.110.58.88:8080/health
curl http://3.110.58.88/api/deals?limit=5
```

### Check logs if something is wrong

```bash
docker compose -f docker-compose.prod.yml logs -f backend
```

## Section 4 - Troubleshooting

### 429 errors

Rate limiting was removed. If 429 errors still occur, check Redis connection and any upstream proxy or CDN rules.

### Empty deals

Run:

```bash
curl http://localhost:8080/api/deals
```

Use this to confirm whether the backend is returning data.

### Database issues

```bash
docker compose logs postgres
```

### Frontend not updating

Hard refresh with `Ctrl+Shift+R` or clear the browser cache.

### Docker networking issues

```bash
docker compose down
docker system prune -f
docker compose up -d
```
