import { createContext, useContext, useState, useCallback, type ReactNode } from 'react'
import { post } from '../lib/api'
import { setAccessToken, clearAccessToken } from '../lib/api'
import type { User, LoginResponse} from '../types/api'

interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
}

interface AuthContextValue extends AuthState {
  login: (username: string, password: string) => Promise<void>
  register: (username: string, email: string, password: string, city: string) => Promise<void>
  logout: () => Promise<void>
  refresh: () => Promise<void>
  isAuthenticated: boolean
}

const AuthContext = createContext<AuthContextValue | null>(null)

function loadStoredAuth(): AuthState {
  try {
    const token = localStorage.getItem('access_token')
    const refreshToken = localStorage.getItem('refresh_token')
    const userStr = localStorage.getItem('user')
    const user = userStr ? (JSON.parse(userStr) as User) : null
    return { user, accessToken: token, refreshToken }
  } catch {
    return { user: null, accessToken: null, refreshToken: null }
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>(loadStoredAuth)

  const register = useCallback(async (username: string, email: string, password: string, city: string) => {
    const res = await post<LoginResponse>('/auth/register', { username, email, password, city })
    
    // 🖥️ FIX: Destructuring token and user data from inside the .data block
    const { access_token, refresh_token, user } = res.data
    
    setAccessToken(access_token)
    localStorage.setItem('refresh_token', refresh_token)
    localStorage.setItem('user', JSON.stringify(user))
    setState({ user, accessToken: access_token, refreshToken: refresh_token })
  }, [])

  const login = useCallback(async (username: string, password: string) => {
    const res = await post<LoginResponse>('/auth/login', { username, password })
    
    // 🖥️ FIX: Destructuring token and user data from inside the .data block
    const { access_token, refresh_token, user } = res.data
    
    setAccessToken(access_token)
    localStorage.setItem('refresh_token', refresh_token)
    localStorage.setItem('user', JSON.stringify(user))
    setState({ user, accessToken: access_token, refreshToken: refresh_token })
  }, [])

  const logout = useCallback(async () => {
    try {
      await post('/auth/logout')
    } catch {
      // ignore errors on logout
    }
    clearAccessToken()
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user')
    setState({ user: null, accessToken: null, refreshToken: null })
  }, [])

  const refresh = useCallback(async () => {
    const storedRefresh = localStorage.getItem('refresh_token')
    if (!storedRefresh) throw new Error('No refresh token')
    const res = await post<any>('/auth/refresh', { refresh_token: storedRefresh }) // Changed generic to <any>
    
    const { access_token, refresh_token } = res.data
    
    setAccessToken(access_token)
    localStorage.setItem('refresh_token', refresh_token)
    setState(prev => ({ ...prev, accessToken: access_token, refreshToken: refresh_token }))
  }, [])

  return (
    <AuthContext.Provider value={{ ...state, login, register, logout, refresh, isAuthenticated: !!state.user }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuthContext(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuthContext must be used within AuthProvider')
  return ctx
}