import { describe, it, expect } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useChessBoard } from './useChessBoard';

const STARTING_FEN = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1';

describe('useChessBoard', () => {
  it('initial fen is the starting position FEN', () => {
    const { result } = renderHook(() => useChessBoard());
    expect(result.current.fen).toBe(STARTING_FEN);
  });

  it('makeMove e2→e4 returns true and fen changes', () => {
    const { result } = renderHook(() => useChessBoard());
    let moved: boolean;
    act(() => {
      moved = result.current.makeMove('e2', 'e4');
    });
    expect(moved!).toBe(true);
    expect(result.current.fen).not.toBe(STARTING_FEN);
  });

  it('makeMove e2→e5 (illegal) returns false and fen is unchanged', () => {
    const { result } = renderHook(() => useChessBoard());
    let moved: boolean;
    act(() => {
      moved = result.current.makeMove('e2', 'e5');
    });
    expect(moved!).toBe(false);
    expect(result.current.fen).toBe(STARTING_FEN);
  });

  it('after makeMove e2→e4, getTurn() returns black', () => {
    const { result } = renderHook(() => useChessBoard());
    act(() => {
      result.current.makeMove('e2', 'e4');
    });
    expect(result.current.getTurn()).toBe('black');
  });

  it('selectSquare e2 sets legalMoves to a non-empty array', () => {
    const { result } = renderHook(() => useChessBoard());
    act(() => {
      result.current.selectSquare('e2');
    });
    expect(result.current.legalMoves.length).toBeGreaterThan(0);
  });

  it('selectSquare e5 (empty square at start) sets selectedSquare to null', () => {
    const { result } = renderHook(() => useChessBoard());
    act(() => {
      result.current.selectSquare('e5');
    });
    expect(result.current.selectedSquare).toBeNull();
  });

  it('loadPosition(fen) updates the fen in state', () => {
    const { result } = renderHook(() => useChessBoard());
    // Use a FEN without en passant to avoid chess.js normalization
    const newFen = 'rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1';
    act(() => {
      result.current.loadPosition(newFen);
    });
    expect(result.current.fen).toBe(newFen);
  });

  it('applyServerMove("e2e4", newFen) sets lastMove = { from: "e2", to: "e4" }', () => {
    const { result } = renderHook(() => useChessBoard());
    const newFen = 'rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1';
    act(() => {
      result.current.applyServerMove('e2e4', newFen);
    });
    expect(result.current.lastMove).toEqual({ from: 'e2', to: 'e4' });
  });

  it('reset() restores starting position fen', () => {
    const { result } = renderHook(() => useChessBoard());
    act(() => {
      result.current.makeMove('e2', 'e4');
    });
    act(() => {
      result.current.reset();
    });
    expect(result.current.fen).toBe(STARTING_FEN);
  });
});
