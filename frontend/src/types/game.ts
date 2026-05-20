/**
 * Game Logic Types
 * Types for board state, moves, and game results
 */

/**
 * Player color on the board
 */
export type Color = 'white' | 'black';

/**
 * Game result
 */
export type GameResult = 'white' | 'black' | 'draw';

/**
 * Game status
 */
export type GameStatus = 'waiting' | 'active' | 'finished';

/**
 * A move on the board
 */
export interface Move {
  from: string;
  to: string;
  promotion?: string;
}

/**
 * Board state snapshot
 */
export interface BoardState {
  fen: string;
  turn: Color;
  isCheck: boolean;
  isCheckmate: boolean;
  isStalemate: boolean;
  isDraw: boolean;
}

/**
 * Square coordinates (e.g., 'e2', 'e4')
 */
export type Square = string;

/**
 * Promotion piece options
 */
export type PromotionPiece = 'q' | 'r' | 'b' | 'n';
