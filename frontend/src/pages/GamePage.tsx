import { useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { GameContainer } from '../components/game/GameContainer'
import { EngineAnalysis } from '../components/analysis/EngineAnalysis'
import { EvaluationGraph } from '../components/analysis/EvaluationGraph'
import { BestMoves } from '../components/analysis/BestMoves'
import { Spinner } from '../components/ui/Spinner'
import { useWebSocketMessage } from '../hooks/useWebSocket'
import { useWebSocketContext } from '../context/WebSocketContext'
import { useGameContext } from '../context/GameContext'
import { useStockfish } from '../hooks/useStockfish'
import { useEvaluation } from '../hooks/useEvaluation'
import { isMoveMade, isGameEnd, isDrawOffered } from '../types/websocket'
import styles from './GamePage.module.css'

export function GamePage() {
  const { t } = useTranslation()
  const { gameId } = useParams<{ gameId: string }>()
  const { send } = useWebSocketContext()
  const { state, dispatch } = useGameContext()
  const { isReady, currentEval, bestMove, analyze, onEval } = useStockfish()
  const { evalHistory, addEval, reset: resetEval } = useEvaluation()

  useEffect(() => {
    resetEval()
  }, [gameId, resetEval])

  useEffect(() => {
    if (!state.fen) return
    analyze(state.fen)
  }, [state.fen, analyze])

  useEffect(() => {
    return onEval(point => {
      addEval({ ...point, moveNumber: state.moveHistory.length })
    })
  }, [onEval, addEval, state.moveHistory.length])

  useWebSocketMessage(msg => {
    if (isMoveMade(msg)) {
      dispatch({
        type: 'MOVE_MADE',
        fen: msg.fen,
        notation: msg.notation,
        isMyTurn: msg.your_turn,
      })
    }
    if (isGameEnd(msg)) {
      dispatch({
        type: 'GAME_END',
        result: msg.result,
        reason: msg.reason,
        whiteRatingDelta: msg.white_rating_delta,
        blackRatingDelta: msg.black_rating_delta,
        myColor: state.myColor ?? 'white',
      })
    }
    if (isDrawOffered(msg)) {
      if (window.confirm(t('game.drawOffered'))) {
        send({ type: 'draw_response', payload: { accepted: true } })
      } else {
        send({ type: 'draw_response', payload: { accepted: false } })
      }
    }
  })

  if (state.status === 'waiting' && !state.gameId) {
    return <Spinner center />
  }

  const playerColor = state.myColor ?? 'white'

  const handleMove = (from: string, to: string) => {
    send({
      type: 'move',
      payload: { game_id: gameId!, from, to },
    })
  }

  const handleResign = () => {
    if (window.confirm(t('game.confirmResign'))) {
      send({ type: 'resign', payload: { game_id: gameId! } })
    }
  }

  const handleOfferDraw = () => {
    send({ type: 'offer_draw', payload: { game_id: gameId! } })
  }

  return (
    <main className={styles.page}>
      <div className={styles.layout}>
        <div className={styles.boardSection}>
          <GameContainer
            playerColor={playerColor}
            whitePlayer={t('game.whitePlayer')}
            blackPlayer={t('game.blackPlayer')}
            onMove={handleMove}
            onResign={handleResign}
            onOfferDraw={handleOfferDraw}
          />
        </div>
        <aside className={styles.analysis}>
          <EngineAnalysis currentEval={currentEval} isReady={isReady} />
          <BestMoves bestMove={bestMove} />
          <EvaluationGraph evalHistory={evalHistory} />
        </aside>
      </div>
    </main>
  )
}
