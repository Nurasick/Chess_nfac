package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chess-nfac/backend/middleware"
	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/utils"
)

// ---------------------------------------------------------------------------
// Mock GameServicer
// ---------------------------------------------------------------------------

type mockGameService struct {
	mock.Mock
}

func (m *mockGameService) GetGame(ctx context.Context, gameID uuid.UUID) (*models.Game, error) {
	args := m.Called(ctx, gameID)
	g, _ := args.Get(0).(*models.Game)
	return g, args.Error(1)
}

func (m *mockGameService) GetMoves(ctx context.Context, gameID uuid.UUID) ([]models.Move, error) {
	args := m.Called(ctx, gameID)
	moves, _ := args.Get(0).([]models.Move)
	return moves, args.Error(1)
}

func (m *mockGameService) ProcessMove(ctx context.Context, gameID, playerID uuid.UUID, moveNotation string) (*models.Game, *models.Move, bool, error) {
	args := m.Called(ctx, gameID, playerID, moveNotation)
	g, _ := args.Get(0).(*models.Game)
	mv, _ := args.Get(1).(*models.Move)
	return g, mv, args.Bool(2), args.Error(3)
}

func (m *mockGameService) ResignGame(ctx context.Context, gameID, playerID uuid.UUID) (*models.Game, error) {
	args := m.Called(ctx, gameID, playerID)
	g, _ := args.Get(0).(*models.Game)
	return g, args.Error(1)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func requestWithGameID(method, path string, body interface{}, gameID uuid.UUID, userID *uuid.UUID) *http.Request {
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", gameID.String())
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	if userID != nil {
		ctx = context.WithValue(ctx, middleware.UserIDKey, *userID)
	}
	return req.WithContext(ctx)
}

// ---------------------------------------------------------------------------
// GET /games/{id}
// ---------------------------------------------------------------------------

func TestGameHandler_GetGame(t *testing.T) {
	gameID := uuid.New()
	sampleGame := &models.Game{ID: gameID, Status: models.GameStatusActive}

	tests := []struct {
		name       string
		pathID     string
		setupMock  func(svc *mockGameService)
		wantStatus int
	}{
		{
			name:   "success 200",
			pathID: gameID.String(),
			setupMock: func(svc *mockGameService) {
				svc.On("GetGame", mock.Anything, gameID).Return(sampleGame, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid UUID → 400",
			pathID:     "bad-id",
			setupMock:  func(svc *mockGameService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "game not found → 404",
			pathID: gameID.String(),
			setupMock: func(svc *mockGameService) {
				svc.On("GetGame", mock.Anything, gameID).
					Return(nil, utils.NewAppError("game_not_found", "Game not found", 404))
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockGameService{}
			tc.setupMock(svc)
			h := NewGameHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/games/"+tc.pathID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.pathID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()
			h.GetGame(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			svc.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// GET /games/{id}/moves
// ---------------------------------------------------------------------------

func TestGameHandler_GetMoves(t *testing.T) {
	gameID := uuid.New()
	sampleMoves := []models.Move{{ID: uuid.New(), GameID: gameID, Notation: "e2e4"}}

	tests := []struct {
		name       string
		pathID     string
		setupMock  func(svc *mockGameService)
		wantStatus int
	}{
		{
			name:   "success 200",
			pathID: gameID.String(),
			setupMock: func(svc *mockGameService) {
				svc.On("GetMoves", mock.Anything, gameID).Return(sampleMoves, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid UUID → 400",
			pathID:     "not-uuid",
			setupMock:  func(svc *mockGameService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "service error → 500",
			pathID: gameID.String(),
			setupMock: func(svc *mockGameService) {
				svc.On("GetMoves", mock.Anything, gameID).Return(nil, assert.AnError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockGameService{}
			tc.setupMock(svc)
			h := NewGameHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/games/"+tc.pathID+"/moves", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.pathID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()
			h.GetMoves(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			svc.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// POST /games/{id}/move
// ---------------------------------------------------------------------------

func TestGameHandler_MakeMove(t *testing.T) {
	gameID := uuid.New()
	userID := uuid.New()
	sampleGame := &models.Game{ID: gameID, Status: models.GameStatusActive}
	sampleMove := &models.Move{ID: uuid.New(), GameID: gameID, Notation: "e2e4"}

	tests := []struct {
		name       string
		injectUser bool
		body       interface{}
		setupMock  func(svc *mockGameService)
		wantStatus int
	}{
		{
			name:       "success 200",
			injectUser: true,
			body:       map[string]string{"move": "e2e4"},
			setupMock: func(svc *mockGameService) {
				svc.On("ProcessMove", mock.Anything, gameID, userID, "e2e4").
					Return(sampleGame, sampleMove, false, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no userID in context → 401",
			injectUser: false,
			body:       map[string]string{"move": "e2e4"},
			setupMock:  func(svc *mockGameService) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid JSON → 400",
			injectUser: true,
			body:       "bad json",
			setupMock:  func(svc *mockGameService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty move field → 400",
			injectUser: true,
			body:       map[string]string{"move": ""},
			setupMock:  func(svc *mockGameService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid move → 400",
			injectUser: true,
			body:       map[string]string{"move": "z9z9"},
			setupMock: func(svc *mockGameService) {
				svc.On("ProcessMove", mock.Anything, gameID, userID, "z9z9").
					Return(nil, nil, false, utils.NewAppError("invalid_move", "That move is not legal", 400))
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockGameService{}
			tc.setupMock(svc)
			h := NewGameHandler(svc)

			var bodyBytes []byte
			if s, ok := tc.body.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes, _ = json.Marshal(tc.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/move", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", gameID.String())
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
			if tc.injectUser {
				ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
			}
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()
			h.MakeMove(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			resp := apiResponse(t, rr.Body)
			if tc.wantStatus == http.StatusOK {
				require.Nil(t, resp["error"])
			} else {
				require.NotNil(t, resp["error"])
			}
			svc.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// POST /games/{id}/resign
// ---------------------------------------------------------------------------

func TestGameHandler_Resign(t *testing.T) {
	gameID := uuid.New()
	userID := uuid.New()
	result := models.GameResultBlackWins
	finishedGame := &models.Game{ID: gameID, Status: models.GameStatusCompleted, Result: &result}

	tests := []struct {
		name       string
		injectUser bool
		setupMock  func(svc *mockGameService)
		wantStatus int
	}{
		{
			name:       "success 200",
			injectUser: true,
			setupMock: func(svc *mockGameService) {
				svc.On("ResignGame", mock.Anything, gameID, userID).Return(finishedGame, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no userID in context → 401",
			injectUser: false,
			setupMock:  func(svc *mockGameService) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "game not active → 400",
			injectUser: true,
			setupMock: func(svc *mockGameService) {
				svc.On("ResignGame", mock.Anything, gameID, userID).
					Return(nil, utils.NewAppError("game_not_active", "Game is not active", 400))
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "forbidden → 403",
			injectUser: true,
			setupMock: func(svc *mockGameService) {
				svc.On("ResignGame", mock.Anything, gameID, userID).
					Return(nil, utils.NewAppError("forbidden", "You are not a player in this game", 403))
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockGameService{}
			tc.setupMock(svc)
			h := NewGameHandler(svc)

			var uid *uuid.UUID
			if tc.injectUser {
				uid = &userID
			}
			req := requestWithGameID(http.MethodPost, "/games/"+gameID.String()+"/resign", nil, gameID, uid)
			rr := httptest.NewRecorder()
			h.Resign(rr, req)

			assert.Equal(t, tc.wantStatus, rr.Code)
			svc.AssertExpectations(t)
		})
	}
}
