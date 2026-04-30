# Meilisearch Upgrade Guide

## Overview
Meilisearch is a fast, relevant, and typo-tolerant search engine that provides better full-text search capabilities than PostgreSQL's built-in search.

## Why Upgrade from PostgreSQL Search?

### Current PostgreSQL Search Limitations
- Limited fuzzy matching capabilities
- Slower performance on large datasets
- Less relevant ranking
- No built-in typo tolerance
- Limited faceted search support

### Meilisearch Benefits
- Lightning-fast search (milliseconds)
- Built-in typo tolerance
- Relevance ranking out of the box
- Faceted search support
- Easy-to-use API
- Real-time indexing

## Meilisearch Deployment Options

### Option 1: Self-Hosted on EC2
```bash
# Install Meilisearch
curl -L https://install.meilisearch.com | sh

# Run Meilisearch
./meilisearch --master-key=${MEILISEARCH_MASTER_KEY}
```

### Option 2: Docker
```bash
docker run -p 7700:7700 \
  -e MEILI_MASTER_KEY=${MEILISEARCH_MASTER_KEY} \
  -v $(pwd)/meili_data:/meili_data \
  getmeili/meilisearch:v1.5
```

### Option 3: Managed (Meilisearch Cloud)
Use Meilisearch Cloud for a fully managed solution.

## Integration Steps

### 1. Add Meilisearch Client to Go
```bash
go get github.com/meilisearch/meilisearch-go
```

### 2. Create Search Service
```go
package services

import (
    "context"
    "github.com/meilisearch/meilisearch-go"
    "github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type MeilisearchService struct {
    client *meilisearch.Client
}

func NewMeilisearchService(url, masterKey string) *MeilisearchService {
    client := meilisearch.NewClient(meilisearch.ClientConfig{
        Host:   url,
        APIKey: masterKey,
    })
    return &MeilisearchService{client: client}
}

func (s *MeilisearchService) IndexGames(ctx context.Context, games []models.Game) error {
    index := s.client.Index("games")
    
    _, err := index.UpdateDocuments(games)
    return err
}

func (s *MeilisearchService) SearchGames(ctx context.Context, query string, filters string, limit, offset int) ([]models.Game, int, error) {
    index := s.client.Index("games")
    
    searchResult, err := index.Search(query, &meilisearch.SearchRequest{
        Limit:  limit,
        Offset: offset,
        Filter: filters,
    })
    
    if err != nil {
        return nil, 0, err
    }
    
    var games []models.Game
    if err := json.Unmarshal(searchResult.Hits, &games); err != nil {
        return nil, 0, err
    }
    
    return games, searchResult.EstimatedTotalHits, nil
}
```

### 3. Configure Index Settings
```go
func (s *MeilisearchService) ConfigureIndex() error {
    index := s.client.Index("games")
    
    settings := map[string]interface{}{
        "searchableAttributes": []string{"title", "platform"},
        "filterableAttributes": []string{"platform", "price_inr", "discount_percent", "review_score"},
        "sortableAttributes":   []string{"price_inr", "discount_percent", "review_score"},
        "rankingRules": []string{
            "words",
            "typo",
            "proximity",
            "attribute",
            "sort",
            "exactness",
        },
    }
    
    _, err := index.UpdateSettings(settings)
    return err
}
```

### 4. Update Search Handler
Replace PostgreSQL search with Meilisearch:
```go
func SearchGamesHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    filters := buildFilters(r.URL.Query())
    limit := parseQueryInt(r.URL.Query().Get("limit"), 30)
    offset := parseQueryInt(r.URL.Query().Get("offset"), 0)
    
    games, total, err := meilisearchService.SearchGames(r.Context(), query, filters, limit, offset)
    // ... rest of handler
}
```

### 5. Sync Data to Meilisearch
Create a job to sync games from PostgreSQL to Meilisearch:
```go
func (s *SyncService) SyncGamesToMeilisearch(ctx context.Context) error {
    games, _, err := s.catalogRepo.ListGames(ctx, "", "", 1000, 0, false, 0)
    if err != nil {
        return err
    }
    
    return s.meilisearchService.IndexGames(ctx, games)
}
```

## Migration Strategy

### Phase 1: Dual Search (No Downtime)
- Keep PostgreSQL search active
- Add Meilisearch alongside
- Compare results and performance
- Gradually shift traffic

### Phase 2: Full Cutover
- Update all search endpoints to use Meilisearch
- Remove PostgreSQL search code
- Monitor performance and relevance

### Phase 3: Optimization
- Fine-tune ranking rules
- Add synonyms
- Configure faceted search
- Enable search analytics

## Environment Variables
```
MEILISEARCH_URL=http://meilisearch:7700
MEILISEARCH_MASTER_KEY=your-master-key
```

## Performance Considerations
- Meilisearch is significantly faster than PostgreSQL full-text search
- Index updates are near real-time
- Consider sharding for very large datasets (>10M documents)
- Enable compression for network efficiency

## Security
- Always use a master key in production
- Restrict access to Meilisearch API
- Use API keys with limited scope for frontend
- Enable HTTPS for all requests

## Monitoring
- Track search latency
- Monitor index size
- Track search analytics (queries, click-through rate)
- Set up alerts for index failures
