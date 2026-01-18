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
