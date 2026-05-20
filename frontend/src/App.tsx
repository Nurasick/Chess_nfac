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

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AuthProvider>
          <WebSocketProvider url={import.meta.env.VITE_WS_URL ?? 'ws://localhost:8080/ws'}>
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
