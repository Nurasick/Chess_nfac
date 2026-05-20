import { useTranslation } from 'react-i18next'
import { Board } from './Board'
import { GameStatus } from './GameStatus'
import { MovesList } from './MovesList'
import { Button } from '../ui/Button'
import { useChessBoard } from '../../hooks/useChessBoard'
import { useGame } from '../../hooks/useGame'
import type { Color } from '../../types/game'
import styles from './GameContainer.module.css'

interface GameContainerProps {
  playerColor: Color
  whitePlayer: string
  blackPlayer: string
  onMove: (from: string, to: string, promotion?: string) => void
  onResign: () => void
  onOfferDraw: () => void
}

export function GameContainer({
  playerColor,
  whitePlayer,
  blackPlayer,
  onMove,
  onResign,
  onOfferDraw,
}: GameContainerProps) {
  const { t } = useTranslation()
  const { state } = useGame()
  const {
    fen,
    selectedSquare,
    legalMoves,
    lastMove,
    isCheck,
    isGameOver,
    selectSquare,
    makeMove,
    getTurn,
    getHistory,
  } = useChessBoard(state.fen)

  const turn = getTurn()
  const isMyTurn = turn === playerColor && !isGameOver
  const history = getHistory()

  const handleSquareClick = (square: string) => {
    if (!isMyTurn) return
    if (selectedSquare && legalMoves.includes(square)) {
      if (makeMove(selectedSquare, square)) {
        onMove(selectedSquare, square)
      }
    } else {
      selectSquare(square)
    }
  }

  const handlePieceDrop = (from: string, to: string): boolean => {
    if (!isMyTurn) return false
    if (makeMove(from, to)) {
      onMove(from, to)
      return true
    }
    return false
  }

  return (
    <div className={styles.container}>
      <div className={styles.boardArea}>
        <Board
          fen={fen}
          orientation={playerColor}
          selectedSquare={selectedSquare}
          legalMoves={legalMoves}
          lastMove={lastMove}
          onSquareClick={handleSquareClick}
          onPieceDrop={handlePieceDrop}
          isDisabled={!isMyTurn}
        />
      </div>
      <div className={styles.sidebar}>
        <GameStatus
          turn={turn}
          isCheck={isCheck}
          isGameOver={isGameOver}
          result={state.result ?? undefined}
          whitePlayer={whitePlayer}
          blackPlayer={blackPlayer}
          playerColor={playerColor}
        />
        <MovesList moves={history} />
        {!isGameOver && (
          <div className={styles.actions}>
            <Button variant="ghost" size="sm" onClick={onOfferDraw}>
              {t('game.offerDraw')}
            </Button>
            <Button variant="danger" size="sm" onClick={onResign}>
              {t('game.resign')}
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
