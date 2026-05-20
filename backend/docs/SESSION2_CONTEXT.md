# Session 2 Context — Backend: Game + Matchmaking + WebSocket

> Read this file at the start of Session 2. It captures what has been built and tested so you
> do NOT need to scan the codebase from scratch. All coverage numbers are current as of Session 1 end.

## Coverage Status

| Package        | Coverage | Status       |
|----------------|----------|--------------|
| `auth/`        | 82.6%    | ✅ done      |
| `service/`     | 84.1%    | ✅ done      |
| `handlers/`    | 85.0%    | ✅ done      |
| `matchmaking/` | 0%       | ❌ Session 2 |
| `websocket/`   | 0%       | ❌ Session 2 |
| `chess/`       | —        | no tests needed (pure wrapper) |

---

## What Exists and Passes

### Test files (all passing)

| File | What it covers |
|------|---------------|
| `auth/password_test.go` | bcrypt hash/verify |
| `auth/jwt_test.go` | generate/parse access+refresh tokens |
| `service/mock_user_repo_test.go` | shared `mockUserRepo` used by user_service_test |
| `service/user_service_test.go` | Register, Login, GetProfile, UpdateProfile |
| `service/rating_service_test.go` | CalculateNewRatings (K-32/16, floor 100), ApplyRatingChange |
| `service/game_service_test.go` | CreateGame, GetGame, GetMoves, ProcessMove (incl. checkmate), ResignGame, determineResult |
| `service/leaderboard_service_test.go` | GetByCity (pagination, city validation), Refresh |
| `handlers/auth_test.go` | POST /register, POST /login, POST /refresh, POST /logout |
| `handlers/users_test.go` | GET /users/me, PUT /users/me |
| `handlers/games_test.go` | POST /games, GET /games/{id}, POST /games/{id}/move, POST /games/{id}/resign |
| `handlers/leaderboard_test.go` | GET /leaderboard?city=&page=&page_size= |

### Key patterns established (re-use these, don't reinvent)

**AppError (value type, not pointer):**
```go
// Define
return utils.AppError{Code: "move_invalid", Message: "That move is not legal"}

// Check in tests
var appErr utils.AppError
require.True(t, errors.As(err, &appErr))
assert.Equal(t, "move_invalid", appErr.Code)
```

**Chi URL params in handler tests:**
```go
rctx := chi.NewRouteContext()
rctx.URLParams.Add("id", gameID.String())
r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
```

**Authenticated user in handler tests:**
```go
r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, userID))
```

**Response envelope** (all handlers use this):
```go
utils.RespondJSON(w, http.StatusOK, data)
utils.RespondError(w, http.StatusBadRequest, "error_code", "Human message")
utils.RespondPaginated(w, http.StatusOK, entries, total, page, pageSize)
```

**Factory function pattern** (prevents pointer mutation across subtests):
```go
newActiveGame := func() *models.Game {
    wr, br := 1200, 1200
    return &models.Game{
        ID: gameID, WhiteID: whiteID, BlackID: blackID,
        Status: models.GameStatusActive, FEN: startingFEN,
        WhiteRatingBefore: &wr, BlackRatingBefore: &br,
    }
}
```

**Fool's Mate FEN** (used in game_service_test for checkmate):
```
rnbqkbnr/pppp1ppp/8/4p3/6P1/5P2/PPPPP2P/RNBQKBNR b KQkq g3 0 2
// move d8h4 → checkmate
```

---

## Session 2 Targets

### `matchmaking/` — needs interface extraction + tests

**Current problem:** `Matcher` and `Queue` both hold concrete `*redis.Client`. Tests require either:
- A real Redis (integration test) — acceptable, or
- Extract a `QueueInterface` and mock it — preferred for unit tests

**`matchmaking/matcher.go` — key structure:**
```go
type Matcher struct {
    queue    *Queue          // ← needs to become QueueInterface
    client   *redis.Client   // ← used only by GetQueuePosition and IsQueuedUser
    interval time.Duration
}

func NewMatcher(client *redis.Client) *Matcher

// Methods:
func (m *Matcher) Start(ctx context.Context, matchCallback func(match *Match) error)
func (m *Matcher) tryMatch(ctx context.Context, matchCallback func(match *Match) error) error
func (m *Matcher) canMatch(player1, player2 QueuedPlayer) bool
func (m *Matcher) performMatch(ctx context.Context, player1, player2 QueuedPlayer, matchCallback func(match *Match) error) error
func (m *Matcher) GetQueuePosition(ctx context.Context, userID uuid.UUID) (int, error)
func (m *Matcher) IsQueuedUser(ctx context.Context, userID uuid.UUID) (bool, error)
```

**`matchmaking/queue.go` — key structure:**
```go
type QueuedPlayer struct {
    UserID uuid.UUID
    Rating int
}

type Queue struct { client *redis.Client }

func NewQueue(client *redis.Client) *Queue

// Methods:
func (q *Queue) Enqueue(ctx context.Context, userID uuid.UUID, rating int) error
func (q *Queue) Dequeue(ctx context.Context, userID uuid.UUID) error
func (q *Queue) GetAll(ctx context.Context) ([]QueuedPlayer, error)
func (q *Queue) GetByRatingRange(ctx context.Context, minRating, maxRating int) ([]QueuedPlayer, error)
```

**Required test cases for `matchmaking/`:**
- `canMatch`: within ±200 ELO → true; outside range → false
- `tryMatch`: fewer than 2 players → no match; 2 matching players → callback called
- `performMatch`: callback success → both dequeued; callback failure → both re-enqueued
- `GetQueuePosition`: user in queue → correct rank; user not in queue → -1
- `IsQueuedUser`: queued user → true; non-queued user → false

### `websocket/` — needs tests without a real WebSocket server

**`websocket/messages.go` — all types:**
```go
type MessageType string

// Constants:
MessageTypeJoin, MessageTypeQueueJoin, MessageTypeQueueLeave,
MessageTypeGameStart, MessageTypeMoveMade, MessageTypeMoveError,
MessageTypeGameEnd, MessageTypeResign, MessageTypeError,
MessageTypeQueueStatus, MessageTypeGameState, MessageTypePing, MessageTypePong

// Wire types:
ClientMessage  { Type, GameID, Move, Payload }
ServerMessage  { Type, GameID, Data, Error, Message }

// Payload types:
GameStartPayload   { GameID, WhiteID, BlackID, WhiteRating, BlackRating, FEN, YourColor }
MoveMadePayload    { GameID, Move, Notation, FEN, MoveCount }
GameEndPayload     { GameID, Result, Reason, WhiteRatingDelta, BlackRatingDelta }
QueueStatusPayload { Queued, Position, QueueSize, EstimateWait }
GameStatePayload   { GameID, FEN, Move (last_move), MoveCount, YourTurn, WhiteRating, BlackRating }
```

**`websocket/hub.go` — key structure:**
```go
type Hub struct {
    clients     map[uuid.UUID]*Client
    gameClients map[uuid.UUID]map[*Client]bool
    // channels: register, unregister, broadcast, joinGame, queueJoin, queueLeave, makeMove, resign
}

func NewHub() *Hub
func (h *Hub) Run()  // owns all state; must run in goroutine

// Public methods for testing:
func (h *Hub) BroadcastToGame(gameID uuid.UUID, msg *ServerMessage)
func (h *Hub) BroadcastToUser(userID uuid.UUID, msg *ServerMessage)
func (h *Hub) RemoveGameClients(gameID uuid.UUID)
func (h *Hub) GetClientByUserID(userID uuid.UUID) *Client
func (h *Hub) IsUserOnline(userID uuid.UUID) bool
func (h *Hub) GetGameClientCount(gameID uuid.UUID) int
```

**Note:** `queueJoin`, `queueLeave`, `makeMove`, `resign` channels are **no-ops** in `Run()` — handled externally. Only `register`, `unregister`, `broadcast`, and `joinGame` need to be tested via Hub.

**Required test cases for `websocket/`:**
- `messages_test.go`: JSON round-trip marshal/unmarshal for ClientMessage, ServerMessage, and all payload types
- `hub_test.go`: register client → IsUserOnline true; unregister → IsUserOnline false; joinGame → GetGameClientCount; BroadcastToGame sends to correct clients only; RemoveGameClients clears state

**Hub testing approach** (use gorilla/websocket test dialer — no real TCP needed):
```go
// Server side: use httptest.NewServer + websocket.Upgrader
// Client side: websocket.DefaultDialer.Dial(server.URL → ws://...)
// Then send hub.register <- client and assert via hub.IsUserOnline
```

---

## Interface Definitions (repository layer)

All services depend on these interfaces, not concrete types:

```go
// repository/user_repo.go
type UserRepository interface {
    Create(ctx, *models.User) error
    FindByUsername(ctx, username string) (*models.User, error)
    FindByID(ctx, id uuid.UUID) (*models.User, error)
    Update(ctx, *models.User) error
}

// repository/game_repo.go
type GameRepository interface {
    Create(ctx, *models.Game) error
    FindByID(ctx, id uuid.UUID) (*models.Game, error)
    UpdateStatus(ctx, gameID uuid.UUID, status models.GameStatus) error
    Finish(ctx, gameID uuid.UUID, result models.GameResult, whiteRatingAfter, blackRatingAfter int) error
    FindActiveByUserID(ctx, userID uuid.UUID) (*models.Game, error)
    UpdateFEN(ctx, gameID uuid.UUID, fen string) error
    GetUserGameCount(ctx, userID uuid.UUID) (int, error)
}

// repository/move_repo.go
type MoveRepository interface {
    Create(ctx, *models.Move) error
    FindByGameID(ctx, gameID uuid.UUID) ([]models.Move, error)
}

// repository/rating_repo.go
type RatingRepository interface {
    ApplyChange(ctx, *models.RatingChange) error
    GetUserRating(ctx, userID uuid.UUID) (int, error)
}

// repository/leaderboard_repo.go
type LeaderboardRepository interface {
    GetByCity(ctx, city string, limit, offset int) ([]models.LeaderboardEntry, int, error)
    Refresh(ctx) error
}
```

---

## Model Enums (needed in tests)

```go
// models/game.go
type GameStatus string
const (
    GameStatusPending   GameStatus = "pending"
    GameStatusActive    GameStatus = "active"
    GameStatusCompleted GameStatus = "completed"
)

type GameResult string
const (
    GameResultWhiteWin GameResult = "1-0"
    GameResultBlackWin GameResult = "0-1"
    GameResultDraw     GameResult = "0.5-0.5"
)
```

---

## What NOT to implement in Session 2

These already exist with passing tests — do NOT recreate:

- `service/game_service_test.go` — full coverage for CreateGame, GetGame, GetMoves, ProcessMove (legal+illegal+checkmate), ResignGame, determineResult
- `service/rating_service_test.go` — full ELO coverage (K-32/16, floor, win/loss/draw)
- `service/leaderboard_service_test.go` — full coverage for GetByCity + Refresh
- All `handlers/` tests and all `auth/` tests

Session 2 exclusively targets: **`matchmaking/`** and **`websocket/`**.
