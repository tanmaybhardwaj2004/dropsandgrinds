# Production Deployment Runbook

This runbook maps the production AWS setup to the CI/CD and Docker assets in this repo.

## AWS Infrastructure

1. EC2
   - Instance: `t2.micro` for the free-tier baseline.
   - AMI: current Amazon Linux 2023.
   - Security group ingress:
     - `22/tcp` from your admin IP only.
     - `80/tcp` and `443/tcp` from the ALB security group only.
   - Security group egress: allow outbound HTTPS for ECR pulls and API calls.
   - Install Docker and Docker Compose plugin.
   - Attach `ec2-app-role`.

2. ECR
   - Repository: `dropsandgrinds`.
   - Enable scan-on-push and tag immutability for SHA tags.
   - Keep `latest` mutable for Watchtower polling.

3. IAM
   - `ec2-app-role`: ECR pull only (`ecr:GetAuthorizationToken`, `ecr:BatchGetImage`, `ecr:GetDownloadUrlForLayer`, `ecr:BatchCheckLayerAvailability`).
   - `github-actions-role`: OIDC trust for this repo, ECR push permissions, and permission to use the EC2 SSH deploy path only if the SSH fallback is used.
   - `dev-user`: CLI access for operators; require MFA and least-privilege scoped policies.

4. ALB
   - Internet-facing ALB on public subnets.
   - Target group points to EC2 instance port `80`.
   - Health check path: `/health`.
   - Listener `80` redirects to `443`.
   - Listener `443` uses the ACM certificate and forwards to the target group.

5. RDS PostgreSQL
   - Engine: PostgreSQL 16.
   - Private subnet group.
   - Security group allows `5432/tcp` from the EC2 app security group only.
   - Set `DATABASE_URL` or `POSTGRES_*` env values on EC2.
   - Run existing migration tooling after restoring data.

6. Redis
   - Preferred AWS path: ElastiCache Redis in private subnets.
   - Low-cost alternative: Upstash free tier.
   - Set `REDIS_URL` to the private ElastiCache endpoint (`host:6379`) or Upstash TLS URL if the client is updated for TLS.

7. CloudFront
   - Origin: S3 static site bucket or ALB frontend origin.
   - Cache static assets with long TTLs and HTML with short TTLs.
   - Forward `/api/*`, `/auth/*`, `/health`, `/metrics`, and `/swagger/*` to the ALB if CloudFront fronts the whole site.

8. ACM TLS
   - Issue one certificate in the ALB region for the API/app domain.
   - Issue CloudFront certificate in `us-east-1`.
   - Enforce HTTPS at both CloudFront and ALB.

## EC2 App Bootstrap

Create Docker secret files outside the repo:

```bash
sudo mkdir -p /opt/dropsandgrinds/secrets
sudo install -m 600 /dev/null /opt/dropsandgrinds/secrets/postgres_password.txt
sudo install -m 600 /dev/null /opt/dropsandgrinds/secrets/jwt_secret.txt
sudo install -m 600 /dev/null /opt/dropsandgrinds/secrets/sentry_dsn.txt
sudo install -m 600 /dev/null /opt/dropsandgrinds/secrets/steam_api_key.txt
```

Set production env values:

```bash
export BACKEND_IMAGE="<account>.dkr.ecr.<region>.amazonaws.com/dropsandgrinds:latest"
export POSTGRES_PASSWORD_FILE=/opt/dropsandgrinds/secrets/postgres_password.txt
export JWT_SECRET_FILE=/opt/dropsandgrinds/secrets/jwt_secret.txt
export SENTRY_DSN_FILE=/opt/dropsandgrinds/secrets/sentry_dsn.txt
export STEAM_API_KEY_FILE=/opt/dropsandgrinds/secrets/steam_api_key.txt
docker compose -f docker-compose.prod.yml up -d
```

## GitHub Settings

Required repository secrets:

- `AWS_GITHUB_ACTIONS_ROLE_ARN`
- `AWS_REGION`
- `ACTIONS_FAILURE_WEBHOOK_URL`
- `WATCHTOWER_HTTP_API_URL` and `WATCHTOWER_HTTP_API_TOKEN`, or `EC2_HOST`, `EC2_USER`, and `EC2_SSH_KEY`
- `DEPLOY_HEALTH_URL`

Required repository variables:

- `ECR_REPOSITORY=dropsandgrinds`

Main branch protection must require:

- `Go Test And Vet`
- `Trivy Docker Image Scan`
