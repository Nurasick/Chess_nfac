import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { get } from '../lib/api'
import { useAuth } from '../hooks/useAuth'
import type { User } from '../types/api'
import { formatElo, formatDate } from '../utils/format'
import { Spinner } from '../components/ui/Spinner'
import { Card } from '../components/ui/Card'
import styles from './ProfilePage.module.css'

export function ProfilePage() {
  const { t } = useTranslation()
  const { user: authUser } = useAuth()

  const { data, isLoading } = useQuery({
    queryKey: ['profile', authUser?.id],
    queryFn: () => get<User>(`/users/${authUser?.id}`),
    enabled: !!authUser,
  })

  const profile = data ?? authUser

  if (isLoading) return <Spinner center />

  return (
    <main className={styles.page}>
      <div className={styles.header}>
        <div className={styles.avatar}>{profile?.username?.[0]?.toUpperCase()}</div>
        <div>
          <h1 className={styles.username}>{profile?.username}</h1>
          <p className={styles.city}>{profile?.city ? t(`cities.${profile.city}`) : ''}</p>
        </div>
      </div>
      <div className={styles.stats}>
        <Card>
          <div className={styles.stat}>
            <span className={styles.statValue}>{formatElo(profile?.rating ?? 1200)}</span>
            <span className={styles.statLabel}>{t('profile.rating')}</span>
          </div>
        </Card>
        <Card>
          <div className={styles.stat}>
            <span className={styles.statValue}>{profile?.games_played ?? 0}</span>
            <span className={styles.statLabel}>{t('profile.gamesPlayed')}</span>
          </div>
        </Card>
        <Card>
          <div className={styles.stat}>
            <span className={styles.statValue}>{formatDate(profile?.created_at ?? '')}</span>
            <span className={styles.statLabel}>{t('profile.memberSince')}</span>
          </div>
        </Card>
      </div>
    </main>
  )
}
