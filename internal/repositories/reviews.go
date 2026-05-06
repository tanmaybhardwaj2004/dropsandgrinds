package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/pkg/reviews"
)

// ReviewRepository handles review score data
type ReviewRepository struct {
	db *pgxpool.Pool
}

// NewReviewRepository creates a new review repository
func NewReviewRepository(db *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// StoreReviewScore stores a review score for a game
func (r *ReviewRepository) StoreReviewScore(ctx context.Context, gameID int64, score reviews.ReviewScore) error {
	query := `
		INSERT INTO review_scores (game_id, source, score, url, fetched_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (game_id, source) DO UPDATE SET
			score = EXCLUDED.score,
			url = EXCLUDED.url,
			fetched_at = EXCLUDED.fetched_at
	`

	_, err := r.db.Exec(ctx, query, gameID, string(score.Source), score.Score, score.URL, score.FetchedAt)
	return err
}

// GetReviewScores retrieves all review scores for a game
func (r *ReviewRepository) GetReviewScores(ctx context.Context, gameID int64) ([]reviews.ReviewScore, error) {
	query := `
		SELECT source, score, url, fetched_at
		FROM review_scores
		WHERE game_id = $1
		ORDER BY fetched_at DESC
	`

	rows, err := r.db.Query(ctx, query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []reviews.ReviewScore
	for rows.Next() {
		var score reviews.ReviewScore
		var source string
		err := rows.Scan(&source, &score.Score, &score.URL, &score.FetchedAt)
		if err != nil {
			return nil, err
		}
		score.Source = reviews.ReviewSource(source)
		scores = append(scores, score)
	}

	return scores, rows.Err()
}

// GetReviewScore retrieves a specific review score for a game
func (r *ReviewRepository) GetReviewScore(ctx context.Context, gameID int64, source reviews.ReviewSource) (*reviews.ReviewScore, error) {
	query := `
		SELECT source, score, url, fetched_at
		FROM review_scores
		WHERE game_id = $1 AND source = $2
	`

	var score reviews.ReviewScore
	err := r.db.QueryRow(ctx, query, gameID, string(source)).Scan(&score.Source, &score.Score, &score.URL, &score.FetchedAt)
	if err != nil {
		return nil, err
	}

	return &score, nil
}

// DeleteReviewScores deletes all review scores for a game
func (r *ReviewRepository) DeleteReviewScores(ctx context.Context, gameID int64) error {
	query := `DELETE FROM review_scores WHERE game_id = $1`
	_, err := r.db.Exec(ctx, query, gameID)
	return err
}

// GetStaleReviews returns games with review scores older than the specified duration
func (r *ReviewRepository) GetStaleReviews(ctx context.Context, olderThan time.Duration) ([]int64, error) {
	query := `
		SELECT DISTINCT game_id
		FROM review_scores
		WHERE fetched_at < $1
	`

	rows, err := r.db.Query(ctx, query, time.Now().Add(-olderThan))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gameIDs []int64
	for rows.Next() {
		var gameID int64
		if err := rows.Scan(&gameID); err != nil {
			return nil, err
		}
		gameIDs = append(gameIDs, gameID)
	}

	return gameIDs, rows.Err()
}

// GetReviewRefreshGameIDs returns games that need review refresh, including never-scored games.
func (r *ReviewRepository) GetReviewRefreshGameIDs(ctx context.Context, olderThan time.Duration) ([]int64, error) {
	query := `
		SELECT g.id
		FROM games g
		LEFT JOIN (
			SELECT game_id, MAX(fetched_at) AS last_fetched
			FROM review_scores
			GROUP BY game_id
		) rs ON rs.game_id = g.id
		WHERE rs.last_fetched IS NULL OR rs.last_fetched < $1
		ORDER BY g.id
	`

	rows, err := r.db.Query(ctx, query, time.Now().Add(-olderThan))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gameIDs []int64
	for rows.Next() {
		var gameID int64
		if err := rows.Scan(&gameID); err != nil {
			return nil, err
		}
		gameIDs = append(gameIDs, gameID)
	}

	return gameIDs, rows.Err()
}

// GetSteamAppID returns the external Steam app ID for a catalog game.
func (r *ReviewRepository) GetSteamAppID(ctx context.Context, gameID int64) (int64, bool, error) {
	var steamAppID int64
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(steam_app_id, 0)
		FROM games
		WHERE id = $1
	`, gameID).Scan(&steamAppID)
	if err != nil {
		return 0, false, err
	}
	return steamAppID, steamAppID > 0, nil
}
