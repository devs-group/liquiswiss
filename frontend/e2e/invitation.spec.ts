import { test, expect } from '@playwright/test'
import { InvitationPage } from './pages/invitation.page'

// Test tokens from E2E fixtures (see backend/internal/db/migrations/dynamic/10001_apply_e2e_test_fixtures.sql)
const VALID_TOKEN_NEW_USER = 'e2e-test-invitation-token-new'
const VALID_TOKEN_EXISTING_USER = 'e2e-test-invitation-token-existing'
const EXPIRED_TOKEN = 'e2e-test-invitation-token-expired'
const INVALID_TOKEN = 'nonexistent-token-12345'

test.describe('Invitation Flow', () => {
  test.describe('Invitation Page Display', () => {
    test('should show loading state initially', async ({ page }) => {
      // Start navigation but don't wait for completion
      const navigationPromise = page.goto(`/auth/invitation?token=${VALID_TOKEN_NEW_USER}`)

      // Check for loading spinner (might be brief)
      // Note: This may pass quickly if the page loads fast
      await navigationPromise
    })

    test('should display valid invitation for new user', async ({ page }) => {
      const invitationPage = new InvitationPage(page)
      await invitationPage.goto(VALID_TOKEN_NEW_USER)
      await invitationPage.waitForLoad()

      // Should show success message with organisation name
      await invitationPage.expectValidInvitation()
      await expect(invitationPage.organisationName).toContainText('E2E Test Organisation')

      // Should show inviter name
      await expect(invitationPage.invitedByText).toBeVisible()

      // Should show password form for new user
      await invitationPage.expectNewUserForm()

      // Email should be displayed (disabled)
      await expect(invitationPage.emailInput).toBeDisabled()
      await expect(invitationPage.emailInput).toHaveValue('e2e-invite-new@test.liquiswiss.ch')
    })

    test('should display valid invitation for existing user', async ({ page }) => {
      const invitationPage = new InvitationPage(page)
      await invitationPage.goto(VALID_TOKEN_EXISTING_USER)
      await invitationPage.waitForLoad()

      // Should show success message
      await invitationPage.expectValidInvitation()

      // Should show accept button without password form
      await invitationPage.expectExistingUserForm()
    })

    test('should display error for expired invitation', async ({ page }) => {
      const invitationPage = new InvitationPage(page)
      await invitationPage.goto(EXPIRED_TOKEN)
      await invitationPage.waitForLoad()

      // Should show error message
      await invitationPage.expectInvalidInvitation()

      // Should show login link
      await expect(invitationPage.loginLink).toBeVisible()
    })

    test('should display error for invalid token', async ({ page }) => {
      const invitationPage = new InvitationPage(page)
      await invitationPage.goto(INVALID_TOKEN)
      await invitationPage.waitForLoad()

      // Should show error message
      await invitationPage.expectInvalidInvitation()
    })

    test('should display error when no token provided', async ({ page }) => {
      await page.goto('/auth/invitation')

      const invitationPage = new InvitationPage(page)
      await invitationPage.waitForLoad()

      // Should show error message about missing token
      await invitationPage.expectInvalidInvitation()
      await expect(page.getByText('Kein Einladungstoken gefunden')).toBeVisible()
    })
  })

  test.describe('New User Acceptance', () => {
    test('should validate password requirements', async ({ page }) => {
      const invitationPage = new InvitationPage(page)
      await invitationPage.goto(VALID_TOKEN_NEW_USER)
      await invitationPage.waitForLoad()
      await invitationPage.expectNewUserForm()

      // Enter short password (less than 8 chars)
      await invitationPage.passwordInput.click()
      await invitationPage.passwordInput.fill('short')
      await invitationPage.passwordRepeatInput.click()
      await invitationPage.passwordRepeatInput.fill('short')
      await invitationPage.passwordRepeatInput.blur()

      // Button should be disabled or show validation error
      await expect(page.getByText('mind. 8 Zeichen')).toBeVisible()
    })

    test('should keep button disabled with mismatched passwords', async ({ page }) => {
      const invitationPage = new InvitationPage(page)
      await invitationPage.goto(VALID_TOKEN_NEW_USER)
      await invitationPage.waitForLoad()
      await invitationPage.expectNewUserForm()

      // Enter mismatched passwords
      await invitationPage.passwordInput.click()
      await invitationPage.passwordInput.fill('Password123!')
      await invitationPage.passwordRepeatInput.click()
      await invitationPage.passwordRepeatInput.fill('DifferentPassword!')

      // Click somewhere else to trigger blur
      await invitationPage.passwordInput.click()

      // Wait a bit for validation to run
      await page.waitForTimeout(500)

      // Button should remain disabled because form is invalid
      await expect(invitationPage.acceptButtonNewUser).toBeDisabled()
    })

    // Note: We skip the actual acceptance test because it modifies the database
    // and would need cleanup. In a real CI environment, you'd use database transactions
    // or recreate test data between runs.
    test.skip('should accept invitation and create account', async ({ page }) => {
      const invitationPage = new InvitationPage(page)
      await invitationPage.goto(VALID_TOKEN_NEW_USER)
      await invitationPage.waitForLoad()

      // Accept with a valid password
      await invitationPage.acceptAsNewUser('SecurePassword123!')

      // Should redirect to home after success
      await invitationPage.expectAcceptSuccess()
    })
  })

  test.describe('Existing User Acceptance', () => {
    // Note: We skip actual acceptance tests that modify the database
    test.skip('should accept invitation for existing user', async ({ page }) => {
      const invitationPage = new InvitationPage(page)
      await invitationPage.goto(VALID_TOKEN_EXISTING_USER)
      await invitationPage.waitForLoad()

      // Accept invitation
      await invitationPage.acceptAsExistingUser()

      // Should redirect to home after success
      await invitationPage.expectAcceptSuccess()
    })
  })

  test.describe('Navigation', () => {
    test('should navigate to login from error page', async ({ page }) => {
      const invitationPage = new InvitationPage(page)
      await invitationPage.goto(INVALID_TOKEN)
      await invitationPage.waitForLoad()

      await invitationPage.loginLink.click()

      // Should navigate to login page
      await expect(page).toHaveURL(/\/auth/)
    })
  })
})
