# Session 4 Context — Frontend: Types, Auth, API Layer

## What Session 4 Must Build

Three test-first modules in `Chess_nfac/frontend/src/`:

| File | Purpose |
|------|---------|
| `src/types/api.ts` | Zod schemas + inferred TS types for every backend shape |
| `src/lib/apiClient.ts` | Typed fetch wrapper (replaces existing `src/lib/api.ts`) |
| `src/hooks/useAuth.ts` | Auth hook: login, logout, token storage, current user |

Test files (write these FIRST, before any implementation):

| Test File | Framework |
|-----------|-----------|
| `src/types/api.test.ts` | Vitest — Zod parse/reject tests |
| `src/lib/apiClient.test.ts` | Vitest + MSW — fetch mock tests |
| `src/hooks/useAuth.test.ts` | Vitest + @testing-library/react — hook tests |

---

## Backend Status (Fully Implemented — Do Not Touch)

The Go backend in `Chess_nfac/backend/` is complete. All routes, handlers, models, and tests are done.

- **REST API contract:** `../backend/docs/API_CONTRACT.md` — open this first, keep it open.
- **Integration tests pass** (requires Docker + PostgreSQL): `go test -tags=integration ./integration/...`
- **Unit test coverage:** ≥80% across all packages

---

## Frontend — Existing Files (Read Before Modifying)

### `src/types/api.ts` — HAS MISMATCHES vs backend (must fix in Session 4)

Current file has these wrong types that must be corrected:

| Field | Current (WRONG) | Correct (matches backend) |
|-------|-----------------|--------------------------|
| `User.city` | `'Almaty' \| 'Astana' \| 'Shymkent'` | `'almaty' \| 'astana' \| 'shymkent'` |
| `Game.status` | `'active' \| 'finished'` | `'waiting' \| 'active' \| 'completed' \| 'abandoned'` |
| `Game.result` | `'white' \| 'black' \| 'draw'` | `'white_wins' \| 'black_wins' \| 'draw' \| null` |
| `Game` | has `time_control`, `ended_at` | these fields DO NOT EXIST in backend |
| `Game` | missing `pgn`, `white_rating_before`, `white_rating_after`, `black_rating_before`, `black_rating_after`, `finished_at` | all required |
| `LeaderboardEntry` | missing `id`, `updated_at` | both required |
| `AuthTokens.expires_in` | present | backend does NOT send this |
| `PaginatedResponse.page_size` | present | backend sends `limit` not `page_size` |

### `src/lib/api.ts` — Existing API client (replace or extend in Session 4)

- Has `apiFetch` with auth header injection
- Error handling is partial — does not use typed error codes from backend
- Session 4 should produce `src/lib/apiClient.ts` as the canonical client

### `src/hooks/useAuth.ts` — Currently just re-exports from AuthContext

Session 4 must implement the full hook logic here.

---

## Frontend — Tooling Already Installed

Check `package.json`. These are confirmed present:

```
zod                       ✓
vitest                    ✓
@testing-library/react    ✓
@testing-library/user-event ✓
jsdom                     ✓
```

**MSW (Mock Service Worker) is NOT installed.** Install before writing apiClient tests:

```bash
npm install --save-dev msw
```

**Vitest config is missing from `vite.config.ts`.** Add before running tests:

```ts
// vite.config.ts — add test block
test: {
  environment: 'jsdom',
  globals: true,
  setupFiles: ['./src/test/setup.ts'],
}
```

Create `src/test/setup.ts` with MSW server setup.

---

## API Response Envelope — CRITICAL

The backend ALWAYS returns this exact shape. Never rename these fields:

```ts
// Standard
{ data: T | null, error: string | null, message: string }

// Paginated
{ data: T[], error: string | null, message: string, total: number, page: number, limit: number }
```

Auth endpoints return `data` with nested objects — see exact shapes in `API_CONTRACT.md`.

---

## TDD Mandate

### Step 1 — Write ALL tests first (RED phase)

**`src/types/api.test.ts`** — Zod schema tests:
- For each schema: test that valid payload parses successfully
- For each schema: test that payload missing a required field is rejected
- Cover: `UserSchema`, `GameSchema`, `MoveSchema`, `LeaderboardEntrySchema`, `ApiResponseSchema<T>`, `PaginatedResponseSchema<T>`, `LoginResponseSchema`, `RefreshResponseSchema`

**`src/lib/apiClient.test.ts`** — MSW fetch mock tests:
- Successful login returns typed `LoginResponse`
- 401 response throws `AuthError` (or similar typed error)
- Network error throws
- Token is attached to authenticated requests

**`src/hooks/useAuth.test.ts`** — React hook tests:
- `login()` sets access token in localStorage and returns user
- `logout()` clears localStorage
- `useAuth()` returns correct `user` after login
- `useAuth()` returns `null` user before login

### Step 2 — Run tests, verify ALL FAIL

```bash
npx vitest run
```

If any test passes before implementation, the test is wrong — fix it.

### Step 3 — Implement (GREEN phase)

Only after all tests are RED:

1. `src/types/api.ts` — Zod schemas + `z.infer<>` types
2. `src/lib/apiClient.ts` — typed fetch wrapper using schemas
3. `src/hooks/useAuth.ts` — hook using apiClient

### Step 4 — Verify coverage

```bash
npx vitest run --coverage
```

Coverage must be ≥80% on `src/types/`, `src/lib/`, `src/hooks/`.

---

## Patterns Already Established in the Codebase

- **Immutable updates** — never mutate state in place
- **Zod for validation** — infer TypeScript types from Zod schemas, never write them separately
- **No `any`** — use `unknown` + narrowing for external data
- **Error types** — define typed error classes, not generic `Error`
- **No `console.log`** in production code

---

## What Session 4 Must NOT Re-implement

- The backend — it is fully done
- Docker or database setup
- The chess game logic or WebSocket hub
- React components (Board, etc.) — only types, API client, and auth hook

---

## Coverage Baseline Going In

From Session 3 (backend):

| Package | Coverage |
|---------|----------|
| `service/` | ≥87% |
| `handlers/` | ≥80% |
| `repository/` | ≥80% |
| `matchmaking/` | ≥80% |
| `websocket/` | ≥80% |

Frontend coverage: 0% (no tests yet). Session 4 target: ≥80% on new modules.
