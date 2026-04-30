# DropsAndGrinds Documentation

This directory contains comprehensive documentation for the DropsAndGrinds project, covering deployment, infrastructure, and future enhancements.

## Documentation Index

### Phase 12: CI/CD
- **[BRANCH_PROTECTION.md](BRANCH_PROTECTION.md)** - Guide for configuring GitHub branch protection rules to ensure code quality before merging to main.

### Phase 13: AWS Infrastructure
- **[AWS_ALB_SETUP.md](AWS_ALB_SETUP.md)** - Complete guide for setting up Application Load Balancer with target groups, listeners, and routing rules.
- **[AWS_RDS_SETUP.md](AWS_RDS_SETUP.md)** - Instructions for configuring Amazon RDS PostgreSQL instance, running migrations, and backup strategies.
- **[AWS_ELASTICACHE_SETUP.md](AWS_ELASTICACHE_SETUP.md)** - Setup guide for ElastiCache Redis cluster including security, monitoring, and performance considerations.
- **[AWS_CLOUDFRONT_SETUP.md](AWS_CLOUDFRONT_SETUP.md)** - CloudFront CDN configuration for serving static assets with S3 origin and cache behaviors.
- **[AWS_IAM_ROLES.md](AWS_IAM_ROLES.md)** - IAM role definitions for EC2, ECS, Lambda, and CodeBuild services with security best practices.
- **[AWS_ACM_SETUP.md](AWS_ACM_SETUP.md)** - AWS Certificate Manager setup for SSL/TLS certificates with DNS validation and HSTS configuration.

### Phase 14: Future Enhancements
- **[MEILISEARCH_UPGRADE.md](MEILISEARCH_UPGRADE.md)** - Guide for upgrading from PostgreSQL full-text search to Meilisearch for better search performance and relevance.
- **[OAUTH_LOGIN.md](OAUTH_LOGIN.md)** - Implementation guide for Google and Steam OAuth 2.0 authentication with security considerations.
- **[ANALYTICS_DASHBOARD.md](ANALYTICS_DASHBOARD.md)** - Analytics implementation using ClickHouse and Metabase for tracking user behavior and platform metrics.
- **[PWA_SETUP.md](PWA_SETUP.md)** - Progressive Web App setup with service workers, offline support, and installability.
- **[POSTGRESQL_READ_REPLICAS.md](POSTGRESQL_READ_REPLICAS.md)** - Guide for setting up PostgreSQL read replicas to scale read capacity and improve performance.

## Quick Start

### For Local Development
1. Use `docker-compose.yml` for local development
2. Run migrations: `migrate -path migrations -database "$DATABASE_URL" up`
3. Start the backend: `go run cmd/server/main.go`
4. Serve frontend: Open `frontend/index.html` in a browser

### For Production Deployment
1. Follow [AWS_ALB_SETUP.md](AWS_ALB_SETUP.md) to set up the load balancer
2. Follow [AWS_RDS_SETUP.md](AWS_RDS_SETUP.md) to configure the database
3. Follow [AWS_ELASTICACHE_SETUP.md](AWS_ELASTICACHE_SETUP.md) to set up Redis
4. Follow [AWS_CLOUDFRONT_SETUP.md](AWS_CLOUDFRONT_SETUP.md) to configure CDN
5. Configure IAM roles per [AWS_IAM_ROLES.md](AWS_IAM_ROLES.md)
6. Set up SSL/TLS per [AWS_ACM_SETUP.md](AWS_ACM_SETUP.md)
7. Deploy using the CI/CD pipeline in `.github/workflows/deploy.yml`

### For CI/CD Setup
1. Configure GitHub Actions workflow in `.github/workflows/deploy.yml`
2. Set up required secrets in GitHub repository settings:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`
   - `AWS_REGION`
   - `EC2_SSH_KEY` (for manual deployment)
   - `EC2_HOST` (for manual deployment)
   - `EC2_USER` (for manual deployment)
3. Configure branch protection per [BRANCH_PROTECTION.md](BRANCH_PROTECTION.md)

## Architecture Overview

```
Internet
    ↓
CloudFront (CDN)
    ↓
ALB (Load Balancer)
    ├─→ Frontend (nginx:80)
    └─→ Backend (Go:8080)
        ↓
    ├─→ RDS PostgreSQL
    └─→ ElastiCache Redis
```

## Security Checklist

- [ ] Enable HTTPS/TLS via ACM
- [ ] Configure security groups for all services
- [ ] Use IAM roles instead of access keys
- [ ] Store secrets in AWS Secrets Manager
- [ ] Enable RDS encryption at rest
- [ ] Enable ElastiCache encryption in transit
- [ ] Configure branch protection rules
- [ ] Enable security scanning in CI/CD

## Monitoring Checklist

- [ ] Enable CloudWatch metrics for all services
- [ ] Set up ALB access logs
- [ ] Enable RDS Enhanced Monitoring
- [ ] Configure ElastiCache metrics
- [ ] Set up CloudFront logging
- [ ] Enable CloudTrail for API audit logs
- [ ] Configure error tracking (Sentry)

## Troubleshooting

### Common Issues

**Database Connection Failed**
- Check RDS security group allows traffic from application
- Verify DATABASE_URL is correct
- Review RDS logs for errors

**Redis Connection Failed**
- Check ElastiCache security group allows traffic
- Verify REDIS_URL is correct
- Ensure AUTH token is configured

**Deployment Failed**
- Check CI/CD workflow logs
- Verify ECR repository exists
- Confirm IAM roles have correct permissions

**High Latency**
- Check CloudFront cache hit ratio
- Review RDS query performance
- Monitor ElastiCache eviction rate
- Check ALB target health

## Support

For issues or questions:
1. Check the relevant documentation file
2. Review GitHub Issues
3. Check CloudWatch logs and metrics
4. Review application logs
