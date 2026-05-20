# Chess Platform — Monorepo Root

## What This Is

Full-stack real-time chess platform with:
- Go WebSocket backend (matchmaking, game engine, ELO ratings)
- React/TypeScript frontend (chess board, Stockfish WASM analysis, eval graph)
- City-based leaderboards for Kazakhstan: Almaty, Astana, Shymkent

## Monorepo Layout

```
Chess_nfac/
├── backend/          Go backend (REST + WebSocket)
├── frontend/         React/TypeScript frontend
├── docker-compose.yml
└── CLAUDE.md
```

## Local Development

```bash
# Start PostgreSQL + Redis
docker-compose up -d

# Backend (from backend/)
go run main.go

# Frontend (from frontend/)
npm run dev
```

## Port Map

| Service    | Port  |
|------------|-------|
| Backend    | 8080  |
| Frontend   | 5173  |
| PostgreSQL | 5432  |
| Redis      | 6379  |

## Tech Stack Summary

| Layer     | Tech                                      |
|-----------|-------------------------------------------|
| Backend   | Go 1.22+, Chi, gorilla/websocket, pgx, go-redis |
| Database  | PostgreSQL 15                             |
| Cache     | Redis 7                                   |
| Frontend  | React 18, TypeScript, Vite, chess.js, Stockfish WASM |

## Architecture Decisions

- **WebSocket hub pattern**: single Hub goroutine owns all connection state; clients communicate via channels
- **Repository pattern**: all DB access goes through repository interfaces; services never touch DB directly
- **JWT auth**: 15-min access token + 7-day refresh token stored in `refresh_tokens` table (revocable)
- **ELO**: standard chess formula with K-factor 32 for new players (<30 games), 16 for established
- **Leaderboard**: `city_leaderboard` table refreshed every 5 minutes by background job (avoid hot queries)
- **Stockfish**: runs in Web Worker off the main thread; evaluations cached by FEN string

## Implementation Phases

| Phase | Focus                          | Status  |
|-------|--------------------------------|---------|
| 1     | Scaffold, DB schema, Docker    | pending |
| 2     | JWT auth                       | pending |
| 3     | Matchmaking + WebSocket loop   | pending |
| 4     | ELO rating + leaderboard       | pending |
| 5     | React board + WebSocket client | pending |
| 6     | Stockfish WASM + eval graph    | pending |

## Shared Conventions

- All API responses use `{ data, error, message }` envelope
- All errors are wrapped with context: `fmt.Errorf("operation: %w", err)`
- No hardcoded secrets — all config via environment variables
- Minimum 80% test coverage on critical paths (auth, move validation, ELO)
- Commit format: `feat:`, `fix:`, `refactor:`, `test:`, `chore:`

## Key Docs

- `backend/docs/API_CONTRACT.md` — all REST endpoints
- `backend/docs/WEBSOCKET_PROTOCOL.md` — WebSocket message shapes
- `backend/db/migrations/001_initial_schema.sql` — full DB schema
