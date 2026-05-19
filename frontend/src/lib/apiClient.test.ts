import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '../test/server'
import { AuthError, getLeaderboard, getMe, login, makeMove, refreshToken } from './apiClient'

const BASE = 'http://localhost:8080'

const validUser = {
  id: 'uuid-1',
  username: 'alice',
  email: 'alice@example.com',
  city: 'almaty',
  rating: 1200,
  games_played: 0,
  created_at: '2026-05-20T00:00:00Z',
  updated_at: '2026-05-20T00:00:00Z',
}

beforeEach(() => {
  localStorage.clear()
})

afterEach(() => {
  localStorage.clear()
})

describe('login()', () => {
  it('returns typed LoginResponse on success', async () => {
    server.use(
      http.post(`${BASE}/auth/login`, () =>
        HttpResponse.json({
          data: {
            user: validUser,
            access_token: 'jwt.access.token',
            refresh_token: 'opaque-refresh',
          },
          error: null,
          message: 'Login successful',
        })
      )
    )

    const result = await login('alice@example.com', 'password123')
    expect(result.data.access_token).toBe('jwt.access.token')
    expect(result.data.user.username).toBe('alice')
    expect(result.data.user.city).toBe('almaty')
  })

  it('throws AuthError on 401 response', async () => {
    server.use(
      http.post(`${BASE}/auth/login`, () =>
        HttpResponse.json(
          { data: null, error: 'invalid_credentials', message: 'Wrong email or password' },
          { status: 401 }
        )
      )
    )

    await expect(login('wrong@example.com', 'wrong')).rejects.toThrow(AuthError)
  })

  it('AuthError is not a plain Error', async () => {
    server.use(
      http.post(`${BASE}/auth/login`, () =>
        HttpResponse.json(
          { data: null, error: 'invalid_credentials', message: 'Unauthorized' },
          { status: 401 }
        )
      )
    )

    let caught: unknown
    try {
      await login('x@x.com', 'bad')
    } catch (err) {
      caught = err
    }
    expect(caught).toBeInstanceOf(AuthError)
    expect((caught as AuthError).status).toBe(401)
    expect((caught as AuthError).code).toBe('invalid_credentials')
  })
})

describe('network failure', () => {
  it('throws when fetch itself fails', async () => {
    server.use(
      http.post(`${BASE}/auth/login`, () => {
        throw new Error('network error')
      })
    )

    await expect(login('alice@example.com', 'password123')).rejects.toThrow()
  })
})

describe('getMe()', () => {
  it('sends Authorization: Bearer header when token is set', async () => {
    localStorage.setItem('access_token', 'test.bearer.token')

    let capturedAuth: string | null = null
    server.use(
      http.get(`${BASE}/users/me`, ({ request }) => {
        capturedAuth = request.headers.get('Authorization')
        return HttpResponse.json({
          data: validUser,
          error: null,
          message: 'User retrieved successfully',
        })
      })
    )

    await getMe()
    expect(capturedAuth).toBe('Bearer test.bearer.token')
  })

  it('throws AuthError on 401 from protected endpoint', async () => {
    server.use(
      http.get(`${BASE}/users/me`, () =>
        HttpResponse.json(
          { data: null, error: 'unauthorized', message: 'Missing token' },
          { status: 401 }
        )
      )
    )

    await expect(getMe()).rejects.toThrow(AuthError)
  })
})

describe('refreshToken()', () => {
  it('returns new tokens on success', async () => {
    server.use(
      http.post(`${BASE}/auth/refresh`, () =>
        HttpResponse.json({
          data: { access_token: 'new.jwt', refresh_token: 'new-refresh' },
          error: null,
          message: 'Token refreshed successfully',
        })
      )
    )

    const result = await refreshToken('old-refresh')
    expect(result.data.access_token).toBe('new.jwt')
    expect(result.data.refresh_token).toBe('new-refresh')
  })
})

describe('makeMove()', () => {
  it('posts move and returns response', async () => {
    server.use(
      http.post(`${BASE}/games/uuid-g1/moves`, () =>
        HttpResponse.json({ data: {}, error: null, message: 'Move recorded' })
      )
    )

    const result = await makeMove('uuid-g1', 'e2e4')
    expect(result.message).toBe('Move recorded')
  })
})

describe('getLeaderboard()', () => {
  it('returns paginated leaderboard', async () => {
    server.use(
      http.get(`${BASE}/leaderboard`, () =>
        HttpResponse.json({
          data: [],
          error: null,
          message: 'ok',
          total: 0,
          page: 1,
          limit: 20,
        })
      )
    )

    const result = await getLeaderboard()
    expect(result.total).toBe(0)
  })

  it('passes city filter in query string', async () => {
    let capturedUrl: string | null = null
    server.use(
      http.get(`${BASE}/leaderboard`, ({ request }) => {
        capturedUrl = request.url
        return HttpResponse.json({ data: [], error: null, message: 'ok', total: 0, page: 1, limit: 20 })
      })
    )

    await getLeaderboard('almaty', 1, 10)
    expect(capturedUrl).toContain('city=almaty')
    expect(capturedUrl).toContain('limit=10')
  })
})
