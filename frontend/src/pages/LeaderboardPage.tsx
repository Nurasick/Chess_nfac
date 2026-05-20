import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { CityLeaderboard } from '../components/leaderboard/CityLeaderboard'
import { GlobalLeaderboard } from '../components/leaderboard/GlobalLeaderboard'
import { CITIES } from '../utils/constants'
import styles from './LeaderboardPage.module.css'

const TABS = ['global', ...CITIES]

export function LeaderboardPage() {
  const { t } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()
  const activeTab = searchParams.get('city') ?? 'global'
  const page = parseInt(searchParams.get('page') ?? '1', 10)

  const setTab = (tab: string) => setSearchParams({ city: tab, page: '1' })
  const setPage = (p: number) => setSearchParams({ city: activeTab, page: String(p) })

  return (
    <main className={styles.page}>
      <h1 className={styles.title}>{t('leaderboard.title')}</h1>
      <div className={styles.tabs} role="tablist">
        {TABS.map(tab => (
          <button
            key={tab}
            role="tab"
            aria-selected={activeTab === tab}
            className={[styles.tab, activeTab === tab ? styles.activeTab : ''].join(' ')}
            onClick={() => setTab(tab)}
          >
            {tab === 'global' ? t('leaderboard.global') : t(`cities.${tab}`)}
          </button>
        ))}
      </div>
      <div className={styles.content}>
        {activeTab === 'global' ? (
          <GlobalLeaderboard page={page} />
        ) : (
          <CityLeaderboard city={activeTab} page={page} />
        )}
      </div>
      <div className={styles.pagination}>
        <button
          className={styles.pageBtn}
          onClick={() => setPage(page - 1)}
          disabled={page <= 1}
        >
          ←
        </button>
        <span className={styles.pageNum}>{t('leaderboard.page', { page })}</span>
        <button
          className={styles.pageBtn}
          onClick={() => setPage(page + 1)}
        >
          →
        </button>
      </div>
    </main>
  )
}
