package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chess-nfac/backend/models"
)

type MoveRepository interface {
	Create(ctx context.Context, move *models.Move) error
	FindByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Move, error)
}

type PostgresMoveRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresMoveRepository(pool *pgxpool.Pool) MoveRepository {
	return &PostgresMoveRepository{pool: pool}
}

func (r *PostgresMoveRepository) Create(ctx context.Context, move *models.Move) error {
	move.ID = uuid.New()
	move.CreatedAt = time.Now()

	query := `
		INSERT INTO moves (id, game_id, player_id, move_number, notation, fen_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	if _, err := r.pool.Exec(ctx, query,
		move.ID, move.GameID, move.PlayerID, move.MoveNumber,
		move.Notation, move.FENAfter, move.CreatedAt,
	); err != nil {
		return fmt.Errorf("repository.PostgresMoveRepository.Create: %w", err)
	}

	return nil
}

func (r *PostgresMoveRepository) FindByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Move, error) {
	query := `
		SELECT id, game_id, player_id, move_number, notation, fen_after, created_at
		FROM moves WHERE game_id = $1
		ORDER BY move_number ASC
	`

	rows, err := r.pool.Query(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("repository.PostgresMoveRepository.FindByGameID: %w", err)
	}
	defer rows.Close()

	var moves []models.Move
	for rows.Next() {
		move := models.Move{}
		if err := rows.Scan(
			&move.ID, &move.GameID, &move.PlayerID, &move.MoveNumber,
			&move.Notation, &move.FENAfter, &move.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository.PostgresMoveRepository.FindByGameID: %w", err)
		}
		moves = append(moves, move)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.PostgresMoveRepository.FindByGameID: %w", err)
	}

	return moves, nil
}
