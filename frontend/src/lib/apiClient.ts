import type { User } from '../types/api'

const BASE_URL: string = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export class AuthError extends Error {
  status: number
  code: string

  constructor(message: string, status: number, code: string) {
    super(message)
    this.name = 'AuthError'
    this.status = status
    this.code = code
    Object.setPrototypeOf(this, AuthError.prototype)
  }
}

function getAuthHeaders(): HeadersInit {
  const token = localStorage.getItem('access_token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

async function handleResponse<T>(res: Response): Promise<T> {
  const body = await res.json()
  if (!res.ok) {
    throw new AuthError(
      body.message ?? 'Request failed',
      res.status,
      body.error ?? 'unknown_error',
    )
  }
  return body as T
}

export async function login(email: string, password: string) {
  const res = await fetch(`${BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  return handleResponse<{
    data: { user: User; access_token: string; refresh_token: string }
    error: string | null
    message: string
  }>(res)
}

export async function logout() {
  const res = await fetch(`${BASE_URL}/auth/logout`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
  })
  return handleResponse<{ data: null; error: string | null; message: string }>(res)
}

export async function refreshToken(token: string) {
  const res = await fetch(`${BASE_URL}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: token }),
  })
  return handleResponse<{
    data: { access_token: string; refresh_token: string }
    error: string | null
    message: string
  }>(res)
}

export async function getMe() {
  const res = await fetch(`${BASE_URL}/users/me`, {
    headers: { ...getAuthHeaders() },
  })
  return handleResponse<{ data: User; error: string | null; message: string }>(res)
}

export async function makeMove(gameId: string, notation: string) {
  const res = await fetch(`${BASE_URL}/games/${gameId}/moves`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
    body: JSON.stringify({ notation }),
  })
  return handleResponse<{ data: unknown; error: string | null; message: string }>(res)
}

export async function getLeaderboard(city?: string, page = 1, limit = 20) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) })
  if (city) params.set('city', city)
  const res = await fetch(`${BASE_URL}/leaderboard?${params}`, {
    headers: { ...getAuthHeaders() },
  })
  return handleResponse<{ data: unknown[]; error: string | null; message: string; total: number; page: number; limit: number }>(res)
}
