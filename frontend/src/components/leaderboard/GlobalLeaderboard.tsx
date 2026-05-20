import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { get } from '../../lib/api'
import type { PaginatedResponse, LeaderboardEntry } from '../../types/api'
import { LeaderboardRow } from './LeaderboardRow'
import { Spinner } from '../ui/Spinner'
import styles from './Leaderboard.module.css'

interface GlobalLeaderboardProps {
  page: number
}

export function GlobalLeaderboard({ page }: GlobalLeaderboardProps) {
  const { t } = useTranslation()

  const { data, isLoading, isError } = useQuery({
    queryKey: ['leaderboard', 'global', page],
    queryFn: () =>
      get<PaginatedResponse<LeaderboardEntry>>(`/leaderboard/global?page=${page}&page_size=20`),
  })

  if (isLoading) return <Spinner center />
  if (isError || !data?.data) {
    return <p className={styles.error}>{t('errors.loadFailed')}</p>
  }

  const entries = data.data ?? []

  return (
    <div className={styles.list}>
      <div className={styles.header}>
        <span>{t('leaderboard.rank')}</span>
        <span>{t('leaderboard.player')}</span>
        <span className={styles.right}>{t('leaderboard.rating')}</span>
        <span className={styles.right}>{t('leaderboard.games')}</span>
      </div>
      {entries.map((entry, idx) => (
        <LeaderboardRow
          key={entry.user_id}
          entry={entry}
          rank={(page - 1) * 20 + idx + 1}
        />
      ))}
      {entries.length === 0 && (
        <p className={styles.empty}>{t('leaderboard.noPlayers')}</p>
      )}
    </div>
  )
}
