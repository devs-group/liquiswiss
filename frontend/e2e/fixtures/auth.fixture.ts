import { test as base, expect, type Page } from '@playwright/test'

// Test credentials - should match E2E test user seeded via dynamic migration
// See: backend/internal/db/migrations/dynamic/10001_apply_e2e_test_fixtures.sql
const TEST_EMAIL = process.env.E2E_TEST_EMAIL || 'e2e@test.liquiswiss.ch'
const TEST_PASSWORD = process.env.E2E_TEST_PASSWORD || 'Test123!'

interface AuthFixtures {
  authenticatedPage: Page
}

/**
 * Extended test with authenticated page fixture.
 * Use this when you need a logged-in user for your tests.
 */
export const test = base.extend<AuthFixtures>({
  authenticatedPage: async ({ page }, use) => {
    // Navigate to login page
    await page.goto('/auth')

    // Fill in credentials
    await page.locator('[data-testid="email-input"]').fill(TEST_EMAIL)
    await page.locator('[data-testid="password-input"]').fill(TEST_PASSWORD)

    // Click login button
    await page.locator('[data-testid="login-button"]').click()

    // Wait for redirect to home page
    await expect(page).toHaveURL('/', { timeout: 10000 })

    // Provide the authenticated page to the test
    await use(page)
  },
})

export { expect }

/**
 * Helper to login programmatically via API (faster than UI login)
 */
export async function loginViaAPI(page: Page, email = TEST_EMAIL, password = TEST_PASSWORD): Promise<void> {
  const apiHost = process.env.NUXT_API_HOST || 'http://localhost:8080'

  // Make login request to get cookies
  const response = await page.request.post(`${apiHost}/api/auth/login`, {
    data: { email, password },
  })

  if (!response.ok()) {
    throw new Error(`Login failed: ${response.status()} ${await response.text()}`)
  }
}
