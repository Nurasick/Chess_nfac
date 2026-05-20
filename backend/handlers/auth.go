package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/chess-nfac/backend/middleware"
	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/utils"
)

// UserServicer is the subset of service.UserService that AuthHandler needs.
type UserServicer interface {
	Register(ctx context.Context, username, email, password, city string) (*models.User, error)
	Login(ctx context.Context, username, password string) (*models.User, string, string, error)
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}

type AuthHandler struct {
	userService UserServicer
}

func NewAuthHandler(userService UserServicer) *AuthHandler {
	return &AuthHandler{userService: userService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		City     string `json:"city"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	user, err := h.userService.Register(r.Context(), req.Username, req.Email, req.Password, req.City)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			utils.RespondAppError(w, appErr)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	user, accessToken, refreshToken, err := h.userService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			utils.RespondAppError(w, appErr)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		utils.RespondError(w, http.StatusBadRequest, "missing_token", "refresh_token is required")
		return
	}

	accessToken, newRefreshToken, err := h.userService.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		if appErr, ok := err.(utils.AppError); ok {
			utils.RespondAppError(w, appErr)
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	if err := h.userService.Logout(r.Context(), userID); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
		return
	}

	utils.RespondJSON(w, http.StatusOK, nil)
}
