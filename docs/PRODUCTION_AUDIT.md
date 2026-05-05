# Production Audit

## Metrics

`/metrics` exposes:

- `http_requests_total` for request count by method, path, and status.
- `http_request_latency_seconds` as a Prometheus summary with p95 quantile.
- `cache_requests_total` for cache hits and misses by hot-path cache name.

Cache hit rate query:

```promql
sum(rate(cache_requests_total{result="hit"}[5m]))
/
sum(rate(cache_requests_total[5m]))
```

## Slow Query Audit

Run these against production-like data before launch and after major catalog growth:

```sql
EXPLAIN ANALYZE SELECT * FROM games ORDER BY title LIMIT 20 OFFSET 0;
EXPLAIN ANALYZE SELECT * FROM deals WHERE is_active = TRUE ORDER BY discount_percent DESC LIMIT 20 OFFSET 0;
EXPLAIN ANALYZE SELECT price_inr, fetched_at FROM prices WHERE game_id = 1 ORDER BY fetched_at DESC LIMIT 30 OFFSET 0;
```

The existing migrations include indexes on critical fields. If `EXPLAIN ANALYZE` shows sequential scans on hot catalog queries with production-sized data, add a focused migration for that specific query shape.

## Redis Cache Audit

Confirmed hot paths with repository-level caching:

- Game search cache: `catalog_search`
- Game list cache: `games_list`
- Game detail cache: `game_detail`
- Deals list cache: `deals_list`

Personalized owned-game filters intentionally skip shared cache keys to avoid cross-user leakage.

## Pagination Audit

Confirmed list-style endpoints expose `limit` and `offset`:

- `GET /api/games`
- `GET /api/games/search`
- `GET /api/deals`
- `GET /api/deals/for-you`
- `GET /api/prices/history/{game_id}`
- `GET /api/wishlist`
- `GET /api/savings/history`

Keep new list endpoints to the same contract: bounded `limit`, non-negative `offset`, and a response that echoes both values.
