# Session 5 Context — Frontend: Chess Board + Game Logic

## What Was Done in Session 4

- `src/types/api.ts` — replaced with Zod schemas + `z.infer<>` types
- `src/lib/apiClient.ts` — typed fetch wrapper with `AuthError`, `login`, `logout`, `refreshToken`, `getMe`, `makeMove`, `getLeaderboard`
- `src/hooks/useAuth.ts` — standalone hook (no context), persists tokens in localStorage
- `vite.config.ts` — added `test.env.VITE_API_URL = 'http://localhost:8080'` override
- All 39 tests pass, coverage ≥88% on `src/types/`, `src/lib/`, `src/hooks/`
- Git checkpoint: `feat: implement Zod schemas, apiClient, and useAuth hook (GREEN)`

---

## WebSocket Contract (authoritative source: `backend/websocket/messages.go`)

### Server → Client message type strings

| Constant              | JSON `"type"` value |
|-----------------------|---------------------|
| `MessageTypeGameStart`   | `"game_start"`   |
| `MessageTypeMoveMade`    | `"move_made"`    |
| `MessageTypeMoveError`   | `"move_error"`   |
| `MessageTypeGameEnd`     | `"game_end"`     |
| `MessageTypeResign`      | `"resign"`       |
| `MessageTypeError`       | `"error"`        |
| `MessageTypeQueueStatus` | `"queue_status"` |
| `MessageTypeGameState`   | `"game_state"`   |
| `MessageTypePong`        | `"pong"`         |

### Client → Server message type strings

| Constant               | JSON `"type"` value |
|------------------------|---------------------|
| `MessageTypeJoin`      | `"join"`            |
| `MessageTypeQueueJoin` | `"queue_join"`      |
| `MessageTypeQueueLeave`| `"queue_leave"`     |
| `MessageTypePing`      | `"ping"`            |
| `MessageTypeResign`    | `"resign"`          |

### Payload field names (exactly as backend sends them)

**`game_start` data:**
```json
{ "game_id": "uuid", "white_id": "uuid", "black_id": "uuid",
  "white_rating": 1500, "black_rating": 1500,
  "fen": "rnbqkbnr/...", "your_color": "white" }
```

**`move_made` data:**
```json
{ "game_id": "uuid", "move": "e2e4", "notation": "e4",
  "fen": "rnbqkbnr/...", "move_count": 1 }
```
NOTE: Backend does NOT send `your_turn` in `move_made`.

**`game_end` data:**
```json
{ "game_id": "uuid", "result": "white_wins", "reason": "checkmate",
  "white_rating_delta": 12, "black_rating_delta": -12 }
```

**`queue_status` data:**
```json
{ "queued": true, "position": 1, "queue_size": 3,
  "estimate_wait_seconds": 15 }
```

**`game_state` data:**
```json
{ "game_id": "uuid", "fen": "rnbqkbnr/...", "last_move": "e2e4",
  "move_count": 1, "your_turn": true,
  "white_rating": 1500, "black_rating": 1500 }
```

### ServerMessage envelope
```json
{ "type": "move_made", "game_id": "uuid", "data": { ... }, "error": "", "message": "" }
```

---

## Mismatches in `frontend/src/types/websocket.ts` to Fix

1. **`MoveMadeMessage.your_turn`** — frontend has it, backend does NOT send it → remove
2. **`DrawOfferedMessage`** (`type: 'draw_offered'`) — not in backend → remove from `WsMessage` union
3. **Client message `join_queue`** — backend expects `"queue_join"` → rename `JoinQueueClientMessage.type` to `'queue_join'`
4. **Client message `leave_queue`** — backend expects `"queue_leave"` → rename to `'leave_queue'` → actually backend uses `"queue_leave"` ✓ but constant is `MessageTypeQueueLeave = "queue_leave"`, so this is fine
5. **Client message `move`** — backend ClientMessage uses `move` field at top level with `type` = unspecified for moves; the frontend should send `{ type: "join", game_id: "...", move: "e2e4" }` — check actual handler in `backend/websocket/handler.go` before changing

Fix 1, 2, and 3 as part of GREEN phase.

---

## Existing Frontend Files to Test Against

### `src/lib/websocket.ts` — `WebSocketClient` class
- `connect()` — creates `new WebSocket(url)`, sets up `onopen/onmessage/onerror/onclose`
- `disconnect()` — sets `isManuallyDisconnected = true`, clears heartbeat, closes socket
- `send(msg)` — JSON.stringify and send if `readyState === OPEN`
- `onMessage(handler)` — registers handler, returns unsubscribe fn
- `onStatusChange(handler)` — registers handler, returns unsubscribe fn
- `isConnected()` — returns `ws.readyState === OPEN`
- Has `console.log`/`console.error` calls — these are code smells but don't block tests

### `src/hooks/useChessBoard.ts` — `useChessBoard(initialFen?)`
Returns: `{ fen, selectedSquare, legalMoves, lastMove, isCheck, isGameOver, selectSquare, makeMove, loadPosition, applyServerMove, getTurn, getHistory, reset }`
- `selectSquare(sq)` — calls `getLegalMovesVerbose(sq)`, sets `selectedSquare` and `legalMoves`
- `makeMove(from, to, promo?)` — calls `validateMove`; returns `false` if illegal
- `loadPosition(fen)` — calls `chess.loadFen(fen)`
- `applyServerMove(move, fen)` — loads FEN, extracts from/to from UCI string

### `src/context/GameContext.tsx` — `GameProvider` / `useGameContext()`
Reducer actions:
- `GAME_START` → sets gameId, fen, myColor, whiteRating, blackRating, isMyTurn=(myColor==='white')
- `MOVE_MADE` → updates fen, appends to moveHistory, sets isMyTurn
- `GAME_END` → sets status='finished', result, resultReason, ratingDelta
- `EVAL_UPDATE` → appends to evalHistory
- `RESET` → returns initialState

### `src/components/game/Board.tsx` — `Board` component
Props: `{ fen, orientation?, selectedSquare, legalMoves, lastMove, onSquareClick, onPieceDrop, isDisabled? }`
Uses `react-chessboard` `<Chessboard>` internally.
- `selectedSquare` → highlights that square
- `legalMoves` → radial-gradient dots on target squares
- `isDisabled` → blocks clicks and drag

### `src/context/WebSocketContext.tsx` — `WebSocketProvider`
- On mount: creates `WebSocketClient(url)`, calls `connect()`
- On unmount: calls `disconnect()`
- Exposes: `{ isConnected, send, onMessage }`

---

## Test Files to Write (RED phase)

### A. `src/hooks/useWebSocket.test.ts`
Mock the global `WebSocket` class. Tests:
1. `WebSocketClient` connects on `connect()` call — `new WebSocket(url)` called
2. `WebSocketClient` calls `disconnect()` — `ws.close()` called, `isConnected()` returns false
3. Message handler receives typed `game_start` message (type string must be exactly `"game_start"`)
4. Message handler receives typed `move_made` message (type string must be exactly `"move_made"`)
5. Message handler receives typed `queue_status` message
6. `send()` calls `ws.send` with JSON-stringified message
7. `onMessage` unsubscribe removes handler

### B. `src/hooks/useChessBoard.test.ts`
No DOM needed (renderHook from @testing-library/react). Tests:
1. Initial FEN is starting position
2. `makeMove('e2', 'e4')` returns true, FEN updates
3. `makeMove('e2', 'e5')` (illegal) returns false, FEN unchanged
4. After white moves, `getTurn()` returns `'black'`
5. `selectSquare('e2')` populates `legalMoves` (white pawn has legal moves)
6. `selectSquare('e5')` (empty square) sets `selectedSquare` to null, `legalMoves` to []
7. `loadPosition(fen)` updates FEN
8. `applyServerMove('e2e4', newFen)` sets `lastMove = { from: 'e2', to: 'e4' }`
9. `reset()` restores starting FEN

### C. `src/components/game/Board.test.tsx`
Use `@testing-library/react`. Tests:
1. Board renders (check the `react-chessboard` renders — can check for role or container div)
2. `onSquareClick` called when square clicked
3. `onPieceDrop` called with correct from/to when piece dropped
4. When `isDisabled=true`, `onSquareClick` is NOT called on click

### D. `src/context/GameContext.test.tsx`
Use `renderHook` with `GameProvider` wrapper. Tests:
1. Initial state: `gameId=null`, `status='waiting'`, `isMyTurn=false`
2. `GAME_START` dispatch → `gameId` set, `myColor` set, `isMyTurn=true` when color=white
3. `MOVE_MADE` dispatch → `fen` updated, `notation` appended to `moveHistory`, `isMyTurn` toggled
4. `GAME_END` dispatch → `status='finished'`, `result` set, `ratingDelta` computed from myColor
5. `RESET` dispatch → state returns to initial
6. `useGameContext` throws if used outside `GameProvider`

---

## Coverage Target

Run: `npx vitest run --coverage`

Minimum 80% on:
- `src/lib/websocket.ts`
- `src/hooks/useChessBoard.ts`
- `src/context/GameContext.tsx`
- `src/components/game/Board.tsx`

---

## Notes for Session 5

- `WEBSOCKET_PROTOCOL.md` does not exist in `backend/docs/` yet — use `backend/websocket/messages.go` as the authoritative source for type strings and payload shapes.
- The heartbeat sends `{ type: 'ping' }` — this matches `MessageTypePing = "ping"` in backend.
- `WebSocketClient` has `console.log`/`console.error` — do NOT remove during session 5 (would be a separate chore commit).
- `react-chessboard` renders an actual board; in jsdom tests it may not render 64 `<div>` squares. Test the component wrapper behavior (props, callbacks) rather than trying to assert on internal chessboard internals.
- `GameContext` uses `status: 'finished'` in `GAME_END` case but `GameStatus` type may define different values — check `src/types/game.ts` before writing tests.
