# Frontend Guide

## Technology Stack

- Nuxt 4 (Vue 3)
- PrimeVue components
- Tailwind CSS
- VeeValidate + Yup for forms

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
