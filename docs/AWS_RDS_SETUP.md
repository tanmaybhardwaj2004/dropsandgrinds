# AWS RDS PostgreSQL Setup

## Overview
Amazon RDS (Relational Database Service) PostgreSQL will host the application's primary database.

## RDS Configuration Steps

### 1. Create RDS Instance
```bash
aws rds create-db-instance \
  --db-instance-identifier dropsandgrinds-db \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --engine-version 16.4 \
  --allocated-storage 20 \
  --storage-type gp2 \
  --master-username dbadmin \
  --master-user-password ${DB_PASSWORD} \
  --vpc-security-group-ids sg-xxx \
  --db-subnet-group-name dropsandgrinds-subnet-group \
  --backup-retention-period 7 \
  --multi-az false \
  --publicly-accessible false \
  --deletion-protection true
```

### 2. Create DB Subnet Group
```bash
aws rds create-db-subnet-group \
  --db-subnet-group-name dropsandgrinds-subnet-group \
  --db-subnet-group-description "Subnet group for DropsAndGrinds RDS" \
  --subnet-ids subnet-xxx subnet-yyy
```

### 3. Configure Security Group
Inbound rules:
- Port 5432 from EC2 security group (only allow application servers)

### 4. Run Database Migrations
Once the RDS instance is available, run migrations:

```bash
# Set environment variables
export DATABASE_URL="postgres://dbadmin:${DB_PASSWORD}@dropsandgrinds-db.xxxx.region.rds.amazonaws.com:5432/dropsandgrinds?sslmode=require"

# Run migrations using migrate tool
migrate -path migrations -database "$DATABASE_URL" up
```

### 5. Migration Files Order
Ensure migrations run in order:
1. `001_create_users.sql`
2. `002_create_games.sql`
3. `003_create_prices.sql`
4. `004_create_deals.sql`
5. `005_add_indexes_for_performance.sql`
6. `006_create_review_scores_table.sql`
7. `007_create_user_steam_library.sql`
8. `008_create_user_savings.sql`
9. `009_create_clicks_table.sql`
10. `010_create_sales_calendar.sql`
11. `011_add_performance_indexes.sql`
12. `012_add_gin_index_for_search.sql`

## Connection String Format
```
postgres://username:password@host:port/database?sslmode=require
```

## Environment Variables
Update `.env` or ECS task definition:
```
DATABASE_URL=postgres://dbadmin:password@dropsandgrinds-db.xxxx.region.rds.amazonaws.com:5432/dropsandgrinds?sslmode=require
```

## Backup Strategy
- **Automated backups**: Enabled with 7-day retention
- **Manual snapshots**: Create before major schema changes
- **Point-in-time recovery**: Available within backup window

## Monitoring
Enable Enhanced Monitoring:
```bash
aws rds modify-db-instance \
  --db-instance-identifier dropsandgrinds-db \
  --monitoring-interval 60 \
  --apply-immediately
```

## Performance Considerations
- Start with `db.t3.micro` for development
- Scale to `db.t3.medium` or `db.r5.large` for production based on load
- Enable Performance Insights for query analysis
- Consider read replicas for read-heavy workloads (Phase 14)

## Security Best Practices
1. Use SSL/TLS for all connections (`sslmode=require`)
2. Store credentials in AWS Secrets Manager
3. Rotate master password regularly
4. Enable deletion protection
5. Restrict access to specific security groups
6. Use parameter groups for PostgreSQL configuration

## Troubleshooting

### Connection Timeout
- Check security group allows traffic from application
- Verify VPC peering if resources are in different VPCs
- Check if RDS is in available state

### Migration Failures
- Check migration file syntax
- Verify database user has necessary permissions
- Review RDS logs for errors

### Slow Queries
- Enable Performance Insights
- Review slow query log
- Add appropriate indexes
- Consider connection pooling
