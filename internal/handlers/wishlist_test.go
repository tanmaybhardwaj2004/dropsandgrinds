package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/middleware"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

type fakeWishlistStoreForHandler struct{}

func (f *fakeWishlistStoreForHandler) CreateWishlistItem(ctx context.Context, userID, gameID int64, targetPriceINR int) (models.WishlistItem, error) {
	return models.WishlistItem{ID: 1, UserID: userID, GameID: gameID, TargetPriceINR: targetPriceINR, Title: "Test Game"}, nil
}

func (f *fakeWishlistStoreForHandler) ListWishlistItems(ctx context.Context, userID int64, limit, offset int) ([]models.WishlistItem, int, error) {
	return []models.WishlistItem{{ID: 1, UserID: userID, GameID: 10, TargetPriceINR: 999, Title: "Test Game"}}, 1, nil
}

func (f *fakeWishlistStoreForHandler) UpdateWishlistTarget(ctx context.Context, userID, wishlistID int64, targetPriceINR int) (models.WishlistItem, bool, error) {
	return models.WishlistItem{ID: wishlistID, UserID: userID, GameID: 10, TargetPriceINR: targetPriceINR, Title: "Test Game"}, true, nil
}

func (f *fakeWishlistStoreForHandler) DeleteWishlistItem(ctx context.Context, userID, wishlistID int64) (bool, error) {
	return true, nil
}

func testBearerToken(t *testing.T, secret string, sub string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": sub})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return "Bearer " + signed
}

func TestWishlistCollectionHandler_Create_OK(t *testing.T) {
	SetWishlistService(services.NewWishlistService(&fakeWishlistStoreForHandler{}))
	body, _ := json.Marshal(models.WishlistCreateRequest{GameID: 10, TargetPriceINR: 999})
	req := httptest.NewRequest(http.MethodPost, "/api/wishlist", bytes.NewReader(body))
	req.Header.Set("Authorization", testBearerToken(t, "test-secret", "1"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	middleware.RequireAuth([]byte("test-secret"), http.HandlerFunc(WishlistCollectionHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}

func TestWishlistItemHandler_Delete_NoContent(t *testing.T) {
	SetWishlistService(services.NewWishlistService(&fakeWishlistStoreForHandler{}))
	req := httptest.NewRequest(http.MethodDelete, "/api/wishlist/1", nil)
	req.Header.Set("Authorization", testBearerToken(t, "test-secret", "1"))
	rr := httptest.NewRecorder()

	middleware.RequireAuth([]byte("test-secret"), http.HandlerFunc(WishlistItemHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}
