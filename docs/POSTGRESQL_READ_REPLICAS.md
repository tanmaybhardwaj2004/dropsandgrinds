# PostgreSQL Read Replicas Setup

## Overview
Read replicas improve database performance by offloading read queries from the primary database instance.

## When to Use Read Replicas

Consider read replicas when:
- Read-heavy workload (many more reads than writes)
- Need to scale read capacity independently
- Want to improve query performance
- Need geographic data distribution
- Want to reduce load on primary instance

## Architecture

```
Application
    |
    ├─→ Primary (Write + Read)
    |
    ├─→ Read Replica 1 (Read only)
    |
    └─→ Read Replica 2 (Read only)
```

## AWS RDS Read Replica Setup

### 1. Create Read Replica
```bash
aws rds create-db-instance-read-replica \
  --db-instance-identifier dropsandgrinds-db-replica-1 \
  --source-db-instance-identifier dropsandgrinds-db
```

### 2. Configure Read Replica
The replica inherits configuration from the primary. You can modify:
- Instance class (scale independently)
- Multi-AZ deployment
- Storage type

### 3. Create Multiple Replicas
Repeat for additional replicas:
```bash
aws rds create-db-instance-read-replica \
  --db-instance-identifier dropsandgrinds-db-replica-2 \
  --source-db-instance-identifier dropsandgrinds-db
```

## Application Configuration

### 1. Update Database Connection Pool
Use a connection pool that supports read/write splitting:
```go
package db

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/jackc/pgx/v5/pgconn"
)

type DBPool struct {
    primary *pgxpool.Pool
    replicas []*pgxpool.Pool
    currentReplica int
}

func NewDBPool(primaryURL string, replicaURLs []string) (*DBPool, error) {
    primary, err := pgxpool.New(context.Background(), primaryURL)
    if err != nil {
        return nil, err
    }
    
    replicas := make([]*pgxpool.Pool, len(replicaURLs))
    for i, url := range replicaURLs {
        replica, err := pgxpool.New(context.Background(), url)
        if err != nil {
            return nil, err
        }
        replicas[i] = replica
    }
    
    return &DBPool{
        primary: primary,
        replicas: replicas,
    }, nil
}

func (p *DBPool) Primary() *pgxpool.Pool {
    return p.primary
}

func (p *DBPool) Replica() *pgxpool.Pool {
    if len(p.replicas) == 0 {
        return p.primary
    }
    
    // Round-robin selection
    replica := p.replicas[p.currentReplica]
    p.currentReplica = (p.currentReplica + 1) % len(p.replicas)
    return replica
}
```

### 2. Update Repository Methods
Direct read queries to replicas:
```go
func (r *CatalogRepository) ListGames(ctx context.Context, query, platform string, limit, offset int, excludeOwned bool, userID int64) ([]models.Game, int, error) {
    // Use replica for read queries
    rows, err := r.dbPool.Replica().Query(ctx, dataQuery, args...)
    // ... rest of implementation
}

func (r *CatalogRepository) CreateGame(ctx context.Context, game *models.Game) error {
    // Use primary for write queries
    _, err := r.dbPool.Primary().Exec(ctx, query, args...)
    // ... rest of implementation
}
```

### 3. Update Environment Variables
```
DATABASE_PRIMARY_URL=postgres://user:pass@primary-host:5432/dbname
DATABASE_REPLICA_URLS=postgres://user:pass@replica1-host:5432/dbname,postgres://user:pass@replica2-host:5432/dbname
```

## Replication Lag

### Monitor Replication Lag
```bash
aws rds describe-db-instances \
  --db-instance-identifier dropsandgrinds-db-replica-1 \
  --query "DBInstances[0].ReadReplicaSourceDBInstanceIdentifier"
```

### Handle Replication Lag in Application
```go
func (p *DBPool) QueryWithLagCheck(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
    // Try replica first
    rows, err := p.Replica().Query(ctx, query, args...)
    if err != nil {
        // Fallback to primary on error
        return p.Primary().Query(ctx, query, args...)
    }
    return rows, nil
}
```

## Performance Considerations

### 1. Connection Pooling
- Use PgBouncer for connection pooling
- Configure appropriate pool sizes
- Monitor connection usage

### 2. Query Optimization
- Ensure read queries are optimized
- Add appropriate indexes
- Avoid long-running queries on replicas

### 3. Monitoring
- Track replication lag
- Monitor replica CPU/memory usage
- Compare primary vs replica performance

## Cost Considerations

- Read replicas incur additional costs
- Scale based on actual read load
- Consider smaller instance classes for replicas
- Use on-demand or reserved instances appropriately

## Troubleshooting

### High Replication Lag
- Check primary instance load
- Verify network bandwidth
- Consider reducing write load
- Scale up replica instance

### Replica Not Syncing
- Check replica status in AWS Console
- Verify replication is enabled
- Review RDS logs for errors

### Connection Failures
- Verify security group allows traffic
- Check DNS resolution
- Verify credentials are correct
