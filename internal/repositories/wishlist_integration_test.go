package repositories

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestWishlistRepositoryIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping wishlist integration tests")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to connect test database: %v", err)
	}
	defer pool.Close()

	if err := resetAndMigrateTestDB(ctx, pool); err != nil {
		t.Fatalf("failed to prepare test database: %v", err)
	}

	repo := NewWishlistRepository(pool)

	t.Run("CreateWishlistItem", func(t *testing.T) {
		item, err := repo.CreateWishlistItem(ctx, 1, 1, 999)
		if err != nil {
			t.Fatalf("CreateWishlistItem returned error: %v", err)
		}
		if item.ID <= 0 || item.UserID != 1 || item.GameID != 1 || item.TargetPriceINR != 999 {
			t.Fatalf("unexpected wishlist item: %+v", item)
		}
	})

	t.Run("ListWishlistItems", func(t *testing.T) {
		items, total, err := repo.ListWishlistItems(ctx, 1, 20, 0)
		if err != nil {
			t.Fatalf("ListWishlistItems returned error: %v", err)
		}
		if total <= 0 || len(items) <= 0 {
			t.Fatalf("expected non-empty wishlist, total=%d len=%d", total, len(items))
		}
	})

	t.Run("UpdateWishlistTarget", func(t *testing.T) {
		item, found, err := repo.UpdateWishlistTarget(ctx, 1, 1, 799)
		if err != nil {
			t.Fatalf("UpdateWishlistTarget returned error: %v", err)
		}
		if !found {
			t.Fatal("expected wishlist id 1 to be found")
		}
		if item.TargetPriceINR != 799 {
			t.Fatalf("expected target price updated to 799, got %d", item.TargetPriceINR)
		}
	})

	t.Run("DeleteWishlistItem", func(t *testing.T) {
		deleted, err := repo.DeleteWishlistItem(ctx, 1, 1)
		if err != nil {
			t.Fatalf("DeleteWishlistItem returned error: %v", err)
		}
		if !deleted {
			t.Fatal("expected deletion to succeed")
		}

		items, _, err := repo.ListWishlistItems(ctx, 1, 20, 0)
		if err != nil {
			t.Fatalf("ListWishlistItems after delete returned error: %v", err)
		}
		if len(items) > 0 {
			t.Fatalf("expected empty wishlist after delete, got %d items", len(items))
		}
	})

	t.Run("CreateWishlistItem_DuplicateGameError", func(t *testing.T) {
		_, err := repo.CreateWishlistItem(ctx, 1, 2, 1299)
		if err != nil {
			t.Fatalf("first CreateWishlistItem returned error: %v", err)
		}
		_, err = repo.CreateWishlistItem(ctx, 1, 2, 999)
		if err == nil {
			t.Fatal("expected duplicate key error")
		}
	})
}
