package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chess-nfac/backend/chess"
	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/repository"
)

// ---------------------------------------------------------------------------
// Mock game repository
// ---------------------------------------------------------------------------

type mockGameRepo struct {
	mock.Mock
}

func (m *mockGameRepo) Create(ctx context.Context, game *models.Game) error {
	args := m.Called(ctx, game)
	return args.Error(0)
}

func (m *mockGameRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	args := m.Called(ctx, id)
	g, _ := args.Get(0).(*models.Game)
	return g, args.Error(1)
}

func (m *mockGameRepo) UpdateStatus(ctx context.Context, gameID uuid.UUID, status models.GameStatus) error {
	args := m.Called(ctx, gameID, status)
	return args.Error(0)
}

func (m *mockGameRepo) Finish(ctx context.Context, gameID uuid.UUID, result models.GameResult, whiteRatingAfter, blackRatingAfter int) error {
	args := m.Called(ctx, gameID, result, whiteRatingAfter, blackRatingAfter)
	return args.Error(0)
}

func (m *mockGameRepo) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*models.Game, error) {
	args := m.Called(ctx, userID)
	g, _ := args.Get(0).(*models.Game)
	return g, args.Error(1)
}

func (m *mockGameRepo) UpdateFEN(ctx context.Context, gameID uuid.UUID, fen string) error {
	args := m.Called(ctx, gameID, fen)
	return args.Error(0)
}

func (m *mockGameRepo) GetUserGameCount(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

var _ repository.GameRepository = (*mockGameRepo)(nil)

// ---------------------------------------------------------------------------
// Mock move repository
// ---------------------------------------------------------------------------

type mockMoveRepo struct {
	mock.Mock
}

func (m *mockMoveRepo) Create(ctx context.Context, move *models.Move) error {
	args := m.Called(ctx, move)
	return args.Error(0)
}

func (m *mockMoveRepo) FindByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Move, error) {
	args := m.Called(ctx, gameID)
	moves, _ := args.Get(0).([]models.Move)
	return moves, args.Error(1)
}

var _ repository.MoveRepository = (*mockMoveRepo)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestGameService(gr *mockGameRepo, mr *mockMoveRepo, rr *mockRatingRepo) *GameService {
	ratingService := NewRatingService(rr)
	engine := chess.NewEngine()
	return NewGameService(gr, mr, ratingService, engine)
}

const startingFENTest = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// FEN where black can deliver checkmate in one move: d8h4 (Fool's Mate finish)
const foolsMateReadyFEN = "rnbqkbnr/pppp1ppp/8/4p3/6P1/5P2/PPPPP2P/RNBQKBNR b KQkq g3 0 2"

// ---------------------------------------------------------------------------
// CreateGame
// ---------------------------------------------------------------------------

func TestGameService_CreateGame(t *testing.T) {
	ctx := context.Background()
	whiteID := uuid.New()
	blackID := uuid.New()

	t.Run("success", func(t *testing.T) {
		gr := &mockGameRepo{}
		mr := &mockMoveRepo{}
		rr := &mockRatingRepo{}
		gr.On("Create", ctx, mock.AnythingOfType("*models.Game")).Return(nil)

		svc := newTestGameService(gr, mr, rr)
		game, err := svc.CreateGame(ctx, whiteID, blackID, 1200, 1200)

		require.NoError(t, err)
		assert.Equal(t, whiteID, game.WhiteID)
		assert.Equal(t, blackID, game.BlackID)
		assert.Equal(t, models.GameStatusActive, game.Status)
		gr.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		gr := &mockGameRepo{}
		mr := &mockMoveRepo{}
		rr := &mockRatingRepo{}
		gr.On("Create", ctx, mock.AnythingOfType("*models.Game")).Return(errors.New("db error"))

		svc := newTestGameService(gr, mr, rr)
		_, err := svc.CreateGame(ctx, whiteID, blackID, 1200, 1200)

		require.Error(t, err)
		gr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// GetGame
// ---------------------------------------------------------------------------

func TestGameService_GetGame(t *testing.T) {
	ctx := context.Background()
	gameID := uuid.New()
	sampleGame := &models.Game{ID: gameID, Status: models.GameStatusActive}

	tests := []struct {
		name      string
		setupMock func(gr *mockGameRepo)
		wantErr   bool
		errCode   string
	}{
		{
			name: "found",
			setupMock: func(gr *mockGameRepo) {
				gr.On("FindByID", ctx, gameID).Return(sampleGame, nil)
			},
		},
		{
			name: "not found",
			setupMock: func(gr *mockGameRepo) {
				gr.On("FindByID", ctx, gameID).Return(nil, nil)
			},
			wantErr: true,
			errCode: "game_not_found",
		},
		{
			name: "repo error",
			setupMock: func(gr *mockGameRepo) {
				gr.On("FindByID", ctx, gameID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gr := &mockGameRepo{}
			tc.setupMock(gr)
			svc := newTestGameService(gr, &mockMoveRepo{}, &mockRatingRepo{})

			game, err := svc.GetGame(ctx, gameID)

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, gameID, game.ID)
			gr.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// GetMoves
// ---------------------------------------------------------------------------

func TestGameService_GetMoves(t *testing.T) {
	ctx := context.Background()
	gameID := uuid.New()
	sampleMoves := []models.Move{{ID: uuid.New(), GameID: gameID, Notation: "e2e4"}}

	tests := []struct {
		name      string
		setupMock func(mr *mockMoveRepo)
		wantErr   bool
		wantLen   int
	}{
		{
			name: "success",
			setupMock: func(mr *mockMoveRepo) {
				mr.On("FindByGameID", ctx, gameID).Return(sampleMoves, nil)
			},
			wantLen: 1,
		},
		{
			name: "repo error",
			setupMock: func(mr *mockMoveRepo) {
				mr.On("FindByGameID", ctx, gameID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mr := &mockMoveRepo{}
			tc.setupMock(mr)
			svc := newTestGameService(&mockGameRepo{}, mr, &mockRatingRepo{})

			moves, err := svc.GetMoves(ctx, gameID)

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, moves, tc.wantLen)
			mr.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// ProcessMove
// ---------------------------------------------------------------------------

func TestGameService_ProcessMove(t *testing.T) {
	ctx := context.Background()
	gameID := uuid.New()
	whiteID := uuid.New()
	blackID := uuid.New()
	whiteRating := 1200
	blackRating := 1200

	activeGame := &models.Game{
		ID:                gameID,
		WhiteID:           whiteID,
		BlackID:           blackID,
		Status:            models.GameStatusActive,
		FEN:               startingFENTest,
		WhiteRatingBefore: &whiteRating,
		BlackRatingBefore: &blackRating,
	}

	t.Run("game not found", func(t *testing.T) {
		gr := &mockGameRepo{}
		gr.On("FindByID", ctx, gameID).Return(nil, nil)

		svc := newTestGameService(gr, &mockMoveRepo{}, &mockRatingRepo{})
		_, _, _, err := svc.ProcessMove(ctx, gameID, whiteID, "e2e4")

		require.Error(t, err)
		gr.AssertExpectations(t)
	})

	t.Run("game not active", func(t *testing.T) {
		gr := &mockGameRepo{}
		finished := &models.Game{ID: gameID, Status: models.GameStatusCompleted, FEN: startingFENTest}
		gr.On("FindByID", ctx, gameID).Return(finished, nil)

		svc := newTestGameService(gr, &mockMoveRepo{}, &mockRatingRepo{})
		_, _, _, err := svc.ProcessMove(ctx, gameID, whiteID, "e2e4")

		require.Error(t, err)
	})

	t.Run("not your turn", func(t *testing.T) {
		gr := &mockGameRepo{}
		gr.On("FindByID", ctx, gameID).Return(activeGame, nil)

		svc := newTestGameService(gr, &mockMoveRepo{}, &mockRatingRepo{})
		// It's white's turn (FEN has "w"), but blackID is trying to move
		_, _, _, err := svc.ProcessMove(ctx, gameID, blackID, "e2e4")

		require.Error(t, err)
	})

	t.Run("invalid move", func(t *testing.T) {
		gr := &mockGameRepo{}
		gr.On("FindByID", ctx, gameID).Return(activeGame, nil)

		svc := newTestGameService(gr, &mockMoveRepo{}, &mockRatingRepo{})
		_, _, _, err := svc.ProcessMove(ctx, gameID, whiteID, "z9z9")

		require.Error(t, err)
	})

	t.Run("valid move game continues", func(t *testing.T) {
		gr := &mockGameRepo{}
		mr := &mockMoveRepo{}
		gr.On("FindByID", ctx, gameID).Return(activeGame, nil)
		mr.On("FindByGameID", ctx, gameID).Return([]models.Move{}, nil)
		mr.On("Create", ctx, mock.AnythingOfType("*models.Move")).Return(nil)
		gr.On("UpdateFEN", ctx, gameID, mock.AnythingOfType("string")).Return(nil)

		svc := newTestGameService(gr, mr, &mockRatingRepo{})
		game, move, over, err := svc.ProcessMove(ctx, gameID, whiteID, "e2e4")

		require.NoError(t, err)
		assert.False(t, over)
		assert.NotNil(t, game)
		assert.NotNil(t, move)
		assert.Equal(t, "e2e4", move.Notation)
		gr.AssertExpectations(t)
		mr.AssertExpectations(t)
	})

	t.Run("valid move checkmate", func(t *testing.T) {
		// Fool's Mate: black delivers checkmate with d8h4
		foolsMateGame := &models.Game{
			ID:                gameID,
			WhiteID:           whiteID,
			BlackID:           blackID,
			Status:            models.GameStatusActive,
			FEN:               foolsMateReadyFEN,
			WhiteRatingBefore: &whiteRating,
			BlackRatingBefore: &blackRating,
		}

		gr := &mockGameRepo{}
		mr := &mockMoveRepo{}
		rr := &mockRatingRepo{}

		gr.On("FindByID", ctx, gameID).Return(foolsMateGame, nil)
		mr.On("FindByGameID", ctx, gameID).Return([]models.Move{}, nil)
		mr.On("Create", ctx, mock.AnythingOfType("*models.Move")).Return(nil)
		// finishGame calls
		gr.On("GetUserGameCount", ctx, whiteID).Return(5, nil)
		gr.On("GetUserGameCount", ctx, blackID).Return(5, nil)
		rr.On("GetUserRating", ctx, whiteID).Return(1200, nil)
		rr.On("GetUserRating", ctx, blackID).Return(1200, nil)
		rr.On("ApplyChange", ctx, mock.MatchedBy(func(c *models.RatingChange) bool { return c.UserID == whiteID })).Return(nil)
		rr.On("ApplyChange", ctx, mock.MatchedBy(func(c *models.RatingChange) bool { return c.UserID == blackID })).Return(nil)
		gr.On("UpdateFEN", ctx, gameID, mock.AnythingOfType("string")).Return(nil)
		gr.On("Finish", ctx, gameID, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		svc := newTestGameService(gr, mr, rr)
		// FEN is black's turn; blackID delivers checkmate
		game, move, over, err := svc.ProcessMove(ctx, gameID, blackID, "d8h4")

		require.NoError(t, err)
		assert.True(t, over)
		assert.NotNil(t, game)
		assert.Equal(t, models.GameStatusCompleted, game.Status)
		assert.NotNil(t, move)
		gr.AssertExpectations(t)
		mr.AssertExpectations(t)
		rr.AssertExpectations(t)
	})
}

// ---------------------------------------------------------------------------
// ResignGame
// ---------------------------------------------------------------------------

func TestGameService_ResignGame(t *testing.T) {
	ctx := context.Background()
	gameID := uuid.New()
	whiteID := uuid.New()
	blackID := uuid.New()

	newActiveGame := func() *models.Game {
		wr, br := 1200, 1200
		return &models.Game{
			ID:                gameID,
			WhiteID:           whiteID,
			BlackID:           blackID,
			Status:            models.GameStatusActive,
			FEN:               startingFENTest,
			WhiteRatingBefore: &wr,
			BlackRatingBefore: &br,
		}
	}

	t.Run("white resigns → black wins", func(t *testing.T) {
		gr := &mockGameRepo{}
		rr := &mockRatingRepo{}

		gr.On("FindByID", ctx, gameID).Return(newActiveGame(), nil)
		gr.On("GetUserGameCount", ctx, whiteID).Return(5, nil)
		gr.On("GetUserGameCount", ctx, blackID).Return(5, nil)
		rr.On("GetUserRating", ctx, whiteID).Return(1200, nil)
		rr.On("GetUserRating", ctx, blackID).Return(1200, nil)
		rr.On("ApplyChange", ctx, mock.MatchedBy(func(c *models.RatingChange) bool { return c.UserID == whiteID })).Return(nil)
		rr.On("ApplyChange", ctx, mock.MatchedBy(func(c *models.RatingChange) bool { return c.UserID == blackID })).Return(nil)
		gr.On("UpdateFEN", ctx, gameID, mock.AnythingOfType("string")).Return(nil)
		gr.On("Finish", ctx, gameID, models.GameResultBlackWins, mock.Anything, mock.Anything).Return(nil)

		svc := newTestGameService(gr, &mockMoveRepo{}, rr)
		game, err := svc.ResignGame(ctx, gameID, whiteID)

		require.NoError(t, err)
		assert.Equal(t, models.GameStatusCompleted, game.Status)
		require.NotNil(t, game.Result)
		assert.Equal(t, models.GameResultBlackWins, *game.Result)
		gr.AssertExpectations(t)
	})

	t.Run("black resigns → white wins", func(t *testing.T) {
		gr := &mockGameRepo{}
		rr := &mockRatingRepo{}

		gr.On("FindByID", ctx, gameID).Return(newActiveGame(), nil)
		gr.On("GetUserGameCount", ctx, whiteID).Return(5, nil)
		gr.On("GetUserGameCount", ctx, blackID).Return(5, nil)
		rr.On("GetUserRating", ctx, whiteID).Return(1200, nil)
		rr.On("GetUserRating", ctx, blackID).Return(1200, nil)
		rr.On("ApplyChange", ctx, mock.MatchedBy(func(c *models.RatingChange) bool { return c.UserID == whiteID })).Return(nil)
		rr.On("ApplyChange", ctx, mock.MatchedBy(func(c *models.RatingChange) bool { return c.UserID == blackID })).Return(nil)
		gr.On("UpdateFEN", ctx, gameID, mock.AnythingOfType("string")).Return(nil)
		gr.On("Finish", ctx, gameID, models.GameResultWhiteWins, mock.Anything, mock.Anything).Return(nil)

		svc := newTestGameService(gr, &mockMoveRepo{}, rr)
		game, err := svc.ResignGame(ctx, gameID, blackID)

		require.NoError(t, err)
		require.NotNil(t, game.Result)
		assert.Equal(t, models.GameResultWhiteWins, *game.Result)
		gr.AssertExpectations(t)
	})

	t.Run("not a player → forbidden", func(t *testing.T) {
		gr := &mockGameRepo{}
		outsider := uuid.New()
		gr.On("FindByID", ctx, gameID).Return(newActiveGame(), nil)

		svc := newTestGameService(gr, &mockMoveRepo{}, &mockRatingRepo{})
		_, err := svc.ResignGame(ctx, gameID, outsider)

		require.Error(t, err)
	})

	t.Run("game not found", func(t *testing.T) {
		gr := &mockGameRepo{}
		gr.On("FindByID", ctx, gameID).Return(nil, nil)

		svc := newTestGameService(gr, &mockMoveRepo{}, &mockRatingRepo{})
		_, err := svc.ResignGame(ctx, gameID, whiteID)

		require.Error(t, err)
	})

	t.Run("game not active", func(t *testing.T) {
		gr := &mockGameRepo{}
		finished := &models.Game{ID: gameID, Status: models.GameStatusCompleted, FEN: startingFENTest}
		gr.On("FindByID", ctx, gameID).Return(finished, nil)

		svc := newTestGameService(gr, &mockMoveRepo{}, &mockRatingRepo{})
		_, err := svc.ResignGame(ctx, gameID, whiteID)

		require.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// determineResult
// ---------------------------------------------------------------------------

func TestGameService_determineResult(t *testing.T) {
	svc := newTestGameService(&mockGameRepo{}, &mockMoveRepo{}, &mockRatingRepo{})

	tests := []struct {
		sideToMove string
		reason     string
		want       models.GameResult
	}{
		{"w", "checkmate", models.GameResultWhiteWins},
		{"b", "checkmate", models.GameResultBlackWins},
		{"w", "stalemate", models.GameResultDraw},
		{"w", "threefold_repetition", models.GameResultDraw},
		{"w", "fifty_move_rule", models.GameResultDraw},
		{"w", "insufficient_material", models.GameResultDraw},
		{"w", "draw_agreement", models.GameResultDraw},
		{"w", "unknown", models.GameResultDraw},
	}

	for _, tc := range tests {
		got := svc.determineResult(tc.sideToMove, tc.reason)
		assert.Equal(t, tc.want, got, "sideToMove=%s reason=%s", tc.sideToMove, tc.reason)
	}
}
