package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chess-nfac/backend/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateRating(ctx context.Context, userID uuid.UUID, newRating int) error
	SaveRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	FindRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	Revoked   bool
}

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) UserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, username, email, password_hash, city, rating, games_played, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	if _, err := r.pool.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.City, user.Rating, user.GamesPlayed, user.CreatedAt, user.UpdatedAt,
	); err != nil {
		return fmt.Errorf("repository.PostgresUserRepository.Create: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, city, rating, games_played, created_at, updated_at
		FROM users WHERE id = $1
	`

	user := &models.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.City, &user.Rating, &user.GamesPlayed, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.PostgresUserRepository.FindByID: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, city, rating, games_played, created_at, updated_at
		FROM users WHERE username = $1
	`

	user := &models.User{}
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.City, &user.Rating, &user.GamesPlayed, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.PostgresUserRepository.FindByUsername: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, city, rating, games_played, created_at, updated_at
		FROM users WHERE email = $1
	`

	user := &models.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.City, &user.Rating, &user.GamesPlayed, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.PostgresUserRepository.FindByEmail: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) UpdateRating(ctx context.Context, userID uuid.UUID, newRating int) error {
	query := `UPDATE users SET rating = $1, updated_at = $2 WHERE id = $3`

	if _, err := r.pool.Exec(ctx, query, newRating, time.Now(), userID); err != nil {
		return fmt.Errorf("repository.PostgresUserRepository.UpdateRating: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) SaveRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, revoked)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err := r.pool.Exec(ctx, query,
		uuid.New(), userID, tokenHash, expiresAt, time.Now(), false,
	); err != nil {
		return fmt.Errorf("repository.PostgresUserRepository.SaveRefreshToken: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) FindRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked
		FROM refresh_tokens WHERE token_hash = $1
	`

	rt := &RefreshToken{}
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt, &rt.Revoked,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.PostgresUserRepository.FindRefreshToken: %w", err)
	}

	return rt, nil
}

func (r *PostgresUserRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1`

	if _, err := r.pool.Exec(ctx, query, tokenHash); err != nil {
		return fmt.Errorf("repository.PostgresUserRepository.RevokeRefreshToken: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`

	if _, err := r.pool.Exec(ctx, query, userID); err != nil {
		return fmt.Errorf("repository.PostgresUserRepository.RevokeAllUserTokens: %w", err)
	}

	return nil
}
