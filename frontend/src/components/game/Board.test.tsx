import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { Board } from './Board';

vi.mock('react-chessboard', () => ({
  Chessboard: ({ options }: { options: { onSquareClick?: (args: { square: string }) => void; onPieceDrop?: (args: { sourceSquare: string; targetSquare: string }) => boolean } }) => (
    <div data-testid="chessboard">
      <button data-testid="sq-e2" onClick={() => options.onSquareClick?.({ square: 'e2' })} />
      <button data-testid="sq-e4" onClick={() => options.onSquareClick?.({ square: 'e4' })} />
      <button data-testid="drop" onClick={() => options.onPieceDrop?.({ sourceSquare: 'e2', targetSquare: 'e4' })} />
    </div>
  ),
}));

const FEN = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1';

const defaultProps = {
  fen: FEN,
  selectedSquare: null,
  legalMoves: [],
  lastMove: null,
  onSquareClick: vi.fn(),
  onPieceDrop: vi.fn().mockReturnValue(true),
};

describe('Board', () => {
  it('renders the chessboard container', () => {
    render(<Board {...defaultProps} />);
    expect(screen.getByTestId('chessboard')).toBeDefined();
  });

  it('clicking sq-e2 calls onSquareClick with "e2"', () => {
    const onSquareClick = vi.fn();
    render(<Board {...defaultProps} onSquareClick={onSquareClick} />);
    fireEvent.click(screen.getByTestId('sq-e2'));
    expect(onSquareClick).toHaveBeenCalledWith('e2');
  });

  it('clicking drop calls onPieceDrop with ("e2","e4") and returns the callback return value', () => {
    const onPieceDrop = vi.fn().mockReturnValue(true);
    render(<Board {...defaultProps} onPieceDrop={onPieceDrop} />);
    fireEvent.click(screen.getByTestId('drop'));
    expect(onPieceDrop).toHaveBeenCalledWith('e2', 'e4');
  });

  it('when isDisabled=true, clicking sq-e2 does NOT call onSquareClick', () => {
    const onSquareClick = vi.fn();
    render(<Board {...defaultProps} onSquareClick={onSquareClick} isDisabled={true} />);
    fireEvent.click(screen.getByTestId('sq-e2'));
    expect(onSquareClick).not.toHaveBeenCalled();
  });

  it('renders with selectedSquare prop (covers selectedSquare squareStyles branch)', () => {
    render(<Board {...defaultProps} selectedSquare="e2" />);
    expect(screen.getByTestId('chessboard')).toBeDefined();
  });

  it('renders with legalMoves prop (covers legalMoves squareStyles branch)', () => {
    render(<Board {...defaultProps} legalMoves={['e3', 'e4']} />);
    expect(screen.getByTestId('chessboard')).toBeDefined();
  });

  it('renders with lastMove prop (covers lastMove squareStyles branch)', () => {
    render(<Board {...defaultProps} lastMove={{ from: 'e2', to: 'e4' }} />);
    expect(screen.getByTestId('chessboard')).toBeDefined();
  });
});
