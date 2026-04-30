# AWS ElastiCache Redis Setup

## Overview
Amazon ElastiCache for Redis provides a managed in-memory data store for caching and session management.

## ElastiCache Configuration Steps

### 1. Create Redis Subnet Group
```bash
aws elasticache create-cache-subnet-group \
  --cache-subnet-group-name dropsandgrinds-redis-subnet-group \
  --cache-subnet-group-description "Subnet group for DropsAndGrinds Redis" \
  --subnet-ids subnet-xxx subnet-yyy
```

### 2. Create Redis Security Group
Inbound rules:
- Port 6379 from EC2 security group (only allow application servers)

### 3. Create Redis Cluster
```bash
aws elasticache create-replication-group \
  --replication-group-id dropsandgrinds-redis \
  --replication-group-description "Redis cluster for DropsAndGrinds" \
  --engine redis \
  --engine-version 7.0 \
  --cache-node-type cache.t3.micro \
  --num-cache-clusters 1 \
  --automatic-failover-enabled \
  --multi-az-enabled \
  --at-rest-encryption-enabled \
  --transit-encryption-enabled \
  --auth-token ${REDIS_AUTH_TOKEN} \
  --cache-subnet-group-name dropsandgrinds-redis-subnet-group \
  --security-group-ids sg-xxx
```

### 4. Get Redis Endpoint
```bash
aws elasticache describe-replication-groups \
  --replication-group-id dropsandgrinds-redis \
  --query "ReplicationGroups[0].PrimaryEndpoint.Address" \
  --output text
```

## Connection String Format
```
redis://:${AUTH_TOKEN}@primary-endpoint:6379/0
```

## Environment Variables
Update `.env` or ECS task definition:
```
REDIS_URL=redis://:${AUTH_TOKEN}@dropsandgrinds-redis.xxxx.cache.amazonaws.com:6379/0
```

## Security Best Practices
1. Enable encryption at rest and in transit
2. Use AUTH token for authentication
3. Store AUTH token in AWS Secrets Manager
4. Restrict access to specific security groups
5. Enable automatic failover for high availability
6. Use VPC endpoints for private connectivity

## Monitoring
Enable CloudWatch metrics:
- CacheHits
- CacheMisses
- Evictions
- Memory usage
- CPU utilization
- Network bytes in/out

## Backup Strategy
- Enable automatic daily backups
- Set retention period to 7 days
- Create manual snapshots before major changes

## Performance Considerations
- Start with `cache.t3.micro` for development
- Scale to `cache.t3.medium` or `cache.r5.large` for production
- Monitor eviction rate - high evictions indicate insufficient memory
- Consider cluster mode for horizontal scaling (Phase 14)

## Troubleshooting

### Connection Refused
- Check security group allows traffic from application
- Verify Redis cluster is in available state
- Check if AUTH token is correct

### High Eviction Rate
- Increase memory allocation
- Review cache TTL settings
- Consider cache partitioning strategy

### Slow Performance
- Monitor CPU utilization
- Check network latency
- Review connection pool settings
