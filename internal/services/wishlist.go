package services

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type WishlistStore interface {
	CreateWishlistItem(ctx context.Context, userID, gameID int64, targetPriceINR int) (models.WishlistItem, error)
	ListWishlistItems(ctx context.Context, userID int64, limit, offset int) ([]models.WishlistItem, int, error)
	UpdateWishlistTarget(ctx context.Context, userID, wishlistID int64, targetPriceINR int) (models.WishlistItem, bool, error)
	DeleteWishlistItem(ctx context.Context, userID, wishlistID int64) (bool, error)
}

type WishlistService struct {
	repo WishlistStore
}

func NewWishlistService(repo WishlistStore) *WishlistService {
	return &WishlistService{repo: repo}
}

func (s *WishlistService) AddItem(ctx context.Context, userID int64, req models.WishlistCreateRequest) (models.WishlistItem, error) {
	if userID <= 0 {
		return models.WishlistItem{}, &ServiceError{StatusCode: http.StatusUnauthorized, Message: "Unauthorized"}
	}
	if req.GameID <= 0 {
		return models.WishlistItem{}, &ServiceError{StatusCode: http.StatusBadRequest, Message: "Invalid game id"}
	}
	if req.TargetPriceINR <= 0 {
		return models.WishlistItem{}, &ServiceError{StatusCode: http.StatusBadRequest, Message: "target_price_inr must be > 0"}
	}

	item, err := s.repo.CreateWishlistItem(ctx, userID, req.GameID, req.TargetPriceINR)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return models.WishlistItem{}, &ServiceError{StatusCode: http.StatusConflict, Message: "Game already in wishlist"}
			}
			if pgErr.Code == "23503" {
				return models.WishlistItem{}, &ServiceError{StatusCode: http.StatusBadRequest, Message: "Game does not exist"}
			}
		}
		return models.WishlistItem{}, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to add wishlist item"}
	}

	return item, nil
}

func (s *WishlistService) ListItems(ctx context.Context, userID int64, limit, offset int) (models.WishlistListResponse, error) {
	if userID <= 0 {
		return models.WishlistListResponse{}, &ServiceError{StatusCode: http.StatusUnauthorized, Message: "Unauthorized"}
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

	items, total, err := s.repo.ListWishlistItems(ctx, userID, limit, offset)
	if err != nil {
		return models.WishlistListResponse{}, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to list wishlist"}
	}

	return models.WishlistListResponse{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *WishlistService) UpdateItem(ctx context.Context, userID, wishlistID int64, req models.WishlistUpdateRequest) (models.WishlistItem, error) {
	if userID <= 0 {
		return models.WishlistItem{}, &ServiceError{StatusCode: http.StatusUnauthorized, Message: "Unauthorized"}
	}
	if wishlistID <= 0 {
		return models.WishlistItem{}, &ServiceError{StatusCode: http.StatusBadRequest, Message: "Invalid wishlist id"}
	}
	if req.TargetPriceINR <= 0 {
		return models.WishlistItem{}, &ServiceError{StatusCode: http.StatusBadRequest, Message: "target_price_inr must be > 0"}
	}

	item, ok, err := s.repo.UpdateWishlistTarget(ctx, userID, wishlistID, req.TargetPriceINR)
	if err != nil {
		return models.WishlistItem{}, &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to update wishlist item"}
	}
	if !ok {
		return models.WishlistItem{}, &ServiceError{StatusCode: http.StatusNotFound, Message: "Wishlist item not found"}
	}

	return item, nil
}

func (s *WishlistService) DeleteItem(ctx context.Context, userID, wishlistID int64) error {
	if userID <= 0 {
		return &ServiceError{StatusCode: http.StatusUnauthorized, Message: "Unauthorized"}
	}
	if wishlistID <= 0 {
		return &ServiceError{StatusCode: http.StatusBadRequest, Message: "Invalid wishlist id"}
	}

	deleted, err := s.repo.DeleteWishlistItem(ctx, userID, wishlistID)
	if err != nil {
		return &ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to delete wishlist item"}
	}
	if !deleted {
		return &ServiceError{StatusCode: http.StatusNotFound, Message: "Wishlist item not found"}
	}

	return nil
}

func parseIDFromPath(path, prefix string) (int64, error) {
	idStr := strings.TrimPrefix(path, prefix)
	if idStr == path {
		return 0, errors.New("missing id")
	}
	if strings.Contains(idStr, "/") || idStr == "" {
		return 0, errors.New("invalid id path")
	}
	var value int64
	for _, ch := range idStr {
		if ch < '0' || ch > '9' {
			return 0, errors.New("invalid id")
		}
		value = (value * 10) + int64(ch-'0')
	}
	if value <= 0 {
		return 0, errors.New("invalid id")
	}
	return value, nil
}
