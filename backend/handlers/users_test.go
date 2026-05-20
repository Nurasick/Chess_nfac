package handlers

import (
	"context"
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
// Mock UserGetterServicer
// ---------------------------------------------------------------------------

type mockUserGetterService struct {
	mock.Mock
}

func (m *mockUserGetterService) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, userID)
	u, _ := args.Get(0).(*models.User)
	return u, args.Error(1)
}

// ---------------------------------------------------------------------------
// GET /users/me
// ---------------------------------------------------------------------------

func TestUserHandler_GetMe(t *testing.T) {
	userID := uuid.New()
	sampleUser := &models.User{ID: userID, Username: "alice", City: "almaty", Rating: 1200}

	tests := []struct {
		name       string
		injectUser bool
		setupMock  func(svc *mockUserGetterService)
		wantStatus int
	}{
		{
			name:       "success 200",
			injectUser: true,
			setupMock: func(svc *mockUserGetterService) {
				svc.On("GetByID", mock.Anything, userID).Return(sampleUser, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no userID in context → 401",
			injectUser: false,
			setupMock:  func(svc *mockUserGetterService) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "user not found → 404",
			injectUser: true,
			setupMock: func(svc *mockUserGetterService) {
				svc.On("GetByID", mock.Anything, userID).
					Return(nil, utils.NewAppError("user_not_found", "User not found", 404))
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserGetterService{}
			tc.setupMock(svc)
			h := NewUserHandler(svc)

			var rr *httptest.ResponseRecorder
			if tc.injectUser {
				req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
				req = req.WithContext(ctx)
				rr = httptest.NewRecorder()
				h.GetMe(rr, req)
			} else {
				req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
				rr = httptest.NewRecorder()
				h.GetMe(rr, req)
			}

			assert.Equal(t, tc.wantStatus, rr.Code)
			svc.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// GET /users/{id}
// ---------------------------------------------------------------------------

func TestUserHandler_GetByID(t *testing.T) {
	userID := uuid.New()
	sampleUser := &models.User{ID: userID, Username: "alice", City: "almaty", Rating: 1200}

	tests := []struct {
		name       string
		pathID     string
		setupMock  func(svc *mockUserGetterService)
		wantStatus int
	}{
		{
			name:   "success 200",
			pathID: userID.String(),
			setupMock: func(svc *mockUserGetterService) {
				svc.On("GetByID", mock.Anything, userID).Return(sampleUser, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid UUID → 400",
			pathID:     "not-a-uuid",
			setupMock:  func(svc *mockUserGetterService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found → 404",
			pathID: userID.String(),
			setupMock: func(svc *mockUserGetterService) {
				svc.On("GetByID", mock.Anything, userID).
					Return(nil, utils.NewAppError("user_not_found", "User not found", 404))
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserGetterService{}
			tc.setupMock(svc)
			h := NewUserHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tc.pathID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.pathID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()
			h.GetByID(rr, req)

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
