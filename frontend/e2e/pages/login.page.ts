import { type Locator, type Page, expect } from '@playwright/test'

/**
 * Page Object for the Login page.
 * Encapsulates all login page interactions for cleaner tests.
 */
export class LoginPage {
  readonly page: Page
  readonly emailInput: Locator
  readonly passwordInput: Locator
  readonly loginButton: Locator
  readonly forgotPasswordLink: Locator
  readonly registerLink: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    this.page = page
    // Use native id attribute for more reliable targeting with PrimeVue
    this.emailInput = page.locator('#email')
    this.passwordInput = page.locator('#password')
    this.loginButton = page.locator('[data-testid="login-button"]')
    this.forgotPasswordLink = page.locator('a[href*="forgot-password"]')
    this.registerLink = page.locator('a[href*="registration"]')
    this.errorMessage = page.locator('.p-toast-message-error, [data-testid="error-message"]')
  }

  async goto(): Promise<void> {
    await this.page.goto('/auth')
  }

  async login(email: string, password: string): Promise<void> {
    await this.emailInput.fill(email)
    await this.passwordInput.fill(password)
    await this.loginButton.click()
  }

  async expectLoginSuccess(): Promise<void> {
    await expect(this.page).toHaveURL('/', { timeout: 10000 })
  }

  async expectLoginError(): Promise<void> {
    await expect(this.errorMessage).toBeVisible({ timeout: 5000 })
  }

  async expectToBeOnLoginPage(): Promise<void> {
    await expect(this.page).toHaveURL('/auth')
    await expect(this.page.getByRole('heading', { name: 'Login' })).toBeVisible()
  }
}
