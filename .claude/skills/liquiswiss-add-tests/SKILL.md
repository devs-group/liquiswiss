---
name: liquiswiss-add-tests
description: Decide whether the work done in this session needs Go integration tests, propose what to add, then add them. Audits existing coverage first and suggests gaps before writing code. Triggers on "add tests", "write tests", "test coverage", "integration test this", "do we have tests for this".
user_invocable: true
auto_trigger:
  - add tests
  - write tests
  - test coverage
  - integration test this
  - do we have tests for this
  - are there tests
  - cover this with tests
---

# Add Tests

Audit what was changed in the session, decide if tests make sense, propose additions, and then add them. Never write tests the user didn't actually ask for — the bar is "does this change need a regression guard or a TDD-style contract?".

Repo rules come from `CLAUDE.md`:
- **All backend changes MUST run `go test -count=1 ./...`** after implementation. The pre-commit hook runs `go test ./...` automatically — failing tests block the commit.
- **Integration tests go through the service or handler layer** with real repos against a real MariaDB (via `docker compose up` running the `database-testing` service on port `3318`).
- **Test environment**: `TESTING_ENVIRONMENT` env var picks `.env.local.testing` (default) or `.env.github.testing` (CI). Files live at `backend/.env.local.testing` and `backend/.env.github.testing`.
- **No build tags** for tests — `go test ./...` runs everything.
- **No frontend test harness** in this repo. Skip frontend test proposals unless the user adds Vitest/similar.

## Flow

### 1. Pinpoint what changed this session

Start from the diff, not the conversation transcript:

```bash
git status --short
git diff --stat main...HEAD          # vs main if on a feature branch
git log --oneline origin/main..HEAD  # commits on this branch
```

If on `main`, fall back to working-tree diff (`git diff` + `git status`). Read the actual changes (`git diff` on the files that matter) — the *why* drives test choices, not filenames.

### 2. Map changes to test layers

Classify each meaningful change:

| Change shape | Test layer |
|---|---|
| New schema migration / column | Integration: real DB round-trip via `db_adapter` + service-level INSERT + every SELECT path |
| New/changed `api_service` method | Integration: `backend/internal/api/handlers/*_test.go` style — call `apiService.X(...)` against real DB |
| New/changed handler (HTTP route) | Handler test using `httptest` + real `apiService` (see `backend/internal/api/handlers/auth_test.go` for mock-style and `invitation_test.go` for real-DB style) |
| New cross-org access path | Add to `*_isolation_test.go` — assert User B can't see User A's data |
| New env var / config field | Usually no test — verification is reading the config + a manual sanity check. Add a test only if behaviour branches on the value |
| Pure frontend change (Nuxt) | No Go test — there is no Nuxt test harness in this repo. Skip |
| Migration file | Implicit coverage: `migrateDatabase` in `main_test.go` runs static + dynamic migrations on every test run. A failing migration fails every test |
| Email template / SMTP wiring | Unit test under `backend/internal/adapter/email_adapter/` (see `email_adapter_test.go` for graceful-skip + render assertions) |
| Currency / fixer fallback | Unit test under `backend/internal/service/fixer_io_service/` (see `fallback_test.go`) |

**Skip writing tests when:**
- The change is a comment / docs / rename / formatting — no behaviour to guard.
- The change is a chart values bump, image pin, or CI tweak.
- The change is a frontend-only Vue component edit.
- The behaviour is already covered by an existing test that would fail if the code regressed. Prove this by grepping for the relevant symbols or table columns.

### 3. Audit existing coverage first

Before proposing anything new, grep for tests that already touch the changed symbols:

```bash
# Files touched in the session
CHANGED=$(git diff --name-only main...HEAD | grep -E '\.go$' | grep -v _test.go)

# For each changed Go file, find tests in the same package or in the handlers test dir
for f in $CHANGED; do
  pkg=$(dirname "$f")
  echo "--- tests in $pkg ---"
  ls "$pkg"/*_test.go 2>/dev/null | head -5
done

# Existing handler tests for a specific symbol
rg -l 'YourNewMethod|new_column_name' backend/internal/api/handlers/*_test.go backend/internal/service/api_service/*_test.go backend/internal/adapter/*/

# Cross-org isolation tests
ls backend/internal/api/handlers/*_isolation_test.go
```

If a test already covers the regression you'd write, **extend it** (add an assertion or a sub-test) instead of creating a new file. One assertion in the right place beats a new test file.

### 4. Propose before writing

State the plan in a short list before touching files. Each item should say:
- **Where** (package + file, new or existing).
- **What it guards** in one sentence — the *regression*, not the feature.
- **Why this layer** — service vs handler vs adapter unit.

Ask the user to confirm or trim. If the user already invoked this skill mid-work ("add tests now"), skip the confirmation for routine integration items but still ask before adding cross-cutting refactors to shared helpers in `main_test.go`.

### 5. Write the tests

Service / handler integration test pattern (see `backend/internal/api/handlers/invitation_test.go` for canonical example):

```go
package handlers_test

import (
    "testing"

    "github.com/stretchr/testify/require"

    "liquiswiss/config"
    "liquiswiss/internal/adapter/db_adapter"
    "liquiswiss/internal/adapter/email_adapter"
    "liquiswiss/internal/service/api_service"
    "liquiswiss/pkg/models"
)

func TestMyFeature_SomeRegression(t *testing.T) {
    conn := SetupTestEnvironment(t)
    defer conn.Close()

    dbAdapter := db_adapter.NewDatabaseAdapter(conn)
    emailService := email_adapter.NewEmailAdapter(config.Config{}) // empty SMTP host -> Send is a no-op
    apiService := api_service.NewAPIService(dbAdapter, emailService)

    _, err := CreateCurrency(apiService, "CHF", "Swiss Franc", "de-CH")
    require.NoError(t, err)

    user, org, err := CreateUserWithOrganisation(
        apiService, dbAdapter, "feature.test@test.com", "test", "Feature Test Org",
    )
    require.NoError(t, err)

    // Call the real service method.
    // Assert on DB state AND on the service return.
}
```

- Use `package handlers_test` (external) — internal access not needed for these flows.
- Each test gets its own `SetupTestEnvironment(t)` + `defer conn.Close()` — tests must not depend on each other (the helper resets the DB schema).
- Pass `config.Config{}` to `NewEmailAdapter` when the test should not actually send mail. Send methods log a warn and return `nil` when `SMTPHost` is empty.
- For cross-org isolation, use `SetupCrossOrgTestEnvironment(t)` and assert User B operations targeting User A resources return an error or empty result.

Adapter / unit test pattern (no DB needed — see `backend/internal/adapter/email_adapter/email_adapter_test.go`):

```go
package email_adapter

import (
    "testing"

    "github.com/stretchr/testify/require"

    "liquiswiss/config"
)

func init() {
    logger.NewZapLogger(false) // many adapters use the global zap logger
}

func TestMyAdapter_Behaviour(t *testing.T) {
    a := newSMTPAdapter(config.Config{ /* ... */ })
    require.NoError(t, a.SomeMethod(...))
}
```

### 6. Run and report

The test suite needs MariaDB running:

```bash
# From repo root: ensure database-testing is up
docker compose ps database-testing
# If not: docker compose up -d database-testing
```

Run only the package(s) you touched first (faster than full suite):

```bash
cd backend
go test -count=1 ./internal/api/handlers -run TestMyFeature -v > /tmp/test-feature.log 2>&1
echo "exit=$?"
grep -E "^--- (PASS|FAIL)|^(PASS|FAIL|ok )" /tmp/test-feature.log | tail -20
```

Then run full suite to catch regressions in other packages:

```bash
cd backend
go test -count=1 ./... > /tmp/test-full.log 2>&1
echo "exit=$?"
grep -E "^(FAIL|ok |---)" /tmp/test-full.log | tail -40
```

Always redirect to a log file and grep — never tail. The handler suite takes ~50s.

If a new method was added to an interface (e.g. `IDatabaseAdapter`, `IAPIService`, `IEmailAdapter`, `IFixerIOService`), regenerate mocks **before** running tests:

```bash
cd backend && go generate ./...
```

Mocks live in `backend/internal/mocks/` and are regenerated from `//go:generate mockgen ...` directives at the top of each interface file.

Report back with:
- Tests added (file:function).
- Result (pass / fail).
- What they guard against in one line each.
- Any follow-up coverage that's recommended but not yet added.

## Rules

- **Never write tests the change doesn't justify.** If the code is a `chore:` or pure-refactor, say "no tests needed" and stop.
- **Never change the code under test to make a test pass.** If a test fails, fix the code or fix the test's premise.
- **Respect CLAUDE.md:** always run `go test -count=1 ./...` before committing; the pre-commit hook will run it anyway. Use `/tmp/test-*.log` for long runs.
- **Regenerate mocks** when interfaces change. Forgetting this breaks every test file that constructs a mock.
- **Don't add frontend tests** — there is no harness. Note this gap in the report instead.
- Keep generated assertions meaningful — asserting that a struct has a field is not a test; assert the *value* round-trips or the *behaviour* changes.
