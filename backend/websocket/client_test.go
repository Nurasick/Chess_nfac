package websocket

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_handleMessage_Ping(t *testing.T) {
	hub := NewHub(nil)
	c := makeTestClient(hub, uuid.New())

	c.handleMessage(&ClientMessage{Type: MessageTypePing})

	msg := drainSend(c)
	require.NotNil(t, msg)
	assert.Equal(t, MessageTypePong, msg.Type)
}

func TestClient_handleMessage_Unknown(t *testing.T) {
	hub := NewHub(nil)
	c := makeTestClient(hub, uuid.New())

	c.handleMessage(&ClientMessage{Type: MessageType("totally_unknown")})

	msg := drainSend(c)
	require.NotNil(t, msg)
	assert.Equal(t, MessageTypeError, msg.Type)
}

func TestClient_handleJoin_NilGameID(t *testing.T) {
	hub := NewHub(nil)
	c := makeTestClient(hub, uuid.New())

	c.handleJoin(&ClientMessage{Type: MessageTypeJoin, GameID: uuid.Nil})

	msg := drainSend(c)
	require.NotNil(t, msg)
	assert.Equal(t, MessageTypeError, msg.Type)
	assert.Equal(t, "invalid_game_id", msg.Error)
}

func TestClient_handleJoin_ValidGameID(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	userID := uuid.New()
	gameID := uuid.New()
	c := makeTestClient(hub, userID)
	hub.register <- c
	time.Sleep(20 * time.Millisecond)

	c.handleJoin(&ClientMessage{Type: MessageTypeJoin, GameID: gameID})
	time.Sleep(20 * time.Millisecond)

	assert.Equal(t, gameID, c.gameID)
	assert.Equal(t, 1, hub.GetGameClientCount(gameID))
}

func TestClient_handleQueueJoin(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	c := makeTestClient(hub, uuid.New())
	hub.register <- c
	time.Sleep(20 * time.Millisecond)

	// queueJoin is drained as no-op by hub.Run — just verify no deadlock/panic
	c.handleQueueJoin(&ClientMessage{Type: MessageTypeQueueJoin})
	time.Sleep(20 * time.Millisecond)
}

func TestClient_handleQueueLeave(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	c := makeTestClient(hub, uuid.New())
	hub.register <- c
	time.Sleep(20 * time.Millisecond)

	c.handleQueueLeave(&ClientMessage{Type: MessageTypeQueueLeave})
	time.Sleep(20 * time.Millisecond)
}

func TestClient_handleMoveMade_NilGameID(t *testing.T) {
	hub := NewHub(nil)
	c := makeTestClient(hub, uuid.New())

	c.handleMoveMade(&ClientMessage{Type: MessageTypeMoveMade, GameID: uuid.Nil, Move: "e2e4"})

	msg := drainSend(c)
	require.NotNil(t, msg)
	assert.Equal(t, MessageTypeError, msg.Type)
}

func TestClient_handleMoveMade_EmptyMove(t *testing.T) {
	hub := NewHub(nil)
	c := makeTestClient(hub, uuid.New())

	c.handleMoveMade(&ClientMessage{Type: MessageTypeMoveMade, GameID: uuid.New(), Move: ""})

	msg := drainSend(c)
	require.NotNil(t, msg)
	assert.Equal(t, MessageTypeError, msg.Type)
}

func TestClient_handleMoveMade_Valid(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	c := makeTestClient(hub, uuid.New())
	hub.register <- c
	time.Sleep(20 * time.Millisecond)

	c.handleMoveMade(&ClientMessage{Type: MessageTypeMoveMade, GameID: uuid.New(), Move: "e2e4"})
	time.Sleep(20 * time.Millisecond)
}

func TestClient_handleResign_NilGameID(t *testing.T) {
	hub := NewHub(nil)
	c := makeTestClient(hub, uuid.New())

	c.handleResign(&ClientMessage{Type: MessageTypeResign, GameID: uuid.Nil})

	msg := drainSend(c)
	require.NotNil(t, msg)
	assert.Equal(t, MessageTypeError, msg.Type)
}

func TestClient_handleResign_Valid(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	c := makeTestClient(hub, uuid.New())
	hub.register <- c
	time.Sleep(20 * time.Millisecond)

	c.handleResign(&ClientMessage{Type: MessageTypeResign, GameID: uuid.New()})
	time.Sleep(20 * time.Millisecond)
}

func TestClient_sendError(t *testing.T) {
	hub := NewHub(nil)
	c := makeTestClient(hub, uuid.New())

	c.sendError("some_code", "some message")

	msg := drainSend(c)
	require.NotNil(t, msg)
	assert.Equal(t, MessageTypeError, msg.Type)
	assert.Equal(t, "some_code", msg.Error)
	assert.Equal(t, "some message", msg.Message)
}

func TestHub_BroadcastToUser(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	userID := uuid.New()
	c := makeTestClient(hub, userID)
	hub.register <- c
	time.Sleep(20 * time.Millisecond)

	hub.BroadcastToUser(userID, &ServerMessage{Type: MessageTypePong})

	msg := drainSend(c)
	require.NotNil(t, msg)
	assert.Equal(t, MessageTypePong, msg.Type)
}

func TestHub_BroadcastToUser_NotOnline(t *testing.T) {
	hub := NewHub(nil)

	// Should not panic when user is not connected
	hub.BroadcastToUser(uuid.New(), &ServerMessage{Type: MessageTypePong})
}

func TestHub_GetClientByUserID(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	userID := uuid.New()
	c := makeTestClient(hub, userID)
	hub.register <- c
	time.Sleep(20 * time.Millisecond)

	got := hub.GetClientByUserID(userID)
	assert.Equal(t, c, got)

	assert.Nil(t, hub.GetClientByUserID(uuid.New()))
}

func TestHub_LogStatus(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	c := makeTestClient(hub, uuid.New())
	hub.register <- c
	time.Sleep(20 * time.Millisecond)

	// just verify no panic
	hub.LogStatus()
}
