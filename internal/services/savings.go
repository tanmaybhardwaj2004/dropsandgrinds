package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

type SavingsStore interface {
	LogPurchase(ctx context.Context, userID, gameID int64, gameTitle string, paidPriceINR, originalPriceINR int) error
	GetTotalSavings(ctx context.Context, userID int64) (int, error)
	GetMonthlyBreakdown(ctx context.Context, userID int64) ([]repositories.MonthlySavings, error)
	GetPurchaseHistory(ctx context.Context, userID int64, limit, offset int) ([]repositories.PurchaseRecord, int, error)
}

type SavingsService struct {
	repo SavingsStore
}

func NewSavingsService(repo SavingsStore) *SavingsService {
	return &SavingsService{repo: repo}
}

// LogPurchaseRequest represents a purchase log request
type LogPurchaseRequest struct {
	GameID           int64  `json:"game_id" binding:"required"`
	GameTitle        string `json:"game_title" binding:"required"`
	PaidPriceINR     int    `json:"paid_price_inr" binding:"required"`
	OriginalPriceINR int    `json:"original_price_inr" binding:"required"`
}

// SavingsResponse represents the savings summary
type SavingsResponse struct {
	TotalSavings     int                           `json:"total_savings"`
	MonthlyBreakdown []repositories.MonthlySavings `json:"monthly_breakdown"`
	EquivalentGames  string                        `json:"equivalent_games"`
}

func (s *SavingsService) LogPurchase(ctx context.Context, userID int64, req LogPurchaseRequest) error {
	if userID <= 0 {
		return &ServiceError{StatusCode: http.StatusUnauthorized, Message: "Unauthorized"}
	}
	if req.GameID <= 0 {
		return &ServiceError{StatusCode: http.StatusBadRequest, Message: "Invalid game id"}
	}
	if req.PaidPriceINR <= 0 {
		return &ServiceError{StatusCode: http.StatusBadRequest, Message: "paid_price_inr must be > 0"}
	}
	if req.OriginalPriceINR <= 0 {
		return &ServiceError{StatusCode: http.StatusBadRequest, Message: "original_price_inr must be > 0"}
	}
	if req.PaidPriceINR > req.OriginalPriceINR {
		return &ServiceError{StatusCode: http.StatusBadRequest, Message: "paid_price_inr cannot be greater than original_price_inr"}
	}

	err := s.repo.LogPurchase(ctx, userID, req.GameID, req.GameTitle, req.PaidPriceINR, req.OriginalPriceINR)
	if err != nil {
		return &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to log purchase"}
	}

	return nil
}

func (s *SavingsService) GetSavings(ctx context.Context, userID int64) (SavingsResponse, error) {
	if userID <= 0 {
		return SavingsResponse{}, &ServiceError{StatusCode: http.StatusUnauthorized, Message: "Unauthorized"}
	}

	totalSavings, err := s.repo.GetTotalSavings(ctx, userID)
	if err != nil {
		return SavingsResponse{}, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to get total savings"}
	}

	monthlyBreakdown, err := s.repo.GetMonthlyBreakdown(ctx, userID)
	if err != nil {
		return SavingsResponse{}, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to get monthly breakdown"}
	}

	// Calculate equivalent free games (assuming average game price ₹500)
	equivalentGames := calculateEquivalentGames(totalSavings)

	return SavingsResponse{
		TotalSavings:     totalSavings,
		MonthlyBreakdown: monthlyBreakdown,
		EquivalentGames:  equivalentGames,
	}, nil
}

func (s *SavingsService) GetPurchaseHistory(ctx context.Context, userID int64, limit, offset int) ([]repositories.PurchaseRecord, int, error) {
	if userID <= 0 {
		return nil, 0, &ServiceError{StatusCode: http.StatusUnauthorized, Message: "Unauthorized"}
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	purchases, total, err := s.repo.GetPurchaseHistory(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to get purchase history"}
	}

	return purchases, total, nil
}

// calculateEquivalentGames calculates how many free games the savings are equivalent to
// Assuming average game price of ₹500
func calculateEquivalentGames(totalSavings int) string {
	if totalSavings <= 0 {
		return "You haven't saved any money yet. Keep tracking your deals!"
	}

	avgGamePrice := 500
	equivalent := totalSavings / avgGamePrice

	switch {
	case equivalent == 0:
		return "You're on your way to your first free game!"
	case equivalent == 1:
		return "You saved enough for 1 free game."
	case equivalent < 5:
		return fmt.Sprintf("You saved enough for %d free games.", equivalent)
	case equivalent < 10:
		return fmt.Sprintf("You saved enough for %d free games. That's impressive.", equivalent)
	default:
		return fmt.Sprintf("You saved enough for %d free games. Excellent deal tracking.", equivalent)
	}
}
