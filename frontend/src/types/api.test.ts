import { describe, expect, it } from 'vitest'
import { z } from 'zod'
import {
  ApiResponseSchema,
  GameSchema,
  LeaderboardEntrySchema,
  LoginResponseSchema,
  MoveSchema,
  PaginatedResponseSchema,
  RefreshResponseSchema,
  UserSchema,
} from './api'

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

const validGame = {
  id: 'uuid-g1',
  white_id: 'uuid-1',
  black_id: 'uuid-2',
  status: 'active',
  result: null,
  pgn: null,
  fen: 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1',
  white_rating_before: 1200,
  white_rating_after: null,
  black_rating_before: 1200,
  black_rating_after: null,
  created_at: '2026-05-20T00:00:00Z',
  updated_at: '2026-05-20T00:00:00Z',
  finished_at: null,
}

const validMove = {
  id: 'uuid-m1',
  game_id: 'uuid-g1',
  player_id: 'uuid-1',
  move_number: 1,
  notation: 'e2e4',
  fen_after: 'rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1',
  created_at: '2026-05-20T00:00:00Z',
}

const validLeaderboardEntry = {
  id: 'uuid-l1',
  user_id: 'uuid-1',
  username: 'alice',
  city: 'almaty',
  rating: 1200,
  rank: 1,
  games_played: 42,
  updated_at: '2026-05-20T00:00:00Z',
}

describe('UserSchema', () => {
  it('parses valid user payload', () => {
    const result = UserSchema.safeParse(validUser)
    expect(result.success).toBe(true)
  })

  it('rejects payload missing required email field', () => {
    const { email: _email, ...withoutEmail } = validUser
    const result = UserSchema.safeParse(withoutEmail)
    expect(result.success).toBe(false)
  })

  it('rejects capitalized city value', () => {
    const result = UserSchema.safeParse({ ...validUser, city: 'Almaty' })
    expect(result.success).toBe(false)
  })
})

describe('GameSchema', () => {
  it('parses valid game payload', () => {
    const result = GameSchema.safeParse(validGame)
    expect(result.success).toBe(true)
  })

  it('rejects game with legacy status "finished"', () => {
    const result = GameSchema.safeParse({ ...validGame, status: 'finished' })
    expect(result.success).toBe(false)
  })

  it('rejects game with legacy result "white"', () => {
    const result = GameSchema.safeParse({ ...validGame, result: 'white' })
    expect(result.success).toBe(false)
  })

  it('rejects payload missing fen field', () => {
    const { fen: _fen, ...withoutFen } = validGame
    const result = GameSchema.safeParse(withoutFen)
    expect(result.success).toBe(false)
  })

  it('accepts completed game with rating fields', () => {
    const result = GameSchema.safeParse({
      ...validGame,
      status: 'completed',
      result: 'white_wins',
      white_rating_after: 1216,
      black_rating_after: 1184,
      finished_at: '2026-05-20T01:00:00Z',
    })
    expect(result.success).toBe(true)
  })
})

describe('MoveSchema', () => {
  it('parses valid move payload', () => {
    const result = MoveSchema.safeParse(validMove)
    expect(result.success).toBe(true)
  })

  it('rejects payload missing notation field', () => {
    const { notation: _notation, ...withoutNotation } = validMove
    const result = MoveSchema.safeParse(withoutNotation)
    expect(result.success).toBe(false)
  })
})

describe('LeaderboardEntrySchema', () => {
  it('parses valid leaderboard entry', () => {
    const result = LeaderboardEntrySchema.safeParse(validLeaderboardEntry)
    expect(result.success).toBe(true)
  })

  it('rejects entry missing id field', () => {
    const { id: _id, ...withoutId } = validLeaderboardEntry
    const result = LeaderboardEntrySchema.safeParse(withoutId)
    expect(result.success).toBe(false)
  })

  it('rejects entry missing updated_at field', () => {
    const { updated_at: _updated_at, ...withoutUpdatedAt } = validLeaderboardEntry
    const result = LeaderboardEntrySchema.safeParse(withoutUpdatedAt)
    expect(result.success).toBe(false)
  })
})

describe('ApiResponseSchema', () => {
  it('parses valid ApiResponse with string data', () => {
    const schema = ApiResponseSchema(z.string())
    const result = schema.safeParse({ data: 'hello', error: null, message: 'ok' })
    expect(result.success).toBe(true)
  })

  it('rejects ApiResponse missing message field', () => {
    const schema = ApiResponseSchema(z.string())
    const result = schema.safeParse({ data: 'hello', error: null })
    expect(result.success).toBe(false)
  })
})

describe('PaginatedResponseSchema', () => {
  it('parses valid paginated response with limit field', () => {
    const schema = PaginatedResponseSchema(LeaderboardEntrySchema)
    const result = schema.safeParse({
      data: [validLeaderboardEntry],
      error: null,
      message: 'ok',
      total: 100,
      page: 1,
      limit: 20,
    })
    expect(result.success).toBe(true)
  })

  it('rejects paginated response missing limit field', () => {
    const schema = PaginatedResponseSchema(LeaderboardEntrySchema)
    const result = schema.safeParse({
      data: [],
      error: null,
      message: 'ok',
      total: 100,
      page: 1,
    })
    expect(result.success).toBe(false)
  })

  it('rejects paginated response with page_size instead of limit', () => {
    const schema = PaginatedResponseSchema(LeaderboardEntrySchema)
    const result = schema.safeParse({
      data: [],
      error: null,
      message: 'ok',
      total: 100,
      page: 1,
      page_size: 20,
    })
    expect(result.success).toBe(false)
  })
})

describe('LoginResponseSchema', () => {
  it('parses valid login response', () => {
    const result = LoginResponseSchema.safeParse({
      data: {
        user: validUser,
        access_token: 'jwt.access.token',
        refresh_token: 'opaque-refresh',
      },
      error: null,
      message: 'Login successful',
    })
    expect(result.success).toBe(true)
  })

  it('rejects login response missing refresh_token', () => {
    const result = LoginResponseSchema.safeParse({
      data: {
        user: validUser,
        access_token: 'jwt.access.token',
      },
      error: null,
      message: 'Login successful',
    })
    expect(result.success).toBe(false)
  })
})

describe('RefreshResponseSchema', () => {
  it('parses valid refresh response', () => {
    const result = RefreshResponseSchema.safeParse({
      data: {
        access_token: 'new.jwt.token',
        refresh_token: 'new-opaque-token',
      },
      error: null,
      message: 'Token refreshed successfully',
    })
    expect(result.success).toBe(true)
  })

  it('rejects refresh response missing access_token', () => {
    const result = RefreshResponseSchema.safeParse({
      data: {
        refresh_token: 'new-opaque-token',
      },
      error: null,
      message: 'Token refreshed successfully',
    })
    expect(result.success).toBe(false)
  })
})
