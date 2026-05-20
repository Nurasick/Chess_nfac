import { useCallback } from 'react'
import { Chessboard } from 'react-chessboard'
import type { PieceDropHandlerArgs, SquareHandlerArgs } from 'react-chessboard'
import styles from './Board.module.css'

interface BoardProps {
  fen: string
  orientation?: 'white' | 'black'
  selectedSquare: string | null
  legalMoves: string[]
  lastMove: { from: string; to: string } | null
  onSquareClick: (square: string) => void
  onPieceDrop: (from: string, to: string) => boolean
  isDisabled?: boolean
}

export function Board({
  fen,
  orientation = 'white',
  selectedSquare,
  legalMoves,
  lastMove,
  onSquareClick,
  onPieceDrop,
  isDisabled = false,
}: BoardProps) {
  const squareStyles: Record<string, React.CSSProperties> = {}

  if (selectedSquare) {
    squareStyles[selectedSquare] = {
      background: 'var(--color-highlight)',
    }
  }

  for (const sq of legalMoves) {
    squareStyles[sq] = {
      background: 'radial-gradient(circle, var(--color-highlight) 30%, transparent 30%)',
    }
  }

  if (lastMove) {
    const lastMoveStyle: React.CSSProperties = { background: 'var(--color-last-move)' }
    squareStyles[lastMove.from] = { ...squareStyles[lastMove.from], ...lastMoveStyle }
    squareStyles[lastMove.to] = { ...squareStyles[lastMove.to], ...lastMoveStyle }
  }

  const handlePieceDrop = useCallback(
    ({ sourceSquare, targetSquare }: PieceDropHandlerArgs) => {
      if (!targetSquare) return false
      return onPieceDrop(sourceSquare, targetSquare)
    },
    [onPieceDrop]
  )

  const handleSquareClick = useCallback(
    ({ square }: SquareHandlerArgs) => {
      if (!isDisabled) onSquareClick(square)
    },
    [onSquareClick, isDisabled]
  )

  return (
    <div className={styles.container}>
      <Chessboard
        options={{
          position: fen,
          boardOrientation: orientation,
          onPieceDrop: handlePieceDrop,
          onSquareClick: handleSquareClick,
          squareStyles,
          darkSquareStyle: { backgroundColor: 'var(--color-board-dark)' },
          lightSquareStyle: { backgroundColor: 'var(--color-board-light)' },
          allowDragging: !isDisabled,
          animationDurationInMs: 150,
        }}
      />
    </div>
  )
}
