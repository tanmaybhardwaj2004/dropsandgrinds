# Deployment Guide

Complete deployment instructions for DropsAndGrinds.

## Deployment Options

1. [Docker Compose (Recommended)](#docker-compose-deployment)
2. [AWS Deployment](#aws-deployment)
3. [Manual Server Deployment](#manual-server-deployment)

## Docker Compose Deployment

### Production Setup

```bash
# Clone repository
git clone <repository-url>
cd dropsandgrinds

# Create production environment file
cp .env.example .env.prod

# Edit .env.prod with production values:
# - Strong JWT_SECRET (use: openssl rand -base64 32)
# - Production database credentials
# - Production Steam API key

# Start production stack
docker-compose -f docker-compose.prod.yml up -d

# Run migrations
docker-compose -f docker-compose.prod.yml exec api go run cmd/migrate/main.go

# Check logs
docker-compose -f docker-compose.prod.yml logs -f
```

### SSL with Let's Encrypt

```bash
# Using nginx-proxy + acme-companion
# Already configured in docker-compose.prod.yml

# Ensure ports 80 and 443 are open
# SSL certificates auto-generated on first request
```

## AWS Deployment

### Prerequisites

- AWS CLI configured
- Terraform installed (optional)

### Services Used

- **ECS/Fargate** - Container orchestration
- **RDS PostgreSQL** - Managed database
- **ElastiCache Redis** - Managed cache
- **ALB** - Application Load Balancer
- **CloudFront** - CDN

### Step-by-Step

1. **Database Setup**
   ```bash
   # Create RDS PostgreSQL instance
   aws rds create-db-instance \
     --db-instance-identifier dropsandgrinds-db \
     --db-instance-class db.t3.micro \
     --engine postgres \
     --master-username admin \
     --master-user-password <secure-password>
   ```

2. **ElastiCache Setup**
   ```bash
   aws elasticache create-cache-cluster \
     --cache-cluster-id dropsandgrinds-cache \
     --engine redis \
     --cache-node-type cache.t3.micro \
     --num-cache-nodes 1
   ```

3. **Deploy to ECS**
   ```bash
   # Build and push images
   docker build -t dropsandgrinds/api:latest .
   docker push dropsandgrinds/api:latest
   
   # Update ECS service
   aws ecs update-service \
     --cluster dropsandgrinds \
     --service api \
     --force-new-deployment
   ```

See [AWS_ALB_SETUP.md](./AWS_ALB_SETUP.md) and [AWS_RDS_SETUP.md](./AWS_RDS_SETUP.md) for detailed instructions.

## Manual Server Deployment

### Server Requirements

- Ubuntu 22.04 LTS (recommended)
- 2GB RAM minimum
- 20GB storage minimum

### Installation

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y postgresql redis-server nginx git

# Install Go
curl -L https://go.dev/dl/go1.21.0.linux-amd64.tar.gz | sudo tar -C /usr/local -xzf -
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Clone and build
git clone <repository-url>
cd dropsandgrinds
go build -o api cmd/api/main.go

# Setup database
sudo -u postgres createdb dropsandgrinds
sudo -u postgres createuser -P dropsandgrinds_user

# Configure environment
sudo mkdir -p /etc/dropsandgrinds
sudo cp .env.example /etc/dropsandgrinds/.env
# Edit /etc/dropsandgrinds/.env with production values

# Create systemd service
sudo tee /etc/systemd/system/dropsandgrinds.service << EOF
[Unit]
Description=DropsAndGrinds API
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/dropsandgrinds
EnvironmentFile=/etc/dropsandgrinds/.env
ExecStart=/opt/dropsandgrinds/api
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Deploy binary
sudo mkdir -p /opt/dropsandgrinds
sudo cp api /opt/dropsandgrinds/
sudo cp -r frontend /opt/dropsandgrinds/

# Start service
sudo systemctl daemon-reload
sudo systemctl enable dropsandgrinds
sudo systemctl start dropsandgrinds
```

### Nginx Configuration

```nginx
# /etc/nginx/sites-available/dropsandgrinds
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL certificates
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Static files
    location / {
        root /opt/dropsandgrinds/frontend;
        try_files $uri $uri/ /index.html;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # API proxy
    location /api/ {
        proxy_pass http://localhost:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/dropsandgrinds /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

## Environment Variables for Production

```bash
# Database
DATABASE_URL=postgres://user:pass@db-host:5432/dropsandgrinds?sslmode=require

# Redis
REDIS_URL=redis://cache-host:6379/0

# Security
JWT_SECRET=<strong-random-key-32-chars-min>
STEAM_API_KEY=<your-steam-api-key>

# Performance
GIN_MODE=release
MAX_WORKERS=4
CACHE_TTL=300

# Monitoring
LOG_LEVEL=warn
ENABLE_METRICS=true
```

## Post-Deployment Checklist

- [ ] Database migrations run successfully
- [ ] API health check passes
- [ ] Frontend loads without errors
- [ ] SSL certificate is valid
- [ ] Rate limiting is active
- [ ] Logs are being collected
- [ ] Backups are configured

## Rollback Procedure

```bash
# If deployment fails, rollback:
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up -d --build

# Or for systemd:
sudo systemctl stop dropsandgrinds
# Restore previous binary
sudo systemctl start dropsandgrinds
```

## Monitoring

### Health Checks

```bash
# API health
curl https://your-domain.com/api/health

# Database connection
curl https://your-domain.com/api/deals?limit=1
```

### Log Locations

- Docker: `docker-compose logs -f`
- Systemd: `sudo journalctl -u dropsandgrinds -f`
- Nginx: `/var/log/nginx/access.log`
