package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chess-nfac/backend/middleware"
	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/utils"
)

// ---------------------------------------------------------------------------
// Mock UserServicer
// ---------------------------------------------------------------------------

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) Register(ctx context.Context, username, email, password, city string) (*models.User, error) {
	args := m.Called(ctx, username, email, password, city)
	u, _ := args.Get(0).(*models.User)
	return u, args.Error(1)
}

func (m *mockUserService) Login(ctx context.Context, username, password string) (*models.User, string, string, error) {
	args := m.Called(ctx, username, password)
	u, _ := args.Get(0).(*models.User)
	return u, args.String(1), args.String(2), args.Error(3)
}

func (m *mockUserService) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockUserService) Logout(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func apiResponse(t *testing.T, body *bytes.Buffer) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(body).Decode(&resp))
	return resp
}

func postJSON(t *testing.T, handler http.HandlerFunc, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler(rr, req)
	return rr
}

func postJSONWithUserID(t *testing.T, handler http.HandlerFunc, path string, body interface{}, userID uuid.UUID) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler(rr, req)
	return rr
}

// ---------------------------------------------------------------------------
// POST /auth/register
// ---------------------------------------------------------------------------

func TestAuthHandler_Register(t *testing.T) {
	sampleUser := &models.User{
		ID:       uuid.New(),
		Username: "alice",
		Email:    "alice@example.com",
		City:     "almaty",
		Rating:   1200,
	}

	tests := []struct {
		name       string
		body       interface{}
		setupMock  func(svc *mockUserService)
		wantStatus int
		wantErrNil bool
	}{
		{
			name: "success 201",
			body: map[string]string{
				"username": "alice",
				"email":    "alice@example.com",
				"password": "password123",
				"city":     "almaty",
			},
			setupMock: func(svc *mockUserService) {
				svc.On("Register", mock.Anything, "alice", "alice@example.com", "password123", "almaty").
					Return(sampleUser, nil)
			},
			wantStatus: http.StatusCreated,
			wantErrNil: true,
		},
		{
			name:       "invalid JSON → 400",
			body:       "not json",
			setupMock:  func(svc *mockUserService) {},
			wantStatus: http.StatusBadRequest,
			wantErrNil: false,
		},
		{
			name: "duplicate username → 409",
			body: map[string]string{
				"username": "alice",
				"email":    "alice@example.com",
				"password": "password123",
				"city":     "almaty",
			},
			setupMock: func(svc *mockUserService) {
				svc.On("Register", mock.Anything, "alice", "alice@example.com", "password123", "almaty").
					Return(nil, utils.NewAppError("user_exists", "Username already taken", 409))
			},
			wantStatus: http.StatusConflict,
			wantErrNil: false,
		},
		{
			name: "validation error → 400",
			body: map[string]string{
				"username": "ab",
				"email":    "bad",
				"password": "short",
				"city":     "mars",
			},
			setupMock: func(svc *mockUserService) {
				svc.On("Register", mock.Anything, "ab", "bad", "short", "mars").
					Return(nil, utils.NewAppError("validation_error", "invalid", 400))
			},
			wantStatus: http.StatusBadRequest,
			wantErrNil: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserService{}
			tc.setupMock(svc)
			h := NewAuthHandler(svc)

			var rr *httptest.ResponseRecorder
			if s, ok := tc.body.(string); ok {
				req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(s))
				rr = httptest.NewRecorder()
				h.Register(rr, req)
			} else {
				rr = postJSON(t, h.Register, "/auth/register", tc.body)
			}

			assert.Equal(t, tc.wantStatus, rr.Code)
			resp := apiResponse(t, rr.Body)
			assert.Equal(t, "success" == resp["message"], tc.wantErrNil)
			if tc.wantErrNil {
				assert.Nil(t, resp["error"])
			} else {
				assert.NotNil(t, resp["error"])
			}
			svc.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// POST /auth/login
// ---------------------------------------------------------------------------

func TestAuthHandler_Login(t *testing.T) {
	sampleUser := &models.User{
		ID:       uuid.New(),
		Username: "alice",
		City:     "almaty",
		Rating:   1200,
	}

	tests := []struct {
		name       string
		body       interface{}
		setupMock  func(svc *mockUserService)
		wantStatus int
		wantErrNil bool
	}{
		{
			name: "success 200",
			body: map[string]string{"username": "alice", "password": "password123"},
			setupMock: func(svc *mockUserService) {
				svc.On("Login", mock.Anything, "alice", "password123").
					Return(sampleUser, "access.token.here", "refresh.token.here", nil)
			},
			wantStatus: http.StatusOK,
			wantErrNil: true,
		},
		{
			name:       "invalid JSON → 400",
			body:       "{bad json",
			setupMock:  func(svc *mockUserService) {},
			wantStatus: http.StatusBadRequest,
			wantErrNil: false,
		},
		{
			name: "wrong credentials → 401",
			body: map[string]string{"username": "alice", "password": "wrongpass"},
			setupMock: func(svc *mockUserService) {
				svc.On("Login", mock.Anything, "alice", "wrongpass").
					Return(nil, "", "", utils.NewAppError("invalid_credentials", "Invalid username or password", 401))
			},
			wantStatus: http.StatusUnauthorized,
			wantErrNil: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserService{}
			tc.setupMock(svc)
			h := NewAuthHandler(svc)

			var rr *httptest.ResponseRecorder
			if s, ok := tc.body.(string); ok {
				req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(s))
				rr = httptest.NewRecorder()
				h.Login(rr, req)
			} else {
				rr = postJSON(t, h.Login, "/auth/login", tc.body)
			}

			assert.Equal(t, tc.wantStatus, rr.Code)
			resp := apiResponse(t, rr.Body)
			if tc.wantErrNil {
				assert.Nil(t, resp["error"])
				data, ok := resp["data"].(map[string]interface{})
				require.True(t, ok, "data should be an object")
				assert.NotEmpty(t, data["access_token"])
				assert.NotEmpty(t, data["refresh_token"])
			} else {
				assert.NotNil(t, resp["error"])
			}
			svc.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// POST /auth/refresh
// ---------------------------------------------------------------------------

func TestAuthHandler_Refresh(t *testing.T) {
	tests := []struct {
		name       string
		body       interface{}
		setupMock  func(svc *mockUserService)
		wantStatus int
		wantErrNil bool
	}{
		{
			name: "success 200",
			body: map[string]string{"refresh_token": "old.refresh.token"},
			setupMock: func(svc *mockUserService) {
				svc.On("RefreshTokens", mock.Anything, "old.refresh.token").
					Return("new.access.token", "new.refresh.token", nil)
			},
			wantStatus: http.StatusOK,
			wantErrNil: true,
		},
		{
			name:       "missing refresh_token field → 400",
			body:       map[string]string{},
			setupMock:  func(svc *mockUserService) {},
			wantStatus: http.StatusBadRequest,
			wantErrNil: false,
		},
		{
			name:       "invalid JSON → 400",
			body:       "notjson",
			setupMock:  func(svc *mockUserService) {},
			wantStatus: http.StatusBadRequest,
			wantErrNil: false,
		},
		{
			name: "invalid token → 401",
			body: map[string]string{"refresh_token": "bad.token"},
			setupMock: func(svc *mockUserService) {
				svc.On("RefreshTokens", mock.Anything, "bad.token").
					Return("", "", utils.NewAppError("invalid_token", "Invalid or expired refresh token", 401))
			},
			wantStatus: http.StatusUnauthorized,
			wantErrNil: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserService{}
			tc.setupMock(svc)
			h := NewAuthHandler(svc)

			var rr *httptest.ResponseRecorder
			if s, ok := tc.body.(string); ok {
				req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(s))
				rr = httptest.NewRecorder()
				h.Refresh(rr, req)
			} else {
				rr = postJSON(t, h.Refresh, "/auth/refresh", tc.body)
			}

			assert.Equal(t, tc.wantStatus, rr.Code)
			resp := apiResponse(t, rr.Body)
			if tc.wantErrNil {
				assert.Nil(t, resp["error"])
				data, ok := resp["data"].(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, data["access_token"])
				assert.NotEmpty(t, data["refresh_token"])
			} else {
				assert.NotNil(t, resp["error"])
			}
			svc.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// POST /auth/logout
// ---------------------------------------------------------------------------

func TestAuthHandler_Logout(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		injectUser bool
		setupMock  func(svc *mockUserService)
		wantStatus int
	}{
		{
			name:       "success 200",
			injectUser: true,
			setupMock: func(svc *mockUserService) {
				svc.On("Logout", mock.Anything, userID).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no userID in context → 401",
			injectUser: false,
			setupMock:  func(svc *mockUserService) {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockUserService{}
			tc.setupMock(svc)
			h := NewAuthHandler(svc)

			var rr *httptest.ResponseRecorder
			if tc.injectUser {
				rr = postJSONWithUserID(t, h.Logout, "/auth/logout", map[string]string{}, userID)
			} else {
				rr = postJSON(t, h.Logout, "/auth/logout", map[string]string{})
			}

			assert.Equal(t, tc.wantStatus, rr.Code)
			resp := apiResponse(t, rr.Body)
			if tc.wantStatus == http.StatusOK {
				assert.Nil(t, resp["error"])
			} else {
				assert.NotNil(t, resp["error"])
			}
			svc.AssertExpectations(t)
		})
	}
}
