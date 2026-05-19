import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '../test/server'
import { useAuth } from './useAuth'

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

describe('useAuth — initial state', () => {
  it('returns null user before login', () => {
    const { result } = renderHook(() => useAuth())
    expect(result.current.user).toBeNull()
  })
})

describe('useAuth — login()', () => {
  it('stores access_token in localStorage after login', async () => {
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

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      await result.current.login('alice@example.com', 'password123')
    })

    expect(localStorage.getItem('access_token')).toBe('jwt.access.token')
  })

  it('stores refresh_token in localStorage after login', async () => {
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

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      await result.current.login('alice@example.com', 'password123')
    })

    expect(localStorage.getItem('refresh_token')).toBe('opaque-refresh')
  })

  it('returns the user after successful login', async () => {
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

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      await result.current.login('alice@example.com', 'password123')
    })

    expect(result.current.user?.username).toBe('alice')
    expect(result.current.user?.city).toBe('almaty')
  })
})

describe('useAuth — logout()', () => {
  it('clears access_token from localStorage after logout', async () => {
    localStorage.setItem('access_token', 'existing.token')
    localStorage.setItem('refresh_token', 'existing-refresh')
    localStorage.setItem('user', JSON.stringify(validUser))

    server.use(
      http.post(`${BASE}/auth/logout`, () =>
        HttpResponse.json({ data: null, error: null, message: 'Logged out' })
      )
    )

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      await result.current.logout()
    })

    expect(localStorage.getItem('access_token')).toBeNull()
  })

  it('sets user to null after logout', async () => {
    localStorage.setItem('access_token', 'existing.token')
    localStorage.setItem('refresh_token', 'existing-refresh')
    localStorage.setItem('user', JSON.stringify(validUser))

    server.use(
      http.post(`${BASE}/auth/logout`, () =>
        HttpResponse.json({ data: null, error: null, message: 'Logged out' })
      )
    )

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      await result.current.logout()
    })

    expect(result.current.user).toBeNull()
  })
})

describe('useAuth — persisted state', () => {
  it('returns user from localStorage on mount if token exists', () => {
    localStorage.setItem('access_token', 'existing.token')
    localStorage.setItem('user', JSON.stringify(validUser))

    const { result } = renderHook(() => useAuth())

    expect(result.current.user?.username).toBe('alice')
  })
})
