# Session 6 Context — Frontend: Stockfish + Eval Graph + Leaderboard

## What was completed in Session 5

All 81 tests pass (GREEN). Coverage on the four target files:

| File | % Stmts | % Branch | % Funcs | % Lines |
|------|---------|----------|---------|---------|
| `src/components/game/Board.tsx` | 94.11 | 90 | 100 | 100 |
| `src/context/GameContext.tsx` | 83.33 | 80 | 75 | 87.5 |
| `src/hooks/useChessBoard.ts` | 94.44 | 83.33 | 92.3 | 96.55 |
| `src/lib/websocket.ts` | 90.66 | 66.66 | 100 | 90.14 |

Git branch: `main`. Last commits:
- `test: add coverage tests for websocket and Board to reach >=80%`
- `feat: chess board and game logic tests passing (GREEN)`
- `test: add RED tests for chess board and game logic`
- `fix: align websocket types with backend contract`

---

## Session 6 Targets

Four new test suites, all targeting files that ALREADY EXIST in the codebase:

| Test file | Implementation file |
|-----------|---------------------|
| `src/hooks/useStockfish.test.ts` | `src/hooks/useStockfish.ts` |
| `src/hooks/useEvaluation.test.ts` | `src/hooks/useEvaluation.ts` |
| `src/components/analysis/EvaluationGraph.test.tsx` | `src/components/analysis/EvaluationGraph.tsx` |
| `src/pages/LeaderboardPage.test.tsx` | `src/pages/LeaderboardPage.tsx` |

---

## Implementation File Summaries

### `src/hooks/useStockfish.ts`

- Creates `new Worker('/stockfish.js')` inside a `useEffect([], ...)` — **not** a constructor arg
- Sends `'uci'` on mount, waits for `'uciok'` → sends `'isready'` → waits for `'readyok'` → sets `isReady: true`
- `analyze(fen, depth=20)`:
  - Checks cache first
  - If not cached, clears debounce timer and sets a new one (`STOCKFISH_DEBOUNCE_MS = 100ms`)
  - Inside the timeout: `worker.postMessage('position fen ' + fen)` then `worker.postMessage('go depth ' + depth)`
  - Guard: only runs if `workerRef.current && state.isReady`
- Parses `'info depth N score cp X'` lines:
  - Regex: `/depth (\d+)/` and `/score cp (-?\d+)/`
  - `score = parseInt(cpMatch[1]) / 100` (centipawns → pawns)
  - Also handles `score mate N` → `mate > 0 ? 99 : -99`
  - Emits `EvalPoint { moveNumber: 0, score, depth, mate }`
- `onEval(handler)` registers a callback, returns unsub function
- `stop()` sends `'stop'` to the worker
- Cleanup: `workerRef.current?.terminate()` on unmount
- State fields returned: `{ isReady, currentEval, bestMove, analyze, onEval, stop }`

### `src/hooks/useEvaluation.ts`

- State: `evalHistory: EvalPoint[]`
- `addEval(point)`: filters out any existing entry with the same `moveNumber`, appends, sorts by moveNumber ascending
- **The 200-entry cap is NOT currently implemented** — the GREEN phase must add it to `addEval`
- `reset()` clears the array

### `src/components/analysis/EvaluationGraph.tsx`

- Path: `src/components/analysis/EvaluationGraph.tsx` — **NOT** `src/components/game/`
- Props: `{ evalHistory: EvalPoint[] }`
- Empty state: renders a `<div>` containing `<span>` with translation key `'analysis.noData'`
- Non-empty state: renders `<ResponsiveContainer>` → `<LineChart>` with data shaped as `{ move, score, label }`
  - `score` is clamped to `[-10, 10]`
  - `label` is `M${mate}` if mate, else `score.toFixed(2)`
- Imports from `recharts`: `LineChart`, `Line`, `XAxis`, `YAxis`, `Tooltip`, `ReferenceLine`, `ResponsiveContainer`
- **Recharts `ResponsiveContainer` fails in jsdom** because it needs real DOM dimensions. Mock it:
  ```tsx
  vi.mock('recharts', async () => {
    const actual = await vi.importActual<typeof import('recharts')>('recharts')
    return {
      ...actual,
      ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
    }
  })
  ```

### `src/pages/LeaderboardPage.tsx`

- Uses `useSearchParams()` from react-router-dom — **must wrap tests with `MemoryRouter`**
- `city` search param selects the active tab (default `'global'`)
- `page` search param controls pagination (default `1`)
- `setTab(tab)` → sets `{ city: tab, page: '1' }`
- `setPage(p)` → sets `{ city: activeTab, page: String(p) }`
- Tabs: `['global', 'Almaty', 'Astana', 'Shymkent']` (from `CITIES` constant)
- Prev button is disabled when `page <= 1`
- Renders `<GlobalLeaderboard page={page} />` when `activeTab === 'global'`
- Renders `<CityLeaderboard city={activeTab} page={page} />` for city tabs
- `GlobalLeaderboard` and `CityLeaderboard` both use `useQuery` from `@tanstack/react-query`

---

## Known Issues to Fix Before Writing Tests

### 1. Missing `PaginatedResponse` type export in `src/types/api.ts`

`GlobalLeaderboard.tsx` and `CityLeaderboard.tsx` import `type { PaginatedResponse, LeaderboardEntry }` from `../../types/api`, but `PaginatedResponse` is NOT exported from `src/types/api.ts`. Add this **before** the RED phase:

```ts
// Add to src/types/api.ts
export type PaginatedResponse<T> = {
  data: T[]
  error: string | null
  message: string
  total: number
  page: number
  limit: number
}
```

Commit as: `fix: export PaginatedResponse type from types/api`

---

## Test Setup (Already in Place from Session 4)

- `src/test/setup.ts` — MSW server + `@testing-library/jest-dom`
- `src/test/server.ts` — `setupServer()` export
- `vite.config.ts` — `test.environment: 'jsdom'`, `test.globals: true`, `test.env.VITE_API_URL`
- Coverage target files listed in `vite.config.ts` → **must be updated** to include the 4 new files

---

## Mocking Patterns for This Session

### Web Worker mock (for useStockfish)

```ts
interface MockWorker {
  postMessage: ReturnType<typeof vi.fn>
  terminate: ReturnType<typeof vi.fn>
  onmessage: ((e: MessageEvent<string>) => void) | null
}

let mockWorker: MockWorker

class MockWorkerClass {
  postMessage = vi.fn()
  terminate = vi.fn()
  onmessage: ((e: MessageEvent<string>) => void) | null = null

  constructor() {
    mockWorker = this as unknown as MockWorker
  }
}

vi.stubGlobal('Worker', MockWorkerClass)
```

**Important**: to simulate the `uci → isready` handshake so `isReady` becomes `true`, fire messages manually:
```ts
act(() => {
  mockWorker.onmessage?.({ data: 'uciok' } as MessageEvent<string>)
})
act(() => {
  mockWorker.onmessage?.({ data: 'readyok' } as MessageEvent<string>)
})
```

### TanStack Query mock (for LeaderboardPage)

Mock the whole module so `useQuery` returns controlled data without a real `QueryClient`:

```tsx
vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn(),
}))

import { useQuery } from '@tanstack/react-query'
const mockUseQuery = vi.mocked(useQuery)

// In each test:
mockUseQuery.mockReturnValue({
  data: { data: [...entries], total: 40, page: 1, limit: 20 },
  isLoading: false,
  isError: false,
} as ReturnType<typeof useQuery>)
```

However since `LeaderboardPage` doesn't call `useQuery` directly (it delegates to `GlobalLeaderboard` / `CityLeaderboard`), it may be easier to mock those child components:

```tsx
vi.mock('../components/leaderboard/GlobalLeaderboard', () => ({
  GlobalLeaderboard: ({ page }: { page: number }) => (
    <div data-testid="global-leaderboard" data-page={page} />
  ),
}))

vi.mock('../components/leaderboard/CityLeaderboard', () => ({
  CityLeaderboard: ({ city, page }: { city: string; page: number }) => (
    <div data-testid="city-leaderboard" data-city={city} data-page={page} />
  ),
}))
```

### MemoryRouter wrapper (for LeaderboardPage)

```tsx
import { MemoryRouter } from 'react-router-dom'

function renderLeaderboard(initialPath = '/leaderboard') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <LeaderboardPage />
    </MemoryRouter>
  )
}
```

### i18n mock (for components using `useTranslation`)

LeaderboardPage uses `t('leaderboard.title')` etc. Add to `src/test/setup.ts` or mock per-file:

```ts
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => {
      if (opts) return `${key}:${JSON.stringify(opts)}`
      return key
    },
    i18n: { language: 'en', changeLanguage: vi.fn() },
  }),
}))
```

---

## Constants Reference

```ts
STOCKFISH_DEBOUNCE_MS = 100   // ms debounce before sending position to worker
STOCKFISH_DEFAULT_DEPTH = 20  // default analysis depth
CITIES = ['Almaty', 'Astana', 'Shymkent']  // City[] — note PascalCase
DEFAULT_PAGE_SIZE = 20
```

---

## `EvalPoint` Type

```ts
export interface EvalPoint {
  moveNumber: number
  score: number      // centipawns / 100 (e.g. 0.35 means +35cp)
  depth: number
  mate?: number | null
}
```

---

## Coverage Config Update Required

In `vite.config.ts`, add the 4 new files to `coverage.include`:

```ts
coverage: {
  include: [
    'src/lib/websocket.ts',
    'src/hooks/useChessBoard.ts',
    'src/context/GameContext.tsx',
    'src/components/game/Board.tsx',
    // NEW for Session 6:
    'src/hooks/useStockfish.ts',
    'src/hooks/useEvaluation.ts',
    'src/components/analysis/EvaluationGraph.tsx',
    'src/pages/LeaderboardPage.tsx',
  ],
},
```
