package websocket

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientMessage_JSONRoundTrip(t *testing.T) {
	gameID := uuid.New()
	original := ClientMessage{
		Type:    MessageTypeMoveMade,
		GameID:  gameID,
		Move:    "e2e4",
		Payload: map[string]interface{}{"key": "value"},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got ClientMessage
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, original.Type, got.Type)
	assert.Equal(t, original.GameID, got.GameID)
	assert.Equal(t, original.Move, got.Move)
}

func TestServerMessage_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		msg  ServerMessage
	}{
		{
			name: "game_start",
			msg:  ServerMessage{Type: MessageTypeGameStart, GameID: uuid.New(), Message: "game started"},
		},
		{
			name: "error",
			msg:  ServerMessage{Type: MessageTypeError, Error: "invalid_move", Message: "That move is not legal"},
		},
		{
			name: "pong",
			msg:  ServerMessage{Type: MessageTypePong},
		},
		{
			name: "queue_status",
			msg:  ServerMessage{Type: MessageTypeQueueStatus},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.msg)
			require.NoError(t, err)

			var got ServerMessage
			require.NoError(t, json.Unmarshal(data, &got))

			assert.Equal(t, tc.msg.Type, got.Type)
			assert.Equal(t, tc.msg.GameID, got.GameID)
			assert.Equal(t, tc.msg.Error, got.Error)
			assert.Equal(t, tc.msg.Message, got.Message)
		})
	}
}

func TestGameStartPayload_JSONRoundTrip(t *testing.T) {
	original := GameStartPayload{
		GameID:      uuid.New(),
		WhiteID:     uuid.New(),
		BlackID:     uuid.New(),
		WhiteRating: 1200,
		BlackRating: 1350,
		FEN:         "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		YourColor:   "white",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got GameStartPayload
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, original.GameID, got.GameID)
	assert.Equal(t, original.WhiteID, got.WhiteID)
	assert.Equal(t, original.BlackID, got.BlackID)
	assert.Equal(t, original.WhiteRating, got.WhiteRating)
	assert.Equal(t, original.BlackRating, got.BlackRating)
	assert.Equal(t, original.FEN, got.FEN)
	assert.Equal(t, original.YourColor, got.YourColor)

	// Verify exact JSON field names the frontend expects
	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))
	assert.Contains(t, raw, "game_id")
	assert.Contains(t, raw, "white_id")
	assert.Contains(t, raw, "black_id")
	assert.Contains(t, raw, "white_rating")
	assert.Contains(t, raw, "black_rating")
	assert.Contains(t, raw, "fen")
	assert.Contains(t, raw, "your_color")
}

func TestMoveMadePayload_JSONRoundTrip(t *testing.T) {
	original := MoveMadePayload{
		GameID:    uuid.New(),
		Move:      "e2e4",
		Notation:  "e4",
		FEN:       "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		MoveCount: 1,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got MoveMadePayload
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, original.GameID, got.GameID)
	assert.Equal(t, original.Move, got.Move)
	assert.Equal(t, original.Notation, got.Notation)
	assert.Equal(t, original.FEN, got.FEN)
	assert.Equal(t, original.MoveCount, got.MoveCount)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))
	assert.Contains(t, raw, "game_id")
	assert.Contains(t, raw, "move")
	assert.Contains(t, raw, "notation")
	assert.Contains(t, raw, "fen")
	assert.Contains(t, raw, "move_count")
}

func TestGameEndPayload_JSONRoundTrip(t *testing.T) {
	original := GameEndPayload{
		GameID:           uuid.New(),
		Result:           "1-0",
		Reason:           "checkmate",
		WhiteRatingDelta: 16,
		BlackRatingDelta: -16,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got GameEndPayload
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, original.GameID, got.GameID)
	assert.Equal(t, original.Result, got.Result)
	assert.Equal(t, original.Reason, got.Reason)
	assert.Equal(t, original.WhiteRatingDelta, got.WhiteRatingDelta)
	assert.Equal(t, original.BlackRatingDelta, got.BlackRatingDelta)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))
	assert.Contains(t, raw, "white_rating_delta")
	assert.Contains(t, raw, "black_rating_delta")
}

func TestQueueStatusPayload_JSONRoundTrip(t *testing.T) {
	original := QueueStatusPayload{
		Queued:       true,
		Position:     3,
		QueueSize:    10,
		EstimateWait: 30,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got QueueStatusPayload
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, original.Queued, got.Queued)
	assert.Equal(t, original.Position, got.Position)
	assert.Equal(t, original.QueueSize, got.QueueSize)
	assert.Equal(t, original.EstimateWait, got.EstimateWait)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))
	assert.Contains(t, raw, "queued")
	assert.Contains(t, raw, "queue_size")
	// position and estimate_wait are omitempty — present since non-zero
	assert.Contains(t, raw, "position")
	assert.Contains(t, raw, "estimate_wait_seconds")
}

func TestGameStatePayload_JSONRoundTrip(t *testing.T) {
	original := GameStatePayload{
		GameID:      uuid.New(),
		FEN:         "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		Move:        "e2e4",
		MoveCount:   1,
		YourTurn:    true,
		WhiteRating: 1200,
		BlackRating: 1350,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got GameStatePayload
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, original.GameID, got.GameID)
	assert.Equal(t, original.FEN, got.FEN)
	assert.Equal(t, original.Move, got.Move)
	assert.Equal(t, original.MoveCount, got.MoveCount)
	assert.Equal(t, original.YourTurn, got.YourTurn)
	assert.Equal(t, original.WhiteRating, got.WhiteRating)
	assert.Equal(t, original.BlackRating, got.BlackRating)

	// The Go field is named Move but the JSON tag is last_move — verify the frontend tag
	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))
	assert.Contains(t, raw, "last_move")
	assert.NotContains(t, raw, "move") // must NOT appear under "move"
}

func TestMessageTypeJSONValues(t *testing.T) {
	// Frontend decodes these exact string values — changing them is a breaking API change.
	tests := []struct {
		msgType  MessageType
		expected string
	}{
		{MessageTypeJoin, "join"},
		{MessageTypeQueueJoin, "queue_join"},
		{MessageTypeQueueLeave, "queue_leave"},
		{MessageTypeGameStart, "game_start"},
		{MessageTypeMoveMade, "move_made"},
		{MessageTypeMoveError, "move_error"},
		{MessageTypeGameEnd, "game_end"},
		{MessageTypeResign, "resign"},
		{MessageTypeError, "error"},
		{MessageTypeQueueStatus, "queue_status"},
		{MessageTypeGameState, "game_state"},
		{MessageTypePing, "ping"},
		{MessageTypePong, "pong"},
	}

	for _, tc := range tests {
		t.Run(string(tc.msgType), func(t *testing.T) {
			msg := ServerMessage{Type: tc.msgType}
			data, err := json.Marshal(msg)
			require.NoError(t, err)

			var raw map[string]interface{}
			require.NoError(t, json.Unmarshal(data, &raw))
			assert.Equal(t, tc.expected, raw["type"])
		})
	}
}
