package services

import (
	"context"
	"sort"
	"strings"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type CatalogStore interface {
	ListGames(ctx context.Context, query, platform string, limit, offset int, excludeOwned bool, userID int64) ([]models.Game, int, error)
	SearchGames(ctx context.Context, query string, platform string, minPrice, maxPrice float64, minDiscount, maxDiscount int, minReviewScore, maxReviewScore float64, paymentMethod string, limit, offset int) ([]models.Game, int, error)
	GetGameByID(ctx context.Context, id int64) (models.Game, bool, error)
	ListDeals(ctx context.Context, limit, offset int) ([]models.Deal, int, error)
	GetPriceHistory(ctx context.Context, gameID int64, limit, offset int) ([]models.PriceHistoryPoint, error)
	GetIndiaArbitrage(ctx context.Context, gameID int64) (models.IndiaArbitrage, error)
}

type GameFilter struct {
	Query        string
	Platform     string
	Limit        int
	Offset       int
	ExcludeOwned bool
	UserID       int64
}

type GamesService struct {
	repo CatalogStore
}

func NewGamesService(repo CatalogStore) *GamesService {
	return &GamesService{repo: repo}
}

func (s *GamesService) ListGames(ctx context.Context, filter GameFilter) (models.GameListResponse, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := strings.ToLower(strings.TrimSpace(filter.Query))
	platform := strings.ToLower(strings.TrimSpace(filter.Platform))

	games, total, err := s.repo.ListGames(ctx, query, platform, limit, offset, filter.ExcludeOwned, filter.UserID)
	if err != nil {
		return models.GameListResponse{}, &ServiceError{StatusCode: 500, Message: "Failed to list games"}
	}

	return models.GameListResponse{
		Games:  games,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *GamesService) GetGameByID(ctx context.Context, id int64) (models.Game, bool, error) {
	game, ok, err := s.repo.GetGameByID(ctx, id)
	if err != nil {
		return models.Game{}, false, &ServiceError{StatusCode: 500, Message: "Failed to get game"}
	}
	return game, ok, nil
}

func (s *GamesService) ListDeals(ctx context.Context, limit, offset int) (models.DealListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	deals, total, err := s.repo.ListDeals(ctx, limit, offset)
	if err != nil {
		return models.DealListResponse{}, &ServiceError{StatusCode: 500, Message: "Failed to list deals"}
	}

	for i := range deals {
		deals[i].DealStatus, deals[i].PotentialSavingsINR = evaluateDeal(deals[i])
		deals[i].DealQuality = dealQuality(deals[i])
	}

	return models.DealListResponse{
		Deals:  deals,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func dealQuality(deal models.Deal) string {
	if deal.IsAllTimeLow || deal.DiscountPercent >= 70 {
		return "hot"
	}
	if deal.DiscountPercent >= 30 {
		return "good"
	}
	return "meh"
}

func evaluateDeal(deal models.Deal) (string, int) {
	if deal.OriginalINR <= 0 || deal.PriceINR <= 0 {
		return "unknown", 0
	}

	savings := deal.OriginalINR - deal.PriceINR
	if savings <= 0 {
		return "poor", 0
	}

	if deal.IsAllTimeLow {
		return "excellent", savings
	}
	if deal.DiscountPercent >= 50 {
		return "good", savings
	}
	if deal.DiscountPercent >= 20 {
		return "fair", savings
	}

	return "poor", savings
}

func (s *GamesService) GetPriceHistory(ctx context.Context, gameID int64, limit, offset int) (models.PriceHistoryResponse, error) {
	if gameID <= 0 {
		return models.PriceHistoryResponse{}, &ServiceError{StatusCode: 400, Message: "Invalid game id"}
	}
	if limit <= 0 {
		limit = 30
	}
	if limit > 365 {
		limit = 365
	}

	history, err := s.repo.GetPriceHistory(ctx, gameID, limit, offset)
	if err != nil {
		return models.PriceHistoryResponse{}, &ServiceError{StatusCode: 500, Message: "Failed to fetch price history"}
	}

	return models.PriceHistoryResponse{GameID: gameID, History: history, Prices: history}, nil
}

func (s *GamesService) GetIndiaArbitrage(ctx context.Context, gameID int64) (models.IndiaArbitrage, error) {
	if gameID <= 0 {
		return models.IndiaArbitrage{}, &ServiceError{StatusCode: 400, Message: "Invalid game id"}
	}

	arbitrage, err := s.repo.GetIndiaArbitrage(ctx, gameID)
	if err != nil {
		return models.IndiaArbitrage{}, &ServiceError{StatusCode: 500, Message: "Failed to fetch India arbitrage data"}
	}

	return arbitrage, nil
}

func (s *GamesService) SearchGames(ctx context.Context, query string, platform string, minPrice, maxPrice float64, minDiscount, maxDiscount int, minReviewScore, maxReviewScore float64, paymentMethod string, limit, offset int) ([]models.Game, int, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	games, total, err := s.repo.SearchGames(ctx, query, platform, minPrice, maxPrice, minDiscount, maxDiscount, minReviewScore, maxReviewScore, strings.ToLower(strings.TrimSpace(paymentMethod)), limit, offset)
	if err != nil {
		return nil, 0, &ServiceError{StatusCode: 500, Message: "Failed to search games"}
	}

	return games, total, nil
}

func (s *GamesService) GetBuyAdvice(ctx context.Context, gameID int64) (models.BuyAdviceResponse, error) {
	if gameID <= 0 {
		return models.BuyAdviceResponse{}, &ServiceError{StatusCode: 400, Message: "Invalid game id"}
	}

	history, err := s.repo.GetPriceHistory(ctx, gameID, 90, 0)
	if err != nil {
		return models.BuyAdviceResponse{}, &ServiceError{StatusCode: 500, Message: "Failed to fetch price history"}
	}
	if len(history) == 0 {
		return models.BuyAdviceResponse{
			GameID:            gameID,
			Recommendation:    "unknown",
			ConfidencePercent: 0,
			Reason:            "No price history available yet.",
		}, nil
	}

	current := history[0].PriceINR
	prices := make([]int, 0, len(history))
	lowest := history[0].PriceINR
	total := 0
	for _, point := range history {
		prices = append(prices, point.PriceINR)
		total += point.PriceINR
		if point.PriceINR < lowest {
			lowest = point.PriceINR
		}
	}
	avg := total / len(prices)

	recommendation := "buy_now"
	confidence := 65
	reason := "Current price is close to historical lows."

	sort.Ints(prices)
	p25 := prices[(len(prices)-1)/4]
	if current > avg || current > p25 {
		recommendation = "wait"
		confidence = 78
		reason = "Current price is above favorable historical range."
	}
	if current <= lowest {
		recommendation = "buy_now"
		confidence = 90
		reason = "Current price matches historical low."
	}

	return models.BuyAdviceResponse{
		GameID:            gameID,
		CurrentPriceINR:   current,
		LowestPriceINR:    lowest,
		AveragePriceINR:   avg,
		Recommendation:    recommendation,
		ConfidencePercent: confidence,
		Reason:            reason,
	}, nil
}
