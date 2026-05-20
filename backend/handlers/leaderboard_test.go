package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/utils"
)

// ---------------------------------------------------------------------------
// Mock LeaderboardServicer
// ---------------------------------------------------------------------------

type mockLeaderboardService struct {
	mock.Mock
}

func (m *mockLeaderboardService) GetByCity(ctx context.Context, city string, page, pageSize int) ([]models.LeaderboardEntry, int, error) {
	args := m.Called(ctx, city, page, pageSize)
	entries, _ := args.Get(0).([]models.LeaderboardEntry)
	return entries, args.Int(1), args.Error(2)
}

// ---------------------------------------------------------------------------
// GET /leaderboard/{city}
// ---------------------------------------------------------------------------

func TestLeaderboardHandler_GetByCity(t *testing.T) {
	sampleEntries := []models.LeaderboardEntry{
		{ID: uuid.New(), UserID: uuid.New(), Username: "alice", City: "almaty", Rating: 1500, Rank: 1, GamesPlayed: 20, UpdatedAt: time.Now()},
	}

	tests := []struct {
		name       string
		city       string
		query      string
		setupMock  func(svc *mockLeaderboardService)
		wantStatus int
	}{
		{
			name:  "success with defaults",
			city:  "almaty",
			query: "",
			setupMock: func(svc *mockLeaderboardService) {
				svc.On("GetByCity", mock.Anything, "almaty", 1, 20).Return(sampleEntries, 1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "with pagination params",
			city:  "astana",
			query: "?page=2&page_size=10",
			setupMock: func(svc *mockLeaderboardService) {
				svc.On("GetByCity", mock.Anything, "astana", 2, 10).Return([]models.LeaderboardEntry{}, 25, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "invalid city → 400",
			city:  "moscow",
			query: "",
			setupMock: func(svc *mockLeaderboardService) {
				svc.On("GetByCity", mock.Anything, "moscow", 1, 20).
					Return(nil, 0, utils.NewAppError("invalid_city", "City not valid", 400))
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "service error → 500",
			city:  "almaty",
			query: "",
			setupMock: func(svc *mockLeaderboardService) {
				svc.On("GetByCity", mock.Anything, "almaty", 1, 20).
					Return(nil, 0, assert.AnError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockLeaderboardService{}
			tc.setupMock(svc)
			h := NewLeaderboardHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/leaderboard/"+tc.city+tc.query, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("city", tc.city)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()
			h.GetByCity(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			svc.AssertExpectations(t)
		})
	}
}
