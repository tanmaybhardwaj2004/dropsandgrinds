# Analytics Dashboard Implementation Guide

## Overview
An analytics dashboard provides insights into user behavior, deal performance, and platform metrics.

## Metrics to Track

### User Metrics
- Daily/Weekly/Monthly Active Users (DAU/WAU/MAU)
- User registration rate
- User retention rate
- Library scan completion rate

### Deal Metrics
- Total deals tracked
- Average discount percentage
- All-time low deal frequency
- Click-through rate on deals

### Search Metrics
- Search query volume
- Popular search terms
- Search result click-through rate
- Zero-result searches

### Platform Metrics
- API response times
- Error rates
- Cache hit/miss ratios
- Database query performance

## Implementation Options

### Option 1: Google Analytics 4
Free, easy integration, limited customization.

**Setup:**
1. Create GA4 property
2. Add tracking code to frontend
3. Configure custom events

### Option 2: Mixpanel
Event-based analytics, user cohorts, funnel analysis.

**Setup:**
1. Create Mixpanel project
2. Add SDK to frontend
3. Track custom events

### Option 3: Self-Hosted (Metabase + ClickHouse)
Full control, no data limits, requires maintenance.

**Setup:**
1. Deploy ClickHouse for analytics database
2. Deploy Metabase for visualization
3. Set up ETL pipeline from PostgreSQL

## Recommended: Self-Hosted Solution

### 1. ClickHouse Setup
```bash
docker run -d \
  --name clickhouse \
  -p 8123:8123 \
  -p 9000:9000 \
  clickhouse/clickhouse-server
```

### 2. Create Analytics Tables
```sql
CREATE TABLE analytics.user_events (
    event_time DateTime,
    user_id UInt64,
    event_type String,
    event_data String,
    page_url String,
    user_agent String
) ENGINE = MergeTree()
ORDER BY (event_time, user_id);

CREATE TABLE analytics.deal_clicks (
    click_time DateTime,
    user_id UInt64,
    game_id UInt64,
    platform String,
    deal_discount UInt8
) ENGINE = MergeTree()
ORDER BY (click_time, game_id);

CREATE TABLE analytics.search_queries (
    query_time DateTime,
    user_id UInt64,
    query_string String,
    result_count UInt32,
    filters String
) ENGINE = MergeTree()
ORDER BY (query_time, query_string);
```

### 3. ETL Pipeline
Create a job to sync data from PostgreSQL to ClickHouse:
```go
package jobs

type AnalyticsETLJob struct {
    pgRepo    *repositories.AnalyticsRepository
    chClient  *clickhouse.Client
}

func (j *AnalyticsETLJob) Run(ctx context.Context) error {
    // Fetch events from PostgreSQL
    events, err := j.pgRepo.GetRecentEvents(ctx, time.Hour)
    if err != nil {
        return err
    }
    
    // Insert into ClickHouse
    for _, event := range events {
        j.chClient.Insert(ctx, "analytics.user_events", event)
    }
    
    return nil
}
```

### 4. Metabase Setup
```bash
docker run -d \
  --name metabase \
  -p 3000:3000 \
  metabase/metabase
```

Connect Metabase to ClickHouse and create dashboards.

## Frontend Event Tracking

### 1. Add Analytics SDK
```javascript
// analytics.js
class Analytics {
    constructor() {
        this.events = [];
        this.flushInterval = 30000; // 30 seconds
    }
    
    track(event, data) {
        this.events.push({
            event_type: event,
            event_data: JSON.stringify(data),
            page_url: window.location.href,
            user_agent: navigator.userAgent,
            event_time: new Date().toISOString()
        });
        
        if (this.events.length >= 10) {
            this.flush();
        }
    }
    
    async flush() {
        if (this.events.length === 0) return;
        
        const response = await fetch('/api/analytics/events', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ events: this.events })
        });
        
        if (response.ok) {
            this.events = [];
        }
    }
}

const analytics = new Analytics();
setInterval(() => analytics.flush(), analytics.flushInterval);
```

### 2. Track Key Events
```javascript
// Track deal clicks
function trackDealClick(gameId, platform, discount) {
    analytics.track('deal_click', { game_id: gameId, platform, discount });
}

// Track searches
function trackSearch(query, resultCount, filters) {
    analytics.track('search', { query, result_count: resultCount, filters });
}

// Track page views
function trackPageView(page) {
    analytics.track('page_view', { page });
}
```

## Backend Event Handler

```go
package handlers

func AnalyticsEventsHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Events []models.AnalyticsEvent `json:"events"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid request"})
        return
    }
    
    // Store events in PostgreSQL
    for _, event := range req.Events {
        if userID := getUserID(r); userID > 0 {
            event.UserID = userID
        }
        analyticsRepo.StoreEvent(r.Context(), event)
    }
    
    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
```

## Dashboard Views

### 1. User Activity Dashboard
- Active users over time
- New user registrations
- User retention cohorts
- Library scan statistics

### 2. Deal Performance Dashboard
- Deals tracked over time
- Average discount trends
- All-time low deals
- Most clicked deals

### 3. Search Analytics Dashboard
- Search volume trends
- Top search terms
- Search result distribution
- Zero-result queries

### 4. System Health Dashboard
- API response times
- Error rates by endpoint
- Cache performance
- Database query times

## Privacy Considerations

1. **Anonymization**: Remove PII from analytics data
2. **Consent**: Respect user analytics consent preferences
3. **Retention**: Set data retention policies (e.g., 90 days)
4. **Aggregation**: Store aggregated data where possible
5. **Access Control**: Restrict dashboard access to authorized users

## Environment Variables
```
CLICKHOUSE_URL=http://clickhouse:8123
CLICKHOUSE_DATABASE=analytics
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=
METABASE_URL=http://metabase:3000
```

## Troubleshooting

### Data Not Appearing
- Check ETL job is running
- Verify ClickHouse connection
- Check event tracking is firing

### Slow Queries
- Add appropriate indexes
- Consider materialized views
- Optimize time range queries

### Dashboard Not Loading
- Verify Metabase connection to ClickHouse
- Check database credentials
- Review Metabase logs
