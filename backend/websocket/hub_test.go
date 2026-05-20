package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeTestClient builds a minimal Client for hub tests — no real WebSocket needed.
// Tests are in the same package so unexported fields are accessible.
func makeTestClient(hub *Hub, userID uuid.UUID) *Client {
	return &Client{
		hub:    hub,
		send:   make(chan []byte, 256),
		userID: userID,
		done:   make(chan struct{}),
	}
}

// drainSend returns the next message from the client's send buffer, or nil if empty.
func drainSend(c *Client) *ServerMessage {
	select {
	case data := <-c.send:
		var msg ServerMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil
		}
		return &msg
	default:
		return nil
	}
}

func TestHub_Register(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	userID := uuid.New()
	client := makeTestClient(hub, userID)

	hub.register <- client
	time.Sleep(20 * time.Millisecond) // wait for Run() to process

	assert.True(t, hub.IsUserOnline(userID))
}

func TestHub_Unregister(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	userID := uuid.New()
	client := makeTestClient(hub, userID)

	hub.register <- client
	time.Sleep(20 * time.Millisecond)

	require.True(t, hub.IsUserOnline(userID))

	hub.unregister <- client
	time.Sleep(20 * time.Millisecond)

	assert.False(t, hub.IsUserOnline(userID))
}

func TestHub_JoinGame(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	userID := uuid.New()
	gameID := uuid.New()
	client := makeTestClient(hub, userID)

	hub.register <- client
	time.Sleep(20 * time.Millisecond)

	hub.joinGame <- &JoinGameRequest{client: client, gameID: gameID}
	time.Sleep(20 * time.Millisecond)

	assert.Equal(t, 1, hub.GetGameClientCount(gameID))
}

func TestHub_BroadcastToGame_OnlyGameRoom(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	gameID := uuid.New()
	otherGameID := uuid.New()

	c1 := makeTestClient(hub, uuid.New()) // in gameID
	c2 := makeTestClient(hub, uuid.New()) // in otherGameID

	hub.register <- c1
	hub.register <- c2
	time.Sleep(20 * time.Millisecond)

	hub.joinGame <- &JoinGameRequest{client: c1, gameID: gameID}
	hub.joinGame <- &JoinGameRequest{client: c2, gameID: otherGameID}
	time.Sleep(20 * time.Millisecond)

	msg := &ServerMessage{Type: MessageTypeMoveMade, GameID: gameID}
	hub.BroadcastToGame(gameID, msg)

	// c1 (in gameID) must receive the broadcast
	got := drainSend(c1)
	require.NotNil(t, got, "c1 should have received the broadcast")
	assert.Equal(t, MessageTypeMoveMade, got.Type)

	// c2 (in a different game) must NOT receive it
	assert.Nil(t, drainSend(c2), "c2 should not receive a broadcast for a different game")
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	c1 := makeTestClient(hub, uuid.New())
	c2 := makeTestClient(hub, uuid.New())

	hub.register <- c1
	hub.register <- c2
	time.Sleep(20 * time.Millisecond)

	hub.broadcast <- &ServerMessage{Type: MessageTypePong}
	time.Sleep(20 * time.Millisecond)

	assert.NotNil(t, drainSend(c1))
	assert.NotNil(t, drainSend(c2))
}

func TestHub_RemoveGameClients(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	gameID := uuid.New()

	c1 := makeTestClient(hub, uuid.New())
	c2 := makeTestClient(hub, uuid.New())

	hub.register <- c1
	hub.register <- c2
	time.Sleep(20 * time.Millisecond)

	hub.joinGame <- &JoinGameRequest{client: c1, gameID: gameID}
	hub.joinGame <- &JoinGameRequest{client: c2, gameID: gameID}
	time.Sleep(20 * time.Millisecond)

	require.Equal(t, 2, hub.GetGameClientCount(gameID))

	hub.RemoveGameClients(gameID)

	assert.Equal(t, 0, hub.GetGameClientCount(gameID))
}
