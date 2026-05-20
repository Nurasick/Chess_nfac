package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chess-nfac/backend/auth"
	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/repository"
	"github.com/chess-nfac/backend/utils"
)

const svcTestSecret = "test_jwt_secret_at_least_32_chars_ok"

func newTestUserService(repo *mockUserRepo) *UserService {
	return NewUserService(repo, svcTestSecret)
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		username  string
		email     string
		password  string
		city      string
		setupMock func(r *mockUserRepo)
		wantErr   bool
		errCode   string
	}{
		{
			name:     "successful registration",
			username: "alice",
			email:    "alice@example.com",
			password: "password123",
			city:     "almaty",
			setupMock: func(r *mockUserRepo) {
				r.On("FindByUsername", ctx, "alice").Return(nil, nil)
				r.On("FindByEmail", ctx, "alice@example.com").Return(nil, nil)
				r.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "duplicate username",
			username: "alice",
			email:    "other@example.com",
			password: "password123",
			city:     "astana",
			setupMock: func(r *mockUserRepo) {
				existing := &models.User{ID: uuid.New(), Username: "alice"}
				r.On("FindByUsername", ctx, "alice").Return(existing, nil)
			},
			wantErr: true,
			errCode: "user_exists",
		},
		{
			name:     "duplicate email",
			username: "bob",
			email:    "taken@example.com",
			password: "password123",
			city:     "shymkent",
			setupMock: func(r *mockUserRepo) {
				r.On("FindByUsername", ctx, "bob").Return(nil, nil)
				existing := &models.User{ID: uuid.New(), Email: "taken@example.com"}
				r.On("FindByEmail", ctx, "taken@example.com").Return(existing, nil)
			},
			wantErr: true,
			errCode: "email_exists",
		},
		{
			name:      "invalid username (too short)",
			username:  "ab",
			email:     "ab@example.com",
			password:  "password123",
			city:      "almaty",
			setupMock: func(r *mockUserRepo) {},
			wantErr:   true,
			errCode:   "validation_error",
		},
		{
			name:      "invalid email",
			username:  "charlie",
			email:     "not-an-email",
			password:  "password123",
			city:      "almaty",
			setupMock: func(r *mockUserRepo) {},
			wantErr:   true,
			errCode:   "validation_error",
		},
		{
			name:      "password too short",
			username:  "dave",
			email:     "dave@example.com",
			password:  "abc",
			city:      "almaty",
			setupMock: func(r *mockUserRepo) {},
			wantErr:   true,
			errCode:   "validation_error",
		},
		{
			name:      "invalid city",
			username:  "eve",
			email:     "eve@example.com",
			password:  "password123",
			city:      "moscow",
			setupMock: func(r *mockUserRepo) {},
			wantErr:   true,
			errCode:   "validation_error",
		},
		{
			name:     "repo error on FindByUsername",
			username: "frank",
			email:    "frank@example.com",
			password: "password123",
			city:     "almaty",
			setupMock: func(r *mockUserRepo) {
				r.On("FindByUsername", ctx, "frank").Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{}
			tc.setupMock(repo)
			svc := newTestUserService(repo)

			user, err := svc.Register(ctx, tc.username, tc.email, tc.password, tc.city)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errCode != "" {
					var appErr utils.AppError
					require.ErrorAs(t, err, &appErr)
					assert.Equal(t, tc.errCode, appErr.Code)
				}
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tc.username, user.Username)
				assert.Equal(t, tc.email, user.Email)
				assert.Equal(t, 1200, user.Rating)
			}

			repo.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// Login
// ---------------------------------------------------------------------------

func TestUserService_Login(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	password := "password123"
	hash, err := auth.HashPassword(password)
	require.NoError(t, err)

	existingUser := &models.User{
		ID:           userID,
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: hash,
		City:         "almaty",
		Rating:       1200,
	}

	tests := []struct {
		name      string
		username  string
		password  string
		setupMock func(r *mockUserRepo)
		wantErr   bool
		errCode   string
	}{
		{
			name:     "successful login",
			username: "alice",
			password: password,
			setupMock: func(r *mockUserRepo) {
				r.On("FindByUsername", ctx, "alice").Return(existingUser, nil)
				r.On("SaveRefreshToken", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "empty username rejected",
			username:  "",
			password:  password,
			setupMock: func(r *mockUserRepo) {},
			wantErr:   true,
			errCode:   "validation_error",
		},
		{
			name:      "empty password rejected",
			username:  "alice",
			password:  "",
			setupMock: func(r *mockUserRepo) {},
			wantErr:   true,
			errCode:   "validation_error",
		},
		{
			name:     "user not found",
			username: "ghost",
			password: password,
			setupMock: func(r *mockUserRepo) {
				r.On("FindByUsername", ctx, "ghost").Return(nil, nil)
			},
			wantErr: true,
			errCode: "invalid_credentials",
		},
		{
			name:     "wrong password",
			username: "alice",
			password: "wrongpass",
			setupMock: func(r *mockUserRepo) {
				r.On("FindByUsername", ctx, "alice").Return(existingUser, nil)
			},
			wantErr: true,
			errCode: "invalid_credentials",
		},
		{
			name:     "repo error",
			username: "alice",
			password: password,
			setupMock: func(r *mockUserRepo) {
				r.On("FindByUsername", ctx, "alice").Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{}
			tc.setupMock(repo)
			svc := newTestUserService(repo)

			user, accessToken, refreshToken, err := svc.Login(ctx, tc.username, tc.password)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errCode != "" {
					var appErr utils.AppError
					require.ErrorAs(t, err, &appErr)
					assert.Equal(t, tc.errCode, appErr.Code)
				}
				assert.Nil(t, user)
				assert.Empty(t, accessToken)
				assert.Empty(t, refreshToken)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.NotEmpty(t, accessToken)
				assert.NotEmpty(t, refreshToken)
			}

			repo.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// RefreshTokens
// ---------------------------------------------------------------------------

func TestUserService_RefreshTokens(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	existingUser := &models.User{
		ID:       userID,
		Username: "alice",
		City:     "almaty",
		Rating:   1200,
	}

	// A plausible token string (content doesn't matter — service only SHA256-hashes it)
	rawToken := "some.refresh.token.string"
	tokenHash := hashRefreshToken(rawToken)

	validRT := &repository.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Revoked:   false,
	}

	tests := []struct {
		name         string
		refreshToken string
		setupMock    func(r *mockUserRepo)
		wantErr      bool
		errCode      string
	}{
		{
			name:         "successful refresh",
			refreshToken: rawToken,
			setupMock: func(r *mockUserRepo) {
				r.On("FindRefreshToken", ctx, tokenHash).Return(validRT, nil)
				r.On("RevokeRefreshToken", ctx, tokenHash).Return(nil)
				r.On("FindByID", ctx, userID).Return(existingUser, nil)
				r.On("SaveRefreshToken", ctx, userID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "token not found",
			refreshToken: rawToken,
			setupMock: func(r *mockUserRepo) {
				r.On("FindRefreshToken", ctx, tokenHash).Return(nil, nil)
			},
			wantErr: true,
			errCode: "invalid_token",
		},
		{
			name:         "revoked token",
			refreshToken: rawToken,
			setupMock: func(r *mockUserRepo) {
				revokedRT := &repository.RefreshToken{
					ID:        uuid.New(),
					UserID:    userID,
					TokenHash: tokenHash,
					ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
					Revoked:   true,
				}
				r.On("FindRefreshToken", ctx, tokenHash).Return(revokedRT, nil)
			},
			wantErr: true,
			errCode: "invalid_token",
		},
		{
			name:         "expired token",
			refreshToken: rawToken,
			setupMock: func(r *mockUserRepo) {
				expiredRT := &repository.RefreshToken{
					ID:        uuid.New(),
					UserID:    userID,
					TokenHash: tokenHash,
					ExpiresAt: time.Now().Add(-1 * time.Hour),
					Revoked:   false,
				}
				r.On("FindRefreshToken", ctx, tokenHash).Return(expiredRT, nil)
			},
			wantErr: true,
			errCode: "invalid_token",
		},
		{
			name:         "user not found after valid token",
			refreshToken: rawToken,
			setupMock: func(r *mockUserRepo) {
				r.On("FindRefreshToken", ctx, tokenHash).Return(validRT, nil)
				r.On("RevokeRefreshToken", ctx, tokenHash).Return(nil)
				r.On("FindByID", ctx, userID).Return(nil, nil)
			},
			wantErr: true,
			errCode: "user_not_found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{}
			tc.setupMock(repo)
			svc := newTestUserService(repo)

			accessToken, newRefresh, err := svc.RefreshTokens(ctx, tc.refreshToken)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errCode != "" {
					var appErr utils.AppError
					require.ErrorAs(t, err, &appErr)
					assert.Equal(t, tc.errCode, appErr.Code)
				}
				assert.Empty(t, accessToken)
				assert.Empty(t, newRefresh)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, accessToken)
				assert.NotEmpty(t, newRefresh)
			}

			repo.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestUserService_GetByID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	existing := &models.User{ID: userID, Username: "alice", City: "almaty"}

	tests := []struct {
		name      string
		userID    uuid.UUID
		setupMock func(r *mockUserRepo)
		wantErr   bool
		errCode   string
	}{
		{
			name:   "user found",
			userID: userID,
			setupMock: func(r *mockUserRepo) {
				r.On("FindByID", ctx, userID).Return(existing, nil)
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: userID,
			setupMock: func(r *mockUserRepo) {
				r.On("FindByID", ctx, userID).Return(nil, nil)
			},
			wantErr: true,
			errCode: "user_not_found",
		},
		{
			name:   "repo error",
			userID: userID,
			setupMock: func(r *mockUserRepo) {
				r.On("FindByID", ctx, userID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{}
			tc.setupMock(repo)
			svc := newTestUserService(repo)

			user, err := svc.GetByID(ctx, tc.userID)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errCode != "" {
					var appErr utils.AppError
					require.ErrorAs(t, err, &appErr)
					assert.Equal(t, tc.errCode, appErr.Code)
				}
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tc.userID, user.ID)
			}

			repo.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// Logout
// ---------------------------------------------------------------------------

func TestUserService_Logout(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name      string
		setupMock func(r *mockUserRepo)
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(r *mockUserRepo) {
				r.On("RevokeAllUserTokens", ctx, userID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			setupMock: func(r *mockUserRepo) {
				r.On("RevokeAllUserTokens", ctx, userID).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{}
			tc.setupMock(repo)
			svc := newTestUserService(repo)

			err := svc.Logout(ctx, userID)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}
