package models

import (
	"time"

	"github.com/google/uuid"
)

type GameStatus string

const (
	GameStatusWaiting   GameStatus = "waiting"
	GameStatusActive    GameStatus = "active"
	GameStatusCompleted GameStatus = "completed"
	GameStatusAbandoned GameStatus = "abandoned"
)

type GameResult string

const (
	GameResultWhiteWins GameResult = "white_wins"
	GameResultBlackWins GameResult = "black_wins"
	GameResultDraw      GameResult = "draw"
)

type Game struct {
	ID                  uuid.UUID   `json:"id"`
	WhiteID             uuid.UUID   `json:"white_id"`
	BlackID             uuid.UUID   `json:"black_id"`
	Status              GameStatus  `json:"status"`
	Result              *GameResult `json:"result"`
	PGN                 *string     `json:"pgn"`
	FEN                 string      `json:"fen"`
	WhiteRatingBefore   *int        `json:"white_rating_before"`
	WhiteRatingAfter    *int        `json:"white_rating_after"`
	BlackRatingBefore   *int        `json:"black_rating_before"`
	BlackRatingAfter    *int        `json:"black_rating_after"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
	FinishedAt          *time.Time  `json:"finished_at"`
}
