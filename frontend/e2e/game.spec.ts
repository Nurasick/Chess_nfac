import { test, expect } from '@playwright/test'

const MOCK_USER = { id: 'u1', username: 'testuser', email: 'test@example.com', city: 'Almaty', rating: 1200, games_played: 5 }

test.beforeEach(async ({ page }) => {
  // Inject auth state so AuthGuard passes
  await page.addInitScript((user) => {
    localStorage.setItem('access_token', 'tok_access')
    localStorage.setItem('refresh_token', 'tok_refresh')
    localStorage.setItem('user', JSON.stringify(user))
  }, MOCK_USER)
})

test('authenticated user can click Join Queue button and queue status appears', async ({ page }) => {
  await page.goto('/')
  await expect(page).toHaveURL('/')

  // Find and click the Join Queue button
  const joinBtn = page.getByRole('button', { name: /найти партию/i })
  await expect(joinBtn).toBeVisible()
  await joinBtn.click()

  // QueueWaiting should render with "Поиск соперника..."
  await expect(page.getByText(/поиск соперника/i)).toBeVisible()
})

test('cancel queue button returns to home screen', async ({ page }) => {
  await page.goto('/')

  await page.getByRole('button', { name: /найти партию/i }).click()
  await expect(page.getByText(/поиск соперника/i)).toBeVisible()

  await page.getByRole('button', { name: /отмена/i }).click()
  await expect(page.getByRole('button', { name: /найти партию/i })).toBeVisible()
})
