/**
 * Chess.js Wrapper
 * Encapsulates chess.js logic with typed methods
 */

import { Chess } from 'chess.js';
import type { Square } from 'chess.js';
import type { Color, Move, BoardState } from '../types/game';

export class ChessWrapper {
  private chess: Chess;

  constructor() {
    this.chess = new Chess();
  }

  /**
   * Get current FEN string
   */
  getFen(): string {
    return this.chess.fen();
  }

  /**
   * Get current player's turn
   */
  getTurn(): Color {
    return this.chess.turn() === 'w' ? 'white' : 'black';
  }

  /**
   * Check if current position is check
   */
  isCheck(): boolean {
    return this.chess.isCheck();
  }

  /**
   * Check if current position is checkmate
   */
  isCheckmate(): boolean {
    return this.chess.isCheckmate();
  }

  /**
   * Check if current position is stalemate
   */
  isStalemate(): boolean {
    return this.chess.isStalemate();
  }

  /**
   * Check if game is drawn (draw, stalemate, insufficient material, etc.)
   */
  isDraw(): boolean {
    return this.chess.isDraw();
  }

  /**
   * Get all legal moves, optionally from a specific square
   */
  getLegalMoves(from?: string): string[] {
    const moves = this.chess.moves({
      square: from as Square | undefined,
      verbose: false,
    });
    return moves as string[];
  }

  /**
   * Get legal moves in verbose format
   */
  getLegalMovesVerbose(from?: string) {
    return this.chess.moves({
      square: from as Square | undefined,
      verbose: true,
    });
  }

  /**
   * Validate and apply a move
   */
  validateMove(from: string, to: string, promotion?: string): Move | null {
    try {
      const move = this.chess.move({
        from,
        to,
        promotion: promotion as 'q' | 'r' | 'b' | 'n' | undefined,
      });

      if (move) {
        return { from, to, promotion };
      }
      return null;
    } catch {
      return null;
    }
  }

  /**
   * Apply a move in algebraic notation or UCI format
   */
  applyMove(move: string): boolean {
    try {
      const result = this.chess.move(move);
      return result !== null;
    } catch {
      return false;
    }
  }

  /**
   * Load a FEN position
   */
  loadFen(fen: string): void {
    this.chess.load(fen);
  }

  /**
   * Get move history
   */
  getHistory(): string[] {
    return this.chess.history();
  }

  /**
   * Get verbose move history
   */
  getHistoryVerbose() {
    return this.chess.history({ verbose: true });
  }

  /**
   * Reset the board to starting position
   */
  reset(): void {
    this.chess.reset();
  }

  /**
   * Get current board state
   */
  getBoardState(): BoardState {
    return {
      fen: this.getFen(),
      turn: this.getTurn(),
      isCheck: this.isCheck(),
      isCheckmate: this.isCheckmate(),
      isStalemate: this.isStalemate(),
      isDraw: this.isDraw(),
    };
  }

  /**
   * Get pieces on the board
   */
  getBoard() {
    return this.chess.board();
  }

  /**
   * Check if game is over
   */
  isGameOver(): boolean {
    return this.chess.isGameOver();
  }

  /**
   * Get the reason game ended
   */
  getGameOverReason(): string | null {
    if (!this.isGameOver()) {
      return null;
    }

    if (this.isCheckmate()) {
      return 'checkmate';
    }
    if (this.isStalemate()) {
      return 'stalemate';
    }
    if (this.chess.isInsufficientMaterial()) {
      return 'insufficient_material';
    }
    if (this.chess.isThreefoldRepetition()) {
      return 'threefold_repetition';
    }
    if (this.chess.isDraw()) {
      return 'fifty_moves';
    }

    return 'unknown';
  }

  /**
   * Move to a specific move in history (0-indexed)
   */
  goToMove(moveNumber: number): void {
    const history = this.getHistory();
    this.reset();
    for (let i = 0; i < moveNumber && i < history.length; i++) {
      this.applyMove(history[i]);
    }
  }

  /**
   * Create a copy of the current chess state
   */
  clone(): ChessWrapper {
    const wrapper = new ChessWrapper();
    wrapper.loadFen(this.getFen());
    return wrapper;
  }
}

/**
 * Singleton instance for global use
 */
let instance: ChessWrapper | null = null;

export function getChessInstance(): ChessWrapper {
  if (!instance) {
    instance = new ChessWrapper();
  }
  return instance;
}

export function resetChessInstance(): void {
  instance = null;
}
