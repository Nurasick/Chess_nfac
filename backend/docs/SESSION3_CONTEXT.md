# Session 3 Context — Backend: Leaderboard StartRefreshJob + Integration Tests

> Read this file at the start of Session 3. It captures exactly what exists, what is already
> tested, and what Session 3 must build. Do NOT scan the codebase from scratch.

## Coverage Status (end of Session 2)

| Package        | Coverage | Status       |
|----------------|----------|--------------|
| `auth/`        | 82.6%    | ✅ done      |
| `service/`     | ~84%     | ⚠️ `StartRefreshJob` at 0% |
| `handlers/`    | 100%     | ✅ done      |
| `matchmaking/` | 82.4%    | ✅ done      |
| `websocket/`   | 87.3%    | ✅ done      |
| `integration/` | —        | ❌ Session 3 |

---

## What Already Exists and Passes (DO NOT RECREATE)

### Passing test files

| File | What it covers |
|------|----------------|
| `auth/password_test.go` | bcrypt hash/verify |
| `auth/jwt_test.go` | GenerateAccessToken, GenerateRefreshToken, ParseToken, ParseRefreshToken |
| `service/user_service_test.go` | Register, Login, GetProfile, UpdateProfile |
| `service/rating_service_test.go` | CalculateNewRatings (K-32/16, floor 100), ApplyRatingChange |
| `service/game_service_test.go` | CreateGame, GetGame, GetMoves, ProcessMove, ResignGame |
| `service/leaderboard_service_test.go` | GetByCity (pagination, cities, clamping, invalid city, repo error), Refresh (success, error) |
| `handlers/auth_test.go` | POST /register, POST /login, POST /refresh, POST /logout |
| `handlers/users_test.go` | GET /users/me, PUT /users/me |
| `handlers/games_test.go` | POST /games, GET /games/{id}, POST /games/{id}/move, POST /games/{id}/resign |
| `handlers/leaderboard_test.go` | GET /leaderboard/{city}?page=&page_size= (success, pagination, invalid city → 400, service error → 500) |
| `matchmaking/queue_test.go` | Enqueue, Dequeue, GetAll, GetByRatingRange via miniredis |
| `matchmaking/matcher_test.go` | canMatch ±200 ELO, tryMatch, performMatch success+rollback, GetQueuePosition, IsQueuedUser |
| `websocket/messages_test.go` | JSON round-trip for all message and payload types |
| `websocket/hub_test.go` | register, unregister, joinGame, BroadcastToGame, Broadcast, RemoveGameClients |
| `websocket/handler_test.go` | missing auth → 401, invalid token → 401, valid Bearer → WS upgrade + online, no-Bearer path |
| `websocket/client_test.go` | NewClient field defaults |
| `websocket/pump_test.go` | ReadPump close-on-bad-message, WritePump buffered-send |

---

## Session 3 Targets

### 1. `service/leaderboard_service_test.go` — add `StartRefreshJob` test

**The only gap in `service/`** is `StartRefreshJob` at 0% coverage.

The function:
```go
func (s *LeaderboardService) StartRefreshJob(ctx context.Context, interval time.Duration) {
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                if err := s.Refresh(ctx); err != nil {
                    fmt.Printf("leaderboard refresh error: %v\n", err)
                }
            }
        }
    }()
}
```

Required cases:
- `ctx.Done()` path: cancel context immediately → goroutine exits (verify via WaitGroup or channel trick)
- `ticker.C` fires → `Refresh` is called (use 1ms interval, wait a few ms)
- `Refresh` returns error → logs it (function still continues; don't assert on stdout, just assert no panic)

**Tip:** Add `StartRefreshJob` cases to the existing `TestLeaderboardService_Refresh` table
or add a new `TestLeaderboardService_StartRefreshJob` function — both approaches are fine.
Use `mockLeaderboardRepo` already defined in that file.

### 2. `integration/` package — two files under `//go:build integration`

Create the `integration/` directory at `backend/integration/`.

Both files must start with:
```go
//go:build integration

package integration
```

Run with: `go test -tags=integration ./integration/...`
Requires: real PostgreSQL accessible at `DATABASE_URL` env var (Docker).

**`integration/auth_flow_test.go`** — full auth lifecycle:
1. Register a new user (`POST /auth/register`)
2. Login (`POST /auth/login`) → get access + refresh tokens
3. Refresh (`POST /auth/refresh`) → get new access token
4. Logout (`POST /auth/logout`) → token revoked
5. Refresh again with revoked token → 401

**`integration/game_flow_test.go`** — full game lifecycle:
1. Register 2 users
2. Login both → get tokens
3. Enqueue both users for matchmaking
4. Poll until match is made (game record appears in DB)
5. Connect both via WebSocket
6. Play 5 legal moves via WS (`make_move` messages)
7. Assert game record in DB has correct FEN and move count

---

## Integration Test Setup Pattern

```go
//go:build integration

package integration

import (
    "context"
    "os"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/stretchr/testify/require"
)

func setupDB(t *testing.T) *pgxpool.Pool {
    t.Helper()
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        t.Skip("DATABASE_URL not set — skipping integration test")
    }
    pool, err := pgxpool.New(context.Background(), dsn)
    require.NoError(t, err)
    t.Cleanup(func() { pool.Close() })
    return pool
}
```

Seed users directly via SQL for speed (avoids HTTP round-trip for setup):
```go
_, err := pool.Exec(ctx, `
    INSERT INTO users (id, username, email, password_hash, city, rating)
    VALUES ($1, $2, $3, $4, $5, $6)`,
    userID, username, email, hashedPw, "almaty", 1200)
```

Run migrations before tests using `golang-migrate`:
```go
import "github.com/golang-migrate/migrate/v4"
// point at db/migrations/, run m.Up()
```

Or rely on the DB already being migrated (Docker Compose runs migrations on start).

---

## Key Interfaces (repository layer)

```go
// repository/leaderboard_repo.go
type LeaderboardRepository interface {
    GetByCity(ctx context.Context, city string, limit, offset int) ([]models.LeaderboardEntry, int, error)
    Refresh(ctx context.Context) error
}

// repository/user_repo.go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    FindByUsername(ctx context.Context, username string) (*models.User, error)
    FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
}

// repository/game_repo.go
type GameRepository interface {
    Create(ctx context.Context, game *models.Game) error
    FindByID(ctx context.Context, id uuid.UUID) (*models.Game, error)
    UpdateStatus(ctx context.Context, gameID uuid.UUID, status models.GameStatus) error
    Finish(ctx context.Context, gameID uuid.UUID, result models.GameResult, whiteRatingAfter, blackRatingAfter int) error
    FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*models.Game, error)
    UpdateFEN(ctx context.Context, gameID uuid.UUID, fen string) error
    GetUserGameCount(ctx context.Context, userID uuid.UUID) (int, error)
}
```

---

## DB Schema (abridged — `001_initial_schema.sql`)

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(32) NOT NULL UNIQUE,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  city VARCHAR(64) NOT NULL,
  rating INTEGER NOT NULL DEFAULT 1200,
  games_played INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE refresh_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE games (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  white_id UUID NOT NULL REFERENCES users(id),
  black_id UUID NOT NULL REFERENCES users(id),
  status VARCHAR(16) NOT NULL DEFAULT 'waiting',
  result VARCHAR(16),
  fen TEXT NOT NULL DEFAULT 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1',
  white_rating_before INTEGER,
  white_rating_after INTEGER,
  black_rating_before INTEGER,
  black_rating_after INTEGER,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE moves (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  game_id UUID NOT NULL REFERENCES games(id),
  player_id UUID NOT NULL REFERENCES users(id),
  move_number INTEGER NOT NULL,
  notation VARCHAR(10) NOT NULL,
  fen_after TEXT NOT NULL
);

CREATE TABLE city_leaderboard (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  city VARCHAR(64) NOT NULL,
  rating INTEGER NOT NULL,
  rank INTEGER NOT NULL,
  games_played INTEGER NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(user_id, city)
);
```

---

## Established Test Patterns (reuse these)

**AppError (value type — `errors.As` with non-pointer target):**
```go
var appErr utils.AppError
require.True(t, errors.As(err, &appErr))
assert.Equal(t, "invalid_city", appErr.Code)
```

**Chi URL params in handler tests:**
```go
rctx := chi.NewRouteContext()
rctx.URLParams.Add("city", "almaty")
req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
```

**Authenticated user in handler tests:**
```go
req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
```

**Response envelope helpers:**
```go
utils.RespondJSON(w, http.StatusOK, data)
utils.RespondError(w, http.StatusBadRequest, "error_code", "Human message")
utils.RespondPaginated(w, http.StatusOK, entries, total, page, pageSize)
```

**testify mock pattern:**
```go
type mockLeaderboardRepo struct{ mock.Mock }
func (m *mockLeaderboardRepo) Refresh(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}
var _ repository.LeaderboardRepository = (*mockLeaderboardRepo)(nil) // compile-time check
```

**miniredis (already in go.mod — used by matchmaking tests):**
```go
import "github.com/alicebob/miniredis/v2"
mr := miniredis.RunT(t)
client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
```

**JWT generation in tests:**
```go
import "github.com/chess-nfac/backend/auth"
const testSecret = "test-secret-for-integration-tests-32chars"
token, err := auth.GenerateAccessToken(userID, "testuser", testSecret)
```

---

## Module and Package Names

- Module: `github.com/chess-nfac/backend`
- Key imports:
  - `github.com/chess-nfac/backend/auth`
  - `github.com/chess-nfac/backend/config`
  - `github.com/chess-nfac/backend/models`
  - `github.com/chess-nfac/backend/repository`
  - `github.com/chess-nfac/backend/service`
  - `github.com/chess-nfac/backend/handlers`
  - `github.com/chess-nfac/backend/middleware`
  - `github.com/chess-nfac/backend/utils`
  - `github.com/chess-nfac/backend/websocket`
- Valid cities (from `utils.ValidateCity`): `almaty`, `astana`, `shymkent`

---

## What NOT to Implement in Session 3

These are complete and passing — do NOT recreate:

- `service/leaderboard_service_test.go` tests for `GetByCity` and `Refresh` — already comprehensive
- `handlers/leaderboard_test.go` — already at 100% handler coverage
- Any `auth/`, `matchmaking/`, or `websocket/` tests — all done in Sessions 1–2
- The `StartRefreshJob` function itself — it already exists in `service/leaderboard_service.go`
- Any new service or handler production code — production code is complete

Session 3 exclusively adds:
1. `StartRefreshJob` unit test (appended to `service/leaderboard_service_test.go`)
2. `integration/auth_flow_test.go`
3. `integration/game_flow_test.go`
