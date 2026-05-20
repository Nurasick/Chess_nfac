package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/chess-nfac/backend/auth"
	"github.com/chess-nfac/backend/config"
	"github.com/chess-nfac/backend/utils"
)

type contextKey string

const UserIDKey contextKey = "userID"

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

func Auth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.RespondError(w, http.StatusUnauthorized, "missing_token", "Authorization header required")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				utils.RespondError(w, http.StatusUnauthorized, "invalid_token", "Invalid authorization format")
				return
			}

			claims, err := auth.ParseToken(parts[1], cfg.JWTSecret)
			if err != nil {
				utils.RespondError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
