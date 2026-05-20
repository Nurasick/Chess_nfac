import { useTranslation } from 'react-i18next'
import type { Color } from '../../types/game'
import styles from './GameStatus.module.css'

interface GameStatusProps {
  turn: Color
  isCheck: boolean
  isGameOver: boolean
  result?: string
  whitePlayer?: string
  blackPlayer?: string
  playerColor?: Color
}

export function GameStatus({
  turn,
  isCheck,
  isGameOver,
  result,
  whitePlayer,
  blackPlayer,
  playerColor,
}: GameStatusProps) {
  const { t } = useTranslation()

  if (isGameOver) {
    return (
      <div className={`${styles.status} ${styles.gameOver}`}>
        <span className={styles.resultText}>{result ?? t('game.gameOver')}</span>
      </div>
    )
  }

  const isYourTurn = playerColor === turn

  return (
    <div className={styles.status}>
      <div className={styles.players}>
        <span
          className={[
            styles.player,
            turn === 'black' ? styles.active : '',
          ]
            .filter(Boolean)
            .join(' ')}
        >
          ♞ {blackPlayer ?? t('game.blackPlayer')}
        </span>
        <span className={styles.vs}>vs</span>
        <span
          className={[
            styles.player,
            turn === 'white' ? styles.active : '',
          ]
            .filter(Boolean)
            .join(' ')}
        >
          ♘ {whitePlayer ?? t('game.whitePlayer')}
        </span>
      </div>
      <div
        className={[
          styles.turnIndicator,
          isCheck ? styles.check : '',
          isYourTurn ? styles.yourTurn : '',
        ]
          .filter(Boolean)
          .join(' ')}
      >
        {isCheck
          ? t('game.check')
          : isYourTurn
          ? t('game.yourTurn')
          : t('game.opponentTurn')}
      </div>
    </div>
  )
}
