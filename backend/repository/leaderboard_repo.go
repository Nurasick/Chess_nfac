package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/chess-nfac/backend/models"
)

type LeaderboardRepository interface {
	GetByCity(ctx context.Context, city string, limit int, offset int) ([]models.LeaderboardEntry, int, error)
	Refresh(ctx context.Context) error
}

type PostgresLeaderboardRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresLeaderboardRepository(pool *pgxpool.Pool) LeaderboardRepository {
	return &PostgresLeaderboardRepository{pool: pool}
}

func (r *PostgresLeaderboardRepository) GetByCity(ctx context.Context, city string, limit int, offset int) ([]models.LeaderboardEntry, int, error) {
	if city == "global" {
		return r.getGlobal(ctx, limit, offset)
	}

	countQuery := `SELECT COUNT(*) FROM city_leaderboard WHERE city = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, city).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("repository.PostgresLeaderboardRepository.GetByCity (count): %w", err)
	}

	query := `
		SELECT cl.id, cl.user_id, u.username, cl.city, cl.rating, cl.rank, cl.games_played, cl.updated_at
		FROM city_leaderboard cl
		JOIN users u ON cl.user_id = u.id
		WHERE cl.city = $1
		ORDER BY cl.rank ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, city, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("repository.PostgresLeaderboardRepository.GetByCity: %w", err)
	}
	defer rows.Close()

	var entries []models.LeaderboardEntry
	for rows.Next() {
		entry := models.LeaderboardEntry{}
		if err := rows.Scan(
			&entry.ID, &entry.UserID, &entry.Username, &entry.City,
			&entry.Rating, &entry.Rank, &entry.GamesPlayed, &entry.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("repository.PostgresLeaderboardRepository.GetByCity: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("repository.PostgresLeaderboardRepository.GetByCity: %w", err)
	}

	return entries, total, nil
}

func (r *PostgresLeaderboardRepository) getGlobal(ctx context.Context, limit int, offset int) ([]models.LeaderboardEntry, int, error) {
	countQuery := `SELECT COUNT(*) FROM city_leaderboard`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("repository.PostgresLeaderboardRepository.getGlobal (count): %w", err)
	}

	query := `
		SELECT cl.id, cl.user_id, u.username, cl.city, cl.rating, cl.rank, cl.games_played, cl.updated_at
		FROM city_leaderboard cl
		JOIN users u ON cl.user_id = u.id
		ORDER BY cl.rating DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("repository.PostgresLeaderboardRepository.getGlobal: %w", err)
	}
	defer rows.Close()

	var entries []models.LeaderboardEntry
	for rows.Next() {
		entry := models.LeaderboardEntry{}
		if err := rows.Scan(
			&entry.ID, &entry.UserID, &entry.Username, &entry.City,
			&entry.Rating, &entry.Rank, &entry.GamesPlayed, &entry.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("repository.PostgresLeaderboardRepository.getGlobal: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("repository.PostgresLeaderboardRepository.getGlobal: %w", err)
	}

	return entries, total, nil
}

func (r *PostgresLeaderboardRepository) Refresh(ctx context.Context) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository.PostgresLeaderboardRepository.Refresh: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM city_leaderboard`); err != nil {
		return fmt.Errorf("repository.PostgresLeaderboardRepository.Refresh: delete: %w", err)
	}

	insertQuery := `
		INSERT INTO city_leaderboard (id, user_id, city, rating, rank, games_played, updated_at)
		SELECT
			gen_random_uuid(),
			u.id,
			u.city,
			u.rating,
			ROW_NUMBER() OVER (PARTITION BY u.city ORDER BY u.rating DESC) AS rank,
			u.games_played,
			CURRENT_TIMESTAMP
		FROM users u
		WHERE u.city IS NOT NULL
	`
	if _, err := tx.Exec(ctx, insertQuery); err != nil {
		return fmt.Errorf("repository.PostgresLeaderboardRepository.Refresh: insert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repository.PostgresLeaderboardRepository.Refresh: commit: %w", err)
	}

	return nil
}
