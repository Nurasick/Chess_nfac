import { test, expect } from '@playwright/test'

const MOCK_USER = { id: 'u1', username: 'testuser', email: 'test@example.com', city: 'Almaty', rating: 1200, games_played: 5 }
const AUTH_RESPONSE = { data: { user: MOCK_USER, access_token: 'tok_access', refresh_token: 'tok_refresh' }, error: null, message: 'OK' }

function mockAuthRoutes(page: import('@playwright/test').Page) {
  page.route('**/auth/register', route =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(AUTH_RESPONSE) })
  )
  page.route('**/auth/login', route =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(AUTH_RESPONSE) })
  )
  page.route('**/auth/logout', route =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ data: null, error: null, message: 'OK' }) })
  )
  page.route('**/users/me', route =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ data: MOCK_USER, error: null, message: 'OK' }) })
  )
}

test('user can register with valid credentials and is redirected', async ({ page }) => {
  await mockAuthRoutes(page)
  await page.goto('/register')

  await page.getByLabel('Имя пользователя').fill('testuser')
  await page.getByLabel('Email').fill('test@example.com')
  await page.getByLabel('Пароль').fill('Password123')
  await page.selectOption('select#city', 'Almaty')
  await page.locator('main').getByRole('button', { name: 'Регистрация' }).click()

  await expect(page).toHaveURL('/')
})

test('user can login with valid credentials and is redirected', async ({ page }) => {
  await mockAuthRoutes(page)
  await page.goto('/login')

  await page.getByLabel('Имя пользователя').fill('testuser')
  await page.getByLabel('Пароль').fill('password123')
  await page.locator('main').getByRole('button', { name: 'Войти' }).click()

  await expect(page).toHaveURL('/')
})

test('login with wrong password shows error message', async ({ page }) => {
  page.route('**/auth/login', route =>
    route.fulfill({
      status: 401,
      contentType: 'application/json',
      body: JSON.stringify({ data: null, error: 'invalid_credentials', message: 'Invalid credentials' }),
    })
  )
  await page.goto('/login')

  await page.getByLabel('Имя пользователя').fill('testuser')
  await page.getByLabel('Пароль').fill('wrongpassword')
  await page.locator('main').getByRole('button', { name: 'Войти' }).click()

  await expect(page.locator('.rootError, [class*="rootError"]')).toBeVisible()
})

test('logout clears session and redirects to /login', async ({ page }) => {
  await mockAuthRoutes(page)

  await page.addInitScript((user) => {
    localStorage.setItem('access_token', 'tok_access')
    localStorage.setItem('refresh_token', 'tok_refresh')
    localStorage.setItem('user', JSON.stringify(user))
  }, MOCK_USER)

  await page.goto('/')
  await expect(page).toHaveURL('/')

  await page.getByRole('button', { name: MOCK_USER.username[0].toUpperCase() + MOCK_USER.username.slice(1), exact: false }).first().click()
  await page.getByRole('menuitem', { name: /выйти/i }).click()

  await expect(page).toHaveURL('/login')
})
