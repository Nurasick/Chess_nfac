package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// dialTestHub spins up a test HTTP server that upgrades connections and runs
// NewClient + WritePump + ReadPump. Returns the hub, server, and dialed WS conn.
func dialTestHub(t *testing.T) (*Hub, *httptest.Server, *websocket.Conn, uuid.UUID) {
	t.Helper()
	hub := NewHub(nil)
	go hub.Run()

	userID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := NewClient(hub, conn, userID)
		hub.register <- client
		go client.WritePump()
		client.ReadPump() // blocks until connection closes
	}))

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	time.Sleep(30 * time.Millisecond) // wait for register + pumps to start
	return hub, server, conn, userID
}

func TestReadPump_PingMessage(t *testing.T) {
	hub, server, conn, _ := dialTestHub(t)
	defer server.Close()
	defer conn.Close()

	// Send ping from client → server ReadPump dispatches → WritePump writes pong back
	msg := map[string]interface{}{"type": "ping"}
	require.NoError(t, conn.WriteJSON(msg))

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var reply ServerMessage
	require.NoError(t, conn.ReadJSON(&reply))
	assert.Equal(t, MessageTypePong, reply.Type)
	_ = hub
}

func TestReadPump_UnknownMessage(t *testing.T) {
	hub, server, conn, _ := dialTestHub(t)
	defer server.Close()
	defer conn.Close()

	msg := map[string]interface{}{"type": "totally_unknown"}
	require.NoError(t, conn.WriteJSON(msg))

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var reply ServerMessage
	require.NoError(t, conn.ReadJSON(&reply))
	assert.Equal(t, MessageTypeError, reply.Type)
	_ = hub
}

func TestReadPump_JoinMessage(t *testing.T) {
	hub, server, conn, _ := dialTestHub(t)
	defer server.Close()
	defer conn.Close()

	gameID := uuid.New()
	msg := map[string]interface{}{"type": "join", "game_id": gameID.String()}
	require.NoError(t, conn.WriteJSON(msg))
	time.Sleep(30 * time.Millisecond)

	assert.Equal(t, 1, hub.GetGameClientCount(gameID))
}

func TestReadPump_QueueJoinLeave(t *testing.T) {
	_, server, conn, _ := dialTestHub(t)
	defer server.Close()
	defer conn.Close()

	// queue_join and queue_leave are no-ops in Run — just verify no panic/deadlock
	require.NoError(t, conn.WriteJSON(map[string]interface{}{"type": "queue_join"}))
	require.NoError(t, conn.WriteJSON(map[string]interface{}{"type": "queue_leave"}))
	time.Sleep(30 * time.Millisecond)
}

func TestReadPump_MoveMadeInvalid(t *testing.T) {
	_, server, conn, _ := dialTestHub(t)
	defer server.Close()
	defer conn.Close()

	// empty move → error response
	msg := map[string]interface{}{"type": "move_made", "game_id": uuid.New().String(), "move": ""}
	require.NoError(t, conn.WriteJSON(msg))

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var reply ServerMessage
	require.NoError(t, conn.ReadJSON(&reply))
	assert.Equal(t, MessageTypeError, reply.Type)
}

func TestReadPump_MoveMadeValid(t *testing.T) {
	_, server, conn, _ := dialTestHub(t)
	defer server.Close()
	defer conn.Close()

	// valid move → sent to makeMove channel (no-op in Run)
	msg := map[string]interface{}{"type": "move_made", "game_id": uuid.New().String(), "move": "e2e4"}
	require.NoError(t, conn.WriteJSON(msg))
	time.Sleep(30 * time.Millisecond)
}

func TestReadPump_ResignValid(t *testing.T) {
	_, server, conn, _ := dialTestHub(t)
	defer server.Close()
	defer conn.Close()

	msg := map[string]interface{}{"type": "resign", "game_id": uuid.New().String()}
	require.NoError(t, conn.WriteJSON(msg))
	time.Sleep(30 * time.Millisecond)
}

func TestReadPump_ConnectionClose(t *testing.T) {
	hub, server, conn, userID := dialTestHub(t)
	defer server.Close()

	require.True(t, hub.IsUserOnline(userID))

	// closing the dialed conn causes ReadPump to detect close and unregister
	conn.Close()
	time.Sleep(50 * time.Millisecond)

	assert.False(t, hub.IsUserOnline(userID))
}

func TestHandleMessage_AllCases(t *testing.T) {
	hub := NewHub(nil)
	go hub.Run()

	userID := uuid.New()
	c := makeTestClient(hub, userID)
	hub.register <- c
	time.Sleep(20 * time.Millisecond)

	cases := []ClientMessage{
		{Type: MessageTypeJoin, GameID: uuid.New()},
		{Type: MessageTypeQueueJoin},
		{Type: MessageTypeQueueLeave},
		{Type: MessageTypeMoveMade, GameID: uuid.New(), Move: "e2e4"},
		{Type: MessageTypeResign, GameID: uuid.New()},
		{Type: MessageTypePing},
		{Type: MessageType("no_such_type")},
	}

	for _, msg := range cases {
		msg := msg
		c.handleMessage(&msg)
		time.Sleep(10 * time.Millisecond) // allow hub to drain channel-based cases
	}
}
