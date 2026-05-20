package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/utils"
)

type LeaderboardServicer interface {
	GetByCity(ctx context.Context, city string, page, pageSize int) ([]models.LeaderboardEntry, int, error)
}

type LeaderboardHandler struct {
	leaderboardService LeaderboardServicer
}

func NewLeaderboardHandler(leaderboardService LeaderboardServicer) *LeaderboardHandler {
	return &LeaderboardHandler{leaderboardService: leaderboardService}
}

func (h *LeaderboardHandler) GetByCity(w http.ResponseWriter, r *http.Request) {
	city := chi.URLParam(r, "city")

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

	entries, total, err := h.leaderboardService.GetByCity(r.Context(), city, page, pageSize)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			utils.RespondAppError(w, appErr)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	utils.RespondPaginated(w, http.StatusOK, entries, total, page, pageSize)
}
