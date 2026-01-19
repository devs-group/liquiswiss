import { test, expect } from '@playwright/test'
import { LoginPage } from './pages/login.page'

test.describe('Authentication', () => {
  test.describe('Login Page', () => {
    test('should display login page with all elements', async ({ page }) => {
      const loginPage = new LoginPage(page)
      await loginPage.goto()

      await expect(page.getByRole('heading', { name: 'Login' })).toBeVisible()
      await expect(loginPage.emailInput).toBeVisible()
      await expect(loginPage.passwordInput).toBeVisible()
      await expect(loginPage.loginButton).toBeVisible()
      await expect(loginPage.forgotPasswordLink).toBeVisible()
      await expect(loginPage.registerLink).toBeVisible()
    })

    test('should have login button disabled when form is empty', async ({ page }) => {
      const loginPage = new LoginPage(page)
      await loginPage.goto()

      // Button should be disabled initially (empty form)
      await expect(loginPage.loginButton).toBeDisabled()
    })

    test('should have login button disabled with only email filled', async ({ page }) => {
      const loginPage = new LoginPage(page)
      await loginPage.goto()

      // Fill only email - button should still be disabled
      await loginPage.emailInput.fill('test@example.com')
      // Need to blur to trigger validation
      await loginPage.emailInput.blur()

      await expect(loginPage.loginButton).toBeDisabled()
    })

    test('should enable login button when form is valid and complete', async ({ page }) => {
      const loginPage = new LoginPage(page)
      await loginPage.goto()

      // Wait for page to be fully loaded
      await page.waitForLoadState('networkidle')

      // Wait for hydration - the input should be interactive
      await expect(loginPage.emailInput).toBeEditable()

      // Fill form fields - click first to focus
      await loginPage.emailInput.click()
      await loginPage.emailInput.fill('test@example.com')
      await loginPage.passwordInput.click()
      await loginPage.passwordInput.fill('password123')

      // Verify values were entered
      await expect(loginPage.emailInput).toHaveValue('test@example.com')
      await expect(loginPage.passwordInput).toHaveValue('password123')

      // Wait for validation to complete and button to be enabled
      await expect(loginPage.loginButton).toBeEnabled({ timeout: 10000 })
    })

    test('should navigate to forgot password page', async ({ page }) => {
      const loginPage = new LoginPage(page)
      await loginPage.goto()

      await loginPage.forgotPasswordLink.click()
      await expect(page).toHaveURL(/forgot-password/)
    })

    test('should navigate to registration page', async ({ page }) => {
      const loginPage = new LoginPage(page)
      await loginPage.goto()

      await loginPage.registerLink.click()
      await expect(page).toHaveURL(/registration/)
    })
  })

  test.describe('Login Flow', () => {
    test('should show error toast for invalid credentials', async ({ page }) => {
      const loginPage = new LoginPage(page)
      await loginPage.goto()

      // Wait for page to be fully loaded
      await page.waitForLoadState('networkidle')

      // Wait for hydration
      await expect(loginPage.emailInput).toBeEditable()

      // Fill credentials - click first to focus
      await loginPage.emailInput.click()
      await loginPage.emailInput.fill('wrong@example.com')
      await loginPage.passwordInput.click()
      await loginPage.passwordInput.fill('wrongpassword')

      // Verify values were entered
      await expect(loginPage.emailInput).toHaveValue('wrong@example.com')
      await expect(loginPage.passwordInput).toHaveValue('wrongpassword')

      // Wait for button to be enabled
      await expect(loginPage.loginButton).toBeEnabled({ timeout: 10000 })
      await loginPage.loginButton.click()

      // Wait for error toast (German: "Login fehlgeschlagen")
      await expect(page.locator('.p-toast-message-error')).toBeVisible({ timeout: 10000 })
      await expect(page.getByText('Login fehlgeschlagen')).toBeVisible()
    })

    test('should redirect to home page on successful login', async ({ page }) => {
      const loginPage = new LoginPage(page)
      await loginPage.goto()

      // Wait for page to be fully loaded
      await page.waitForLoadState('networkidle')

      // Use E2E test user credentials (seeded via dynamic migration)
      const testEmail = process.env.E2E_TEST_EMAIL || 'e2e@test.liquiswiss.ch'
      const testPassword = process.env.E2E_TEST_PASSWORD || 'Test123!'

      // Fill credentials
      await loginPage.emailInput.click()
      await loginPage.emailInput.fill(testEmail)
      await loginPage.passwordInput.click()
      await loginPage.passwordInput.fill(testPassword)

      // Verify and submit
      await expect(loginPage.emailInput).toHaveValue(testEmail)
      await expect(loginPage.passwordInput).toHaveValue(testPassword)
      await expect(loginPage.loginButton).toBeEnabled({ timeout: 10000 })
      await loginPage.loginButton.click()

      // Should redirect to home page after successful login
      await loginPage.expectLoginSuccess()
    })
  })

  test.describe('Protected Routes', () => {
    test('should redirect unauthenticated users to login without showing session expired message', async ({ page }) => {
      // First-time visitor accessing a protected route should NOT see any "session expired" message
      // They never had a session, so there's nothing to expire

      // Try to access a protected route directly (as first-time visitor with no cookies)
      await page.goto('/employees')

      // Should be redirected to login
      await expect(page).toHaveURL(/\/auth/)

      // Wait for page to be fully loaded
      await page.waitForLoadState('networkidle')

      // No session expired message should appear (neither toast nor dialog)
      const sessionExpiredMessage = page.getByText('Sitzung abgelaufen')
      await expect(sessionExpiredMessage).not.toBeVisible({ timeout: 2000 })
    })

    test('should redirect to login when accessing settings', async ({ page }) => {
      await page.goto('/settings')
      await expect(page).toHaveURL(/\/auth/)
    })

    test('should redirect to login when accessing transactions', async ({ page }) => {
      await page.goto('/transactions')
      await expect(page).toHaveURL(/\/auth/)
    })
  })

  test.describe('Session Expiry', () => {
    // Helper to log in and clear auth cookies
    async function loginAndClearAuthCookies(page: import('@playwright/test').Page) {
      const loginPage = new LoginPage(page)
      await loginPage.goto()
      await page.waitForLoadState('networkidle')

      const testEmail = process.env.E2E_TEST_EMAIL || 'e2e@test.liquiswiss.ch'
      const testPassword = process.env.E2E_TEST_PASSWORD || 'Test123!'

      await loginPage.emailInput.click()
      await loginPage.emailInput.fill(testEmail)
      await loginPage.passwordInput.click()
      await loginPage.passwordInput.fill(testPassword)
      await expect(loginPage.loginButton).toBeEnabled({ timeout: 10000 })
      await loginPage.loginButton.click()

      await loginPage.expectLoginSuccess()

      // Clear only auth cookies, keep hadSession cookie
      const cookies = await page.context().cookies()
      const authCookieNames = ['liq-access-token', 'liq-refresh-token']
      const nonAuthCookies = cookies
        .filter(c => !authCookieNames.includes(c.name))
        .map(c => ({
          ...c,
          // Override secure to false for HTTP testing environment
          secure: false,
        }))

      await page.context().clearCookies()
      if (nonAuthCookies.length > 0) {
        await page.context().addCookies(nonAuthCookies)
      }
    }

    // Note: Client-side navigation with deleted cookies cannot be tested in E2E because
    // auth cookies are HTTP-only (cannot be read by JavaScript). Session expiry is detected
    // when the next API call fails with 401, which triggers the fetchInterceptor to show
    // the session expired dialog. Manual testing:
    // 1. Log in to the app
    // 2. Open browser DevTools and delete liq-access-token and liq-refresh-token cookies
    // 3. Perform any action that makes an API call (e.g., navigate to a page that loads data)
    // 4. The "Sitzung abgelaufen" dialog should appear when the API call fails

    test('should show session expired toast when page is reloaded after session expires', async ({ page }) => {
      // Scenario: User logged in, session expires, user reloads the page
      // Expected: Redirect to login page, toast message appears
      await loginAndClearAuthCookies(page)

      // Reload the page (this is a full page load, not client-side navigation)
      await page.reload()

      // Should be redirected to login
      await expect(page).toHaveURL(/\/auth/)
      await page.waitForLoadState('networkidle')

      // Session expired toast SHOULD appear on login page
      const sessionExpiredToast = page.getByText('Sitzung abgelaufen')
      await expect(sessionExpiredToast).toBeVisible({ timeout: 5000 })
    })

    test('should show session expired toast when loading a protected page after session expires', async ({ page }) => {
      // Scenario: User logged in, session expires, user types URL directly (page load)
      // Expected: Redirect to login page, toast message appears
      await loginAndClearAuthCookies(page)

      // Navigate via URL (full page load)
      await page.goto('/employees')

      // Should be redirected to login
      await expect(page).toHaveURL(/\/auth/)
      await page.waitForLoadState('networkidle')

      // Session expired toast SHOULD appear on login page
      const sessionExpiredToast = page.getByText('Sitzung abgelaufen')
      await expect(sessionExpiredToast).toBeVisible({ timeout: 5000 })
    })
  })
})
