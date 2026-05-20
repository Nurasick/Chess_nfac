package websocket

import "github.com/google/uuid"

type MessageType string

const (
	MessageTypeJoin         MessageType = "join"
	MessageTypeQueueJoin    MessageType = "queue_join"
	MessageTypeQueueLeave   MessageType = "queue_leave"
	MessageTypeGameStart    MessageType = "game_start"
	MessageTypeMove         MessageType = "move"
	MessageTypeMoveMade     MessageType = "move_made"
	MessageTypeMoveError    MessageType = "move_error"
	MessageTypeGameEnd      MessageType = "game_end"
	MessageTypeResign       MessageType = "resign"
	MessageTypeError        MessageType = "error"
	MessageTypeQueueStatus  MessageType = "queue_status"
	MessageTypeGameState    MessageType = "game_state"
	MessageTypePing         MessageType = "ping"
	MessageTypePong         MessageType = "pong"
)

type ClientMessage struct {
	Type    MessageType            `json:"type"`
	GameID  uuid.UUID              `json:"game_id,omitempty"`
	Move    string                 `json:"move,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type ServerMessage struct {
	Type    MessageType            `json:"type"`
	GameID  uuid.UUID              `json:"game_id,omitempty"`
	Data    interface{}            `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Message string                 `json:"message,omitempty"`
}

type GameStartPayload struct {
	GameID      uuid.UUID `json:"game_id"`
	WhiteID     uuid.UUID `json:"white_id"`
	BlackID     uuid.UUID `json:"black_id"`
	WhiteRating int       `json:"white_rating"`
	BlackRating int       `json:"black_rating"`
	FEN         string    `json:"fen"`
	YourColor   string    `json:"your_color"`
}

type MoveMadePayload struct {
	GameID    uuid.UUID `json:"game_id"`
	Move      string    `json:"move"`
	Notation  string    `json:"notation"`
	FEN       string    `json:"fen"`
	MoveCount int       `json:"move_count"`
}

type GameEndPayload struct {
	GameID           uuid.UUID `json:"game_id"`
	Result           string    `json:"result"`
	Reason           string    `json:"reason"`
	WhiteRatingDelta int       `json:"white_rating_delta"`
	BlackRatingDelta int       `json:"black_rating_delta"`
}

type QueueStatusPayload struct {
	Queued       bool `json:"queued"`
	Position     int  `json:"position,omitempty"`
	QueueSize    int  `json:"queue_size"`
	EstimateWait int  `json:"estimate_wait_seconds,omitempty"`
}

type GameStatePayload struct {
	GameID      uuid.UUID `json:"game_id"`
	FEN         string    `json:"fen"`
	Move        string    `json:"last_move,omitempty"`
	MoveCount   int       `json:"move_count"`
	YourTurn    bool      `json:"your_turn"`
	WhiteRating int       `json:"white_rating"`
	BlackRating int       `json:"black_rating"`
}

// GameStartServerMessage is sent flat (no "data" wrapper) so the frontend can
// read game_id, fen, your_color etc. directly off the message object.
type GameStartServerMessage struct {
	Type        MessageType `json:"type"`
	GameID      uuid.UUID   `json:"game_id"`
	WhiteID     uuid.UUID   `json:"white_id"`
	BlackID     uuid.UUID   `json:"black_id"`
	WhiteRating int         `json:"white_rating"`
	BlackRating int         `json:"black_rating"`
	FEN         string      `json:"fen"`
	YourColor   string      `json:"your_color"`
}

// MoveMadeServerMessage is sent flat to each player with a personalised your_turn flag.
type MoveMadeServerMessage struct {
	Type      MessageType `json:"type"`
	GameID    uuid.UUID   `json:"game_id"`
	Move      string      `json:"move"`
	Notation  string      `json:"notation"`
	FEN       string      `json:"fen"`
	MoveCount int         `json:"move_count"`
	YourTurn  bool        `json:"your_turn"`
}

// GameEndServerMessage is sent flat to both players when the game finishes.
type GameEndServerMessage struct {
	Type             MessageType `json:"type"`
	GameID           uuid.UUID   `json:"game_id"`
	Result           string      `json:"result"`
	Reason           string      `json:"reason"`
	WhiteRatingDelta int         `json:"white_rating_delta"`
	BlackRatingDelta int         `json:"black_rating_delta"`
}

// GameMoveResult is returned by the GameMover interface after processing a move.
type GameMoveResult struct {
	FEN       string
	Move      string
	Notation  string
	MoveCount int
	WhiteID   uuid.UUID
	BlackID   uuid.UUID
	GameOver  bool
	Result    string
	Reason    string
}
