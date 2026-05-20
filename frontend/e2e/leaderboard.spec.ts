import { test, expect } from '@playwright/test'

const GLOBAL_RESPONSE = {
  data: [
    { user_id: 'u1', username: 'alice', rating: 1800, games_played: 100, city: 'Almaty' },
    { user_id: 'u2', username: 'bob', rating: 1700, games_played: 80, city: 'Astana' },
  ],
  error: null,
  message: 'OK',
  total: 2,
  page: 1,
  page_size: 20,
}

const CITY_RESPONSE = {
  data: [
    { user_id: 'u1', username: 'alice', rating: 1800, games_played: 100, city: 'Almaty' },
  ],
  error: null,
  message: 'OK',
  total: 1,
  page: 1,
  page_size: 20,
}

function mockLeaderboard(page: import('@playwright/test').Page) {
  page.route('**/leaderboard/global**', route =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(GLOBAL_RESPONSE) })
  )
  page.route('**/leaderboard/Almaty**', route =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(CITY_RESPONSE) })
  )
  page.route('**/leaderboard/Astana**', route =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ ...GLOBAL_RESPONSE, data: [] }) })
  )
  page.route('**/leaderboard/Shymkent**', route =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ ...GLOBAL_RESPONSE, data: [] }) })
  )
}

test('leaderboard page has 4 tabs visible', async ({ page }) => {
  await mockLeaderboard(page)
  await page.goto('/leaderboard')

  const tabs = page.getByRole('tab')
  await expect(tabs).toHaveCount(4)
})

test('global tab is active by default', async ({ page }) => {
  await mockLeaderboard(page)
  await page.goto('/leaderboard')

  const globalTab = page.getByRole('tab').first()
  await expect(globalTab).toHaveAttribute('aria-selected', 'true')
})

test('clicking Almaty tab shows city leaderboard', async ({ page }) => {
  await mockLeaderboard(page)
  await page.goto('/leaderboard')

  await page.getByRole('tab', { name: 'Алматы' }).click()

  await expect(page.locator('[data-testid="city-leaderboard"]')).toBeVisible()
})

test('prev page button is disabled on page 1 and next page increments', async ({ page }) => {
  await mockLeaderboard(page)
  await page.goto('/leaderboard')

  const prevBtn = page.getByRole('button', { name: '←' })
  await expect(prevBtn).toBeDisabled()

  const nextBtn = page.getByRole('button', { name: '→' })
  await nextBtn.click()

  await expect(page).toHaveURL(/page=2/)
})
