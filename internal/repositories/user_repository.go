package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
)

// UserRepository handles user data operations
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	query := `
		SELECT id, email, username, first_name, last_name, 
		       is_active, email_verified, created_at, updated_at
		FROM users 
		WHERE id = $1
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found: %d", userID)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserEmail retrieves only the email for a user by ID
func (r *UserRepository) GetUserEmail(ctx context.Context, userID int64) (string, error) {
	query := `SELECT email FROM users WHERE id = $1 AND is_active = true`

	var email string
	err := r.pool.QueryRow(ctx, query, userID).Scan(&email)

	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("active user not found: %d", userID)
		}
		return "", fmt.Errorf("failed to get user email: %w", err)
	}

	return email, nil
}

// GetUserEmails retrieves emails for multiple users by their IDs
func (r *UserRepository) GetUserEmails(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	if len(userIDs) == 0 {
		return make(map[int64]string), nil
	}

	// Build the query with IN clause
	query := `
		SELECT id, email 
		FROM users 
		WHERE id = ANY($1) AND is_active = true
	`

	rows, err := r.pool.Query(ctx, query, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get user emails: %w", err)
	}
	defer rows.Close()

	userEmails := make(map[int64]string)
	for rows.Next() {
		var userID int64
		var email string
		if err := rows.Scan(&userID, &email); err != nil {
			return nil, fmt.Errorf("failed to scan user email: %w", err)
		}
		userEmails[userID] = email
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user emails: %w", err)
	}

	return userEmails, nil
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (email, username, first_name, last_name, is_active, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		user.Email,
		user.Username,
		user.FirstName,
		user.LastName,
		user.IsActive,
		user.EmailVerified,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET email = $1, username = $2, first_name = $3, last_name = $4, 
		    is_active = $5, email_verified = $6, updated_at = NOW()
		WHERE id = $7
	`

	result, err := r.pool.Exec(ctx, query,
		user.Email,
		user.Username,
		user.FirstName,
		user.LastName,
		user.IsActive,
		user.EmailVerified,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %d", user.ID)
	}

	return nil
}

// DeleteUser soft deletes a user by setting is_active to false
func (r *UserRepository) DeleteUser(ctx context.Context, userID int64) error {
	query := `
		UPDATE users 
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %d", userID)
	}

	return nil
}

// GetUserByEmail retrieves a user by their email address
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, username, first_name, last_name, 
		       is_active, email_verified, created_at, updated_at
		FROM users 
		WHERE email = $1
	`

	var user models.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetActiveUsersCount returns the count of active users
func (r *UserRepository) GetActiveUsersCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE is_active = true`

	var count int64
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get active users count: %w", err)
	}

	return count, nil
}
