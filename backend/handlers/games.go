package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/chess-nfac/backend/middleware"
	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/utils"
)

type GameServicer interface {
	GetGame(ctx context.Context, gameID uuid.UUID) (*models.Game, error)
	GetMoves(ctx context.Context, gameID uuid.UUID) ([]models.Move, error)
	GetUserGames(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Game, int, error)
	ProcessMove(ctx context.Context, gameID, playerID uuid.UUID, moveNotation string) (*models.Game, *models.Move, bool, error)
	ResignGame(ctx context.Context, gameID, playerID uuid.UUID) (*models.Game, error)
}

type GameHandler struct {
	gameService GameServicer
}

func NewGameHandler(gameService GameServicer) *GameHandler {
	return &GameHandler{gameService: gameService}
}

func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	gameID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid_id", "Invalid game ID")
		return
	}

	game, err := h.gameService.GetGame(r.Context(), gameID)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			utils.RespondAppError(w, appErr)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	utils.RespondJSON(w, http.StatusOK, game)
}

func (h *GameHandler) GetMoves(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	gameID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid_id", "Invalid game ID")
		return
	}

	moves, err := h.gameService.GetMoves(r.Context(), gameID)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			utils.RespondAppError(w, appErr)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	utils.RespondJSON(w, http.StatusOK, moves)
}

func (h *GameHandler) MakeMove(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	gameID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid_id", "Invalid game ID")
		return
	}

	var req struct {
		Move string `json:"move"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}
	if req.Move == "" {
		utils.RespondError(w, http.StatusBadRequest, "missing_move", "move is required")
		return
	}

	game, move, over, err := h.gameService.ProcessMove(r.Context(), gameID, userID, req.Move)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			utils.RespondAppError(w, appErr)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"game":     game,
		"move":     move,
		"game_over": over,
	})
}

func (h *GameHandler) GetUserGames(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID")
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	pageSize := 20
	if s := r.URL.Query().Get("page_size"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	games, total, err := h.gameService.GetUserGames(r.Context(), userID, page, pageSize)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			utils.RespondAppError(w, appErr)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	utils.RespondPaginated(w, http.StatusOK, games, total, page, pageSize)
}

func (h *GameHandler) Resign(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	gameID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid_id", "Invalid game ID")
		return
	}

	game, err := h.gameService.ResignGame(r.Context(), gameID, userID)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			utils.RespondAppError(w, appErr)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	utils.RespondJSON(w, http.StatusOK, game)
}
