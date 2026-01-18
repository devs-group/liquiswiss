# Cross-User/Cross-Organisation Data Isolation Test Plan

## Overview

This test plan covers comprehensive data isolation testing for the LiquiSwiss application. The application uses a multi-tenant architecture where users can belong to multiple organizations, and all data is scoped to organizations via the `get_current_user_organisation_id(user_id)` database function.

## Data Isolation Mechanism

### Primary Isolation Method
- The `get_current_user_organisation_id(user_id)` function retrieves the user's `current_organisation_id` from the `users` table
- All organization-scoped queries include `WHERE ... organisation_id = get_current_user_organisation_id(?)`

### Organisation-Scoped Entities

**Direct organisation_id column:**

| Entity | organisation_id | Notes |
|--------|-----------------|-------|
| employees | Required | Direct FK to organisations |
| transactions | Required | Direct FK to organisations |
| bank_accounts | Required | Direct FK to organisations |
| salary_cost_labels | Required | Direct FK to organisations |
| forecasts | Required | Direct FK to organisations |
| forecast_details | Required | Direct FK to organisations |
| vat_settings | Required (1:1) | One per organisation |
| vats | Optional (nullable) | NULL = system-wide, ID = org-specific |
| categories | Optional (nullable) | NULL = system-wide, ID = org-specific |

**Indirectly scoped (via employee):**

| Entity | Parent Relationship | Isolation Path |
|--------|---------------------|----------------|
| salaries | employee_id -> employees | employee.organisation_id |
| salary_costs | salary_id -> salaries -> employees | employee.organisation_id |
| salary_exclusions | salary_id -> salaries -> employees | employee.organisation_id |
| salary_cost_exclusions | label_id -> salary_cost_labels | salary_cost_labels.organisation_id |

**Indirectly scoped (via transaction):**

| Entity | Parent Relationship | Isolation Path |
|--------|---------------------|----------------|
| transaction_exclusions | transaction_id -> transactions | transaction.organisation_id |

**Global entities (no isolation needed):**
- currencies
- fiat_rates
- users (accessed via auth context)
- registrations (pre-auth)
- refresh_tokens (per-user)
- reset_password (pre-auth)

## Priority Order

| Priority | Entities | Risk Level |
|----------|----------|------------|
| P0 | Employees, Transactions, Bank Accounts | Critical - direct org isolation |
| P1 | Salaries, Salary Costs | High - complex chain, potential vulnerabilities |
| P2 | VATs, Categories, Salary Cost Labels | Medium - nullable org_id handling |
| P3 | Forecasts, Exclusions, VAT Settings | Lower |

## Test Types Per Entity

1. **List** - User A can't see User B's data in list results
2. **Get** - User A can't fetch User B's specific resource by ID
3. **Update** - User A can't modify User B's resource
4. **Delete** - User A can't delete User B's resource
5. **Cross-Reference** - Can't reference resources from other orgs (e.g., transaction referencing another org's employee)

## Detailed Test Scenarios

### P0: Employee Isolation Tests

**File:** `employees_isolation_test.go`

```
TestListEmployees_CrossOrgIsolation
  - User A sees only Org A employees
  - User B sees only Org B employees

TestGetEmployee_CrossOrgIsolation
  - User A can GET own employee
  - User B cannot GET User A's employee (404/ErrNoRows)

TestUpdateEmployee_CrossOrgIsolation
  - User A can update own employee
  - User B cannot update User A's employee (0 rows affected)

TestDeleteEmployee_CrossOrgIsolation
  - User B cannot delete User A's employee
  - Employee still exists after User B's attempt
```

### P0: Transaction Isolation Tests

**File:** `transactions_isolation_test.go`

```
TestListTransactions_CrossOrgIsolation
TestGetTransaction_CrossOrgIsolation
TestUpdateTransaction_CrossOrgIsolation
TestDeleteTransaction_CrossOrgIsolation

TestCreateTransaction_WithCrossOrgEmployee
  - Cannot create transaction referencing employee from different org

TestCreateTransaction_WithCrossOrgCategory
  - Can use system categories (organisation_id = NULL)
  - Cannot use categories from another org
```

### P0: Bank Account Isolation Tests

**File:** `bank_accounts_isolation_test.go`

```
TestListBankAccounts_CrossOrgIsolation
TestGetBankAccount_CrossOrgIsolation
TestUpdateBankAccount_CrossOrgIsolation
TestDeleteBankAccount_CrossOrgIsolation
```

### P1: Salary Isolation Tests

**File:** `salaries_isolation_test.go`

```
TestListSalaries_CrossOrgIsolation
  - User B cannot list salaries for User A's employee

TestGetSalary_CrossOrgIsolation
  - User B cannot GET a salary belonging to User A's employee

TestCreateSalary_CrossOrgEmployee
  - User B cannot create salary for User A's employee

TestUpdateSalary_CrossOrgIsolation
  - User B cannot update User A's employee's salary

TestDeleteSalary_CrossOrgIsolation
```

### P1: Salary Cost Isolation Tests

**File:** `salary_costs_isolation_test.go`

```
TestListSalaryCosts_CrossOrgIsolation
TestGetSalaryCost_CrossOrgIsolation

TestCreateSalaryCost_CrossOrgSalary
  - Cannot create cost for salary in different org

TestUpdateSalaryCost_CrossOrgIsolation
  - NOTE: Current query uses EXISTS check - verify it properly isolates

TestDeleteSalaryCost_CrossOrgIsolation

TestCopySalaryCosts_CrossOrgIsolation
  - Cannot copy costs from a salary in different org
```

### P2: Salary Cost Label Isolation Tests

**File:** `salary_cost_labels_isolation_test.go`

```
TestListSalaryCostLabels_CrossOrgIsolation
TestGetSalaryCostLabel_CrossOrgIsolation
TestUpdateSalaryCostLabel_CrossOrgIsolation
TestDeleteSalaryCostLabel_CrossOrgIsolation

TestCreateSalaryCost_WithCrossOrgLabel
  - Cannot use label from different org when creating salary cost
```

### P2: VAT Isolation Tests

**File:** `vats_isolation_test.go`

```
TestListVats_ShowsSystemAndOwnOrg
  - User sees system VATs (organisation_id = NULL)
  - User sees own organisation's VATs
  - User does NOT see other org's custom VATs

TestGetVat_CrossOrgIsolation
  - Can access system VATs
  - Can access own org's VATs
  - Cannot access other org's VATs

TestUpdateVat_CrossOrgIsolation
  - Cannot update system VATs (canEdit = false)
  - Cannot update other org's VATs

TestDeleteVat_CrossOrgIsolation
```

### P2: Category Isolation Tests

**File:** `categories_isolation_test.go`

```
TestListCategories_ShowsSystemAndOwnOrg
TestGetCategory_CrossOrgIsolation

TestUpdateCategory_CrossOrgIsolation
  - Cannot update system categories
  - Cannot update other org's categories
```

### P3: VAT Settings Isolation Tests

**File:** `vat_settings_isolation_test.go`

```
TestGetVatSetting_CrossOrgIsolation
TestUpdateVatSetting_CrossOrgIsolation
TestDeleteVatSetting_CrossOrgIsolation
```

### P3: Forecast Isolation Tests

**File:** `forecasts_isolation_test.go`

```
TestListForecasts_CrossOrgIsolation
TestListForecastDetails_CrossOrgIsolation
TestCalculateForecast_OnlyUsesOwnOrgData
```

### P3: Organisation Isolation Tests

**File:** `organisations_isolation_test.go`

```
TestListOrganisations_OnlyShowsOwnMemberships
  - User only sees organisations they are a member of

TestGetOrganisation_MembershipRequired
  - Cannot GET organisation user is not a member of

TestUpdateOrganisation_MembershipRequired
  - Cannot update organisation user is not a member of
```

## Edge Cases

### Multi-Organisation User Tests

```
TestUserSwitchOrganisation_DataChanges
  - Create User A with both Organisation A and Organisation B
  - Create Employee in each organisation
  - Set current organisation to A -> only see Org A employees
  - Set current organisation to B -> only see Org B employees
  - Previous Org A employee ID returns 404
```

### Foreign Key Reference Tests

```
TestTransaction_CannotReferenceOtherOrgEmployee
TestSalaryCost_CannotReferenceOtherOrgLabel
TestTransaction_CanReferenceSystemCategory
TestTransaction_CanReferenceSystemVat
```

## Potential Vulnerabilities Identified

### 1. update_salary_cost.sql

**Location:** `backend/internal/adapter/db_adapter/queries/update_salary_cost.sql`

**Issue:** Uses a generic EXISTS check:
```sql
AND EXISTS (
    SELECT 1
    FROM employees AS e
    WHERE e.organisation_id = get_current_user_organisation_id(?)
)
```

This only checks if ANY employee exists in the user's org, not that the specific salary_cost belongs to that org.

**Test:** `TestUpdateSalaryCost_VulnerabilityCheck`

### 2. UpdateSalary DB Layer

**Location:** `backend/internal/adapter/db_adapter/salary.go:329`

**Issue:** WHERE clause uses `employee_id` without organization check:
```sql
WHERE id = ? AND employee_id = ?
```

**Defense:** Service layer calls GetSalary first (which has org check) and uses its employeeID.

**Test:** Verify service layer protection is sufficient.

### 3. delete_salary_costs_by_salary.sql

**Location:** `backend/internal/adapter/db_adapter/queries/delete_salary_costs_by_salary.sql`

**Issue:** No organization check:
```sql
DELETE FROM salary_costs WHERE salary_id = ?
```

**Defense:** Called only from CopySalaryCosts which validates the salary first.

**Test:** Verify the service layer protection chain is complete.

## Test Helper Functions Needed

```go
// Extend main_test.go with:

type CrossOrgTestEnv struct {
    Conn      *sql.DB
    APIService api_service.IAPIService
    DBAdapter  db_adapter.IDatabaseAdapter
    UserA     *models.User
    UserB     *models.User
    Currency  *models.Currency
}

func SetupCrossOrgTestEnvironment(t *testing.T) *CrossOrgTestEnv {
    conn := SetupTestEnvironment(t)

    dbAdapter := db_adapter.NewDatabaseAdapter(conn)
    sendgridService := sendgrid_adapter.NewSendgridAdapter("")
    apiService := api_service.NewAPIService(dbAdapter, sendgridService)

    currency, _ := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")

    userA, _, _ := CreateUserWithOrganisation(
        apiService, dbAdapter, "userA@test.com", "test", "Organisation A",
    )
    userB, _, _ := CreateUserWithOrganisation(
        apiService, dbAdapter, "userB@test.com", "test", "Organisation B",
    )

    return &CrossOrgTestEnv{
        Conn: conn,
        APIService: apiService,
        DBAdapter: dbAdapter,
        UserA: userA,
        UserB: userB,
        Currency: currency,
    }
}
```

## Estimated Scope

- ~40-50 test functions
- ~10 test files
- P0 tests: ~12 functions
- P1 tests: ~12 functions
- P2 tests: ~12 functions
- P3 tests: ~10 functions
