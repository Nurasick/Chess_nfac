package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/repository"
	"github.com/chess-nfac/backend/utils"
)

type LeaderboardService struct {
	leaderboardRepo repository.LeaderboardRepository
}

func NewLeaderboardService(leaderboardRepo repository.LeaderboardRepository) *LeaderboardService {
	return &LeaderboardService{
		leaderboardRepo: leaderboardRepo,
	}
}

func (s *LeaderboardService) GetByCity(ctx context.Context, city string, page, pageSize int) ([]models.LeaderboardEntry, int, error) {
	if err := utils.ValidateCity(city); err != nil {
		return nil, 0, utils.NewAppError("invalid_city", err.Error(), 400)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	entries, total, err := s.leaderboardRepo.GetByCity(ctx, city, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("leaderboard_service.GetByCity: %w", err)
	}

	return entries, total, nil
}

func (s *LeaderboardService) Refresh(ctx context.Context) error {
	if err := s.leaderboardRepo.Refresh(ctx); err != nil {
		return fmt.Errorf("leaderboard_service.Refresh: %w", err)
	}
	return nil
}

// StartRefreshJob runs leaderboard refresh every interval in a background goroutine.
func (s *LeaderboardService) StartRefreshJob(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.Refresh(ctx); err != nil {
					fmt.Printf("leaderboard refresh error: %v\n", err)
				}
			}
		}
	}()
}
