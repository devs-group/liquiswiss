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

### Running Both Servers (Docker Compose)

**IMPORTANT**: All services run via `docker compose` (no more native `air` or `npm run dev`). At the start of each session, start the stack in the background:

```bash
docker compose up -d
```

To follow logs:

```bash
docker compose logs -f backend nuxt
```

To restart a single service after code changes that require image rebuild (rare — `air` and Vite hot-reload most changes):

```bash
docker compose up -d --build <service>
```

**Note**: During hot-reloading, refreshing, or in-between states, you may see transient errors in the logs until the code changes are complete or fixed.

### Docker Compose Setup

- **Root `docker-compose.yml`**: local dev (nuxt, backend, database, database-testing, phpmyadmin). Backend uses `target: dev` with `air` hot-reload. Nuxt runs `npm run dev-host`.
- Local ports: nuxt `3007`, backend `8087`, database `3317`, database-testing `3318`, phpmyadmin `8097`.
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
- **Always `git push` after a successful commit** (as a separate command, never combined with commit)

**Examples:**
- `Add link field to transactions to enter url`
- `Fix calculation of VAT for quarterly transactions`
- `Update employee form validation`

## Configuration

See `.env.example`, `backend/.env.example`, and `frontend/.env.example` for required environment variables.

## External Services

- **Fixer.io**: Currency exchange rates (synced every 12 hours)
- **SendGrid**: Transactional emails (requires API key and dynamic template)

## Plans

- Store implementation plans in `docs/plans/`
- Delete plan files once fully implemented

## Context from Previous Sessions

- Check for `claude_chat_history.txt` in root for context from the previous session

## General Guidelines

- **Current year is 2026**: Always search for up-to-date methods and documentation (2025-2026) to prevent outdated implementations. Libraries and frameworks evolve rapidly.
- **Always verify current directory**: Before running any shell command, verify the working directory using `pwd` or by using absolute paths. This prevents errors from running commands in the wrong directory.
