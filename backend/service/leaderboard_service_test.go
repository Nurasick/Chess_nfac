package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/repository"
	"github.com/chess-nfac/backend/utils"
)

// ---------------------------------------------------------------------------
// Mock leaderboard repository
// ---------------------------------------------------------------------------

type mockLeaderboardRepo struct {
	mock.Mock
}

func (m *mockLeaderboardRepo) GetByCity(ctx context.Context, city string, limit, offset int) ([]models.LeaderboardEntry, int, error) {
	args := m.Called(ctx, city, limit, offset)
	entries, _ := args.Get(0).([]models.LeaderboardEntry)
	return entries, args.Int(1), args.Error(2)
}

func (m *mockLeaderboardRepo) Refresh(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

var _ repository.LeaderboardRepository = (*mockLeaderboardRepo)(nil)

// ---------------------------------------------------------------------------
// GetByCity
// ---------------------------------------------------------------------------

func TestLeaderboardService_GetByCity(t *testing.T) {
	ctx := context.Background()

	sampleEntries := []models.LeaderboardEntry{
		{ID: uuid.New(), UserID: uuid.New(), Username: "alice", City: "almaty", Rating: 1500, Rank: 1, GamesPlayed: 20, UpdatedAt: time.Now()},
		{ID: uuid.New(), UserID: uuid.New(), Username: "bob", City: "almaty", Rating: 1400, Rank: 2, GamesPlayed: 15, UpdatedAt: time.Now()},
	}

	tests := []struct {
		name      string
		city      string
		page      int
		pageSize  int
		setupMock func(r *mockLeaderboardRepo)
		wantErr   bool
		errCode   string
		wantCount int
		wantTotal int
	}{
		{
			name:     "success default pagination",
			city:     "almaty",
			page:     1,
			pageSize: 20,
			setupMock: func(r *mockLeaderboardRepo) {
				r.On("GetByCity", ctx, "almaty", 20, 0).Return(sampleEntries, 2, nil)
			},
			wantErr:   false,
			wantCount: 2,
			wantTotal: 2,
		},
		{
			name:     "page 2",
			city:     "astana",
			page:     2,
			pageSize: 10,
			setupMock: func(r *mockLeaderboardRepo) {
				r.On("GetByCity", ctx, "astana", 10, 10).Return([]models.LeaderboardEntry{}, 25, nil)
			},
			wantErr:   false,
			wantCount: 0,
			wantTotal: 25,
		},
		{
			name:      "invalid city",
			city:      "moscow",
			page:      1,
			pageSize:  20,
			setupMock: func(r *mockLeaderboardRepo) {},
			wantErr:   true,
			errCode:   "invalid_city",
		},
		{
			name:     "page < 1 defaults to 1",
			city:     "shymkent",
			page:     0,
			pageSize: 20,
			setupMock: func(r *mockLeaderboardRepo) {
				r.On("GetByCity", ctx, "shymkent", 20, 0).Return(sampleEntries[:1], 1, nil)
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:     "pageSize > 100 defaults to 20",
			city:     "almaty",
			page:     1,
			pageSize: 200,
			setupMock: func(r *mockLeaderboardRepo) {
				r.On("GetByCity", ctx, "almaty", 20, 0).Return(sampleEntries, 2, nil)
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:     "repo error",
			city:     "almaty",
			page:     1,
			pageSize: 20,
			setupMock: func(r *mockLeaderboardRepo) {
				r.On("GetByCity", ctx, "almaty", 20, 0).Return(nil, 0, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockLeaderboardRepo{}
			tc.setupMock(repo)
			svc := NewLeaderboardService(repo)

			entries, total, err := svc.GetByCity(ctx, tc.city, tc.page, tc.pageSize)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errCode != "" {
					var appErr utils.AppError
					require.True(t, errors.As(err, &appErr))
					assert.Equal(t, tc.errCode, appErr.Code)
				}
				return
			}
			require.NoError(t, err)
			assert.Len(t, entries, tc.wantCount)
			if tc.wantTotal > 0 {
				assert.Equal(t, tc.wantTotal, total)
			}
			repo.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// Refresh
// ---------------------------------------------------------------------------

func TestLeaderboardService_Refresh(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setupMock func(r *mockLeaderboardRepo)
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(r *mockLeaderboardRepo) {
				r.On("Refresh", ctx).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			setupMock: func(r *mockLeaderboardRepo) {
				r.On("Refresh", ctx).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockLeaderboardRepo{}
			tc.setupMock(repo)
			svc := NewLeaderboardService(repo)

			err := svc.Refresh(ctx)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// StartRefreshJob
// ---------------------------------------------------------------------------

func TestLeaderboardService_StartRefreshJob(t *testing.T) {
	t.Run("ctx cancelled before ticker fires exits cleanly", func(t *testing.T) {
		repo := &mockLeaderboardRepo{}
		svc := NewLeaderboardService(repo)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately — goroutine should observe ctx.Done before ticker fires

		svc.StartRefreshJob(ctx, time.Second) // long interval so ticker won't fire first

		time.Sleep(10 * time.Millisecond) // give goroutine time to see ctx.Done
		repo.AssertNotCalled(t, "Refresh", mock.Anything)
	})

	t.Run("ticker fires calls Refresh", func(t *testing.T) {
		repo := &mockLeaderboardRepo{}

		var once sync.Once
		called := make(chan struct{})
		repo.On("Refresh", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			once.Do(func() { close(called) })
		})

		svc := NewLeaderboardService(repo)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		svc.StartRefreshJob(ctx, time.Millisecond)

		select {
		case <-called:
			// Refresh was called — pass
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timed out waiting for Refresh to be called")
		}
	})

	t.Run("Refresh error does not panic", func(t *testing.T) {
		repo := &mockLeaderboardRepo{}

		var once sync.Once
		called := make(chan struct{})
		repo.On("Refresh", mock.Anything).Return(errors.New("db error")).Run(func(args mock.Arguments) {
			once.Do(func() { close(called) })
		})

		svc := NewLeaderboardService(repo)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		svc.StartRefreshJob(ctx, time.Millisecond)

		select {
		case <-called:
			// Refresh returned an error — no panic occurred
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timed out waiting for Refresh to be called")
		}
	})
}
