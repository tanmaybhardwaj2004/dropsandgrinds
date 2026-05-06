package services

import (
	"context"
	"fmt"
	"log"

	"github.com/meilisearch/meilisearch-go"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

// MeilisearchService handles search functionality using Meilisearch
type MeilisearchService struct {
	client    meilisearch.ServiceManager
	indexName string
}

func NewMeilisearchService(url, masterKey string) *MeilisearchService {
	client := meilisearch.New(url, meilisearch.WithAPIKey(masterKey))

	return &MeilisearchService{
		client:    client,
		indexName: "games",
	}
}

// ConfigureIndex sets up the Meilisearch index with fuzzy search configuration
func (s *MeilisearchService) ConfigureIndex() error {
	index := s.client.Index(s.indexName)

	// Configure searchable fields with weights
	searchableAttributes := []string{
		"title",
		"platform",
	}

	// Configure filterable attributes
	filterableAttributes := []string{
		"platform",
		"is_all_time_low",
	}

	// Configure sortable attributes
	sortableAttributes := []string{
		"price_inr",
		"discount_percent",
		"review_score",
		"release_date",
	}

	// Configure ranking rules
	rankingRules := []string{
		"words",
		"typo",
		"proximity",
		"attribute",
		"sort",
		"exactness",
	}

	// Configure typo tolerance for fuzzy search
	typoTolerance := meilisearch.TypoTolerance{
		Enabled: true,
		MinWordSizeForTypos: meilisearch.MinWordSizeForTypos{
			OneTypo:  4,
			TwoTypos: 8,
		},
		DisableOnWords:      []string{},
		DisableOnAttributes: []string{},
	}

	task, err := index.UpdateSettings(&meilisearch.Settings{
		SearchableAttributes: searchableAttributes,
		FilterableAttributes: filterableAttributes,
		SortableAttributes:   sortableAttributes,
		RankingRules:         rankingRules,
		TypoTolerance:        &typoTolerance,
	})

	if err != nil {
		return fmt.Errorf("failed to update Meilisearch settings: %w", err)
	}

	log.Printf("Meilisearch configure settings task created: %v", task.TaskUID)
	return nil
}

// IndexGames indexes game data in Meilisearch
func (s *MeilisearchService) IndexGames(ctx context.Context, games []models.Game) error {
	index := s.client.Index(s.indexName)

	var documents []map[string]interface{}
	for _, game := range games {
		// Convert game to Meilisearch document
		doc := map[string]interface{}{
			"id":               game.ID,
			"title":            game.Title,
			"platform":         game.Platform,
			"cover_url":        game.CoverURL,
			"price_inr":        game.PriceINR,
			"lowest_price_inr": game.LowestPriceINR,
			"original_inr":     game.OriginalINR,
			"discount_percent": game.DiscountPercent,
			"review_score":     game.ReviewScore,
			"is_all_time_low":  game.IsAllTimeLow,
		}

		documents = append(documents, doc)
	}

	task, err := index.AddDocuments(documents, nil)
	if err != nil {
		return fmt.Errorf("failed to add documents to Meilisearch: %w", err)
	}

	log.Printf("Meilisearch index %d games task created: %v", len(games), task.TaskUID)
	return nil
}

// SearchGames performs fuzzy search with filters
func (s *MeilisearchService) SearchGames(ctx context.Context, query string, filters string, limit, offset int) ([]models.Game, int, error) {
	index := s.client.Index(s.indexName)

	searchReq := &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
	}

	if filters != "" {
		// Example filter string: "genres = RPG AND platforms = PC"
		// Meilisearch filter string is passed directly if formatted properly
		searchReq.Filter = filters
	}

	resp, err := index.Search(query, searchReq)
	if err != nil {
		return nil, 0, fmt.Errorf("meilisearch search failed: %w", err)
	}

	var games []models.Game
	if err := resp.Hits.Decode(&games); err != nil {
		return nil, 0, fmt.Errorf("failed to decode hits: %w", err)
	}

	return games, int(resp.EstimatedTotalHits), nil
}

// DeleteAllGames removes all games from the index
func (s *MeilisearchService) DeleteAllGames(ctx context.Context) error {
	index := s.client.Index(s.indexName)
	task, err := index.DeleteAllDocuments(nil)
	if err != nil {
		return fmt.Errorf("failed to delete all documents: %w", err)
	}
	
	log.Printf("Meilisearch delete all documents task created: %v", task.TaskUID)
	return nil
}
