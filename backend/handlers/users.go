package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/chess-nfac/backend/middleware"
	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/utils"
)

type UserGetterServicer interface {
	GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

type UserHandler struct {
	userService UserGetterServicer
}

func NewUserHandler(userService UserGetterServicer) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			utils.RespondAppError(w, appErr)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			utils.RespondAppError(w, appErr)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}
