import { useTranslation } from 'react-i18next'
import type { EvalPoint } from '../../types/evaluation'
import styles from './EngineAnalysis.module.css'

interface EngineAnalysisProps {
  currentEval: EvalPoint | null
  isReady: boolean
}

export function EngineAnalysis({ currentEval, isReady }: EngineAnalysisProps) {
  const { t } = useTranslation()

  if (!isReady) {
    return (
      <div className={styles.container}>
        <span className={styles.loading}>{t('analysis.loading')}</span>
      </div>
    )
  }

  const formatScore = (e: EvalPoint) => {
    if (e.mate != null) {
      return e.mate > 0 ? `+M${e.mate}` : `-M${Math.abs(e.mate)}`
    }
    const sign = e.score > 0 ? '+' : ''
    return `${sign}${e.score.toFixed(2)}`
  }

  const scoreValue = currentEval?.score ?? 0
  const whiteAdv = Math.max(0, Math.min(100, 50 + scoreValue * 5))

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <span className={styles.label}>{t('analysis.engineEval')}</span>
        {currentEval && (
          <span className={styles.score}>{formatScore(currentEval)}</span>
        )}
      </div>
      <div className={styles.bar} aria-label={t('analysis.advantage')}>
        <div
          className={styles.whiteBar}
          style={{ width: `${whiteAdv}%` }}
        />
      </div>
      {currentEval && (
        <div className={styles.depth}>
          {t('analysis.depth')}: {currentEval.depth}
        </div>
      )}
    </div>
  )
}
