import { test, expect } from '@playwright/test'
import { LoginPage } from './pages/login.page'
import { OrganisationPage, SettingsOrganisationsPage } from './pages/organisation-settings.page'

// E2E test organisation from fixtures (see backend/internal/db/migrations/dynamic/10001_apply_e2e_test_fixtures.sql)
const E2E_ORG_NAME = 'E2E Test Organisation'

test.describe('Organisation Page', () => {
  // Helper to log in before each test
  async function loginAsTestUser(page: import('@playwright/test').Page) {
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
  }

  test.describe('Main Organisation Page', () => {
    test('should display both General and Members sections on one page', async ({ page }) => {
      await loginAsTestUser(page)

      const orgPage = new OrganisationPage(page)
      await orgPage.goto()
      await orgPage.expectPageLoaded()

      // Both panels should be visible
      await expect(orgPage.generalSettingsPanel).toBeVisible()
      await expect(orgPage.membersPanel).toBeVisible()
    })

    test('should display organisation name in page header', async ({ page }) => {
      await loginAsTestUser(page)

      const orgPage = new OrganisationPage(page)
      await orgPage.goto()
      await orgPage.waitForLoad()

      const orgName = await orgPage.getOrganisationName()
      expect(orgName.trim()).toBe(E2E_ORG_NAME)
    })

    test('should load organisation data with SSR - no infinite spinner', async ({ page }) => {
      await loginAsTestUser(page)

      const orgPage = new OrganisationPage(page)
      await orgPage.goto()

      // Key test: Page should load without infinite spinner
      await orgPage.expectPageLoaded()

      // Name input should have the organisation name
      await expect(orgPage.nameInput).toHaveValue(E2E_ORG_NAME)

      // Should have at least one member (the E2E test user)
      const memberCount = await orgPage.getMemberCount()
      expect(memberCount).toBeGreaterThan(0)
    })

    test('should have save button disabled when form is unchanged', async ({ page }) => {
      await loginAsTestUser(page)

      const orgPage = new OrganisationPage(page)
      await orgPage.goto()
      await orgPage.expectPageLoaded()

      // Save button should be disabled (form not dirty)
      await expect(orgPage.saveButton).toBeDisabled()
    })

    test('should show invite button for owner', async ({ page }) => {
      await loginAsTestUser(page)

      const orgPage = new OrganisationPage(page)
      await orgPage.goto()
      await orgPage.waitForLoad()

      // Owner should see the invite button
      await orgPage.expectInviteButtonVisible()
    })

    test('should display pending invitations when they exist', async ({ page }) => {
      await loginAsTestUser(page)

      const orgPage = new OrganisationPage(page)
      await orgPage.goto()
      await orgPage.waitForLoad()

      // E2E fixtures include pending invitations
      await orgPage.expectInvitationsSection()

      // Should have at least one invitation (from fixtures)
      const invitationCount = await orgPage.getInvitationCount()
      expect(invitationCount).toBeGreaterThan(0)
    })

    test('should display member cards with member information', async ({ page }) => {
      await loginAsTestUser(page)

      const orgPage = new OrganisationPage(page)
      await orgPage.goto()
      await orgPage.waitForLoad()

      // Should display the E2E test user's email in a member card
      await expect(orgPage.memberCards.first()).toBeVisible()

      // E2E test user should be visible
      await expect(page.getByText('e2e@test.liquiswiss.ch')).toBeVisible()
    })
  })

  test.describe('Settings Organisations Page', () => {
    test('should display organisation cards with switch buttons', async ({ page }) => {
      await loginAsTestUser(page)

      const settingsPage = new SettingsOrganisationsPage(page)
      await settingsPage.goto()
      await settingsPage.waitForLoad()

      // Should have at least one organisation
      const orgCount = await settingsPage.getOrganisationCount()
      expect(orgCount).toBeGreaterThan(0)
    })

    test('should show add organisation button', async ({ page }) => {
      await loginAsTestUser(page)

      const settingsPage = new SettingsOrganisationsPage(page)
      await settingsPage.goto()
      await settingsPage.waitForLoad()

      await settingsPage.expectAddButtonVisible()
    })

    test('should display organisation name and role', async ({ page }) => {
      await loginAsTestUser(page)

      const settingsPage = new SettingsOrganisationsPage(page)
      await settingsPage.goto()
      await settingsPage.waitForLoad()

      // Should show organisation name in the card
      await expect(settingsPage.organisationCards.first().getByText(E2E_ORG_NAME)).toBeVisible()

      // Should show role (Eigentümer for owner)
      await expect(page.getByText('Eigentümer')).toBeVisible()
    })

    test('should show Active tag for current organisation', async ({ page }) => {
      await loginAsTestUser(page)

      const settingsPage = new SettingsOrganisationsPage(page)
      await settingsPage.goto()
      await settingsPage.waitForLoad()

      // Current organisation should show "Aktiv" tag
      await expect(page.getByText('Aktiv')).toBeVisible()
    })
  })

  test.describe('Navigation', () => {
    test('should navigate to organisation page from main menu', async ({ page }) => {
      await loginAsTestUser(page)

      // Click Organisation in main menu
      await page.getByRole('link', { name: 'Organisation' }).click()

      // Should be on organisation page
      await expect(page).toHaveURL('/organisation')

      const orgPage = new OrganisationPage(page)
      await orgPage.expectPageLoaded()
    })
  })
})
