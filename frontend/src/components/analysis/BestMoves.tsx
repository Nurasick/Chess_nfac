import { useTranslation } from 'react-i18next'
import styles from './BestMoves.module.css'

interface BestMovesProps {
  bestMove: string | null
}

export function BestMoves({ bestMove }: BestMovesProps) {
  const { t } = useTranslation()

  return (
    <div className={styles.container}>
      <span className={styles.label}>{t('analysis.bestMove')}</span>
      {bestMove ? (
        <span className={styles.move}>{bestMove}</span>
      ) : (
        <span className={styles.empty}>—</span>
      )}
    </div>
  )
}
