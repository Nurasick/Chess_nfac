package websocket

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)


// Queuer is the subset of matchmaking.QueueInterface the hub needs.
type Queuer interface {
	Enqueue(ctx context.Context, userID uuid.UUID, rating int) error
	Dequeue(ctx context.Context, userID uuid.UUID) error
}

// GameMover is the subset of the game service the hub needs to process moves.
type GameMover interface {
	HandleMove(ctx context.Context, gameID, playerID uuid.UUID, move string) (*GameMoveResult, error)
}

type moveResultEvent struct {
	req    *MakeMoveRequest
	result *GameMoveResult
	err    error
}

type JoinGameRequest struct {
	client *Client
	gameID uuid.UUID
}

type QueueRequest struct {
	client *Client
	userID uuid.UUID
}

type MakeMoveRequest struct {
	client *Client
	gameID uuid.UUID
	move   string
}

type ResignRequest struct {
	client *Client
	gameID uuid.UUID
}

// GameStartedEvent carries data needed to notify both players that a game has begun.
type GameStartedEvent struct {
	GameID      uuid.UUID
	WhiteID     uuid.UUID
	BlackID     uuid.UUID
	WhiteRating int
	BlackRating int
	FEN         string
}

type Hub struct {
	clients     map[uuid.UUID]*Client
	gameClients map[uuid.UUID]map[*Client]bool
	register    chan *Client
	unregister  chan *Client
	broadcast   chan *ServerMessage
	joinGame    chan *JoinGameRequest
	queueJoin   chan *QueueRequest
	queueLeave  chan *QueueRequest
	makeMove    chan *MakeMoveRequest
	resign      chan *ResignRequest
	gameStarted chan *GameStartedEvent
	moveResult  chan *moveResultEvent
	queue       Queuer
	gameService GameMover
}

func NewHub(queue Queuer, gameMover ...GameMover) *Hub {
	var gm GameMover
	if len(gameMover) > 0 {
		gm = gameMover[0]
	}
	return &Hub{
		clients:     make(map[uuid.UUID]*Client),
		gameClients: make(map[uuid.UUID]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *ServerMessage),
		joinGame:    make(chan *JoinGameRequest),
		queueJoin:   make(chan *QueueRequest),
		queueLeave:  make(chan *QueueRequest),
		makeMove:    make(chan *MakeMoveRequest),
		resign:      make(chan *ResignRequest),
		gameStarted: make(chan *GameStartedEvent, 16),
		moveResult:  make(chan *moveResultEvent, 32),
		queue:       queue,
		gameService: gm,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.userID] = client

		case client := <-h.unregister:
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				client.Close()
			}

		case msg := <-h.broadcast:
			for _, client := range h.clients {
				client.sendMessage(msg)
			}

		case req := <-h.joinGame:
			if _, ok := h.gameClients[req.gameID]; !ok {
				h.gameClients[req.gameID] = make(map[*Client]bool)
			}
			h.gameClients[req.gameID][req.client] = true

		case req := <-h.queueJoin:
			if h.queue != nil {
				if err := h.queue.Enqueue(context.Background(), req.userID, req.client.rating); err != nil {
					req.client.sendError("queue_error", "Failed to join queue")
				}
			}

		case req := <-h.queueLeave:
			if h.queue != nil {
				_ = h.queue.Dequeue(context.Background(), req.userID)
			}

		case event := <-h.gameStarted:
			h.handleGameStarted(event)

		case req := <-h.makeMove:
			if h.gameService != nil {
				go func(r *MakeMoveRequest) {
					res, err := h.gameService.HandleMove(context.Background(), r.gameID, r.client.userID, r.move)
					h.moveResult <- &moveResultEvent{req: r, result: res, err: err}
				}(req)
			}

		case ev := <-h.moveResult:
			h.handleMoveResult(ev)

		case <-h.resign:
			// handled externally by game service
		}
	}
}

func (h *Hub) handleGameStarted(event *GameStartedEvent) {
	fmt.Printf("[Hub] game_start event: game=%s white=%s black=%s\n", event.GameID, event.WhiteID, event.BlackID)
	if white, ok := h.clients[event.WhiteID]; ok {
		fmt.Printf("[Hub] sending game_start to white player %s\n", event.WhiteID)
		white.sendJSON(&GameStartServerMessage{
			Type:        MessageTypeGameStart,
			GameID:      event.GameID,
			WhiteID:     event.WhiteID,
			BlackID:     event.BlackID,
			WhiteRating: event.WhiteRating,
			BlackRating: event.BlackRating,
			FEN:         event.FEN,
			YourColor:   "white",
		})
	}
	if black, ok := h.clients[event.BlackID]; ok {
		fmt.Printf("[Hub] sending game_start to black player %s\n", event.BlackID)
		black.sendJSON(&GameStartServerMessage{
			Type:        MessageTypeGameStart,
			GameID:      event.GameID,
			WhiteID:     event.WhiteID,
			BlackID:     event.BlackID,
			WhiteRating: event.WhiteRating,
			BlackRating: event.BlackRating,
			FEN:         event.FEN,
			YourColor:   "black",
		})
	}
}

// NotifyGameStarted sends a game_start event to both matched players.
// Safe to call from any goroutine.
func (h *Hub) NotifyGameStarted(event *GameStartedEvent) {
	h.gameStarted <- event
}

func (h *Hub) BroadcastToGame(gameID uuid.UUID, msg *ServerMessage) {
	if clients, ok := h.gameClients[gameID]; ok {
		for client := range clients {
			client.sendMessage(msg)
		}
	}
}

func (h *Hub) BroadcastToUser(userID uuid.UUID, msg *ServerMessage) {
	if client, ok := h.clients[userID]; ok {
		client.sendMessage(msg)
	}
}

func (h *Hub) RemoveGameClients(gameID uuid.UUID) {
	if clients, ok := h.gameClients[gameID]; ok {
		for client := range clients {
			delete(clients, client)
		}
		delete(h.gameClients, gameID)
	}
}

func (h *Hub) GetClientByUserID(userID uuid.UUID) *Client {
	return h.clients[userID]
}

func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	_, ok := h.clients[userID]
	return ok
}

func (h *Hub) GetGameClientCount(gameID uuid.UUID) int {
	if clients, ok := h.gameClients[gameID]; ok {
		return len(clients)
	}
	return 0
}

func (h *Hub) LogStatus() {
	fmt.Printf("Hub status: %d clients, %d active games\n", len(h.clients), len(h.gameClients))
}

func (h *Hub) handleMoveResult(ev *moveResultEvent) {
	if ev.err != nil {
		ev.req.client.sendError("move_error", "Illegal move or not your turn")
		return
	}
	r := ev.result

	fenParts := strings.Fields(r.FEN)
	whiteTurn := len(fenParts) >= 2 && fenParts[1] == "w"

	send := func(playerID uuid.UUID, yourTurn bool) {
		client, ok := h.clients[playerID]
		if !ok {
			return
		}
		client.sendJSON(&MoveMadeServerMessage{
			Type:      MessageTypeMoveMade,
			GameID:    ev.req.gameID,
			Move:      r.Move,
			Notation:  r.Notation,
			FEN:       r.FEN,
			MoveCount: r.MoveCount,
			YourTurn:  yourTurn,
		})
	}

	send(r.WhiteID, whiteTurn)
	send(r.BlackID, !whiteTurn)

	if r.GameOver {
		endMsg := &GameEndServerMessage{
			Type:   MessageTypeGameEnd,
			GameID: ev.req.gameID,
			Result: r.Result,
			Reason: r.Reason,
		}
		if white, ok := h.clients[r.WhiteID]; ok {
			white.sendJSON(endMsg)
		}
		if black, ok := h.clients[r.BlackID]; ok {
			black.sendJSON(endMsg)
		}
	}
}
