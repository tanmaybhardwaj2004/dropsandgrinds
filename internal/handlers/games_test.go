package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

type fakeCatalogStoreForHandler struct{}

func (f *fakeCatalogStoreForHandler) ListGames(ctx context.Context, query, platform string, limit, offset int, excludeOwned bool, userID int64) ([]models.Game, int, error) {
	return []models.Game{{ID: 1, Title: "Cyberpunk 2077", Platform: "Steam", PriceINR: 1499, LowestPriceINR: 999, IsAllTimeLow: false}}, 1, nil
}

func (f *fakeCatalogStoreForHandler) SearchGames(ctx context.Context, query string, platform string, minPrice, maxPrice float64, minDiscount, maxDiscount int, minReviewScore, maxReviewScore float64, limit, offset int) ([]models.Game, int, error) {
	return []models.Game{{ID: 1, Title: "Cyberpunk 2077", Platform: "Steam", PriceINR: 1499, LowestPriceINR: 999, IsAllTimeLow: false}}, 1, nil
}

func (f *fakeCatalogStoreForHandler) GetGameByID(ctx context.Context, id int64) (models.Game, bool, error) {
	if id == 1 {
		return models.Game{ID: 1, Title: "Cyberpunk 2077", Platform: "Steam", PriceINR: 1499, LowestPriceINR: 999, IsAllTimeLow: false}, true, nil
	}
	return models.Game{}, false, nil
}

func (f *fakeCatalogStoreForHandler) ListDeals(ctx context.Context, limit, offset int) ([]models.Deal, int, error) {
	return []models.Deal{{Game: models.Game{ID: 1, Title: "Cyberpunk 2077", Platform: "Steam", PriceINR: 1499, LowestPriceINR: 999, IsAllTimeLow: false, OriginalINR: 2999, DiscountPercent: 50}}}, 1, nil
}

func (f *fakeCatalogStoreForHandler) GetPriceHistory(ctx context.Context, gameID int64, limit, offset int) ([]models.PriceHistoryPoint, error) {
	return []models.PriceHistoryPoint{{PriceINR: 1499, FetchedAt: "2026-04-21T10:00:00Z"}}, nil
}

func (f *fakeCatalogStoreForHandler) GetIndiaArbitrage(ctx context.Context, gameID int64) (models.ArbitrageData, error) {
	return models.ArbitrageData{}, nil
}

func TestGamesListHandler_OK(t *testing.T) {
	SetGamesService(services.NewGamesService(&fakeCatalogStoreForHandler{}))
	req := httptest.NewRequest(http.MethodGet, "/api/games?limit=5&offset=0", nil)
	rr := httptest.NewRecorder()

	GamesListHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var payload models.GameListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if payload.Total != 1 || len(payload.Games) != 1 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestGameDetailHandler_NotFound(t *testing.T) {
	SetGamesService(services.NewGamesService(&fakeCatalogStoreForHandler{}))
	req := httptest.NewRequest(http.MethodGet, "/api/games/99", nil)
	rr := httptest.NewRecorder()

	GameDetailHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDealsListHandler_MethodNotAllowed(t *testing.T) {
	SetGamesService(services.NewGamesService(&fakeCatalogStoreForHandler{}))
	req := httptest.NewRequest(http.MethodPost, "/api/deals", nil)
	rr := httptest.NewRecorder()

	DealsListHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestGameDetailHandler_BuyAdviceSubPath_OK(t *testing.T) {
	SetGamesService(services.NewGamesService(&fakeCatalogStoreForHandler{}))
	req := httptest.NewRequest(http.MethodGet, "/api/games/1/buy-advice", nil)
	rr := httptest.NewRecorder()

	GameDetailHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var payload models.BuyAdviceResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if payload.GameID != 1 {
		t.Fatalf("expected game id 1, got %d", payload.GameID)
	}
}
