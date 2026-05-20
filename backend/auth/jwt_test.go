package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "supersecretkey_at_least_32_chars_long"

func TestGenerateAccessToken(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name     string
		userID   uuid.UUID
		username string
		secret   string
		wantErr  bool
	}{
		{
			name:     "valid token generated",
			userID:   userID,
			username: "alice",
			secret:   testSecret,
			wantErr:  false,
		},
		{
			name:     "empty secret still generates",
			userID:   userID,
			username: "alice",
			secret:   "",
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := GenerateAccessToken(tc.userID, tc.username, tc.secret)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, token)
		})
	}
}

func TestParseToken(t *testing.T) {
	userID := uuid.New()

	validToken, err := GenerateAccessToken(userID, "alice", testSecret)
	require.NoError(t, err)

	expiredClaims := Claims{
		UserID:   userID,
		Username: "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
		},
	}
	expiredRaw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte(testSecret))
	require.NoError(t, err)

	tests := []struct {
		name     string
		token    string
		secret   string
		wantErr  bool
		wantUser uuid.UUID
	}{
		{
			name:     "valid token parsed",
			token:    validToken,
			secret:   testSecret,
			wantErr:  false,
			wantUser: userID,
		},
		{
			name:    "expired token rejected",
			token:   expiredRaw,
			secret:  testSecret,
			wantErr: true,
		},
		{
			name:    "wrong secret rejected",
			token:   validToken,
			secret:  "wrong_secret_that_is_at_least_32chars",
			wantErr: true,
		},
		{
			name:    "garbage token rejected",
			token:   "not.a.token",
			secret:  testSecret,
			wantErr: true,
		},
		{
			name:    "empty token rejected",
			token:   "",
			secret:  testSecret,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := ParseToken(tc.token, tc.secret)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantUser, claims.UserID)
			assert.Equal(t, "alice", claims.Username)
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	userID := uuid.New()

	token, err := GenerateRefreshToken(userID, testSecret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestParseRefreshToken(t *testing.T) {
	userID := uuid.New()

	validToken, err := GenerateRefreshToken(userID, testSecret)
	require.NoError(t, err)

	expiredRaw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(-1 * time.Minute).Unix(),
		"iat":     time.Now().Add(-2 * time.Minute).Unix(),
	}).SignedString([]byte(testSecret))
	require.NoError(t, err)

	badIDRaw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "not-a-uuid",
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}).SignedString([]byte(testSecret))
	require.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		secret  string
		wantErr bool
		wantID  uuid.UUID
	}{
		{
			name:    "valid refresh token parsed",
			token:   validToken,
			secret:  testSecret,
			wantErr: false,
			wantID:  userID,
		},
		{
			name:    "expired refresh token rejected",
			token:   expiredRaw,
			secret:  testSecret,
			wantErr: true,
		},
		{
			name:    "wrong secret rejected",
			token:   validToken,
			secret:  "wrong_secret_that_is_at_least_32chars",
			wantErr: true,
		},
		{
			name:    "garbage token rejected",
			token:   "garbage",
			secret:  testSecret,
			wantErr: true,
		},
		{
			name:    "invalid user_id in claims rejected",
			token:   badIDRaw,
			secret:  testSecret,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id, err := ParseRefreshToken(tc.token, tc.secret)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantID, id)
		})
	}
}

func TestAccessTokenExpiry(t *testing.T) {
	userID := uuid.New()
	token, err := GenerateAccessToken(userID, "alice", testSecret)
	require.NoError(t, err)

	claims, err := ParseToken(token, testSecret)
	require.NoError(t, err)

	expectedExpiry := time.Now().Add(15 * time.Minute)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 5*time.Second)
}

func TestRefreshTokenExpiry(t *testing.T) {
	userID := uuid.New()
	token, err := GenerateRefreshToken(userID, testSecret)
	require.NoError(t, err)

	id, err := ParseRefreshToken(token, testSecret)
	require.NoError(t, err)
	assert.Equal(t, userID, id)

	parsed, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(tok *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	require.NoError(t, err)
	mc := parsed.Claims.(jwt.MapClaims)
	exp := mc["exp"].(float64)
	expTime := time.Unix(int64(exp), 0)
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, expTime, 5*time.Second)
}
