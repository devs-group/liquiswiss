import { test, expect, type APIRequestContext, type Page } from '@playwright/test'
import { LoginPage } from './pages/login.page'

// API requests use relative paths so they go through the Nuxt proxy and follow the
// configured baseURL, which differs between the local docker stack and CI.

// Names are prefixed so they sort to the end and cannot collide with real data
const PAID_EMPLOYEE = 'ZZZ E2E Paid'
const ZERO_EMPLOYEE = 'ZZZ E2E Zero'
const SALARY_CATEGORY = 'Löhne'

async function loginAsTestUser(page: Page) {
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

async function createEmployee(request: APIRequestContext, name: string): Promise<number> {
  const response = await request.post(`/api/employees`, { data: { name } })
  expect(response.ok(), `creating employee ${name}: ${response.status()}`).toBeTruthy()
  return (await response.json()).id
}

function firstOfCurrentMonth(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-01`
}

async function createSalary(
  request: APIRequestContext,
  employeeID: number,
  amount: number,
  fromDate: string,
  isTermination = false,
) {
  const response = await request.post(`/api/employees/${employeeID}/salary`, {
    data: {
      hoursPerMonth: isTermination ? 0 : 160,
      amount,
      cycle: 'monthly',
      currencyID: 1,
      vacationDaysPerYear: 25,
      fromDate,
      toDate: null,
      isTermination,
    },
  })
  expect(response.ok(), `creating salary for ${employeeID}: ${response.status()}`).toBeTruthy()
}

async function calculateForecast(request: APIRequestContext) {
  const response = await request.get(`/api/forecasts/calculate`)
  expect(response.ok(), `calculating forecast: ${response.status()}`).toBeTruthy()
}

async function deleteEmployee(request: APIRequestContext, employeeID: number) {
  await request.delete(`/api/employees/${employeeID}`)
}

async function gotoForecast(page: Page) {
  await page.goto('/')
  // The dev server keeps a HMR socket open, so networkidle never settles here
  await expect(page.getByText('Zeitraum:')).toBeVisible({ timeout: 30000 })
}

/**
 * Clicks the zero row toggle until the expected label sticks. The organisation settings
 * are loaded asynchronously after hydration and can reset a click that happens too early.
 */
async function toggleZeroRows(page: Page, currentLabel: string, targetLabel: string) {
  for (let attempt = 0; attempt < 3; attempt++) {
    await expect(page.getByRole('button', { name: currentLabel })).toBeVisible({ timeout: 15000 })

    try {
      // Waiting for the PATCH matters: a reload right after the click would otherwise
      // race the save, and the setting would silently stay on its old value.
      await Promise.all([
        page.waitForResponse(
          response =>
            response.url().includes('/api/user-organisation-settings')
            && response.request().method() === 'PATCH'
            && response.ok(),
          { timeout: 5000 },
        ),
        page.getByRole('button', { name: currentLabel }).click(),
      ])

      await expect(page.getByRole('button', { name: targetLabel })).toBeVisible()
      return
    }
    catch {
      // The click landed before the settings were loaded, so it was reverted. Retry.
    }
  }
  throw new Error(`Toggle did not switch from "${currentLabel}" to "${targetLabel}"`)
}

/**
 * Expanding the expense section and the "Löhne" category is a persisted setting, so it is
 * set through the API instead of by clicking. Clicking would toggle whatever state the
 * previous run left behind.
 */
async function expandSalaryRows(request: APIRequestContext, showZeroRows = false) {
  const response = await request.patch(`/api/user-organisation-settings`, {
    data: {
      forecastExpenseDetails: true,
      forecastChildDetails: [SALARY_CATEGORY],
      forecastShowZeroRows: showZeroRows,
    },
  })
  expect(response.ok(), `updating forecast settings: ${response.status()}`).toBeTruthy()
}

async function expectSalaryRowsVisible(page: Page) {
  await expect(page.getByText(SALARY_CATEGORY, { exact: true })).toBeVisible({ timeout: 15000 })
  await expect(page.getByText(PAID_EMPLOYEE, { exact: true })).toBeVisible({ timeout: 15000 })
}

test.describe('Prognose: Zeilen ohne Beträge', () => {
  let paidID: number
  let zeroID: number

  test.beforeEach(async ({ page }) => {
    await loginAsTestUser(page)

    paidID = await createEmployee(page.request, PAID_EMPLOYEE)
    zeroID = await createEmployee(page.request, ZERO_EMPLOYEE)

    await createSalary(page.request, paidID, 500000, firstOfCurrentMonth())

    // A terminated employee keeps its row in the forecast but shows 0.00 in every
    // month from now on, which is exactly the case the filter has to hide
    await createSalary(page.request, zeroID, 300000, `${new Date().getFullYear()}-01-01`)
    await createSalary(page.request, zeroID, 0, firstOfCurrentMonth(), true)

    await calculateForecast(page.request)
    await expandSalaryRows(page.request)
  })

  test.afterEach(async ({ page }) => {
    await deleteEmployee(page.request, paidID)
    await deleteEmployee(page.request, zeroID)
    await calculateForecast(page.request)
  })

  test('hides employees without amounts by default and shows them after toggling', async ({ page }) => {
    await gotoForecast(page)

    await expectSalaryRowsVisible(page)

    // Default: the permanent zero row is gone, the paid one stays
    await expect(page.getByText(ZERO_EMPLOYEE, { exact: true })).toHaveCount(0)

    // Toggle makes it visible again
    await toggleZeroRows(page, 'Nullzeilen ausgeblendet', 'Nullzeilen sichtbar')
    await expect(page.getByText(ZERO_EMPLOYEE, { exact: true })).toBeVisible()
    await expect(page.getByText(PAID_EMPLOYEE, { exact: true })).toBeVisible()

    // And hides it again
    await toggleZeroRows(page, 'Nullzeilen sichtbar', 'Nullzeilen ausgeblendet')
    await expect(page.getByText(ZERO_EMPLOYEE, { exact: true })).toHaveCount(0)
  })
})
