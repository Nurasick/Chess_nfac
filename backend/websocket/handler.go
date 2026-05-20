package websocket

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/chess-nfac/backend/config"
	"github.com/chess-nfac/backend/repository"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const defaultRating = 1200

// HandleWebSocket upgrades the connection, authenticates via ?token= query param
// (browsers cannot send headers during WebSocket handshake), and registers the client.
// ratingRepo is optional — if omitted, connected clients start with defaultRating.
func HandleWebSocket(hub *Hub, cfg *config.Config, ratingRepos ...repository.RatingRepository) http.HandlerFunc {
	var ratingRepo repository.RatingRepository
	if len(ratingRepos) > 0 {
		ratingRepo = ratingRepos[0]
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Browsers can't send Authorization headers during WS handshake; read from ?token=
		tokenString := r.URL.Query().Get("token")
		if tokenString == "" {
			// Fallback: Authorization header (useful for non-browser clients / tests)
			tokenString = r.Header.Get("Authorization")
			if strings.HasPrefix(tokenString, "Bearer ") {
				tokenString = tokenString[7:]
			}
		}
		if tokenString == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		claims := &jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		userIDStr, ok := (*claims)["user_id"].(string)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "invalid user id", http.StatusUnauthorized)
			return
		}

		rating := defaultRating
		if ratingRepo != nil {
			if r, err := ratingRepo.GetUserRating(context.Background(), userID); err == nil {
				rating = r
			}
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("websocket upgrade error: %v\n", err)
			return
		}

		client := NewClient(hub, conn, userID)
		client.rating = rating
		hub.register <- client

		go client.ReadPump()
		go client.WritePump()
	}
}
