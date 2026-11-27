# Database & Migrations

## Critical Rules

1. **Prefer migrations over manual mysql commands** - always use migrations for schema changes
2. When migrations fail due to version conflicts, **always try `make goose-static-down-to <version>` first** to roll back, then reapply
3. If old migration records exist from other branches (versions in `goose_db_version` without corresponding files), **ask the user before deleting** them via `docker compose exec database mysql`
4. Only after user approval: `DELETE FROM goose_db_version WHERE version_id > <target_version>;`

## Two-Tier Migration System

LiquiSwiss uses a unique migration approach with Goose:

### Static Migrations
- **Location**: `backend/internal/db/migrations/static/`
- **Purpose**: Permanent schema changes (tables, columns, indexes)
- **Behavior**: Applied once per environment using Goose versioning
- **Rule**: Never modify after applied to production; create new migrations instead

### Dynamic Migrations
- **Location**: `backend/internal/db/migrations/dynamic/`
- **Purpose**: Database functions, views, and reference data
- **Behavior**: Dropped and recreated on every app startup (unversioned)
- **Rule**: Safe to edit directly; changes apply on next restart

### Testing Fixtures
- **Location**: `backend/internal/db/migrations/testing/`
- **Purpose**: E2E accounts with well-known credentials
- **Behavior**: Unversioned like dynamic, but only applied when `TESTING_ENVIRONMENT` is
  set, so they can never reach a real environment
- **Note**: `backend/internal/db/migrations/excluded/` holds migrations kept out of the
  embed on purpose

Pending static migrations are logged on startup (air runs with `--no-migrate`); apply
them with `make migrate` from the repo root.

## When to Use Which

| Change Type | Migration Type |
|-------------|----------------|
| Add table | Static |
| Add/modify column | Static |
| Add index | Static |
| Create/modify view | Dynamic |
| Create/modify function | Dynamic |
| Insert reference data | Dynamic |

## Key Database Objects

| Object | Purpose | Location |
|--------|---------|----------|
| `get_current_user_organisation_id()` | Returns organisation ID for current user | Dynamic migrations |
| `get_current_user_scenario_id()` | Returns the user's active scenario in the current organisation, falls back to the base scenario | Dynamic migrations |
| `trg_organisation_base_scenario` | Creates the base scenario ("Grundszenario") for every new organisation | Dynamic migrations |
| `ranked_salaries` view | Salaries with ranking for active/terminated status | Dynamic migrations |

## Scenarios

Every organisation has one base scenario (`scenarios.is_base`) and any number of copies.

- **Scenario-bound** (column `scenario_id`): `employees`, `transactions`, `bank_accounts`,
  `forecasts`, `forecast_details`, `salary_cost_exclusions`. Salaries, salary costs, their
  details, base links and the transaction/salary exclusions hang on these through their parent.
- **Shared by the organisation**: categories, VATs, VAT settings, salary cost labels, members.
- Queries on scenario-bound tables filter with `scenario_id = get_current_user_scenario_id(?)`
  instead of the organisation function. Inserts use `INSERT ... SELECT ... FROM scenarios s
  WHERE s.id = get_current_user_scenario_id(?)` to set both `organisation_id` and `scenario_id`.
- Checks that must see every scenario (e.g. "is this category still used?") keep the
  organisation filter.
- The active scenario is per user (`user_organisation_settings.current_scenario_id`).
- Creating a scenario deep-copies the source in `db_adapter/scenario.go`; the column list comes
  from `information_schema`, so new columns are copied without code changes. A new
  scenario-bound **table** must be added to the copier.
- Static `Down` migrations must run cleanly on a database where the migration never ran: the
  tests execute all `Down` migrations unversioned before every test (use `IF EXISTS`).

## Commands

Run from `backend/` directory:

```bash
make goose-static-create <name>    # Create new static migration
make goose-dynamic-create <name>   # Create new dynamic migration
make goose-static-up               # Apply static migrations
make goose-dynamic-up              # Apply dynamic migrations
make goose-static-down             # Rollback one static migration
make goose-dynamic-down            # Rollback one dynamic migration
```

See [backend/Makefile](../../backend/Makefile) for all available commands.

## Database Access

Always access the database via docker compose:

```bash
docker compose exec database mysql -u liquiswiss -ppassword liquiswiss -e "YOUR SQL QUERY;"
```
