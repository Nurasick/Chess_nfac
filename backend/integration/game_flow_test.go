//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chess-nfac/backend/chess"
	"github.com/chess-nfac/backend/config"
	"github.com/chess-nfac/backend/handlers"
	"github.com/chess-nfac/backend/middleware"
	"github.com/chess-nfac/backend/repository"
	"github.com/chess-nfac/backend/service"
)

// legalMoves is a valid 5-move sequence from the starting position (UCI notation).
// White: e2e4, g1f3, f1b5   Black: e7e5, b8c6
var legalMoves = []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1b5"}

type gameTestCtx struct {
	srv         *httptest.Server
	gameSvc     *service.GameService
	userSvc     *service.UserService
	whiteID     uuid.UUID
	blackID     uuid.UUID
	whiteToken  string
	blackToken  string
}

func setupGameServer(t *testing.T) *gameTestCtx {
	t.Helper()
	pool := setupDB(t)
	cfg := &config.Config{JWTSecret: testJWTSecret}

	userRepo := repository.NewPostgresUserRepository(pool)
	gameRepo := repository.NewPostgresGameRepository(pool)
	moveRepo := repository.NewPostgresMoveRepository(pool)
	ratingRepo := repository.NewPostgresRatingRepository(pool)

	userSvc := service.NewUserService(userRepo, testJWTSecret)
	ratingSvc := service.NewRatingService(ratingRepo)
	engine := chess.NewEngine()
	gameSvc := service.NewGameService(gameRepo, moveRepo, ratingSvc, engine)

	authHandler := handlers.NewAuthHandler(userSvc)
	gameHandler := handlers.NewGameHandler(gameSvc)

	r := chi.NewRouter()
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)
	r.With(middleware.Auth(cfg)).Group(func(r chi.Router) {
		r.Post("/games/{id}/move", gameHandler.MakeMove)
		r.Get("/games/{id}", gameHandler.GetGame)
		r.Get("/games/{id}/moves", gameHandler.GetMoves)
	})

	srv := httptest.NewServer(r)

	return &gameTestCtx{
		srv:     srv,
		gameSvc: gameSvc,
		userSvc: userSvc,
	}
}

func registerAndLogin(t *testing.T, srv *httptest.Server, username, email, password string) (uuid.UUID, string) {
	t.Helper()

	// Register
	body, _ := json.Marshal(map[string]string{
		"username": username, "email": email, "password": password, "city": "almaty",
	})
	resp, err := http.Post(srv.URL+"/auth/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var regEnv map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&regEnv))
	resp.Body.Close()

	regData := regEnv["data"].(map[string]interface{})
	userIDStr := regData["id"].(string)
	userID, err := uuid.Parse(userIDStr)
	require.NoError(t, err)

	// Login
	body, _ = json.Marshal(map[string]string{"username": username, "password": password})
	resp, err = http.Post(srv.URL+"/auth/login", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var loginEnv map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&loginEnv))
	resp.Body.Close()

	loginData := loginEnv["data"].(map[string]interface{})
	accessToken := loginData["access_token"].(string)

	return userID, accessToken
}

func makeMove(t *testing.T, srv *httptest.Server, gameID uuid.UUID, move, token string) int {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"move": move})
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/games/%s/move", srv.URL, gameID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	return resp.StatusCode
}

func TestGameFlow(t *testing.T) {
	gtx := setupGameServer(t)
	defer gtx.srv.Close()

	suffix := fmt.Sprintf("%d", uniqueSuffix())
	whiteUsername := "white_" + suffix
	blackUsername := "black_" + suffix

	pool := setupDB(t)
	t.Cleanup(func() {
		pool.Exec(context.Background(),
			"DELETE FROM users WHERE username IN ($1, $2)", whiteUsername, blackUsername)
	})

	whiteID, whiteToken := registerAndLogin(t, gtx.srv, whiteUsername,
		"white_"+suffix+"@example.com", "Password123!")
	blackID, blackToken := registerAndLogin(t, gtx.srv, blackUsername,
		"black_"+suffix+"@example.com", "Password123!")

	// Create game directly — bypasses matchmaking (Matcher.interval is unexported/hardcoded).
	game, err := gtx.gameSvc.CreateGame(context.Background(), whiteID, blackID, 1200, 1200)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, game.ID)

	t.Cleanup(func() {
		pool.Exec(context.Background(),
			"DELETE FROM moves WHERE game_id = $1", game.ID)
		pool.Exec(context.Background(),
			"DELETE FROM games WHERE id = $1", game.ID)
	})

	// Play 5 legal moves alternating white / black.
	tokens := map[int]string{0: whiteToken, 1: blackToken, 2: whiteToken, 3: blackToken, 4: whiteToken}
	for i, mv := range legalMoves {
		status := makeMove(t, gtx.srv, game.ID, mv, tokens[i])
		assert.Equal(t, http.StatusOK, status, "move %d (%s) failed", i+1, mv)
	}

	// Verify: game record exists and has 5 moves in DB.
	req, _ := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/games/%s/moves", gtx.srv.URL, game.ID), nil)
	req.Header.Set("Authorization", "Bearer "+whiteToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var movesEnv map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&movesEnv))
	movesData, ok := movesEnv["data"].([]interface{})
	require.True(t, ok, "expected moves array in data")
	assert.Len(t, movesData, 5, "expected 5 moves recorded")
}
