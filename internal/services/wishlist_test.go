package services

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

type fakeWishlistStore struct {
	createFunc func(ctx context.Context, userID, gameID int64, targetPriceINR int) (models.WishlistItem, error)
	listFunc   func(ctx context.Context, userID int64, limit, offset int) ([]models.WishlistItem, int, error)
	updateFunc func(ctx context.Context, userID, wishlistID int64, targetPriceINR int) (models.WishlistItem, bool, error)
	deleteFunc func(ctx context.Context, userID, wishlistID int64) (bool, error)
}

func (f *fakeWishlistStore) CreateWishlistItem(ctx context.Context, userID, gameID int64, targetPriceINR int) (models.WishlistItem, error) {
	return f.createFunc(ctx, userID, gameID, targetPriceINR)
}

func (f *fakeWishlistStore) ListWishlistItems(ctx context.Context, userID int64, limit, offset int) ([]models.WishlistItem, int, error) {
	return f.listFunc(ctx, userID, limit, offset)
}

func (f *fakeWishlistStore) UpdateWishlistTarget(ctx context.Context, userID, wishlistID int64, targetPriceINR int) (models.WishlistItem, bool, error) {
	return f.updateFunc(ctx, userID, wishlistID, targetPriceINR)
}

func (f *fakeWishlistStore) DeleteWishlistItem(ctx context.Context, userID, wishlistID int64) (bool, error) {
	return f.deleteFunc(ctx, userID, wishlistID)
}

func TestWishlistService_AddItem_MapsDuplicateConflict(t *testing.T) {
	svc := NewWishlistService(&fakeWishlistStore{
		createFunc: func(context.Context, int64, int64, int) (models.WishlistItem, error) {
			return models.WishlistItem{}, &pgconn.PgError{Code: "23505"}
		},
		listFunc: func(context.Context, int64, int, int) ([]models.WishlistItem, int, error) { return nil, 0, nil },
		updateFunc: func(context.Context, int64, int64, int) (models.WishlistItem, bool, error) {
			return models.WishlistItem{}, false, nil
		},
		deleteFunc: func(context.Context, int64, int64) (bool, error) { return false, nil },
	})

	_, err := svc.AddItem(context.Background(), 1, models.WishlistCreateRequest{GameID: 10, TargetPriceINR: 999})
	if err == nil {
		t.Fatal("expected error")
	}
	se, ok := err.(*ServiceError)
	if !ok || se.StatusCode != 409 {
		t.Fatalf("expected 409 service error, got %#v", err)
	}
}

func TestWishlistService_ListItems_NormalizesPagination(t *testing.T) {
	svc := NewWishlistService(&fakeWishlistStore{
		createFunc: func(context.Context, int64, int64, int) (models.WishlistItem, error) {
			return models.WishlistItem{}, nil
		},
		listFunc: func(context.Context, int64, int, int) ([]models.WishlistItem, int, error) {
			return []models.WishlistItem{{ID: 1, GameID: 10}}, 1, nil
		},
		updateFunc: func(context.Context, int64, int64, int) (models.WishlistItem, bool, error) {
			return models.WishlistItem{}, false, nil
		},
		deleteFunc: func(context.Context, int64, int64) (bool, error) { return false, nil },
	})

	res, err := svc.ListItems(context.Background(), 1, 1000, -4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Limit != 100 || res.Offset != 0 {
		t.Fatalf("unexpected pagination normalization: %+v", res)
	}
}

func TestWishlistService_DeleteItem_NotFound(t *testing.T) {
	svc := NewWishlistService(&fakeWishlistStore{
		createFunc: func(context.Context, int64, int64, int) (models.WishlistItem, error) {
			return models.WishlistItem{}, nil
		},
		listFunc: func(context.Context, int64, int, int) ([]models.WishlistItem, int, error) { return nil, 0, nil },
		updateFunc: func(context.Context, int64, int64, int) (models.WishlistItem, bool, error) {
			return models.WishlistItem{}, false, nil
		},
		deleteFunc: func(context.Context, int64, int64) (bool, error) { return false, errors.New("db down") },
	})

	err := svc.DeleteItem(context.Background(), 1, 7)
	if err == nil {
		t.Fatal("expected error")
	}
	se, ok := err.(*ServiceError)
	if !ok || se.StatusCode != 500 {
		t.Fatalf("expected 500 service error, got %#v", err)
	}
}
