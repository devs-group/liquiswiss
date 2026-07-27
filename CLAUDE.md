# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LiquiSwiss is a liquidity planning application with a Go backend and Nuxt 4 frontend. It helps organizations forecast cash flow based on employee salaries, salary costs, transactions, and multi-currency exchange rates.

## Setup (after cloning)

```bash
git config core.hooksPath .githooks
```

This enables pre-commit hooks that run `npm run lint:fix` and `go test ./...` before each commit.

**Important**: At the start of each session, verify git hooks are configured by running the command above. The pre-commit hook runs `npm run lint:fix` and `go test` automatically on commit, so there's no need to run these separately right before committing.

## Quick Commands

### Frontend (`frontend/`)
```bash
nvm use              # Use correct Node version (required first)
npm install          # Install dependencies
npm run dev          # Dev server at http://localhost:3000
npm run build        # Production build
```

Note: `npm run lint:fix` runs automatically via pre-commit hook.

### Backend (`backend/`)
```bash
go mod tidy          # Install dependencies
go run .             # Run dev server (or use `air` for hot-reloading)
go test -count=1 ./...  # Run all tests (requires docker compose up)
go test -count=1 ./internal/api/handlers -run TestName  # Run specific test
go vet ./...         # Static analysis
go generate ./...    # Regenerate mocks
make modernize       # Apply Go modernize suggestions
```

**Testing Process (MANDATORY for all backend changes)**:
1. After implementing backend changes, ALWAYS run `go test -count=1 ./...`
2. Evaluate if new tests are required (new endpoints, new parameters, business logic changes)
3. Write new tests following existing patterns (see `*_test.go` files)
4. Run tests again to verify all pass before committing

**Test Requirements**:
- Tests require MariaDB running (`docker compose up`) and `.env.local.testing` configured
- Test environment determined by `TESTING_ENVIRONMENT` env var (uses `.env.local.testing` locally, `.env.github.testing` in CI)
- Optional fixtures available at `backend/internal/adapter/db_adapter/fixtures/`

### Database
```bash
docker compose up    # Start MariaDB
make goose-static-create <name>   # Create schema migration (from backend/)
make goose-dynamic-create <name>  # Create function/view migration (from backend/)
```

### Running Both Servers (Docker Compose via Make)

**IMPORTANT**: All services run via `docker compose` wrapped by `make` targets. The Makefile pulls the BWSM access token from macOS keychain and runs `bws run -- docker compose ...`. No native `air`/`npm run dev`.

```bash
make up              # docker compose up -d
make down
make restart
make logs            # follow all logs
make ps
make build           # up -d --build
make dc CMD="logs -f backend"   # any compose passthrough
```

**No required BWSM secrets locally**: all env vars have dev fallbacks in compose. Without a BWSM token, `make` falls through to bare `docker compose` and starts the stack with `SMTP_HOST=mailpit` (mail captured locally — see below) and `FIXER_IO_KEY=""` (rates load from `backend/internal/service/fixer_io_service/fallback_rates.json`). Everything works without external accounts. With a BWSM token, the Makefile auto-injects real values via `bws run`.

**Local mail capture (Mailpit)**: outbound emails are routed to the `mailpit` compose service. Inspect them at <http://localhost:8025>. SMTP listener is `mailpit:1025` (no auth, no TLS). Production uses any SMTP relay (SendGrid SMTP, Brevo, AWS SES, Mailgun, Postmark, etc.) via the same `SMTP_*` env vars.

**Mail env vars** (all default sensibly; override only for prod or non-Mailpit local relay):

| Var | Local default | Notes |
|---|---|---|
| `SMTP_HOST` | `mailpit` | `""` disables mail send entirely (graceful skip) |
| `SMTP_PORT` | `1025` | `587` (STARTTLS) or `465` (implicit TLS) for most relays |
| `SMTP_USER` | `""` | e.g. `apikey` for SendGrid SMTP |
| `SMTP_PASS` | `""` | API key / password |
| `SMTP_FROM_ADDRESS` | `no-reply@liquiswiss.local` | sender address |
| `SMTP_FROM_NAME` | `LiquiSwiss` | sender display name |
| `SMTP_TLS` | _(auto)_ | `off` \| `starttls` \| `implicit`. Empty → derived from port (465→implicit, 587→starttls, else→off). |

**Local currency rates fallback**: when `FIXER_IO_KEY` is empty, `FetchFiatRates()` upserts pairs from `backend/internal/service/fixer_io_service/fallback_rates.json` instead of calling Fixer.io. Refresh the JSON manually if drift becomes problematic (one-off `curl` against Fixer.io with a real key, paste in).

**Bitwarden Secrets Manager (BWSM)**: local dev uses the `liquiswiss-dev` project. Project id and access token come from the vault (web vault → Secrets Manager), they are not documented here.

**Keychain access token setup** (token = `0.<uuid>...:<key>` from BWSM machine account):

```bash
# Set / first time:
security add-generic-password -a "$USER" -s "bws-liquiswiss-dev" -w
#   -> prompts twice for password, paste token

# Update (rotate token):
security delete-generic-password -a "$USER" -s "bws-liquiswiss-dev"
security add-generic-password -a "$USER" -s "bws-liquiswiss-dev" -w

# Delete:
security delete-generic-password -a "$USER" -s "bws-liquiswiss-dev"
```

**Token expiry**: BWSM access tokens expire 60 days after creation. When `make` stops injecting secrets, the token has lapsed: mint a new one via web vault → Secrets Manager → Machine accounts → access tokens, then re-run keychain `delete` + `add` above. Without a token everything still runs on the dev fallbacks described above.

To follow logs of specific services:

```bash
make dc CMD="logs -f backend nuxt"
```

**Note**: During hot-reloading, refreshing, or in-between states, you may see transient errors in the logs until the code changes are complete or fixed.

### Docker Compose Setup

- **Root `docker-compose.yml`**: local dev (nuxt, backend, database, database-testing, mailpit, phpmyadmin). Backend uses `target: dev` with `air` hot-reload. Nuxt runs `npm run dev-host`.
- Local ports: nuxt `3007`, backend `8087`, database `3317`, database-testing `3318`, mailpit SMTP `1025`, mailpit UI `8025`, phpmyadmin `8097`.
- Inside compose network, services reach each other by service name (e.g. `backend:8080`, `database:3306`).
- **`_deployment/docker-compose.yml`**: mirrors prod (nuxt, backend, landing/wordpress, database-app, database-wordpress, phpmyadmin) with Traefik labels and resource limits. Use for prod-parity local testing: `docker compose -f _deployment/docker-compose.yml --env-file .env up`.
- **Single `.env` at root** (transitional): flat namespace (`APP_DB_*`, `WP_DB_*`, etc.) mirrors future Bitwarden Secrets Manager injection. Both compose files load from it. Once BWSM wired: `bws run -- docker compose up` replaces `.env`.
- **Native run is deprecated**: `backend/.env` and `frontend/.env*` removed. `backend/.env.local.testing` + `backend/.env.github.testing` retained only for `go test` (not picked up by compose).
- **Named volumes** (`nuxt_node_modules`, `backend_go_cache`, `backend_build_cache`): cache for fast rebuilds. `nuxt_node_modules` is critical on macOS/Docker Desktop — bind-mounting node_modules across VM boundary is slow. Tradeoff: host IDE won't see container's node_modules. After changing `package.json`, run `docker compose build nuxt` or `docker compose run --rm nuxt npm install`.

## Documentation

Detailed documentation is in [docs/ai/](docs/ai/):

- [Architecture Overview](docs/ai/architecture.md) - System design and key entry points
- [Backend Guide](docs/ai/backend.md) - Go API structure and patterns
- [Frontend Guide](docs/ai/frontend.md) - Nuxt 4 structure and patterns
- [Database & Migrations](docs/ai/database.md) - Two-tier migration system
- [Business Logic](docs/ai/business-logic.md) - Salary costs, forecasts, VAT calculations
- [Authentication](docs/ai/authentication.md) - JWT dual-token flow

## Git Commits

- Always a single line, no multi-line messages
- No "Co-Authored-By" or similar footers
- Start with capital letter
- Write as if completing: "(This commit will) ..."
- Commit in small logical groups, but keep the total commit count low
- **Never `git push` without asking first** (unless explicitly instructed to push for a while)

**Examples:**
- `Add link field to transactions to enter url`
- `Fix calculation of VAT for quarterly transactions`
- `Update employee form validation`

## Configuration

All environment variables and their local dev defaults are declared inline in `docker-compose.yml` (local) and `_deployment/docker-compose.yml` (production). There are no `.env.example` files anymore; production values come from BWSM.

## External Services

- **Fixer.io**: Currency exchange rates (synced every 12 hours)
- **SMTP relay** (production): transactional emails via any provider — SendGrid SMTP, Brevo, AWS SES, Mailgun, Postmark, etc. Configured via `SMTP_*` env vars. Locally: Mailpit captures everything.

## Plans

- Store implementation plans in `docs/plans/`
- Delete plan files once fully implemented

## Context from Previous Sessions

- Check for `claude_chat_history.txt` in root for context from the previous session

## General Guidelines

- **Current year is 2026**: Always search for up-to-date methods and documentation (2025-2026) to prevent outdated implementations. Libraries and frameworks evolve rapidly.
- **Always verify current directory**: Before running any shell command, verify the working directory using `pwd` or by using absolute paths. This prevents errors from running commands in the wrong directory.
