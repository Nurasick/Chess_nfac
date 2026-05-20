# Backend — Go Service

## Tech Stack

| Concern        | Library / Tool                        |
|----------------|---------------------------------------|
| HTTP router    | `github.com/go-chi/chi/v5`            |
| WebSocket      | `github.com/gorilla/websocket`        |
| PostgreSQL     | `github.com/jackc/pgx/v5`            |
| Redis          | `github.com/redis/go-redis/v9`        |
| JWT            | `github.com/golang-jwt/jwt/v5`        |
| Password hash  | `golang.org/x/crypto/bcrypt`          |
| Chess rules    | `github.com/notnil/chess`             |
| Migrations     | `github.com/golang-migrate/migrate/v4` |
| Testing        | `github.com/stretchr/testify`         |

## Folder Structure

```
backend/
├── main.go                  Entry point — wires everything, starts HTTP server
├── config/
│   └── config.go            Load env vars into Config struct
├── db/
│   ├── postgres.go          Connection pool setup (pgxpool)
│   └── migrations/
│       └── 001_initial_schema.sql
├── models/                  Plain Go structs (no DB logic)
│   ├── user.go
│   ├── game.go
│   ├── move.go
│   ├── rating.go
│   └── leaderboard.go
├── repository/              DB access only — no business logic
│   ├── user_repo.go
│   ├── game_repo.go
│   ├── move_repo.go
│   ├── rating_repo.go
│   └── leaderboard_repo.go
├── service/                 Business logic — calls repos, never touches DB directly
│   ├── user_service.go
│   ├── game_service.go
│   ├── rating_service.go
│   └── leaderboard_service.go
├── handlers/                HTTP handlers — parse request, call service, return response
│   ├── auth.go
│   ├── users.go
│   ├── games.go
│   └── leaderboard.go
├── websocket/
│   ├── hub.go               Owns all active connections; single goroutine with select loop
│   ├── client.go            Per-connection read/write pumps
│   ├── messages.go          Message type definitions + JSON marshal/unmarshal
│   └── handler.go           HTTP → WebSocket upgrade, auth check, register with hub
├── matchmaking/
│   ├── queue.go             Redis sorted-set queue (score = ELO rating)
│   └── matcher.go           Pairing algorithm — match within ±200 ELO, widen after 30s
├── chess/
│   └── engine.go            Thin wrapper over notnil/chess for move validation
├── auth/
│   ├── jwt.go               GenerateAccessToken (15min), GenerateRefreshToken (7d), ParseToken
│   └── password.go          HashPassword (bcrypt cost 12), VerifyPassword
├── cache/
│   ├── redis.go             Redis client setup
│   ├── game_cache.go        Cache active game state by game_id (TTL 2h)
│   └── session_cache.go     Cache user online status (TTL 5min)
├── middleware/
│   ├── auth.go              Validate JWT, inject userID into context
│   ├── cors.go              CORS headers
│   └── logging.go           Request/response logging
├── router/
│   └── router.go            Register all routes and middleware
├── utils/
│   ├── errors.go            AppError type with code + message
│   ├── response.go          JSON response helpers (Success, Error envelopes)
│   └── validators.go        Input validation (city enum, username length, etc.)
└── docs/
    ├── API_CONTRACT.md
    └── WEBSOCKET_PROTOCOL.md
```

## Environment Variables

| Variable       | Required | Description                              |
|----------------|----------|------------------------------------------|
| `DATABASE_URL` | yes      | `postgres://user:pass@localhost:5432/chess` |
| `REDIS_URL`    | yes      | `redis://localhost:6379`                 |
| `JWT_SECRET`   | yes      | min 32 chars, random string              |
| `PORT`         | no       | default `8080`                           |
| `ENVIRONMENT`  | no       | `dev` or `prod`                          |
| `LOG_LEVEL`    | no       | `debug`, `info`, `warn`, `error`         |
| `CORS_ORIGINS` | no       | comma-separated allowed origins          |

## Running

```bash
# Start dependencies
docker-compose up -d

# Run migrations
go run main.go migrate

# Start server
go run main.go

# All tests with coverage
go test -cover ./...

# Coverage report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Lint
golangci-lint run ./...
```

## Key Conventions

### Error handling
Always wrap with context:
```go
if err != nil {
    return fmt.Errorf("game_service.ProcessMove: %w", err)
}
```

Use `utils.AppError` for user-facing errors:
```go
return utils.AppError{Code: "move_invalid", Message: "That move is not legal"}
```

### Response envelope
```go
// Success
utils.RespondJSON(w, http.StatusOK, data)

// Error
utils.RespondError(w, http.StatusBadRequest, "error_code", "Human message")
```

### Repository pattern
- Repositories take `context.Context` as first arg
- Return domain models, not raw DB rows
- Never log inside repositories — let the caller log

### WebSocket hub
- Hub owns a `map[userID]*Client` — never access from outside the hub goroutine
- Clients send to Hub via `hub.Register`, `hub.Unregister`, `hub.Broadcast` channels
- Each client has `send chan []byte` — write pump drains it

### Concurrency rules
- Use `context.Context` for cancellation everywhere
- No raw `go func()` without a WaitGroup or done channel
- Redis operations: use atomic Lua scripts for queue pop to avoid race conditions

## Database Access Patterns

- All queries use parameterized statements — never string concatenation
- Use `pgx` named args for readability on INSERT/UPDATE
- Wrap multi-step operations (e.g., save move + update rating) in a transaction
- `city_leaderboard` table is refreshed by a background goroutine every 5 minutes — never recompute on request

## Testing Strategy

- Unit tests: mock repository interfaces with testify/mock
- Integration tests: use a real PostgreSQL instance (test container or local Docker)
- WebSocket tests: use `gorilla/websocket` test client
- Table-driven tests for ELO calculator and move validator

## Code Review Checklist

- [ ] `go test ./...` passes
- [ ] Coverage >= 80% on changed packages
- [ ] No `fmt.Print*` — use structured logging
- [ ] No hardcoded secrets or IPs
- [ ] All SQL parameterized
- [ ] WebSocket messages validated before processing
- [ ] Goroutines have cleanup (context cancel or WaitGroup)
- [ ] Errors wrapped with context
