import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    env: {
      VITE_API_URL: 'http://localhost:8080',
    },
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      include: [
        'src/lib/websocket.ts',
        'src/hooks/useChessBoard.ts',
        'src/context/GameContext.tsx',
        'src/components/game/Board.tsx',
        'src/hooks/useStockfish.ts',
        'src/hooks/useEvaluation.ts',
        'src/components/analysis/EvaluationGraph.tsx',
        'src/pages/LeaderboardPage.tsx',
      ],
    },
    testTimeout: 30000,
    hookTimeout: 30000,
  },
})
