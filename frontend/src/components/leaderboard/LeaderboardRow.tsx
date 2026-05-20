import { formatElo } from '../../utils/format'
import type { LeaderboardEntry } from '../../types/api'
import styles from './LeaderboardRow.module.css'

interface LeaderboardRowProps {
  entry: LeaderboardEntry
  rank: number
}

export function LeaderboardRow({ entry, rank }: LeaderboardRowProps) {
  const rankClass =
    rank === 1 ? styles.gold : rank === 2 ? styles.silver : rank === 3 ? styles.bronze : ''

  return (
    <div className={`${styles.row} ${rankClass}`}>
      <span className={styles.rank}>{rank}</span>
      <span className={styles.username}>{entry.username}</span>
      <span className={styles.rating}>{formatElo(entry.rating)}</span>
      <span className={styles.games}>{entry.games_played}</span>
    </div>
  )
}
