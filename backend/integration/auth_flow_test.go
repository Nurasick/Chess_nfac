//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chess-nfac/backend/config"
	"github.com/chess-nfac/backend/handlers"
	"github.com/chess-nfac/backend/middleware"
	"github.com/chess-nfac/backend/repository"
	"github.com/chess-nfac/backend/service"
)

const testJWTSecret = "test-secret-for-integration-tests-minimum32ch"

func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set — skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(context.Background()))
	t.Cleanup(pool.Close)
	return pool
}

func setupAuthServer(t *testing.T, pool *pgxpool.Pool) *httptest.Server {
	t.Helper()
	cfg := &config.Config{JWTSecret: testJWTSecret}

	userRepo := repository.NewPostgresUserRepository(pool)
	userSvc := service.NewUserService(userRepo, testJWTSecret)
	authHandler := handlers.NewAuthHandler(userSvc)

	r := chi.NewRouter()
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)
	r.Post("/auth/refresh", authHandler.Refresh)
	r.With(middleware.Auth(cfg)).Post("/auth/logout", authHandler.Logout)

	return httptest.NewServer(r)
}

func postJSON(t *testing.T, url string, body interface{}, token string) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decodeBody(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	var out map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	return out
}

func TestAuthFlow(t *testing.T) {
	pool := setupDB(t)
	srv := setupAuthServer(t, pool)
	defer srv.Close()

	suffix := fmt.Sprintf("%d", uniqueSuffix())
	username := "testuser_auth_" + suffix
	email := "testauth_" + suffix + "@example.com"
	password := "Password123!"

	// Clean up test data after the test.
	t.Cleanup(func() {
		pool.Exec(context.Background(),
			"DELETE FROM users WHERE username = $1", username)
	})

	// 1. Register
	resp := postJSON(t, srv.URL+"/auth/register", map[string]string{
		"username": username,
		"email":    email,
		"password": password,
		"city":     "almaty",
	}, "")
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// 2. Login
	resp = postJSON(t, srv.URL+"/auth/login", map[string]string{
		"username": username,
		"password": password,
	}, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	loginBody := decodeBody(t, resp)

	data, ok := loginBody["data"].(map[string]interface{})
	require.True(t, ok, "expected data envelope")
	accessToken, ok := data["access_token"].(string)
	require.True(t, ok && accessToken != "", "expected access_token")
	refreshToken, ok := data["refresh_token"].(string)
	require.True(t, ok && refreshToken != "", "expected refresh_token")

	// 3. Refresh
	resp = postJSON(t, srv.URL+"/auth/refresh", map[string]string{
		"refresh_token": refreshToken,
	}, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	refreshBody := decodeBody(t, resp)
	refreshData, ok := refreshBody["data"].(map[string]interface{})
	require.True(t, ok)
	newAccessToken, ok := refreshData["access_token"].(string)
	require.True(t, ok && newAccessToken != "", "expected new access_token")
	newRefreshToken, ok := refreshData["refresh_token"].(string)
	require.True(t, ok && newRefreshToken != "", "expected new refresh_token")

	// 4. Logout
	resp = postJSON(t, srv.URL+"/auth/logout", nil, newAccessToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// 5. Refresh with revoked token → 401
	resp = postJSON(t, srv.URL+"/auth/refresh", map[string]string{
		"refresh_token": newRefreshToken,
	}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}
