# Chess Platform

Real-time multiplayer chess platform with ELO-based matchmaking, Stockfish analysis, and city-based leaderboards for Kazakhstan.

## What Has Been Built

### Backend (Go)

- JWT authentication — 15-minute access tokens, 7-day refresh tokens stored in PostgreSQL (revocable)
- REST API covering registration, login, logout, token refresh, user profiles, game history, moves, and leaderboards
- WebSocket server using the hub pattern — single goroutine owns all connection state, clients communicate via channels
- Matchmaking queue backed by Redis sorted sets, pairing players within ±200 ELO, widening the window after 30 seconds
- Move validation and game state via the `notnil/chess` library
- ELO rating updates after each game (K-factor 32 for players under 30 games, 16 for established players)
- City leaderboard for Almaty, Astana, and Shymkent — refreshed every 5 minutes by a background job
- Game cache in Redis (TTL 2 hours) to avoid hot database reads during active games
- Database migrations with `golang-migrate`
- Structured request/response logging middleware
- CORS middleware with configurable origins

### Frontend (React + TypeScript)

- Authentication flow — register, login, logout, token refresh with silent retry
- Interactive chess board with legal move highlighting and drag-and-drop
- Real-time move sync over WebSocket — both players see moves instantly
- Stockfish WASM running in a Web Worker (off the main thread) for engine analysis
- Evaluation graph showing centipawn score over the course of the game
- Best move suggestions from Stockfish
- Move history list with algebraic notation
- Game history page showing past games per user
- City leaderboard page
- User profile page
- Queue waiting screen with live position updates
- i18n setup (react-i18next)

### Infrastructure

- Docker Compose for local development (PostgreSQL 15, Redis 7)
- Multi-stage Dockerfiles for backend and frontend
- Production docker-compose with environment variable injection
- Deployed: backend on Render, frontend on Vercel

## What Still Needs to Be Implemented

### Correctness bugs

- Game history shows "Loss" for both players — `formatResult` compares `"white_wins"` against `"white"` (type mismatch between backend constants and frontend `GameResult` type)
- Game screen shows translation keys ("game.whitePlayer", "game.blackPlayer") instead of actual usernames — backend does not send usernames in the `game_start` WebSocket message
- Game history shows raw UUID instead of opponent username — the `FindByUserID` query does not JOIN the users table

### Missing features

- Clock / time control — the time control is selected in the queue UI but no countdown timer runs during the game and no timeout loss is enforced on the backend
- Draw offer UI — the backend can receive `offer_draw` and broadcasts `draw_offered`, but there is no visible draw offer notification in the game screen (uses `window.confirm` as a placeholder)
- Resign confirmation — currently uses `window.confirm`, needs a proper modal
- Reconnection handling — if a player disconnects mid-game and reconnects, the game state is not restored to the client
- Spectator mode — no support for watching a game in progress
- Game analysis page — post-game Stockfish analysis with move-by-move evaluation is not wired up as a separate route
- Notifications for opponent going offline or disconnecting
- Admin tooling — no way to inspect or manage active games, users, or the queue outside the database

### Quality

- Test coverage for matchmaking, ELO calculation, and WebSocket message handling is below 80%
- No E2E tests covering the full matchmaking → game → result flow in a deployed environment
- Frontend has no error boundary — a crash in one component takes down the whole page
- No rate limiting on WebSocket message handling (move spam is not throttled)

## Local Development

```bash
# Start PostgreSQL and Redis
docker-compose up -d

# Backend
cd backend
go run main.go

# Frontend
cd frontend
npm install
npm run dev
```

Backend runs on `http://localhost:8080`, frontend on `http://localhost:5173`.

## Environment Variables

| Variable       | Where        | Description                                      |
|----------------|--------------|--------------------------------------------------|
| DATABASE_URL   | backend      | PostgreSQL connection string                     |
| REDIS_URL      | backend      | Redis connection string                          |
| JWT_SECRET     | backend      | Min 32 characters                                |
| CORS_ORIGINS   | backend      | Comma-separated allowed origins                  |
| ENVIRONMENT    | backend      | `dev` or `prod`                                  |
| VITE_API_URL   | frontend     | Backend base URL, e.g. `https://host.com/api/v1` |
| VITE_WS_URL    | frontend     | WebSocket URL, e.g. `wss://host.com/ws`          |

## Tech Stack

| Layer     | Technology                                              |
|-----------|---------------------------------------------------------|
| Backend   | Go 1.23, Chi, gorilla/websocket, pgx/v5, go-redis       |
| Database  | PostgreSQL 15                                           |
| Cache     | Redis 7                                                 |
| Frontend  | React 18, TypeScript, Vite, chess.js, Stockfish WASM    |
| Auth      | JWT (golang-jwt/jwt v5), bcrypt cost 12                 |
| Deploy    | Render (backend + DB + Redis), Vercel (frontend)        |
