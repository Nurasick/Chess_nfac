import { describe, it, expect } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { GameProvider, useGameContext } from './GameContext';

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <GameProvider>{children}</GameProvider>
);

describe('GameContext', () => {
  it('initial state has gameId=null, status="waiting", isMyTurn=false', () => {
    const { result } = renderHook(() => useGameContext(), { wrapper });
    expect(result.current.state.gameId).toBeNull();
    expect(result.current.state.status).toBe('waiting');
    expect(result.current.state.isMyTurn).toBe(false);
  });

  it('GAME_START sets gameId, myColor, isMyTurn=true when myColor="white"', () => {
    const { result } = renderHook(() => useGameContext(), { wrapper });
    act(() => {
      result.current.dispatch({
        type: 'GAME_START',
        gameId: 'game-1',
        fen: 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1',
        myColor: 'white',
        whiteRating: 1200,
        blackRating: 1300,
      });
    });
    expect(result.current.state.gameId).toBe('game-1');
    expect(result.current.state.myColor).toBe('white');
    expect(result.current.state.isMyTurn).toBe(true);
  });

  it('GAME_START sets isMyTurn=false when myColor="black"', () => {
    const { result } = renderHook(() => useGameContext(), { wrapper });
    act(() => {
      result.current.dispatch({
        type: 'GAME_START',
        gameId: 'game-2',
        fen: 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1',
        myColor: 'black',
        whiteRating: 1200,
        blackRating: 1300,
      });
    });
    expect(result.current.state.isMyTurn).toBe(false);
  });

  it('MOVE_MADE updates fen and appends notation to moveHistory', () => {
    const { result } = renderHook(() => useGameContext(), { wrapper });
    act(() => {
      result.current.dispatch({
        type: 'MOVE_MADE',
        fen: 'rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1',
        notation: 'e4',
        isMyTurn: false,
      });
    });
    expect(result.current.state.fen).toBe('rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1');
    expect(result.current.state.moveHistory).toContain('e4');
  });

  it('MOVE_MADE toggles isMyTurn', () => {
    const { result } = renderHook(() => useGameContext(), { wrapper });
    act(() => {
      result.current.dispatch({
        type: 'MOVE_MADE',
        fen: 'some-fen',
        notation: 'e4',
        isMyTurn: true,
      });
    });
    expect(result.current.state.isMyTurn).toBe(true);
    act(() => {
      result.current.dispatch({
        type: 'MOVE_MADE',
        fen: 'some-fen-2',
        notation: 'e5',
        isMyTurn: false,
      });
    });
    expect(result.current.state.isMyTurn).toBe(false);
  });

  it('GAME_END sets status="finished" and result', () => {
    const { result } = renderHook(() => useGameContext(), { wrapper });
    act(() => {
      result.current.dispatch({
        type: 'GAME_END',
        result: 'white',
        reason: 'checkmate',
        whiteRatingDelta: 10,
        blackRatingDelta: -10,
        myColor: 'white',
      });
    });
    expect(result.current.state.status).toBe('finished');
    expect(result.current.state.result).toBe('white');
  });

  it('GAME_END computes ratingDelta from myColor (white delta vs black delta)', () => {
    const { result } = renderHook(() => useGameContext(), { wrapper });
    act(() => {
      result.current.dispatch({
        type: 'GAME_END',
        result: 'white',
        reason: 'checkmate',
        whiteRatingDelta: 15,
        blackRatingDelta: -15,
        myColor: 'black',
      });
    });
    expect(result.current.state.ratingDelta).toBe(-15);
  });

  it('RESET returns to initial state', () => {
    const { result } = renderHook(() => useGameContext(), { wrapper });
    act(() => {
      result.current.dispatch({
        type: 'GAME_START',
        gameId: 'game-x',
        fen: 'fen',
        myColor: 'white',
        whiteRating: 1000,
        blackRating: 1000,
      });
    });
    act(() => {
      result.current.dispatch({ type: 'RESET' });
    });
    expect(result.current.state.gameId).toBeNull();
    expect(result.current.state.status).toBe('waiting');
  });

  it('useGameContext() throws Error when used outside GameProvider', () => {
    expect(() => {
      renderHook(() => useGameContext());
    }).toThrow('useGameContext must be used within GameProvider');
  });
});
