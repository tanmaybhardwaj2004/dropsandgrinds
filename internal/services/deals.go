package services

import (
	"context"
	"fmt"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

// DealQuality represents the quality tier of a deal
type DealQuality string

const (
	DealQualityHot  DealQuality = "hot"  // Excellent deal (high discount + all-time low)
	DealQualityGood DealQuality = "good" // Good deal (decent discount)
	DealQualityMeh  DealQuality = "meh"  // Average deal (low discount)
)

// DealEvaluationService evaluates deal quality
type DealEvaluationService struct {
	repo *repositories.CatalogRepository
}

// NewDealEvaluationService creates a new deal evaluation service
func NewDealEvaluationService(repo *repositories.CatalogRepository) *DealEvaluationService {
	return &DealEvaluationService{repo: repo}
}

// EvaluateDeal evaluates the quality of a deal
func (s *DealEvaluationService) EvaluateDeal(ctx context.Context, gameID int64) (DealQuality, string, error) {
	// Get game with current price and historical low
	game, found, err := s.repo.GetGameByID(ctx, gameID)
	if err != nil {
		return DealQualityMeh, "", err
	}
	if !found {
		return DealQualityMeh, "", fmt.Errorf("game not found")
	}

	// Determine deal quality based on discount and historical low
	discountPercent := game.DiscountPercent
	isAllTimeLow := game.IsAllTimeLow

	quality := DealQualityMeh
	reason := ""

	if isAllTimeLow && discountPercent >= 50 {
		quality = DealQualityHot
		reason = "All-time low with 50%+ discount"
	} else if isAllTimeLow && discountPercent >= 30 {
		quality = DealQualityHot
		reason = "All-time low with 30%+ discount"
	} else if discountPercent >= 70 {
		quality = DealQualityHot
		reason = "Massive 70%+ discount"
	} else if discountPercent >= 50 {
		quality = DealQualityGood
		reason = "Great 50%+ discount"
	} else if discountPercent >= 30 {
		quality = DealQualityGood
		reason = "Good 30%+ discount"
	} else if discountPercent >= 15 {
		quality = DealQualityGood
		reason = "Decent discount"
	} else {
		quality = DealQualityMeh
		reason = "Low discount"
	}

	return quality, reason, nil
}

// GetDealsForYou returns personalized deals based on user's wishlist and click history
func (s *DealEvaluationService) GetDealsForYou(ctx context.Context, userID int64, limit, offset int) ([]models.Deal, int, error) {
	// For MVP: return top deals sorted by discount
	// In production: filter by wishlist + click history
	return s.repo.ListDeals(ctx, limit, offset)
}
