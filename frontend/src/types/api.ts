import { z } from 'zod'

export const UserSchema = z.object({
  id: z.string(),
  username: z.string(),
  email: z.string(),
  city: z.enum(['almaty', 'astana', 'shymkent']),
  rating: z.number(),
  games_played: z.number(),
  created_at: z.string(),
  updated_at: z.string(),
})

export const GameSchema = z.object({
  id: z.string(),
  white_id: z.string(),
  black_id: z.string(),
  status: z.enum(['waiting', 'active', 'completed', 'abandoned']),
  result: z.enum(['white_wins', 'black_wins', 'draw']).nullable(),
  pgn: z.string().nullable(),
  fen: z.string(),
  white_rating_before: z.number(),
  white_rating_after: z.number().nullable(),
  black_rating_before: z.number(),
  black_rating_after: z.number().nullable(),
  created_at: z.string(),
  updated_at: z.string(),
  finished_at: z.string().nullable(),
})

export const MoveSchema = z.object({
  id: z.string(),
  game_id: z.string(),
  player_id: z.string(),
  move_number: z.number(),
  notation: z.string(),
  fen_after: z.string(),
  created_at: z.string(),
})

export const LeaderboardEntrySchema = z.object({
  id: z.string(),
  user_id: z.string(),
  username: z.string(),
  city: z.enum(['almaty', 'astana', 'shymkent']),
  rating: z.number(),
  rank: z.number(),
  games_played: z.number(),
  updated_at: z.string(),
})

export function ApiResponseSchema<T extends z.ZodTypeAny>(dataSchema: T) {
  return z.object({
    data: dataSchema.nullable(),
    error: z.string().nullable(),
    message: z.string(),
  })
}

export function PaginatedResponseSchema<T extends z.ZodTypeAny>(itemSchema: T) {
  return z.object({
    data: z.array(itemSchema),
    error: z.string().nullable(),
    message: z.string(),
    total: z.number(),
    page: z.number(),
    limit: z.number(),
  })
}

export const LoginResponseSchema = z.object({
  data: z.object({
    user: UserSchema,
    access_token: z.string(),
    refresh_token: z.string(),
  }),
  error: z.string().nullable(),
  message: z.string(),
})

export const RefreshResponseSchema = z.object({
  data: z.object({
    access_token: z.string(),
    refresh_token: z.string(),
  }),
  error: z.string().nullable(),
  message: z.string(),
})

export type User = z.infer<typeof UserSchema>
export type Game = z.infer<typeof GameSchema>
export type Move = z.infer<typeof MoveSchema>
export type LeaderboardEntry = z.infer<typeof LeaderboardEntrySchema>
export type LoginResponse = z.infer<typeof LoginResponseSchema>
export type RefreshResponse = z.infer<typeof RefreshResponseSchema>

export type PaginatedResponse<T> = {
  data: T[]
  error: string | null
  message: string
  total: number
  page: number
  limit: number
}
