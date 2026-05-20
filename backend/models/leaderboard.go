package models

import (
	"time"

	"github.com/google/uuid"
)

type LeaderboardEntry struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	City        string    `json:"city"`
	Rating      int       `json:"rating"`
	Rank        int       `json:"rank"`
	GamesPlayed int       `json:"games_played"`
	UpdatedAt   time.Time `json:"updated_at"`
}
