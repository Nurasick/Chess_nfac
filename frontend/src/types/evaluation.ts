/**
 * Engine Evaluation Types
 * Types for Stockfish analysis results
 */

/**
 * A single evaluation point in the game
 */
export interface EvalPoint {
  moveNumber: number;
  score: number;
  depth: number;
  mate?: number | null;
}
