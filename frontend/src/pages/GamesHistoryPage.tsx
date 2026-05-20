import { useSearchParams, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { get } from '../lib/api'
import type { PaginatedResponse, Game } from '../types/api'
import type { Color } from '../types/game'
import { useAuth } from '../hooks/useAuth'
import { formatDate, formatResult } from '../utils/format'
import { Spinner } from '../components/ui/Spinner'
import { Card } from '../components/ui/Card'
import styles from './GamesHistoryPage.module.css'

export function GamesHistoryPage() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()
  const page = parseInt(searchParams.get('page') ?? '1', 10)

  const { data, isLoading, isError } = useQuery({
    queryKey: ['games', 'history', page],
    queryFn: () =>
      get<PaginatedResponse<Game>>(`/users/${user?.id}/games?page=${page}&page_size=20`),
    enabled: !!user,
  })

  const games = data?.data ?? []

  return (
    <main className={styles.page}>
      <h1 className={styles.title}>{t('history.title')}</h1>
      {isLoading && <Spinner center />}
      {isError && <p className={styles.error}>{t('errors.loadFailed')}</p>}
      <div className={styles.list}>
        {games.map((game: Game) => {
          const isWhite = game.white_id === user?.id
          const myColor: Color = isWhite ? 'white' : 'black'
          const opponent = isWhite ? game.black_id : game.white_id
          const resultText = formatResult(game.result, myColor)
          return (
            <Link key={game.id} to={`/game/${game.id}`} className={styles.gameLink}>
              <Card interactive>
                <div className={styles.gameRow}>
                  <span className={styles.color}>{isWhite ? '♘' : '♞'}</span>
                  <span className={styles.opponent}>{opponent}</span>
                  <span className={[styles.result, resultCls(resultText, styles)].join(' ')}>
                    {t(`game.result.${resultText}`)}
                  </span>
                  <span className={styles.date}>{formatDate(game.created_at)}</span>
                </div>
              </Card>
            </Link>
          )
        })}
        {!isLoading && games.length === 0 && (
          <p className={styles.empty}>{t('history.noGames')}</p>
        )}
      </div>
      <div className={styles.pagination}>
        <button
          className={styles.pageBtn}
          onClick={() => setSearchParams({ page: String(page - 1) })}
          disabled={page <= 1}
        >←</button>
        <span>{t('leaderboard.page', { page })}</span>
        <button
          className={styles.pageBtn}
          onClick={() => setSearchParams({ page: String(page + 1) })}
          disabled={games.length < 20}
        >→</button>
      </div>
    </main>
  )
}

function resultCls(text: string, s: Record<string, string>): string {
  if (text === 'Win') return s.win ?? ''
  if (text === 'Loss') return s.loss ?? ''
  return s.draw ?? ''
}
