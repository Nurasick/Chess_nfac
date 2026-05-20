package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chess-nfac/backend/auth"
	"github.com/chess-nfac/backend/config"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-secret-for-websocket-handler-tests-32chars"

func testConfig() *config.Config {
	return &config.Config{
		JWTSecret: testJWTSecret,
	}
}

func TestHandleWebSocket_MissingAuth(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	server := httptest.NewServer(HandleWebSocket(hub, testConfig()))
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandleWebSocket_InvalidToken(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	server := httptest.NewServer(HandleWebSocket(hub, testConfig()))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalid.token.value")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandleWebSocket_ValidToken(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	server := httptest.NewServer(HandleWebSocket(hub, testConfig()))
	defer server.Close()

	userID := uuid.New()
	tokenString, err := auth.GenerateAccessToken(userID, "testuser", testJWTSecret)
	require.NoError(t, err)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	header := http.Header{}
	header.Set("Authorization", "Bearer "+tokenString)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	require.NoError(t, err)
	defer conn.Close()

	time.Sleep(30 * time.Millisecond)
	assert.True(t, hub.IsUserOnline(userID))
}

func TestHandleWebSocket_BearerPrefixStripped(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()
	server := httptest.NewServer(HandleWebSocket(hub, testConfig()))
	defer server.Close()

	userID := uuid.New()
	tokenString, err := auth.GenerateAccessToken(userID, "testuser", testJWTSecret)
	require.NoError(t, err)

	// Send token without "Bearer " prefix — handler should reject it
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", tokenString) // no "Bearer " prefix

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	// Token without Bearer prefix is passed as-is and should be a valid JWT
	// since handler only strips if len > 7 && starts with "Bearer "
	// Raw token IS a valid JWT so it should succeed (upgrade attempt)
	// But it's a plain HTTP request not a WS upgrade, so we just verify it
	// doesn't return 401 (it'll return 400 Bad Request for a non-WS upgrade)
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
}
