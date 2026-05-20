package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chess-nfac/backend/models"
)

type RatingRepository interface {
	ApplyChange(ctx context.Context, change *models.RatingChange) error
	GetUserRating(ctx context.Context, userID uuid.UUID) (int, error)
}

type PostgresRatingRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRatingRepository(pool *pgxpool.Pool) RatingRepository {
	return &PostgresRatingRepository{pool: pool}
}

func (r *PostgresRatingRepository) ApplyChange(ctx context.Context, change *models.RatingChange) error {
	query := `
		UPDATE users
		SET rating = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	if _, err := r.pool.Exec(ctx, query, change.NewRating, change.UserID); err != nil {
		return fmt.Errorf("repository.PostgresRatingRepository.ApplyChange: %w", err)
	}

	return nil
}

func (r *PostgresRatingRepository) GetUserRating(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT rating FROM users WHERE id = $1`

	var rating int
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&rating); err != nil {
		return 0, fmt.Errorf("repository.PostgresRatingRepository.GetUserRating: %w", err)
	}

	return rating, nil
}
