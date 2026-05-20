import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, it, expect, vi } from 'vitest'
import { LeaderboardPage } from './LeaderboardPage'

vi.mock('../components/leaderboard/GlobalLeaderboard', () => ({
  GlobalLeaderboard: ({ page }: { page: number }) => (
    <div data-testid="global-leaderboard" data-page={page} />
  ),
}))

vi.mock('../components/leaderboard/CityLeaderboard', () => ({
  CityLeaderboard: ({ city, page }: { city: string; page: number }) => (
    <div data-testid="city-leaderboard" data-city={city} data-page={page} />
  ),
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => {
      if (opts) return `${key}:${JSON.stringify(opts)}`
      return key
    },
    i18n: { language: 'en', changeLanguage: vi.fn() },
  }),
}))

function renderLeaderboard(initialPath = '/leaderboard') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <LeaderboardPage />
    </MemoryRouter>
  )
}

describe('LeaderboardPage', () => {
  it('renders 4 tab buttons (global + 3 cities)', () => {
    renderLeaderboard()
    const tabs = screen.getAllByRole('tab')
    expect(tabs).toHaveLength(4)
  })

  it('global tab is selected by default (aria-selected=true)', () => {
    renderLeaderboard()
    const globalTab = screen.getAllByRole('tab')[0]
    expect(globalTab).toHaveAttribute('aria-selected', 'true')
  })

  it('GlobalLeaderboard is rendered by default', () => {
    renderLeaderboard()
    expect(screen.getByTestId('global-leaderboard')).toBeInTheDocument()
  })

  it('clicking a city tab renders CityLeaderboard with correct city prop', () => {
    renderLeaderboard()
    const almaty = screen.getAllByRole('tab')[1]
    fireEvent.click(almaty)
    const cityBoard = screen.getByTestId('city-leaderboard')
    expect(cityBoard).toBeInTheDocument()
    expect(cityBoard).toHaveAttribute('data-city', 'Almaty')
  })

  it('clicking a city tab sets aria-selected on that tab', () => {
    renderLeaderboard()
    const almaty = screen.getAllByRole('tab')[1]
    fireEvent.click(almaty)
    expect(almaty).toHaveAttribute('aria-selected', 'true')
    expect(screen.getAllByRole('tab')[0]).toHaveAttribute('aria-selected', 'false')
  })

  it('next page button increments page prop passed to GlobalLeaderboard', () => {
    renderLeaderboard()
    const globalBoard = screen.getByTestId('global-leaderboard')
    expect(globalBoard).toHaveAttribute('data-page', '1')

    const nextBtn = screen.getByText('→')
    fireEvent.click(nextBtn)

    expect(screen.getByTestId('global-leaderboard')).toHaveAttribute('data-page', '2')
  })

  it('prev page button is disabled when page=1', () => {
    renderLeaderboard()
    const prevBtn = screen.getByText('←')
    expect(prevBtn).toBeDisabled()
  })

  it('prev page button is enabled when page>1', () => {
    renderLeaderboard('/leaderboard?page=2')
    const prevBtn = screen.getByText('←')
    expect(prevBtn).not.toBeDisabled()
  })
})
