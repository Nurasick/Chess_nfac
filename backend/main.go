package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chess-nfac/backend/cache"
	"github.com/chess-nfac/backend/chess"
	"github.com/chess-nfac/backend/config"
	"github.com/chess-nfac/backend/db"
	"github.com/chess-nfac/backend/handlers"
	"github.com/chess-nfac/backend/matchmaking"
	"github.com/chess-nfac/backend/repository"
	"github.com/chess-nfac/backend/router"
	"github.com/chess-nfac/backend/service"
	"github.com/chess-nfac/backend/websocket"
	"github.com/google/uuid"
)

type gameMoverAdapter struct {
	svc *service.GameService
}

func (a *gameMoverAdapter) HandleMove(ctx context.Context, gameID, playerID uuid.UUID, move string) (*websocket.GameMoveResult, error) {
	game, m, over, err := a.svc.ProcessMove(ctx, gameID, playerID, move)
	if err != nil {
		return nil, err
	}
	result := ""
	if game.Result != nil {
		result = string(*game.Result)
	}
	reason := ""
	return &websocket.GameMoveResult{
		FEN:       game.FEN,
		Move:      m.Notation,
		Notation:  m.Notation,
		MoveCount: m.MoveNumber,
		WhiteID:   game.WhiteID,
		BlackID:   game.BlackID,
		GameOver:  over,
		Result:    result,
		Reason:    reason,
	}, nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	databaseURL := os.Getenv("DATABASE_URL")
	log.Println("Running database migrations...")
	if err := runMigrations(databaseURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Database
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Redis
	redisClient, err := cache.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer redisClient.Close()

	// Repositories
	userRepo := repository.NewPostgresUserRepository(pool)
	gameRepo := repository.NewPostgresGameRepository(pool)
	moveRepo := repository.NewPostgresMoveRepository(pool)
	ratingRepo := repository.NewPostgresRatingRepository(pool)
	leaderboardRepo := repository.NewPostgresLeaderboardRepository(pool)

	// Services
	ratingService := service.NewRatingService(ratingRepo)
	engine := chess.NewEngine()
	userService := service.NewUserService(userRepo, cfg.JWTSecret)
	gameService := service.NewGameService(gameRepo, moveRepo, ratingService, engine)
	leaderboardService := service.NewLeaderboardService(leaderboardRepo)

	// Background jobs
	leaderboardService.StartRefreshJob(ctx, 5*time.Minute)

	// WebSocket hub — wire up the matchmaking queue so queue_join/leave messages work
	queue := matchmaking.NewQueue(redisClient)
	hub := websocket.NewHub(queue, &gameMoverAdapter{svc: gameService})
	go hub.Run()

	// Matchmaking
	matcher := matchmaking.NewMatcher(redisClient)
	go matcher.Start(ctx, func(match *matchmaking.Match) error {
		whiteRating, err := ratingRepo.GetUserRating(ctx, match.WhiteUserID)
		if err != nil {
			return fmt.Errorf("matchmaking: get white rating: %w", err)
		}
		blackRating, err := ratingRepo.GetUserRating(ctx, match.BlackUserID)
		if err != nil {
			return fmt.Errorf("matchmaking: get black rating: %w", err)
		}

		game, err := gameService.CreateGame(ctx, match.WhiteUserID, match.BlackUserID, whiteRating, blackRating)
		if err != nil {
			return fmt.Errorf("matchmaking: create game: %w", err)
		}

		log.Printf("game started: %s (white: %s vs black: %s)", game.ID, match.WhiteUserID, match.BlackUserID)

		hub.NotifyGameStarted(&websocket.GameStartedEvent{
			GameID:      game.ID,
			WhiteID:     game.WhiteID,
			BlackID:     game.BlackID,
			WhiteRating: whiteRating,
			BlackRating: blackRating,
			FEN:         game.FEN,
		})
		return nil
	})

	// Handlers
	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	gameHandler := handlers.NewGameHandler(gameService)
	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardService)

	// Router
	h := router.New(cfg, authHandler, userHandler, gameHandler, leaderboardHandler, hub, ratingRepo)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
