package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

// ArbitrageService handles India vs Global price comparison
type ArbitrageService struct {
	catalogRepo  *repositories.CatalogRepository
	logger       *slog.Logger
	exchangeRate float64 // USD to INR exchange rate
	gstRate      float64 // GST rate (typically 0.18 for 18%)
}

// NewArbitrageService creates a new arbitrage service
func NewArbitrageService(
	catalogRepo *repositories.CatalogRepository,
	logger *slog.Logger,
	exchangeRate, gstRate float64,
) *ArbitrageService {
	return &ArbitrageService{
		catalogRepo:  catalogRepo,
		logger:       logger,
		exchangeRate: exchangeRate,
		gstRate:      gstRate,
	}
}

// CalculateArbitrage calculates India vs Global price comparison
func (s *ArbitrageService) CalculateArbitrage(ctx context.Context, gameID int64) (*models.ArbitrageData, error) {
	// Get game details
	game, found, err := s.catalogRepo.GetGameByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch game: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("game not found")
	}

	// Current India price (from database)
	indiaBaseINR := float64(game.PriceINR)

	// Calculate GST on India price (18%)
	indiaGSTINR := indiaBaseINR * s.gstRate
	indiaTotalINR := indiaBaseINR + indiaGSTINR

	// Global base price (assuming USD base, convert to INR)
	// For MVP, we'll estimate global base from original price
	// In production, this would fetch actual USD price from Steam API
	globalBaseUSD := float64(game.OriginalINR) / s.exchangeRate
	globalBaseINR := globalBaseUSD * s.exchangeRate

	// Calculate GST on global price (18%)
	globalGSTINR := globalBaseINR * s.gstRate
	globalTotalINR := globalBaseINR + globalGSTINR

	// Determine cheapest region
	cheapestRegion := "india"
	verdict := fmt.Sprintf("Buy from India - saves ₹%.0f", globalTotalINR-indiaTotalINR)

	if globalTotalINR < indiaTotalINR {
		cheapestRegion = "global"
		verdict = fmt.Sprintf("Buy from Global region - saves ₹%.0f", indiaTotalINR-globalTotalINR)
	} else if indiaTotalINR == globalTotalINR {
		verdict = "Prices are equal - buy from preferred region"
	}

	return &models.ArbitrageData{
		IndiaBaseINR:   indiaBaseINR,
		IndiaGSTINR:    indiaGSTINR,
		IndiaTotalINR:  indiaTotalINR,
		GlobalBaseINR:  globalBaseINR,
		GlobalGSTINR:   globalGSTINR,
		GlobalTotalINR: globalTotalINR,
		CheapestRegion: cheapestRegion,
		Verdict:        verdict,
	}, nil
}
