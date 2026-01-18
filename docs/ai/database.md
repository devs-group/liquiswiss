# Database & Migrations

## Critical Rules

1. **NEVER manually adjust the database with mysql commands** unless explicitly allowed by the user - always use migrations
2. When migrations fail due to version conflicts, **always try `goose down-to <version>` first** to roll back, then reapply
3. If old migration records exist from other branches (versions in `goose_db_version` without corresponding files), **ask the user before deleting** them via mysql
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
| `ranked_salaries` view | Salaries with ranking for active/terminated status | Dynamic migrations |

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
