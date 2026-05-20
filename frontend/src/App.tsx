import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import { WebSocketProvider } from './context/WebSocketContext'
import { GameProvider } from './context/GameContext'
import { AuthGuard } from './components/auth/AuthGuard'
import { Header } from './components/navigation/Header'
import { LoginPage } from './pages/LoginPage'
import { RegisterPage } from './pages/RegisterPage'
import { HomePage } from './pages/HomePage'
import { GamePage } from './pages/GamePage'
import { GamesHistoryPage } from './pages/GamesHistoryPage'
import { LeaderboardPage } from './pages/LeaderboardPage'
import { ProfilePage } from './pages/ProfilePage'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 30_000,
    },
  },
})

// Dynamic calculation of the production/local WebSocket URL
const calculateWsUrl = (): string => {
  if (import.meta.env.VITE_WS_URL) {
    const raw: string = import.meta.env.VITE_WS_URL;
    return raw.endsWith('/ws') ? raw : raw.replace(/\/?$/, '/ws');
  }

  const apiBaseUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  let cleanHost = apiBaseUrl.replace(/^https?:\/\//, '');
  if (cleanHost.endsWith('/api/v1')) {
    cleanHost = cleanHost.slice(0, -7);
  }
  return `${protocol}//${cleanHost}/ws`;
}

const WS_URL = calculateWsUrl();

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AuthProvider>
          {/* Pass the dynamically generated, clean connection string */}
          <WebSocketProvider url={WS_URL}>
            <GameProvider>
              <Header />
              <Routes>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/register" element={<RegisterPage />} />
                <Route path="/" element={<AuthGuard><HomePage /></AuthGuard>} />
                <Route path="/game/:gameId" element={<AuthGuard><GamePage /></AuthGuard>} />
                <Route path="/history" element={<AuthGuard><GamesHistoryPage /></AuthGuard>} />
                <Route path="/leaderboard" element={<LeaderboardPage />} />
                <Route path="/profile" element={<AuthGuard><ProfilePage /></AuthGuard>} />
                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </GameProvider>
          </WebSocketProvider>
        </AuthProvider>
      </BrowserRouter>
    </QueryClientProvider>
  )
}