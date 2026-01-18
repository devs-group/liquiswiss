# Backend Guide

## Technology Stack

- Go with Gin web framework
- MariaDB database
- JWT authentication
- Goose migrations

## Key Files to Understand

| Purpose | File |
|---------|------|
| All API routes | [internal/api/router.go](../../backend/internal/api/router.go) |
| Service interface (50+ methods) | [internal/service/api_service/api_service.go](../../backend/internal/service/api_service/api_service.go) |
| Domain models | [pkg/models/](../../backend/pkg/models/) |
| JWT handling | [pkg/auth/auth.go](../../backend/pkg/auth/auth.go) |
| Auth middleware | [internal/middleware/auth.go](../../backend/internal/middleware/auth.go) |
| Environment config | [config/config.go](../../backend/config/config.go) |

## Patterns

**Embedded SQL**: Queries live in `internal/adapter/db_adapter/queries/*.sql` and are embedded into the binary.

**Handler structure**: Each handler file in `internal/api/handlers/` corresponds to a domain entity.

**Modernize**: Always apply Go modernize suggestions (e.g., use `any` instead of `interface{}`). Run `make modernize` to auto-fix.

## Testing

**Always run tests after backend changes:**
```bash
go test -count=1 ./...
```

The `-count=1` flag disables test caching to ensure tests always run.

**After making changes, always check if new tests are required** - especially for new endpoints, business logic, or edge cases.

Mocks are generated via `go generate ./...` using mockgen. Tests require a running MariaDB instance (`docker compose up`).

Test environment is configured via:
- `.env.local.testing` for local development
- `.env.github.testing` for CI

### Time Simulation in Tests

When writing tests that depend on date/time (forecasts, scheduled transactions, salary calculations), you **MUST** set both:

1. **Database time** - `SetDatabaseTime(conn, "2025-01-01")` - sets MariaDB's `NOW()`/`CURDATE()`
2. **Go clock time** - `utils.DefaultClock.SetFixedTime(&parsedTime)` - sets the application clock

Example:
```go
simulatedTime := "2025-01-01"
err := SetDatabaseTime(conn, simulatedTime)
require.NoError(t, err)
parsedTime, err := time.Parse(utils.InternalDateFormat, simulatedTime)
require.NoError(t, err)
utils.DefaultClock.SetFixedTime(&parsedTime)
defer utils.DefaultClock.SetFixedTime(nil) // Always reset after test
```

Failing to set both will cause inconsistent test behavior between Go code and database queries.

## Cross-Organisation Data Isolation

This is a multi-tenant application where each user belongs to one or more organisations. **All data must be scoped to the user's current organisation.**

### Validation Rules

When creating or updating entities that reference other entities (e.g., transactions referencing employees, categories, VATs), you **MUST** validate that the referenced entity belongs to the user's organisation:

```go
// Example: Validate employee belongs to user's org before creating transaction
if payload.Employee != nil {
    if _, err := a.dbService.GetEmployee(userID, *payload.Employee); err != nil {
        return nil, fmt.Errorf("invalid employee: not found")
    }
}
```

**Key points:**
- Use the existing `Get*` functions (e.g., `GetEmployee`, `GetCategory`, `GetVat`) which already filter by organisation
- Return generic "not found" errors - never reveal that an entity exists in another organisation
- Apply validation in both Create and Update functions

### Isolation Tests

All new endpoints and entity relationships **MUST** have cross-organisation isolation tests. These tests verify that:
1. Users can only see/access their own organisation's data
2. Users cannot reference entities from other organisations
3. List operations only return data from the user's organisation

Test files follow the pattern `*_isolation_test.go` and use `SetupCrossOrgTestEnvironment()` to create two users in separate organisations.

Example test structure:
```go
func TestCreateEntity_WithCrossOrgReference(t *testing.T) {
    env := SetupCrossOrgTestEnvironment(t)
    defer env.Conn.Close()

    // Create entity for User A
    entityA := createEntityForUserA()

    // User B attempts to reference User A's entity - should fail
    _, err := env.APIService.CreateSomething(..., entityA.ID, env.UserB.ID)
    require.Error(t, err)
}
```
