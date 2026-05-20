import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Spinner } from '../ui/Spinner'
import { Button } from '../ui/Button'
import styles from './QueueWaiting.module.css'

interface QueueWaitingProps {
  onCancel: () => void
  queuePosition?: number
}

export function QueueWaiting({ onCancel, queuePosition }: QueueWaitingProps) {
  const { t } = useTranslation()
  const [elapsed, setElapsed] = useState(0)

  useEffect(() => {
    const interval = setInterval(() => setElapsed(s => s + 1), 1000)
    return () => clearInterval(interval)
  }, [])

  const formatTime = (s: number) => {
    const m = Math.floor(s / 60)
    const sec = s % 60
    return `${m}:${sec.toString().padStart(2, '0')}`
  }

  return (
    <div className={styles.container}>
      <Spinner size="lg" />
      <h2 className={styles.title}>{t('queue.searching')}</h2>
      <p className={styles.elapsed}>{formatTime(elapsed)}</p>
      {queuePosition != null && (
        <p className={styles.position}>
          {t('queue.position', { position: queuePosition })}
        </p>
      )}
      <Button variant="secondary" onClick={onCancel}>
        {t('queue.cancel')}
      </Button>
    </div>
  )
}
