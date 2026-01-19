# E2E Test Plan

## Current Coverage

**Existing tests** (`frontend/e2e/auth.spec.ts`): 8 tests
- Login page elements display
- Button disabled/enabled states
- Navigation to forgot password/registration
- Invalid credentials error handling
- Successful login redirect
- Protected routes redirect unauthenticated users

## Proposed Test Structure

```
frontend/e2e/
├── auth.spec.ts              ✓ exists
├── forecast.spec.ts          ○ new
├── employees.spec.ts         ○ new
├── transactions.spec.ts      ○ new
├── bank-accounts.spec.ts     ○ new
├── settings.spec.ts          ○ new
├── pages/
│   ├── login.page.ts         ✓ exists
│   ├── forecast.page.ts      ○ new
│   ├── employees.page.ts     ○ new
│   ├── transactions.page.ts  ○ new
│   ├── bank-accounts.page.ts ○ new
│   └── settings.page.ts      ○ new
└── fixtures/
    └── auth.fixture.ts       ✓ exists
```

## Test Specifications

### 1. Forecast Dashboard (`forecast.spec.ts`)

**Priority: High** - Core business feature

| Test | Description |
|------|-------------|
| Display forecast data | Verify forecast loads with revenue, expenses, cashflow |
| Adjust forecast period | Change months slider, verify data updates |
| Adjust performance | Change performance slider, verify calculations update |
| Recalculate forecast | Click recalculate button, verify new data loads |
| Expand revenue details | Toggle revenue expansion, verify categories shown |
| Expand expense details | Toggle expense expansion, verify categories shown |
| Chart display | Verify chart renders with data |

### 2. Employees (`employees.spec.ts`)

**Priority: High** - Core CRUD functionality

| Test | Description |
|------|-------------|
| Display employee list | Verify employees load in grid/table view |
| Search employees | Enter search term, verify filtered results |
| Toggle grid/table view | Switch views, verify display changes |
| Hide terminated employees | Toggle filter, verify list updates |
| Add employee | Open dialog, fill form, save, verify appears in list |
| Edit employee | Navigate to detail, edit fields, save |
| Delete employee | Delete employee, verify removed from list |
| Pagination | Load more employees, verify list extends |

### 3. Transactions (`transactions.spec.ts`)

**Priority: High** - Core CRUD functionality

| Test | Description |
|------|-------------|
| Display transaction list | Verify transactions load in grid/table view |
| Search transactions | Enter search term, verify filtered results |
| Toggle grid/table view | Switch views, verify display changes |
| Hide disabled transactions | Toggle filter, verify list updates |
| Add transaction | Open dialog, fill form, save, verify appears in list |
| Edit transaction | Edit fields, save, verify changes |
| Clone transaction | Clone existing, verify copy created |
| Delete transaction | Delete transaction, verify removed from list |

### 4. Bank Accounts (`bank-accounts.spec.ts`)

**Priority: Medium** - Supporting feature

| Test | Description |
|------|-------------|
| Display bank account list | Verify accounts load with balances |
| Total balance display | Verify total in organisation currency shown |
| Search bank accounts | Enter search term, verify filtered results |
| Toggle grid/table view | Switch views, verify display changes |
| Add bank account | Open dialog, fill form with currency, save |
| Edit bank account | Edit fields, save, verify changes |
| Clone bank account | Clone existing, verify copy created |
| Delete bank account | Delete account, verify removed from list |

### 5. Settings (`settings.spec.ts`)

**Priority: Medium** - User preferences and configuration

| Test | Description |
|------|-------------|
| **Profile** | |
| Update name | Change name, save, verify success message |
| Email read-only | Verify email field is not editable |
| Change password | Enter current/new password, save |
| **Organisations** | |
| List organisations | Verify user's organisations displayed |
| Create organisation | Open dialog, fill form, create |
| Edit organisation | Navigate to edit, change name, save |
| Delete organisation | Delete organisation, verify removed |
| **App Settings** | |
| Toggle skip org switch | Enable/disable, verify persists |
| Change color mode | Switch dark/light/system, verify applies |
| **Automation** | |
| Enable VAT settlement | Toggle VAT, verify form appears |
| Set VAT billing date | Select date, save |
| Enter company number | Fill field, save |

## Implementation Priority

### Phase 1: Critical Path (Recommended First)
1. `forecast.spec.ts` - Core business value
2. `employees.spec.ts` - Primary data management

### Phase 2: Data Management
3. `transactions.spec.ts` - Financial data
4. `bank-accounts.spec.ts` - Account management

### Phase 3: Configuration
5. `settings.spec.ts` - User preferences

## Test Data Requirements

Current E2E test user (`e2e@test.liquiswiss.ch`) needs:
- [ ] At least 1 employee with salary
- [ ] At least 1 transaction (positive and negative)
- [ ] At least 1 bank account with balance
- [ ] Forecast data populated

**Option A**: Extend dynamic migration `10001_apply_e2e_test_fixtures.sql`
**Option B**: Create test data via API in test setup/beforeAll

## Shared Patterns

### Page Object Pattern
Each page object should include:
- Locators for key elements (use `#id` over `[data-testid]`)
- `goto()` method with `waitForLoadState('networkidle')`
- Common actions (fill form, submit, etc.)

### Form Interaction Pattern
```typescript
// Click before fill for PrimeVue + vee-validate
await input.click()
await input.fill('value')
await expect(input).toHaveValue('value')
```

### Authenticated Tests
```typescript
import { test } from '../fixtures/auth.fixture'

test('should do something', async ({ authenticatedPage }) => {
  // authenticatedPage is already logged in
})
```

## Definition of Done

- [ ] All tests pass locally
- [ ] All tests pass in CI (GitHub Actions)
- [ ] Page objects created for each module
- [ ] Test data seeded via migration or API
- [ ] Documentation updated in `docs/ai/frontend.md`
