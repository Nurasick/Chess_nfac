package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingInterval   = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   uuid.UUID
	gameID   uuid.UUID
	done     chan struct{}
	rating   int
}

func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		done:   make(chan struct{}),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	c.conn.SetReadLimit(maxMessageSize)

	for {
		var msg ClientMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("websocket error: %v\n", err)
			}
			return
		}

		c.handleMessage(&msg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.done:
			return
		}
	}
}

func (c *Client) handleMessage(msg *ClientMessage) {
	switch msg.Type {
	case MessageTypeJoin:
		c.handleJoin(msg)
	case MessageTypeQueueJoin:
		c.handleQueueJoin(msg)
	case MessageTypeQueueLeave:
		c.handleQueueLeave(msg)
	case MessageTypeMove:
		c.handleMove(msg)
	case MessageTypeMoveMade:
		c.handleMoveMade(msg)
	case MessageTypeResign:
		c.handleResign(msg)
	case MessageTypePing:
		c.sendMessage(&ServerMessage{Type: MessageTypePong})
	default:
		c.sendError("unknown_message_type", "Unknown message type")
	}
}

func (c *Client) handleJoin(msg *ClientMessage) {
	if msg.GameID == uuid.Nil {
		c.sendError("invalid_game_id", "Game ID is required")
		return
	}

	c.gameID = msg.GameID
	c.hub.joinGame <- &JoinGameRequest{
		client: c,
		gameID: msg.GameID,
	}
}

func (c *Client) handleQueueJoin(msg *ClientMessage) {
	fmt.Printf("[WS] user %s joining queue (rating %d)\n", c.userID, c.rating)
	c.hub.queueJoin <- &QueueRequest{
		client: c,
		userID: c.userID,
	}
}

func (c *Client) handleQueueLeave(msg *ClientMessage) {
	c.hub.queueLeave <- &QueueRequest{
		client: c,
		userID: c.userID,
	}
}

func (c *Client) handleMove(msg *ClientMessage) {
	if msg.Payload == nil {
		c.sendError("invalid_move", "Move payload required")
		return
	}

	gameIDStr, _ := msg.Payload["game_id"].(string)
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil || gameID == uuid.Nil {
		c.sendError("invalid_game_id", "Valid game_id required in payload")
		return
	}

	from, _ := msg.Payload["from"].(string)
	to, _ := msg.Payload["to"].(string)
	if from == "" || to == "" {
		c.sendError("invalid_move", "from and to squares required")
		return
	}

	c.hub.makeMove <- &MakeMoveRequest{
		client: c,
		gameID: gameID,
		move:   from + to,
	}
}

func (c *Client) handleMoveMade(msg *ClientMessage) {
	if msg.GameID == uuid.Nil {
		c.sendError("invalid_game_id", "Game ID is required")
		return
	}

	if msg.Move == "" {
		c.sendError("invalid_move", "Move notation is required")
		return
	}

	c.hub.makeMove <- &MakeMoveRequest{
		client: c,
		gameID: msg.GameID,
		move:   msg.Move,
	}
}

func (c *Client) handleResign(msg *ClientMessage) {
	if msg.GameID == uuid.Nil {
		c.sendError("invalid_game_id", "Game ID is required")
		return
	}

	c.hub.resign <- &ResignRequest{
		client: c,
		gameID: msg.GameID,
	}
}

func (c *Client) sendMessage(msg *ServerMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("error marshaling message: %v\n", err)
		return
	}

	select {
	case c.send <- data:
	default:
		fmt.Printf("message channel full for user %s\n", c.userID)
	}
}

func (c *Client) sendJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Printf("error marshaling message: %v\n", err)
		return
	}
	select {
	case c.send <- data:
	default:
		fmt.Printf("message channel full for user %s\n", c.userID)
	}
}

func (c *Client) sendError(code, message string) {
	c.sendMessage(&ServerMessage{
		Type:    MessageTypeError,
		Error:   code,
		Message: message,
	})
}

func (c *Client) Close() {
	close(c.done)
	close(c.send)
}
