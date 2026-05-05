package services

import (
	"context"
	"errors"
	"testing"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type fakeCatalogStore struct {
	listGamesFunc       func(ctx context.Context, query, platform string, limit, offset int, excludeOwned bool, userID int64) ([]models.Game, int, error)
	searchGamesFunc     func(ctx context.Context, query string, platform string, minPrice, maxPrice float64, minDiscount, maxDiscount int, minReviewScore, maxReviewScore float64, limit, offset int) ([]models.Game, int, error)
	getGameByIDFunc     func(ctx context.Context, id int64) (models.Game, bool, error)
	listDealsFunc       func(ctx context.Context, limit, offset int) ([]models.Deal, int, error)
	getPriceHistoryFunc func(ctx context.Context, gameID int64, limit, offset int) ([]models.PriceHistoryPoint, error)
}

func (f *fakeCatalogStore) ListGames(ctx context.Context, query, platform string, limit, offset int, excludeOwned bool, userID int64) ([]models.Game, int, error) {
	return f.listGamesFunc(ctx, query, platform, limit, offset, excludeOwned, userID)
}

func (f *fakeCatalogStore) SearchGames(ctx context.Context, query string, platform string, minPrice, maxPrice float64, minDiscount, maxDiscount int, minReviewScore, maxReviewScore float64, limit, offset int) ([]models.Game, int, error) {
	if f.searchGamesFunc == nil {
		return nil, 0, nil
	}
	return f.searchGamesFunc(ctx, query, platform, minPrice, maxPrice, minDiscount, maxDiscount, minReviewScore, maxReviewScore, limit, offset)
}

func (f *fakeCatalogStore) GetGameByID(ctx context.Context, id int64) (models.Game, bool, error) {
	return f.getGameByIDFunc(ctx, id)
}

func (f *fakeCatalogStore) ListDeals(ctx context.Context, limit, offset int) ([]models.Deal, int, error) {
	return f.listDealsFunc(ctx, limit, offset)
}

func (f *fakeCatalogStore) GetPriceHistory(ctx context.Context, gameID int64, limit, offset int) ([]models.PriceHistoryPoint, error) {
	return f.getPriceHistoryFunc(ctx, gameID, limit, offset)
}

func (f *fakeCatalogStore) GetIndiaArbitrage(ctx context.Context, gameID int64) (models.IndiaArbitrage, error) {
	return models.IndiaArbitrage{}, nil
}

func TestGamesService_ListGames_NormalizesLimitOffset(t *testing.T) {
	store := &fakeCatalogStore{
		listGamesFunc: func(ctx context.Context, query, platform string, limit, offset int, excludeOwned bool, userID int64) ([]models.Game, int, error) {
			if limit != 100 {
				t.Fatalf("expected limit 100, got %d", limit)
			}
			if offset != 0 {
				t.Fatalf("expected offset 0, got %d", offset)
			}
			return []models.Game{{ID: 1, Title: "Test"}}, 1, nil
		},
		getGameByIDFunc:     func(context.Context, int64) (models.Game, bool, error) { return models.Game{}, false, nil },
		listDealsFunc:       func(context.Context, int, int) ([]models.Deal, int, error) { return nil, 0, nil },
		getPriceHistoryFunc: func(context.Context, int64, int, int) ([]models.PriceHistoryPoint, error) { return nil, nil },
	}

	svc := NewGamesService(store)
	res, err := svc.ListGames(context.Background(), GameFilter{Limit: 1000, Offset: -5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Total != 1 || len(res.Games) != 1 {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestGamesService_ListGames_ErrorMapping(t *testing.T) {
	svc := NewGamesService(&fakeCatalogStore{
		listGamesFunc: func(context.Context, string, string, int, int, bool, int64) ([]models.Game, int, error) {
			return nil, 0, errors.New("db down")
		},
		getGameByIDFunc:     func(context.Context, int64) (models.Game, bool, error) { return models.Game{}, false, nil },
		listDealsFunc:       func(context.Context, int, int) ([]models.Deal, int, error) { return nil, 0, nil },
		getPriceHistoryFunc: func(context.Context, int64, int, int) ([]models.PriceHistoryPoint, error) { return nil, nil },
	})

	_, err := svc.ListGames(context.Background(), GameFilter{})
	if err == nil {
		t.Fatal("expected error")
	}
	serviceErr, ok := err.(*ServiceError)
	if !ok || serviceErr.StatusCode != 500 {
		t.Fatalf("expected service error 500, got %#v", err)
	}
}

func TestGamesService_GetPriceHistory_ValidatesInput(t *testing.T) {
	svc := NewGamesService(&fakeCatalogStore{
		listGamesFunc: func(context.Context, string, string, int, int, bool, int64) ([]models.Game, int, error) {
			return nil, 0, nil
		},
		getGameByIDFunc:     func(context.Context, int64) (models.Game, bool, error) { return models.Game{}, false, nil },
		listDealsFunc:       func(context.Context, int, int) ([]models.Deal, int, error) { return nil, 0, nil },
		getPriceHistoryFunc: func(context.Context, int64, int, int) ([]models.PriceHistoryPoint, error) { return nil, nil },
	})

	_, err := svc.GetPriceHistory(context.Background(), 0, 10, 0)
	if err == nil {
		t.Fatal("expected error")
	}
	serviceErr, ok := err.(*ServiceError)
	if !ok || serviceErr.StatusCode != 400 {
		t.Fatalf("expected service error 400, got %#v", err)
	}
}

func TestGamesService_ListDeals_AddsDealEvaluation(t *testing.T) {
	svc := NewGamesService(&fakeCatalogStore{
		listGamesFunc: func(context.Context, string, string, int, int, bool, int64) ([]models.Game, int, error) {
			return nil, 0, nil
		},
		getGameByIDFunc: func(context.Context, int64) (models.Game, bool, error) { return models.Game{}, false, nil },
		listDealsFunc: func(context.Context, int, int) ([]models.Deal, int, error) {
			return []models.Deal{{
				Game: models.Game{
					ID:              1,
					Title:           "Cyberpunk 2077",
					PriceINR:        1499,
					OriginalINR:     2999,
					DiscountPercent: 50,
					IsAllTimeLow:    true,
				},
			}}, 1, nil
		},
		getPriceHistoryFunc: func(context.Context, int64, int, int) ([]models.PriceHistoryPoint, error) { return nil, nil },
	})

	resp, err := svc.ListDeals(context.Background(), 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Deals) != 1 {
		t.Fatalf("expected one deal, got %d", len(resp.Deals))
	}
	if resp.Deals[0].DealStatus != "excellent" {
		t.Fatalf("expected deal status excellent, got %q", resp.Deals[0].DealStatus)
	}
	if resp.Deals[0].PotentialSavingsINR != 1500 {
		t.Fatalf("expected potential savings 1500, got %d", resp.Deals[0].PotentialSavingsINR)
	}
}

func TestGamesService_GetBuyAdvice_ReturnsRecommendation(t *testing.T) {
	svc := NewGamesService(&fakeCatalogStore{
		listGamesFunc: func(context.Context, string, string, int, int, bool, int64) ([]models.Game, int, error) {
			return nil, 0, nil
		},
		getGameByIDFunc: func(context.Context, int64) (models.Game, bool, error) { return models.Game{}, false, nil },
		listDealsFunc:   func(context.Context, int, int) ([]models.Deal, int, error) { return nil, 0, nil },
		getPriceHistoryFunc: func(context.Context, int64, int, int) ([]models.PriceHistoryPoint, error) {
			return []models.PriceHistoryPoint{
				{PriceINR: 1500, FetchedAt: "2026-04-22T10:00:00Z"},
				{PriceINR: 1300, FetchedAt: "2026-04-20T10:00:00Z"},
				{PriceINR: 1700, FetchedAt: "2026-04-18T10:00:00Z"},
			}, nil
		},
	})

	advice, err := svc.GetBuyAdvice(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if advice.Recommendation == "" || advice.ConfidencePercent <= 0 {
		t.Fatalf("unexpected advice payload: %+v", advice)
	}
}
