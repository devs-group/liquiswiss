import { type Locator, type Page, expect } from '@playwright/test'

/**
 * Page Object for the Invitation acceptance page.
 * Handles invitation checking and acceptance for both new and existing users.
 */
export class InvitationPage {
  readonly page: Page
  readonly loadingSpinner: Locator
  readonly successMessage: Locator
  readonly errorMessage: Locator
  readonly organisationName: Locator
  readonly invitedByText: Locator
  readonly emailInput: Locator
  readonly passwordInput: Locator
  readonly passwordRepeatInput: Locator
  readonly acceptButtonNewUser: Locator
  readonly acceptButtonExistingUser: Locator
  readonly loginLink: Locator

  constructor(page: Page) {
    this.page = page
    this.loadingSpinner = page.locator('.p-progress-spinner')
    this.successMessage = page.locator('.p-message-success')
    this.errorMessage = page.locator('.p-message-error')
    this.organisationName = page.locator('.p-message-success strong')
    this.invitedByText = page.getByText('Eingeladen von:')
    this.emailInput = page.locator('#email')
    this.passwordInput = page.locator('#password')
    this.passwordRepeatInput = page.locator('#passwordRepeat')
    this.acceptButtonNewUser = page.getByRole('button', { name: 'Konto erstellen & beitreten' })
    this.acceptButtonExistingUser = page.getByRole('button', { name: 'Einladung annehmen' })
    this.loginLink = page.getByRole('link', { name: 'Zum Login' })
  }

  async goto(token: string): Promise<void> {
    await this.page.goto(`/auth/invitation?token=${token}`)
  }

  async waitForLoad(): Promise<void> {
    // Wait for loading spinner to disappear
    await expect(this.loadingSpinner).not.toBeVisible({ timeout: 10000 })
  }

  async expectValidInvitation(): Promise<void> {
    await this.waitForLoad()
    await expect(this.successMessage).toBeVisible()
  }

  async expectInvalidInvitation(): Promise<void> {
    await this.waitForLoad()
    await expect(this.errorMessage).toBeVisible()
  }

  async expectNewUserForm(): Promise<void> {
    await expect(this.passwordInput).toBeVisible()
    await expect(this.passwordRepeatInput).toBeVisible()
    await expect(this.acceptButtonNewUser).toBeVisible()
  }

  async expectExistingUserForm(): Promise<void> {
    await expect(this.acceptButtonExistingUser).toBeVisible()
    // Password inputs should not be visible for existing users
    await expect(this.passwordInput).not.toBeVisible()
  }

  async fillNewUserPassword(password: string): Promise<void> {
    await this.passwordInput.click()
    await this.passwordInput.fill(password)
    await this.passwordRepeatInput.click()
    await this.passwordRepeatInput.fill(password)
  }

  async acceptAsNewUser(password: string): Promise<void> {
    await this.fillNewUserPassword(password)
    await expect(this.acceptButtonNewUser).toBeEnabled({ timeout: 5000 })
    await this.acceptButtonNewUser.click()
  }

  async acceptAsExistingUser(): Promise<void> {
    await expect(this.acceptButtonExistingUser).toBeEnabled()
    await this.acceptButtonExistingUser.click()
  }

  async expectAcceptSuccess(): Promise<void> {
    // After successful acceptance, user should be redirected to home
    // and see a success toast
    await expect(this.page.locator('.p-toast-message-success')).toBeVisible({ timeout: 15000 })
    await expect(this.page).toHaveURL('/', { timeout: 15000 })
  }
}
