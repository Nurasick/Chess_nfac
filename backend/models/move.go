package models

import (
	"time"

	"github.com/google/uuid"
)

type Move struct {
	ID        uuid.UUID `json:"id"`
	GameID    uuid.UUID `json:"game_id"`
	PlayerID  uuid.UUID `json:"player_id"`
	MoveNumber int      `json:"move_number"`
	Notation  string    `json:"notation"`
	FENAfter  string    `json:"fen_after"`
	CreatedAt time.Time `json:"created_at"`
}
