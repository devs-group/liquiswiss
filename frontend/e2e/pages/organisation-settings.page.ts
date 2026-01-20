import { type Locator, type Page, expect } from '@playwright/test'

/**
 * Page Object for the Organisation page.
 * Handles the combined general settings and members sections.
 */
export class OrganisationPage {
  readonly page: Page

  // Page Header
  readonly pageTitle: Locator

  // General Settings Panel
  readonly generalSettingsPanel: Locator
  readonly nameInput: Locator
  readonly currencySelect: Locator
  readonly saveButton: Locator

  // Members Panel
  readonly membersPanel: Locator
  readonly inviteMemberButton: Locator
  readonly membersGrid: Locator
  readonly memberCards: Locator
  readonly invitationsCount: Locator
  readonly invitationsGrid: Locator
  readonly invitationCards: Locator

  constructor(page: Page) {
    this.page = page

    // Page Header
    this.pageTitle = page.locator('h1')

    // General Settings Panel
    this.generalSettingsPanel = page.locator('[data-testid="general-settings-panel"]')
    this.nameInput = page.locator('[data-testid="organisation-name-input"]')
    this.currencySelect = page.locator('[data-testid="organisation-currency-select"]')
    this.saveButton = page.locator('[data-testid="organisation-save-button"]')

    // Members Panel
    this.membersPanel = page.locator('[data-testid="members-panel"]')
    this.inviteMemberButton = page.locator('[data-testid="invite-member-button"]')
    this.membersGrid = page.locator('[data-testid="members-grid"]')
    this.memberCards = page.locator('[data-testid="member-card"]')
    this.invitationsCount = page.locator('[data-testid="invitations-count"]')
    this.invitationsGrid = page.locator('[data-testid="invitations-grid"]')
    this.invitationCards = page.locator('[data-testid="invitation-card"]')
  }

  async goto(): Promise<void> {
    await this.page.goto('/organisation')
  }

  async waitForLoad(): Promise<void> {
    await expect(this.generalSettingsPanel).toBeVisible({ timeout: 10000 })
    await expect(this.membersPanel).toBeVisible({ timeout: 10000 })
  }

  async expectPageLoaded(): Promise<void> {
    await this.waitForLoad()
    await expect(this.nameInput).toBeVisible()
    await expect(this.currencySelect).toBeVisible()
    await expect(this.saveButton).toBeVisible()
  }

  async expectMembersSectionLoaded(): Promise<void> {
    await expect(this.membersPanel).toBeVisible()
  }

  async getOrganisationName(): Promise<string> {
    return await this.pageTitle.textContent() ?? ''
  }

  async getMemberCount(): Promise<number> {
    return await this.memberCards.count()
  }

  async getInvitationCount(): Promise<number> {
    return await this.invitationCards.count()
  }

  async expectInviteButtonVisible(): Promise<void> {
    await expect(this.inviteMemberButton).toBeVisible()
  }

  async expectInviteButtonNotVisible(): Promise<void> {
    await expect(this.inviteMemberButton).not.toBeVisible()
  }

  async expectInvitationsSection(): Promise<void> {
    await expect(this.invitationsGrid).toBeVisible()
  }

  async expectNoInvitationsSection(): Promise<void> {
    await expect(this.invitationsGrid).not.toBeVisible()
  }
}

/**
 * Page Object for the Settings Organisations page.
 * Shows organisation cards with switch functionality.
 */
export class SettingsOrganisationsPage {
  readonly page: Page
  readonly organisationCards: Locator
  readonly addOrganisationButton: Locator
  readonly switchButtons: Locator

  constructor(page: Page) {
    this.page = page
    this.organisationCards = page.locator('[data-testid="organisation-card"]')
    this.addOrganisationButton = page.getByRole('button', { name: 'Organisation hinzufügen' })
    this.switchButtons = page.locator('[data-testid="switch-organisation-button"]')
  }

  async goto(): Promise<void> {
    await this.page.goto('/settings/organisations')
  }

  async waitForLoad(): Promise<void> {
    await expect(this.page.getByText('Ihre Organisationen')).toBeVisible({ timeout: 10000 })
  }

  async getOrganisationCount(): Promise<number> {
    return await this.organisationCards.count()
  }

  async expectAddButtonVisible(): Promise<void> {
    await expect(this.addOrganisationButton).toBeVisible()
  }
}
