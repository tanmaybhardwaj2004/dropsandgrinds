package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

// BuyTimingRecommendation represents a buy timing recommendation
type BuyTimingRecommendation struct {
	GameID         int64                    `json:"game_id"`
	Recommendation string                   `json:"recommendation"` // "buy_now", "wait_soon", "wait_next"
	Reason         string                   `json:"reason"`
	ActiveSale     *repositories.SaleEvent  `json:"active_sale,omitempty"`
	NextSale       *repositories.SaleEvent  `json:"next_sale,omitempty"`
	DaysUntilSale  int                      `json:"days_until_sale,omitempty"`
	SaleCalendar   []repositories.SaleEvent `json:"sale_calendar"`
	CheckedAt      time.Time                `json:"checked_at"`
}

// BuyTimingService handles buy timing recommendations
type BuyTimingService struct {
	salesRepo *repositories.SalesCalendarRepository
	logger    *slog.Logger
}

// NewBuyTimingService creates a new buy timing service
func NewBuyTimingService(salesRepo *repositories.SalesCalendarRepository, logger *slog.Logger) *BuyTimingService {
	return &BuyTimingService{
		salesRepo: salesRepo,
		logger:    logger,
	}
}

// GetBuyTiming returns a buy timing recommendation for a game
func (s *BuyTimingService) GetBuyTiming(ctx context.Context, gameID int64) (*BuyTimingRecommendation, error) {
	now := time.Now()

	// Get active sales
	activeSales, err := s.salesRepo.GetActiveSales(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sales: %w", err)
	}

	// Get upcoming sales
	upcomingSales, err := s.salesRepo.GetUpcomingSales(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming sales: %w", err)
	}

	// Get all sales for calendar view
	allSales, err := s.salesRepo.GetAllSales(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales calendar: %w", err)
	}

	recommendation := &BuyTimingRecommendation{
		GameID:       gameID,
		SaleCalendar: allSales,
		CheckedAt:    now,
	}

	// Check if there's an active sale
	if len(activeSales) > 0 {
		// Find the sale ending soonest
		soonestEnding := activeSales[0]
		for _, sale := range activeSales {
			if sale.EndDate.Before(soonestEnding.EndDate) {
				soonestEnding = sale
			}
		}

		daysRemaining := int(soonestEnding.EndDate.Sub(now).Hours() / 24)
		recommendation.ActiveSale = &soonestEnding
		recommendation.Recommendation = "buy_now"
		recommendation.Reason = fmt.Sprintf("%s is ON SALE NOW - ends in %d days", soonestEnding.Name, daysRemaining)
		return recommendation, nil
	}

	// Check for upcoming sales within 30 days
	if len(upcomingSales) > 0 {
		nextSale := upcomingSales[0]
		daysUntil := int(nextSale.StartDate.Sub(now).Hours() / 24)

		recommendation.NextSale = &nextSale
		recommendation.DaysUntilSale = daysUntil

		if daysUntil <= 30 {
			recommendation.Recommendation = "wait_soon"
			recommendation.Reason = fmt.Sprintf("Wait - %s starts in %d days", nextSale.Name, daysUntil)
		} else {
			recommendation.Recommendation = "wait_next"
			recommendation.Reason = fmt.Sprintf("Wait for %s (%d days away)", nextSale.Name, daysUntil)
		}
		return recommendation, nil
	}

	// No upcoming sales
	recommendation.Recommendation = "wait_next"
	recommendation.Reason = "No upcoming sales scheduled. Check back later."
	return recommendation, nil
}

// GetActiveSales returns currently active sales
func (s *BuyTimingService) GetActiveSales(ctx context.Context) ([]repositories.SaleEvent, error) {
	return s.salesRepo.GetActiveSales(ctx)
}

// GetSalesCalendar returns the full sales calendar
func (s *BuyTimingService) GetSalesCalendar(ctx context.Context) ([]repositories.SaleEvent, error) {
	return s.salesRepo.GetAllSales(ctx)
}
