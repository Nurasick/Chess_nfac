package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/chess-nfac/backend/config"
	"github.com/chess-nfac/backend/handlers"
	"github.com/chess-nfac/backend/middleware"
	"github.com/chess-nfac/backend/repository"
	"github.com/chess-nfac/backend/websocket"
)

func New(
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	gameHandler *handlers.GameHandler,
	leaderboardHandler *handlers.LeaderboardHandler,
	hub *websocket.Hub,
	ratingRepo repository.RatingRepository,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.Logging)
	r.Use(middleware.CORS(cfg))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.Refresh)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg))

			r.Post("/auth/logout", authHandler.Logout)

			r.Get("/users/me", userHandler.GetMe)
			r.Get("/users/{id}", userHandler.GetByID)
			r.Get("/users/{id}/games", gameHandler.GetUserGames)

			r.Get("/games/{id}", gameHandler.GetGame)
			r.Get("/games/{id}/moves", gameHandler.GetMoves)
			r.Post("/games/{id}/move", gameHandler.MakeMove)
			r.Post("/games/{id}/resign", gameHandler.Resign)

			r.Get("/leaderboard/{city}", leaderboardHandler.GetByCity)
		})
	})

	r.Get("/ws", websocket.HandleWebSocket(hub, cfg, ratingRepo))

	return r
}
