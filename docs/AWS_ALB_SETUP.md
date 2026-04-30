# AWS Application Load Balancer (ALB) Configuration

## Overview
The ALB serves as the entry point for all traffic, distributing requests between frontend and backend services.

## Architecture
```
Internet → CloudFront → ALB → Target Groups
                              ├─ Frontend (nginx:80)
                              └─ Backend (Go:8080)
```

## ALB Configuration Steps

### 1. Create the Application Load Balancer
```bash
aws elbv2 create-load-balancer \
  --name dropsandgrinds-alb \
  --subnets subnet-xxx subnet-yyy \
  --security-groups sg-xxx \
  --scheme internet-facing \
  --type application
```

### 2. Create Target Groups

#### Frontend Target Group
```bash
aws elbv2 create-target-group \
  --name dropsandgrinds-frontend \
  --protocol HTTP \
  --port 80 \
  --vpc-id vpc-xxx \
  --target-type ip \
  --health-check-path / \
  --health-check-interval-seconds 30 \
  --health-check-timeout-seconds 5 \
  --healthy-threshold 2 \
  --unhealthy-threshold 3
```

#### Backend Target Group
```bash
aws elbv2 create-target-group \
  --name dropsandgrinds-backend \
  --protocol HTTP \
  --port 8080 \
  --vpc-id vpc-xxx \
  --target-type ip \
  --health-check-path /health \
  --health-check-interval-seconds 30 \
  --health-check-timeout-seconds 5 \
  --healthy-threshold 2 \
  --unhealthy-threshold 3
```

### 3. Register Targets (EC2 instances or ECS tasks)
```bash
# Register frontend target
aws elbv2 register-targets \
  --target-group-arn arn:aws:elasticloadbalancing:region:account:targetgroup/dropsandgrinds-frontend/xxx \
  --targets Id=instance-id,Port=80

# Register backend target
aws elbv2 register-targets \
  --target-group-arn arn:aws:elasticloadbalancing:region:account:targetgroup/dropsandgrinds-backend/xxx \
  --targets Id=instance-id,Port=8080
```

### 4. Create Listeners

#### HTTP Listener (redirects to HTTPS)
```bash
aws elbv2 create-listener \
  --load-balancer-arn arn:aws:elasticloadbalancing:region:account:loadbalancer/app/dropsandgrinds-alb/xxx \
  --protocol HTTP \
  --port 80 \
  --default-actions Type=redirect,RedirectConfig="{Protocol=HTTPS,Port=443,StatusCode=HTTP_301}"
```

#### HTTPS Listener
```bash
aws elbv2 create-listener \
  --load-balancer-arn arn:aws:elasticloadbalancing:region:account:loadbalancer/app/dropsandgrinds-alb/xxx \
  --protocol HTTPS \
  --port 443 \
  --certificates CertificateArn=arn:aws:acm:region:account:certificate/xxx \
  --ssl-policy ELBSecurityPolicy-TLS-1-2-2017-01 \
  --default-actions Type=forward,TargetGroupArn=arn:aws:elasticloadbalancing:region:account:targetgroup/dropsandgrinds-frontend/xxx
```

### 5. Create Listener Rules

#### Rule for API requests (backend)
```bash
aws elbv2 create-rule \
  --listener-arn arn:aws:elasticloadbalancing:region:account:listener/app/dropsandgrinds-alb/xxx/xxx \
  --priority 1 \
  --conditions Field=path-pattern,Values="{/api/*}" \
  --actions Type=forward,TargetGroupArn=arn:aws:elasticloadbalancing:region:account:targetgroup/dropsandgrinds-backend/xxx
```

#### Rule for Swagger (backend)
```bash
aws elbv2 create-rule \
  --listener-arn arn:aws:elasticloadbalancing:region:account:listener/app/dropsandgrinds-alb/xxx/xxx \
  --priority 2 \
  --conditions Field=path-pattern,Values="{/swagger/*}" \
  --actions Type=forward,TargetGroupArn=arn:aws:elasticloadbalancing:region:account:targetgroup/dropsandgrinds-backend/xxx
```

### 6. Configure Security Groups

ALB Security Group (Inbound):
- Port 80 from 0.0.0.0/0 (HTTP)
- Port 443 from 0.0.0.0/0 (HTTPS)

ALB Security Group (Outbound):
- All traffic to target group security group

Target Group Security Group (Inbound):
- Port 80 from ALB security group (frontend)
- Port 8080 from ALB security group (backend)

## Health Checks

### Frontend Health Check
- **Path**: `/`
- **Interval**: 30 seconds
- **Timeout**: 5 seconds
- **Healthy threshold**: 2 consecutive successes
- **Unhealthy threshold**: 3 consecutive failures

### Backend Health Check
- **Path**: `/health`
- **Interval**: 30 seconds
- **Timeout**: 5 seconds
- **Healthy threshold**: 2 consecutive successes
- **Unhealthy threshold**: 3 consecutive failures

## Routing Summary

| Path Pattern | Target Group | Notes |
|-------------|--------------|-------|
| `/api/*` | Backend | All API endpoints |
| `/swagger/*` | Backend | Swagger documentation |
| `/*` | Frontend | All other requests (static files) |

## Monitoring

Enable access logs for the ALB:
```bash
aws elbv2 modify-load-balancer-attributes \
  --load-balancer-arn arn:aws:elasticloadbalancing:region:account:loadbalancer/app/dropsandgrinds-alb/xxx \
  --attributes Key=access_logs.s3.enabled,Value=true \
  --attributes Key=access_logs.s3.bucket,Value=dropsandgrinds-logs \
  --attributes Key=access_logs.s3.prefix,Value=alb-logs
```

## Troubleshooting

### Unhealthy Targets
1. Check security group rules allow traffic from ALB
2. Verify health check path is correct
3. Check application logs for errors
4. Ensure target is listening on correct port

### 502 Bad Gateway
- Backend target may be unhealthy
- Check backend application logs
- Verify backend is running on port 8080

### 503 Service Unavailable
- No healthy targets in target group
- All targets may be in draining state
- Check target health status
