import { useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import styles from './MovesList.module.css'

interface MovesListProps {
  moves: string[]
  currentMoveIndex?: number
}

export function MovesList({ moves, currentMoveIndex }: MovesListProps) {
  const { t } = useTranslation()
  const endRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  }, [moves.length])

  const pairs: Array<[string, string | undefined]> = []
  for (let i = 0; i < moves.length; i += 2) {
    pairs.push([moves[i], moves[i + 1]])
  }

  if (pairs.length === 0) {
    return (
      <div className={styles.empty}>
        <span>{t('game.noMoves')}</span>
      </div>
    )
  }

  return (
    <div className={styles.list} role="log" aria-label={t('game.moveHistory')}>
      {pairs.map(([white, black], idx) => {
        const moveNum = idx + 1
        const whiteIdx = idx * 2
        const blackIdx = idx * 2 + 1
        return (
          <div key={moveNum} className={styles.row}>
            <span className={styles.moveNumber}>{moveNum}.</span>
            <span
              className={[
                styles.move,
                currentMoveIndex === whiteIdx ? styles.active : '',
              ]
                .filter(Boolean)
                .join(' ')}
            >
              {white}
            </span>
            {black && (
              <span
                className={[
                  styles.move,
                  currentMoveIndex === blackIdx ? styles.active : '',
                ]
                  .filter(Boolean)
                  .join(' ')}
              >
                {black}
              </span>
            )}
          </div>
        )
      })}
      <div ref={endRef} />
    </div>
  )
}
