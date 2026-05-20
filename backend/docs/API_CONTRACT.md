# Chess Platform — REST API Contract

> Source of truth for all frontend API integration. Field names are **exact** — match them verbatim in TypeScript types.

## Base URL

```
http://localhost:8080
```

## Response Envelope

All REST endpoints return the same JSON envelope.

### Standard Response

```json
{
  "data":    <T | null>,
  "error":   <string | null>,
  "message": <string>
}
```

### Paginated Response

Used by `GET /leaderboard/:city`.

```json
{
  "data":    <T[]>,
  "error":   <string | null>,
  "message": <string>,
  "total":   <number>,
  "page":    <number>,
  "limit":   <number>
}
```

**CRITICAL for frontend:** Never rename these fields. `data`, `error`, `message`, `total`, `page`, `limit` — exact keys, always.

---

## Authentication

All protected routes require:

```
Authorization: Bearer <access_token>
```

### Token Lifetimes

| Token         | Lifetime |
|---------------|----------|
| `access_token`  | 15 minutes |
| `refresh_token` | 7 days     |

---

## Models

### User

```json
{
  "id":           "uuid",
  "username":     "string",
  "email":        "string",
  "city":         "almaty | astana | shymkent",
  "rating":       1200,
  "games_played": 0,
  "created_at":   "2026-05-20T00:00:00Z",
  "updated_at":   "2026-05-20T00:00:00Z"
}
```

> `city` is always **lowercase**: `"almaty"`, `"astana"`, `"shymkent"`.

### Game

```json
{
  "id":                   "uuid",
  "white_id":             "uuid",
  "black_id":             "uuid",
  "status":               "waiting | active | completed | abandoned",
  "result":               "white_wins | black_wins | draw | null",
  "pgn":                  "string | null",
  "fen":                  "string",
  "white_rating_before":  1200,
  "white_rating_after":   null,
  "black_rating_before":  1200,
  "black_rating_after":   null,
  "created_at":           "2026-05-20T00:00:00Z",
  "updated_at":           "2026-05-20T00:00:00Z",
  "finished_at":          "2026-05-20T00:00:00Z | null"
}
```

> `result` is `null` when game is not yet finished.  
> `white_rating_after` / `black_rating_after` are `null` until game completes.  
> `finished_at` is `null` for active/waiting games.

### Move

```json
{
  "id":          "uuid",
  "game_id":     "uuid",
  "player_id":   "uuid",
  "move_number": 1,
  "notation":    "e2e4",
  "fen_after":   "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
  "created_at":  "2026-05-20T00:00:00Z"
}
```

> `notation` is UCI format: `"e2e4"`, `"g1f3"`, `"e1g1"` (castling), `"e7e8q"` (promotion).

### LeaderboardEntry

```json
{
  "id":           "uuid",
  "user_id":      "uuid",
  "username":     "string",
  "city":         "almaty | astana | shymkent",
  "rating":       1200,
  "rank":         1,
  "games_played": 42,
  "updated_at":   "2026-05-20T00:00:00Z"
}
```

---

## Endpoints

### POST /auth/register

Register a new user.

**Request body:**

```json
{
  "username": "string (3–32 chars, alphanumeric + underscore)",
  "email":    "string (valid email)",
  "password": "string (8–128 chars)",
  "city":     "almaty | astana | shymkent"
}
```

**Response `201 Created`:**

```json
{
  "data": {
    "user": {
      "id":           "uuid",
      "username":     "alice",
      "email":        "alice@example.com",
      "city":         "almaty",
      "rating":       1200,
      "games_played": 0,
      "created_at":   "2026-05-20T00:00:00Z",
      "updated_at":   "2026-05-20T00:00:00Z"
    },
    "access_token":  "jwt.string.here",
    "refresh_token": "opaque-token-string"
  },
  "error":   null,
  "message": "User registered successfully"
}
```

**Errors:**

| Status | `error` code      | Condition                     |
|--------|-------------------|-------------------------------|
| 400    | `invalid_input`   | Missing / malformed fields    |
| 400    | `invalid_city`    | City not in allowed values    |
| 409    | `user_exists`     | Username already taken        |
| 409    | `email_exists`    | Email already registered      |

---

### POST /auth/login

Authenticate an existing user.

**Request body:**

```json
{
  "email":    "alice@example.com",
  "password": "secret123"
}
```

**Response `200 OK`:**

```json
{
  "data": {
    "user": {
      "id":           "uuid",
      "username":     "alice",
      "email":        "alice@example.com",
      "city":         "almaty",
      "rating":       1200,
      "games_played": 0,
      "created_at":   "2026-05-20T00:00:00Z",
      "updated_at":   "2026-05-20T00:00:00Z"
    },
    "access_token":  "jwt.string.here",
    "refresh_token": "opaque-token-string"
  },
  "error":   null,
  "message": "Login successful"
}
```

**Errors:**

| Status | `error` code          | Condition                     |
|--------|-----------------------|-------------------------------|
| 400    | `invalid_input`       | Missing fields                |
| 401    | `invalid_credentials` | Wrong email or password       |

---

### POST /auth/refresh

Exchange a refresh token for new token pair.

**Request body:**

```json
{
  "refresh_token": "opaque-token-string"
}
```

**Response `200 OK`:**

```json
{
  "data": {
    "access_token":  "new.jwt.here",
    "refresh_token": "new-opaque-token"
  },
  "error":   null,
  "message": "Token refreshed successfully"
}
```

**Errors:**

| Status | `error` code    | Condition                              |
|--------|-----------------|----------------------------------------|
| 400    | `invalid_input` | Missing refresh_token field            |
| 401    | `invalid_token` | Token expired, revoked, or not found   |

---

### POST /auth/logout

Revoke the current refresh token. **Requires Bearer token.**

**Request body:**

```json
{
  "refresh_token": "opaque-token-string"
}
```

**Response `200 OK`:**

```json
{
  "data":    null,
  "error":   null,
  "message": "Logged out successfully"
}
```

**Errors:**

| Status | `error` code    | Condition              |
|--------|-----------------|------------------------|
| 400    | `invalid_input` | Missing refresh_token  |
| 401    | `unauthorized`  | Missing/invalid Bearer |

---

### GET /users/me

Get the currently authenticated user. **Requires Bearer token.**

**Response `200 OK`:**

```json
{
  "data": {
    "id":           "uuid",
    "username":     "alice",
    "email":        "alice@example.com",
    "city":         "almaty",
    "rating":       1200,
    "games_played": 0,
    "created_at":   "2026-05-20T00:00:00Z",
    "updated_at":   "2026-05-20T00:00:00Z"
  },
  "error":   null,
  "message": "User retrieved successfully"
}
```

**Errors:**

| Status | `error` code   | Condition              |
|--------|----------------|------------------------|
| 401    | `unauthorized` | Missing/invalid Bearer |

---

### GET /users/:id

Get a user by ID. **Requires Bearer token.**

**Response `200 OK`:** Same shape as `GET /users/me`.

**Errors:**

| Status | `error` code   | Condition              |
|--------|----------------|------------------------|
| 401    | `unauthorized` | Missing/invalid Bearer |
| 404    | `not_found`    | No user with that ID   |

---

### GET /games/:id

Get a game by ID. **Requires Bearer token.**

**Response `200 OK`:**

```json
{
  "data":    { /* Game object */ },
  "error":   null,
  "message": "Game retrieved successfully"
}
```

**Errors:**

| Status | `error` code     | Condition              |
|--------|------------------|------------------------|
| 401    | `unauthorized`   | Missing/invalid Bearer |
| 404    | `game_not_found` | No game with that ID   |

---

### POST /games/:id/move

Make a move in an active game. **Requires Bearer token.**

**Request body:**

```json
{
  "notation": "e2e4"
}
```

> `notation` must be UCI format: `"e2e4"`, `"g1f3"`, `"e7e8q"` (promotion), `"e1g1"` (castling as king move).

**Response `200 OK`:**

```json
{
  "data": {
    "game": { /* Game object with updated fen, status, result */ },
    "move": { /* Move object */ },
    "game_over": false
  },
  "error":   null,
  "message": "Move made successfully"
}
```

**Errors:**

| Status | `error` code     | Condition                          |
|--------|------------------|------------------------------------|
| 400    | `invalid_input`  | Missing notation                   |
| 400    | `invalid_move`   | Illegal move for current position  |
| 401    | `unauthorized`   | Missing/invalid Bearer             |
| 403    | `not_your_turn`  | Not this player's turn             |
| 404    | `game_not_found` | No game with that ID               |
| 409    | `game_not_active`| Game is not in `"active"` status   |

---

### GET /games/:id/moves

Get all moves for a game. **Requires Bearer token.**

**Response `200 OK`:**

```json
{
  "data":    [ /* Move[] */ ],
  "error":   null,
  "message": "Moves retrieved successfully"
}
```

**Errors:**

| Status | `error` code     | Condition              |
|--------|------------------|------------------------|
| 401    | `unauthorized`   | Missing/invalid Bearer |
| 404    | `game_not_found` | No game with that ID   |

---

### GET /leaderboard/:city

Get ranked leaderboard for a city. **Requires Bearer token.**

**Path parameter:** `city` — one of `almaty`, `astana`, `shymkent` (lowercase).

**Query parameters:**

| Param     | Type | Default | Max |
|-----------|------|---------|-----|
| `page`    | int  | 1       | —   |
| `page_size` | int | 20     | 100 |

**Response `200 OK`:**

```json
{
  "data": [ /* LeaderboardEntry[] */ ],
  "error":   null,
  "message": "Leaderboard retrieved successfully",
  "total":   150,
  "page":    1,
  "limit":   20
}
```

> Note: The pagination key is `"limit"` (not `"page_size"`). The query param is `page_size` but the response field is `limit`.

**Errors:**

| Status | `error` code   | Condition                         |
|--------|----------------|-----------------------------------|
| 400    | `invalid_city` | City not in allowed values        |
| 401    | `unauthorized` | Missing/invalid Bearer            |

---

### GET /ws

WebSocket upgrade endpoint. **Requires Bearer token** (passed as query param or header).

```
ws://localhost:8080/ws?token=<access_token>
```

See the WebSocket Protocol section below.

---

## WebSocket Protocol

All WebSocket messages are JSON objects with a `"type"` discriminator.

### Client → Server Messages

#### `join` — Announce presence for an active game

```json
{
  "type":    "join",
  "payload": { "game_id": "uuid" }
}
```

#### `queue_join` — Enter matchmaking queue

```json
{
  "type":    "queue_join",
  "payload": {}
}
```

#### `queue_leave` — Leave matchmaking queue

```json
{
  "type":    "queue_leave",
  "payload": {}
}
```

#### `move_made` — Send a chess move (alternative to REST)

```json
{
  "type":    "move_made",
  "payload": {
    "game_id":  "uuid",
    "notation": "e2e4"
  }
}
```

#### `resign` — Resign from a game

```json
{
  "type":    "resign",
  "payload": { "game_id": "uuid" }
}
```

#### `ping` — Keepalive

```json
{
  "type":    "ping",
  "payload": {}
}
```

---

### Server → Client Messages

#### `game_start` — Match found, game beginning

```json
{
  "type": "game_start",
  "payload": {
    "game_id":    "uuid",
    "white_id":   "uuid",
    "black_id":   "uuid",
    "fen":        "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
  }
}
```

#### `move_made` — Move was applied to the game

```json
{
  "type": "move_made",
  "payload": {
    "game_id":     "uuid",
    "player_id":   "uuid",
    "notation":    "e2e4",
    "fen":         "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
    "move_number": 1
  }
}
```

#### `move_error` — Attempted move was rejected

```json
{
  "type": "move_error",
  "payload": {
    "game_id": "uuid",
    "error":   "invalid move"
  }
}
```

#### `game_end` — Game finished

```json
{
  "type": "game_end",
  "payload": {
    "game_id": "uuid",
    "result":  "white_wins | black_wins | draw",
    "reason":  "checkmate | stalemate | resignation | timeout"
  }
}
```

#### `queue_status` — Matchmaking queue update

```json
{
  "type": "queue_status",
  "payload": {
    "in_queue":     true,
    "queue_size":   3,
    "waiting_since": "2026-05-20T00:00:00Z"
  }
}
```

#### `game_state` — Full game state (on reconnect / join)

```json
{
  "type": "game_state",
  "payload": {
    "game_id":     "uuid",
    "fen":         "string",
    "moves":       [ /* Move[] */ ],
    "white_id":    "uuid",
    "black_id":    "uuid",
    "status":      "active",
    "your_color":  "white | black"
  }
}
```

#### `pong` — Response to client `ping`

```json
{
  "type":    "pong",
  "payload": {}
}
```

#### `error` — Generic server error

```json
{
  "type": "error",
  "payload": {
    "code":    "error_code_string",
    "message": "human readable message"
  }
}
```

---

## Error Codes Reference

| Code                 | HTTP Status | Meaning                                  |
|----------------------|-------------|------------------------------------------|
| `unauthorized`       | 401         | Missing or invalid Bearer token          |
| `forbidden`          | 403         | Authenticated but not permitted          |
| `not_found`          | 404         | Resource does not exist                  |
| `conflict`           | 409         | Resource already exists                  |
| `invalid_input`      | 400         | Request body missing/malformed           |
| `user_exists`        | 409         | Username already taken                   |
| `email_exists`       | 409         | Email already registered                 |
| `invalid_credentials`| 401         | Wrong email or password                  |
| `invalid_token`      | 401         | Refresh token expired, revoked, not found|
| `validation_error`   | 400         | Field-level validation failure           |
| `invalid_city`       | 400         | City not in `almaty|astana|shymkent`     |
| `game_not_found`     | 404         | No game with that ID                     |
| `game_not_active`    | 409         | Game status is not `"active"`            |
| `not_your_turn`      | 403         | Move submitted by wrong player           |
| `invalid_move`       | 400         | Move is illegal in current position      |

---

## Validation Rules

### Username
- 3–32 characters
- Alphanumeric and underscore only: `^[a-zA-Z0-9_]+$`

### Password
- 8–128 characters

### Email
- Standard email format

### City
- Exactly one of: `"almaty"`, `"astana"`, `"shymkent"` (lowercase)

### Move notation
- UCI format: source square + destination square + optional promotion piece
- Examples: `"e2e4"`, `"g1f3"`, `"e1g1"` (castling), `"e7e8q"` (queen promotion)
