# Business Logic

## Salary Cost Calculation

**Location**: [backend/internal/service/api_service/salary_cost.go](../../backend/internal/service/api_service/salary_cost.go)

Salaries have base amounts. Salary costs attach to salaries with:

| Field | Description |
|-------|-------------|
| Amount & AmountType | Fixed amount or percentage of base salary |
| Distribution | `employee`, `employer`, or `both` (multiplies by 2) |
| Cycle | Monthly, quarterly, annually, or one-time |
| BaseSalaryCostIDs | Dependencies on other costs (recursive calculation) |

**Key rules**:
- Circular dependencies are validated and rejected
- Costs on termination salaries are not allowed
- Base costs are resolved recursively before calculating dependent costs

**Example**: A 13% pension cost split between employee and employer uses `distribution = "both"`, resulting in `base_salary * 0.13 * 2`.

## Forecast Calculation

**Location**: [backend/internal/service/api_service/forecast.go](../../backend/internal/service/api_service/forecast.go)

| Component | Calculation |
|-----------|-------------|
| Revenue | Positive transactions + positive salary amounts |
| Expenses | Negative transactions + salary costs + VAT |
| Cashflow | Revenue - Expenses |

**Features**:
- Projects 12 months ahead by default
- Users can exclude specific items from specific forecast months
- Performance slider adjusts displayed income values and VAT

## VAT Calculation

**Location**: [backend/internal/service/api_service/vat.go](../../backend/internal/service/api_service/vat.go)

- Automatically calculated based on positive transactions (revenue)
- Configurable per organisation via [vat_setting.go](../../backend/internal/service/api_service/vat_setting.go)
- Affects forecast cashflow calculations
