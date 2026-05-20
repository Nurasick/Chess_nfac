/**
 * Zod Validation Schemas
 */

import { z } from 'zod';

/**
 * Login form schema
 */
export const loginSchema = z.object({
  username: z.string()
    .min(3, 'Username must be at least 3 characters')
    .max(20, 'Username must be at most 20 characters'),
  password: z.string()
    .min(6, 'Password must be at least 6 characters'),
});

export type LoginInput = z.infer<typeof loginSchema>;

/**
 * Register form schema
 */
export const registerSchema = z.object({
  username: z.string()
    .min(3, 'Username must be at least 3 characters')
    .max(20, 'Username must be at most 20 characters')
    .regex(/^[a-zA-Z0-9_]+$/, 'Username can only contain letters, numbers, and underscores'),
  email: z.string()
    .email('Please enter a valid email address'),
  password: z.string()
    .min(8, 'Password must be at least 8 characters')
    .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Password must contain at least one number'),
  city: z.enum(['Almaty', 'Astana', 'Shymkent'] as const),
});

export type RegisterInput = z.infer<typeof registerSchema>;

/**
 * Move notation schema (algebraic notation or UCI)
 */
export const moveSchema = z.string()
  .min(2, 'Invalid move')
  .max(5, 'Invalid move');

/**
 * Game ID schema
 */
export const gameIdSchema = z.string()
  .uuid('Invalid game ID');

/**
 * User ID schema
 */
export const userIdSchema = z.string()
  .uuid('Invalid user ID');
