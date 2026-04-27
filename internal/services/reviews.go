package services

import (
	"context"
	"fmt"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/pkg/reviews"
)

// ReviewService handles review score aggregation
type ReviewService struct {
	repo       *repositories.ReviewRepository
	aggregator *reviews.ReviewAggregator
}

// NewReviewService creates a new review service
func NewReviewService(repo *repositories.ReviewRepository, steamAPIKey, gameSpotAPIKey string) *ReviewService {
	return &ReviewService{
		repo:       repo,
		aggregator: reviews.NewReviewAggregator(steamAPIKey, gameSpotAPIKey),
	}
}

// AggregatedReview represents the aggregated review data
type AggregatedReview struct {
	GameID       int64                      `json:"game_id"`
	Score        int                        `json:"score"`
	Label        string                     `json:"label"`
	Color        string                     `json:"color"`
	SourceCount  int                        `json:"source_count"`
	Sources      []reviews.ReviewScore      `json:"sources"`
	Reason       string                     `json:"reason,omitempty"`
}

// GetAggregatedReview fetches and aggregates review scores for a game
func (s *ReviewService) GetAggregatedReview(ctx context.Context, gameID int64) (*AggregatedReview, error) {
	// Try to get cached scores from database
	scores, err := s.repo.GetReviewScores(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch review scores: %w", err)
	}
	
	// If we have enough cached scores, return them
	if len(scores) >= 2 {
		return s.calculateAggregation(gameID, scores)
	}
	
	// Otherwise, fetch fresh scores from all sources
	freshScores, err := s.aggregator.FetchAllScores(ctx, fmt.Sprintf("%d", gameID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fresh review scores: %w", err)
	}
	
	// Store fresh scores in database
	for _, score := range freshScores {
		if err := s.repo.StoreReviewScore(ctx, gameID, score); err != nil {
			// Log error but continue
			continue
		}
	}
	
	return s.calculateAggregation(gameID, freshScores)
}

// calculateAggregation calculates the aggregated review score
func (s *ReviewService) calculateAggregation(gameID int64, scores []reviews.ReviewScore) (*AggregatedReview, error) {
	// Calculate weighted average
	average, reason, err := reviews.CalculateWeightedAverage(scores, reviews.DefaultWeights)
	if err != nil {
		return &AggregatedReview{
			GameID:      gameID,
			Score:       0,
			Label:       "N/A",
			Color:       "gray",
			SourceCount: len(scores),
			Sources:     scores,
			Reason:      reason,
		}, nil
	}
	
	return &AggregatedReview{
		GameID:      gameID,
		Score:       average,
		Label:       reviews.GetScoreLabel(average),
		Color:       reviews.GetScoreColor(average),
		SourceCount: len(scores),
		Sources:     scores,
		Reason:      reason,
	}, nil
}

// RefreshReviewScores refreshes review scores for a specific game
func (s *ReviewService) RefreshReviewScores(ctx context.Context, gameID int64) error {
	freshScores, err := s.aggregator.FetchAllScores(ctx, fmt.Sprintf("%d", gameID))
	if err != nil {
		return fmt.Errorf("failed to fetch fresh review scores: %w", err)
	}
	
	// Store fresh scores in database
	for _, score := range freshScores {
		if err := s.repo.StoreReviewScore(ctx, gameID, score); err != nil {
			return fmt.Errorf("failed to store review score: %w", err)
		}
	}
	
	return nil
}

// RefreshAllStaleReviews refreshes review scores for all stale games
func (s *ReviewService) RefreshAllStaleReviews(ctx context.Context) error {
	// Get games with stale review scores (older than 24 hours)
	staleGameIDs, err := s.repo.GetStaleReviews(ctx, 24*60*60) // 24 hours
	if err != nil {
		return fmt.Errorf("failed to get stale reviews: %w", err)
	}
	
	// Refresh each game
	for _, gameID := range staleGameIDs {
		if err := s.RefreshReviewScores(ctx, gameID); err != nil {
			// Log error but continue with other games
			continue
		}
	}
	
	return nil
}
