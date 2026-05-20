import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { TimeControlSelect } from '../components/queue/TimeControlSelect'
import { QueueWaiting } from '../components/queue/QueueWaiting'
import { Button } from '../components/ui/Button'
import { useWebSocketMessage } from '../hooks/useWebSocket'
import { useWebSocketContext } from '../context/WebSocketContext'
import { useGameContext } from '../context/GameContext'
import { isGameStart, isQueueStatus } from '../types/websocket'
import { TIME_CONTROLS } from '../utils/constants'
import styles from './HomePage.module.css'

export function HomePage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { send } = useWebSocketContext()
  const { dispatch } = useGameContext()
  const [selectedTimeControl, setSelectedTimeControl] = useState<string>(TIME_CONTROLS[1].value)
  const [isInQueue, setIsInQueue] = useState(false)
  const [queuePosition, setQueuePosition] = useState<number | undefined>()

  useWebSocketMessage(msg => {
    if (isQueueStatus(msg)) {
      setQueuePosition(msg.queue_size)
    }
    if (isGameStart(msg)) {
      dispatch({
        type: 'GAME_START',
        gameId: msg.game_id,
        fen: msg.fen,
        myColor: msg.your_color,
        whiteRating: msg.white_rating,
        blackRating: msg.black_rating,
      })
      navigate(`/game/${msg.game_id}`)
    }
  })

  const handleJoinQueue = () => {
    const tc = TIME_CONTROLS.find(t => t.value === selectedTimeControl)
    if (!tc) return
    send({ type: 'queue_join', payload: { time_control: tc.value } })
    setIsInQueue(true)
  }

  const handleLeaveQueue = () => {
    send({ type: 'queue_leave', payload: {} })
    setIsInQueue(false)
    setQueuePosition(undefined)
  }

  if (isInQueue) {
    return (
      <main className={styles.page}>
        <QueueWaiting onCancel={handleLeaveQueue} queuePosition={queuePosition} />
      </main>
    )
  }

  return (
    <main className={styles.page}>
      <div className={styles.hero}>
        <h1 className={styles.headline}>{t('home.headline')}</h1>
        <p className={styles.subline}>{t('home.subline')}</p>
      </div>
      <div className={styles.card}>
        <TimeControlSelect selected={selectedTimeControl} onChange={setSelectedTimeControl} />
        <Button size="lg" fullWidth onClick={handleJoinQueue}>
          {t('home.playNow')}
        </Button>
      </div>
    </main>
  )
}
