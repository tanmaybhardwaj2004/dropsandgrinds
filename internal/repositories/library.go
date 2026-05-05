package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// LibraryRepository handles Steam library operations
type LibraryRepository struct {
	pool *pgxpool.Pool
}

// NewLibraryRepository creates a new library repository
func NewLibraryRepository(pool *pgxpool.Pool) *LibraryRepository {
	return &LibraryRepository{pool: pool}
}

// ImportOwnedGames imports a list of Steam app IDs for a user
func (r *LibraryRepository) ImportOwnedGames(ctx context.Context, userID int64, appIDs []int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing owned games for this user
	_, err = tx.Exec(ctx, "DELETE FROM user_steam_library WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete existing library: %w", err)
	}

	// Insert new owned games
	for _, appID := range appIDs {
		_, err = tx.Exec(ctx, `
			INSERT INTO user_steam_library (user_id, steam_app_id, imported_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (user_id, steam_app_id) DO NOTHING
		`, userID, appID)
		if err != nil {
			return fmt.Errorf("failed to insert app ID %d: %w", appID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetOwnedGames returns the list of owned Steam app IDs for a user
func (r *LibraryRepository) GetOwnedGames(ctx context.Context, userID int64) ([]int64, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT steam_app_id FROM user_steam_library WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query owned games: %w", err)
	}
	defer rows.Close()

	var appIDs []int64
	for rows.Next() {
		var appID int64
		if err := rows.Scan(&appID); err != nil {
			return nil, fmt.Errorf("failed to scan app ID: %w", err)
		}
		appIDs = append(appIDs, appID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return appIDs, nil
}

// IsGameOwned checks if a user owns a specific game (by game_id)
func (r *LibraryRepository) IsGameOwned(ctx context.Context, userID int64, gameID int64) (bool, error) {
	var owned bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM user_steam_library 
			WHERE user_id = $1 AND game_id = $2
		)
	`, userID, gameID).Scan(&owned)
	if err != nil {
		return false, fmt.Errorf("failed to check if game is owned: %w", err)
	}
	return owned, nil
}

// LinkSteamAppToGame links a Steam app ID to a game_id
func (r *LibraryRepository) LinkSteamAppToGame(ctx context.Context, userID int64, steamAppID int64, gameID int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE user_steam_library 
		SET game_id = $1 
		WHERE user_id = $2 AND steam_app_id = $3
	`, gameID, userID, steamAppID)
	if err != nil {
		return fmt.Errorf("failed to link Steam app to game: %w", err)
	}
	return nil
}

// GetOwnedGameIDs returns the list of game_ids (not steam_app_ids) for a user
func (r *LibraryRepository) GetOwnedGameIDs(ctx context.Context, userID int64) ([]int64, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT game_id FROM user_steam_library 
		WHERE user_id = $1 AND game_id IS NOT NULL
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query owned game IDs: %w", err)
	}
	defer rows.Close()

	var gameIDs []int64
	for rows.Next() {
		var gameID int64
		if err := rows.Scan(&gameID); err != nil {
			return nil, fmt.Errorf("failed to scan game ID: %w", err)
		}
		gameIDs = append(gameIDs, gameID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return gameIDs, nil
}

// GetLibraryCount returns the count of owned games for a user
func (r *LibraryRepository) GetLibraryCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM user_steam_library WHERE user_id = $1
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get library count: %w", err)
	}
	return count, nil
}

func (r *LibraryRepository) GetOwnedGameTitles(ctx context.Context, userID int64) (map[int64]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT g.id, g.title
		FROM user_steam_library l
		JOIN games g ON g.id = l.game_id
		WHERE l.user_id = $1 AND l.game_id IS NOT NULL
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]string{}
	for rows.Next() {
		var id int64
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			return nil, err
		}
		out[id] = title
	}
	return out, rows.Err()
}
