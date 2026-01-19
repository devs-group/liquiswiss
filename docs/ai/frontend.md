# Frontend Guide

## Technology Stack

- Nuxt 4 (Vue 3)
- PrimeVue components
- Tailwind CSS
- VeeValidate + Yup for forms
- Node version managed via `.nvmrc` - always run `nvm use` before npm commands

## Key Files to Understand

| Purpose | File |
|---------|------|
| App entry | [app/app.vue](../../frontend/app/app.vue) |
| Auth composable | [app/composables/useAuth.ts](../../frontend/app/composables/useAuth.ts) |
| Global state | [app/composables/useGlobalData.ts](../../frontend/app/composables/useGlobalData.ts) |
| Auth middleware | [app/middleware/auth.global.ts](../../frontend/app/middleware/auth.global.ts) |
| TypeScript models | [app/models/](../../frontend/app/models/) |
| Utility functions | [app/utils/](../../frontend/app/utils/) |
| Theme config | [app/config/](../../frontend/app/config/) |

## Patterns

**Composables**: Each feature has a composable that encapsulates state (`useState`), API calls (`$fetch`/`useFetch`), and computed properties. No Pinia/Vuex needed.

**Global middleware**: `auth.global.ts` runs on every route, protecting pages and auto-refreshing tokens.

**Constants**: Always use `app/utils/constants.ts` as the first choice for adding cookie names, magic strings, and configuration values. This centralizes all constants and makes them easy to find and maintain.

**Cookies over localStorage**: Always use cookies (`useCookie`) instead of localStorage for storing state. localStorage doesn't work with SSR since it's only available on the client side. Cookies work on both server and client.

**Cookie value types**: Nuxt's `useCookie` automatically serializes/deserializes values to JSON. This means:
- `cookie.value = true` → stored as JSON `true` → read back as boolean `true`
- `cookie.value = 'hello'` → stored as JSON `"hello"` → read back as string `'hello'`
- `cookie.value = { foo: 1 }` → stored as JSON object → read back as object

Always use truthy checks (`if (cookie.value)`) instead of string comparisons (`if (cookie.value === 'true')`). Custom `encode`/`decode` options can be passed to `useCookie` if you need different serialization behavior. See [Nuxt useCookie docs](https://nuxt.com/docs/api/composables/use-cookie).

## Composables Overview

| Composable | Purpose |
|------------|---------|
| `useAuth` | Login, logout, token refresh |
| `useEmployees` | Employee CRUD |
| `useSalaries` | Salary management |
| `useSalaryCosts` | Cost calculations |
| `useForecasts` | Forecast data + exclusions |
| `useTransactions` | Transaction CRUD |
| `useBankAccounts` | Bank account management |
| `useVat` / `useVatSettings` | VAT calculations and config |
| `useGlobalData` | Currencies, categories, fiat rates |
| `useCharts` | Chart data preparation |

## E2E Testing (Playwright)

**Location**: `frontend/e2e/`

### Key Patterns

**Wait for SSR hydration**: Nuxt uses SSR, so pages render on the server first and then Vue "hydrates" them on the client. Always wait for hydration before interacting:

```typescript
await page.waitForLoadState('networkidle')
await expect(locator).toBeEditable()
```

**Click before fill**: For PrimeVue inputs with vee-validate, click to focus before filling to ensure proper Vue reactivity:

```typescript
await emailInput.click()
await emailInput.fill('test@example.com')
```

**Use native IDs over data-testid**: PrimeVue components sometimes have complex DOM structures. Use native `id` attributes when available:

```typescript
// Prefer this (more reliable)
this.emailInput = page.locator('#email')

// Over this (can be flaky with component wrappers)
this.emailInput = page.locator('[data-testid="email-input"]')
```

### Running Tests

```bash
npm run test:e2e           # Headless
npm run test:e2e:ui        # Interactive UI mode
npm run test:e2e:headed    # Visible browser
npm run test:e2e:debug     # Debug mode
```

### Test Structure

- `e2e/fixtures/` - Reusable test fixtures (e.g., authenticated user)
- `e2e/pages/` - Page object models
- `e2e/*.spec.ts` - Test files

### E2E Test User

A test user is automatically seeded via dynamic migration:
- **Email**: `e2e@test.liquiswiss.ch`
- **Password**: `Test123!`
- **Migration**: `backend/internal/db/migrations/dynamic/10001_apply_e2e_test_fixtures.sql`

Override credentials with environment variables: `E2E_TEST_EMAIL`, `E2E_TEST_PASSWORD`
