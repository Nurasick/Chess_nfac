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

type GameRepository interface {
	Create(ctx context.Context, game *models.Game) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Game, error)
	UpdateStatus(ctx context.Context, gameID uuid.UUID, status models.GameStatus) error
	Finish(ctx context.Context, gameID uuid.UUID, result models.GameResult, whiteRatingAfter, blackRatingAfter int) error
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*models.Game, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Game, int, error)
	UpdateFEN(ctx context.Context, gameID uuid.UUID, fen string) error
	GetUserGameCount(ctx context.Context, userID uuid.UUID) (int, error)
}

type PostgresGameRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresGameRepository(pool *pgxpool.Pool) GameRepository {
	return &PostgresGameRepository{pool: pool}
}

func (r *PostgresGameRepository) Create(ctx context.Context, game *models.Game) error {
	game.ID = uuid.New()
	game.CreatedAt = time.Now()
	game.UpdatedAt = time.Now()

	query := `
		INSERT INTO games (id, white_id, black_id, status, fen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	if _, err := r.pool.Exec(ctx, query,
		game.ID, game.WhiteID, game.BlackID, game.Status,
		game.FEN, game.CreatedAt, game.UpdatedAt,
	); err != nil {
		return fmt.Errorf("repository.PostgresGameRepository.Create: %w", err)
	}

	return nil
}

func (r *PostgresGameRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	query := `
		SELECT id, white_id, black_id, status, result, pgn, fen,
		       white_rating_before, white_rating_after, black_rating_before, black_rating_after,
		       created_at, updated_at, finished_at
		FROM games WHERE id = $1
	`

	game := &models.Game{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&game.ID, &game.WhiteID, &game.BlackID, &game.Status, &game.Result, &game.PGN, &game.FEN,
		&game.WhiteRatingBefore, &game.WhiteRatingAfter, &game.BlackRatingBefore, &game.BlackRatingAfter,
		&game.CreatedAt, &game.UpdatedAt, &game.FinishedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.PostgresGameRepository.FindByID: %w", err)
	}

	return game, nil
}

func (r *PostgresGameRepository) UpdateStatus(ctx context.Context, gameID uuid.UUID, status models.GameStatus) error {
	query := `UPDATE games SET status = $1, updated_at = $2 WHERE id = $3`

	if _, err := r.pool.Exec(ctx, query, status, time.Now(), gameID); err != nil {
		return fmt.Errorf("repository.PostgresGameRepository.UpdateStatus: %w", err)
	}

	return nil
}

func (r *PostgresGameRepository) Finish(ctx context.Context, gameID uuid.UUID, result models.GameResult, whiteRatingAfter, blackRatingAfter int) error {
	query := `
		UPDATE games
		SET status = $1, result = $2, white_rating_after = $3, black_rating_after = $4,
		    updated_at = $5, finished_at = $6
		WHERE id = $7
	`

	now := time.Now()
	if _, err := r.pool.Exec(ctx, query,
		models.GameStatusCompleted, result, whiteRatingAfter, blackRatingAfter,
		now, now, gameID,
	); err != nil {
		return fmt.Errorf("repository.PostgresGameRepository.Finish: %w", err)
	}

	return nil
}

func (r *PostgresGameRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*models.Game, error) {
	query := `
		SELECT id, white_id, black_id, status, result, pgn, fen,
		       white_rating_before, white_rating_after, black_rating_before, black_rating_after,
		       created_at, updated_at, finished_at
		FROM games
		WHERE (white_id = $1 OR black_id = $1) AND status = $2
		LIMIT 1
	`

	game := &models.Game{}
	err := r.pool.QueryRow(ctx, query, userID, models.GameStatusActive).Scan(
		&game.ID, &game.WhiteID, &game.BlackID, &game.Status, &game.Result, &game.PGN, &game.FEN,
		&game.WhiteRatingBefore, &game.WhiteRatingAfter, &game.BlackRatingBefore, &game.BlackRatingAfter,
		&game.CreatedAt, &game.UpdatedAt, &game.FinishedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.PostgresGameRepository.FindActiveByUserID: %w", err)
	}

	return game, nil
}

func (r *PostgresGameRepository) UpdateFEN(ctx context.Context, gameID uuid.UUID, fen string) error {
	query := `UPDATE games SET fen = $1, updated_at = $2 WHERE id = $3`

	if _, err := r.pool.Exec(ctx, query, fen, time.Now(), gameID); err != nil {
		return fmt.Errorf("repository.PostgresGameRepository.UpdateFEN: %w", err)
	}

	return nil
}

func (r *PostgresGameRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Game, int, error) {
	countQuery := `SELECT COUNT(*) FROM games WHERE white_id = $1 OR black_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("repository.PostgresGameRepository.FindByUserID count: %w", err)
	}

	query := `
		SELECT id, white_id, black_id, status, result, pgn, fen,
		       white_rating_before, white_rating_after, black_rating_before, black_rating_after,
		       created_at, updated_at, finished_at
		FROM games
		WHERE white_id = $1 OR black_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("repository.PostgresGameRepository.FindByUserID: %w", err)
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		g := models.Game{}
		if err := rows.Scan(
			&g.ID, &g.WhiteID, &g.BlackID, &g.Status, &g.Result, &g.PGN, &g.FEN,
			&g.WhiteRatingBefore, &g.WhiteRatingAfter, &g.BlackRatingBefore, &g.BlackRatingAfter,
			&g.CreatedAt, &g.UpdatedAt, &g.FinishedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("repository.PostgresGameRepository.FindByUserID scan: %w", err)
		}
		games = append(games, g)
	}

	if games == nil {
		games = []models.Game{}
	}

	return games, total, nil
}

func (r *PostgresGameRepository) GetUserGameCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT games_played FROM users WHERE id = $1`

	var count int
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("repository.PostgresGameRepository.GetUserGameCount: %w", err)
	}

	return count, nil
}
