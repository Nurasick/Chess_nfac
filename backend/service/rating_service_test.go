package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/repository"
)

// ---------------------------------------------------------------------------
// Mock rating repository
// ---------------------------------------------------------------------------

type mockRatingRepo struct {
	mock.Mock
}

func (m *mockRatingRepo) ApplyChange(ctx context.Context, change *models.RatingChange) error {
	args := m.Called(ctx, change)
	return args.Error(0)
}

func (m *mockRatingRepo) GetUserRating(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

var _ repository.RatingRepository = (*mockRatingRepo)(nil)

// ---------------------------------------------------------------------------
// CalculateNewRatings — pure math, no repo
// ---------------------------------------------------------------------------

func TestRatingService_CalculateNewRatings(t *testing.T) {
	svc := NewRatingService(&mockRatingRepo{})

	tests := []struct {
		name           string
		whiteRating    int
		blackRating    int
		result         string
		whiteGameCount int
		blackGameCount int
		wantErr        bool
		checkWhite     func(t *testing.T, newRating int)
		checkBlack     func(t *testing.T, newRating int)
	}{
		{
			name:           "white wins — higher rated white gains less",
			whiteRating:    1400,
			blackRating:    1200,
			result:         "1-0",
			whiteGameCount: 50,
			blackGameCount: 50,
			wantErr:        false,
			checkWhite: func(t *testing.T, r int) {
				assert.Greater(t, r, 1400, "white should gain rating after win")
			},
			checkBlack: func(t *testing.T, r int) {
				assert.Less(t, r, 1200, "black should lose rating after loss")
			},
		},
		{
			name:           "black wins",
			whiteRating:    1200,
			blackRating:    1200,
			result:         "0-1",
			whiteGameCount: 5,
			blackGameCount: 5,
			wantErr:        false,
			checkWhite: func(t *testing.T, r int) { assert.Less(t, r, 1200) },
			checkBlack: func(t *testing.T, r int) { assert.Greater(t, r, 1200) },
		},
		{
			name:           "draw — equal ratings, no change",
			whiteRating:    1200,
			blackRating:    1200,
			result:         "0.5-0.5",
			whiteGameCount: 50,
			blackGameCount: 50,
			wantErr:        false,
			checkWhite: func(t *testing.T, r int) { assert.Equal(t, 1200, r) },
			checkBlack: func(t *testing.T, r int) { assert.Equal(t, 1200, r) },
		},
		{
			name:           "new player uses K=32",
			whiteRating:    1200,
			blackRating:    1200,
			result:         "1-0",
			whiteGameCount: 5,
			blackGameCount: 5,
			wantErr:        false,
			checkWhite: func(t *testing.T, r int) { assert.Equal(t, 1216, r) },
			checkBlack: func(t *testing.T, r int) { assert.Equal(t, 1184, r) },
		},
		{
			name:           "established player uses K=16",
			whiteRating:    1200,
			blackRating:    1200,
			result:         "1-0",
			whiteGameCount: 50,
			blackGameCount: 50,
			wantErr:        false,
			checkWhite: func(t *testing.T, r int) { assert.Equal(t, 1208, r) },
			checkBlack: func(t *testing.T, r int) { assert.Equal(t, 1192, r) },
		},
		{
			name:           "rating cannot go below 100",
			whiteRating:    100,
			blackRating:    3000,
			result:         "0-1",
			whiteGameCount: 50,
			blackGameCount: 50,
			wantErr:        false,
			checkWhite: func(t *testing.T, r int) { assert.GreaterOrEqual(t, r, 100) },
			checkBlack: func(t *testing.T, r int) { assert.GreaterOrEqual(t, r, 3000) },
		},
		{
			name:    "invalid result string",
			result:  "invalid",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			newWhite, newBlack, err := svc.CalculateNewRatings(
				tc.whiteRating, tc.blackRating,
				tc.result,
				tc.whiteGameCount, tc.blackGameCount,
			)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.checkWhite != nil {
				tc.checkWhite(t, newWhite)
			}
			if tc.checkBlack != nil {
				tc.checkBlack(t, newBlack)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ApplyRatingChange
// ---------------------------------------------------------------------------

func TestRatingService_ApplyRatingChange(t *testing.T) {
	ctx := context.Background()
	whiteID := uuid.New()
	blackID := uuid.New()

	tests := []struct {
		name            string
		newWhiteRating  int
		newBlackRating  int
		setupMock       func(r *mockRatingRepo)
		wantErr         bool
		wantWhiteDelta  int
		wantBlackDelta  int
	}{
		{
			name:           "success",
			newWhiteRating: 1220,
			newBlackRating: 1180,
			setupMock: func(r *mockRatingRepo) {
				r.On("GetUserRating", ctx, whiteID).Return(1200, nil)
				r.On("GetUserRating", ctx, blackID).Return(1200, nil)
				r.On("ApplyChange", ctx, mock.MatchedBy(func(c *models.RatingChange) bool {
					return c.UserID == whiteID
				})).Return(nil)
				r.On("ApplyChange", ctx, mock.MatchedBy(func(c *models.RatingChange) bool {
					return c.UserID == blackID
				})).Return(nil)
			},
			wantErr:        false,
			wantWhiteDelta: 20,
			wantBlackDelta: -20,
		},
		{
			name: "get white rating error",
			setupMock: func(r *mockRatingRepo) {
				r.On("GetUserRating", ctx, whiteID).Return(0, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "get black rating error",
			setupMock: func(r *mockRatingRepo) {
				r.On("GetUserRating", ctx, whiteID).Return(1200, nil)
				r.On("GetUserRating", ctx, blackID).Return(0, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:           "apply white change error",
			newWhiteRating: 1220,
			newBlackRating: 1180,
			setupMock: func(r *mockRatingRepo) {
				r.On("GetUserRating", ctx, whiteID).Return(1200, nil)
				r.On("GetUserRating", ctx, blackID).Return(1200, nil)
				r.On("ApplyChange", ctx, mock.MatchedBy(func(c *models.RatingChange) bool {
					return c.UserID == whiteID
				})).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockRatingRepo{}
			tc.setupMock(repo)
			svc := NewRatingService(repo)

			whiteDelta, blackDelta, err := svc.ApplyRatingChange(ctx, whiteID, blackID, tc.newWhiteRating, tc.newBlackRating)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantWhiteDelta, whiteDelta)
			assert.Equal(t, tc.wantBlackDelta, blackDelta)
			repo.AssertExpectations(t)
		})
	}
}
