# LiquiSwiss

Liquidity Planning

- [Landing Page](https://liquiswiss.ch/)
- [Discord](https://discord.gg/7ckBNzskYh)

## Setup (after cloning)

```bash
git config core.hooksPath .githooks
```

This enables pre-commit hooks that run lint and tests before each commit.

## Configuration

No configuration is needed to start: every variable has a local dev default in `docker-compose.yml`, so
`make up` brings up the full stack without any accounts or secrets.

- **E-mail**: outbound mail goes to the local [Mailpit](https://mailpit.axllent.org/) service. Read it at
  http://localhost:8025. Production uses any SMTP relay via the `SMTP_*` variables.
- **Currency rates**: with an empty `FIXER_IO_KEY`, rates are loaded from
  [fallback_rates.json](backend/internal/service/fixer_io_service/fallback_rates.json) instead of calling
  [Fixer.io](https://fixer.io/).
- **Secrets**: production values are injected from Bitwarden Secrets Manager by the `make` targets. See
  [CLAUDE.md](CLAUDE.md) for the setup.

## Admin

We are using [phpMyAdmin](https://www.phpmyadmin.net/) with Docker to provide an interface to the database.

- Configure it with the `PMA_*` variables in `docker-compose.yml`, see the
  [image documentation](https://hub.docker.com/_/phpmyadmin)
- You can check out your database (locally) at: http://localhost:8097/

# Frontend (Nuxt 4)

Look at the [Nuxt documentation](https://nuxt.com/docs/getting-started/introduction) to learn more.

> Make sure you are in the [frontend](frontend) directory for all the following actions

## Setup

```
npm install
```

## Development Server

```
npm run dev or npm run dev-host (to expose host and be able to connect from another device)
```

# Backend (Golang)

> Make sure you are in the [backend](backend) directory for all the following actions

## Setup

```
go get OR go mod tidy
```

## Development Server

1. Install [Air](https://github.com/air-verse/air): `go install github.com/air-verse/air@latest`

```
air
```

## Migrations

1. Install [Goose](https://github.com/pressly/goose): `go install github.com/pressly/goose/v3/cmd/goose@latest`

We differentiate between static and dynamic migrations whereas **static migrations** are all migrations that
actually hold data later and a migration down would lead to data loss such as table creation or alterations.

Dynamic migrations are stored functions, views or triggers, basically things that can be removed entirely and reapplied.

> The placeholder <directory> must either be replaced by "static" or "dynamic" (without quotes)

- Create Migration: `goose --dir internal/db/migrations/<directory> create <name-of-migration> sql`
    - Follow up with: `goose --dir internal/db/migrations/<directory> fix` to apply sequential numbering

> We run auto migrations on each app start. Since "air" will restart the app on any changes also in the .sql files
> the migrations will apply automatically but it might be helpful sometimes to rollback and reapply.

- Apply Migration: `goose --dir internal/db/migrations/<directory> mysql liquiswiss:password@/liquiswiss up`
- Rollback Migration: `goose --dir internal/db/migrations/<directory> mysql liquiswiss:password@/liquiswiss down`
- Or check out the [Makefile](backend/Makefile)

## Fixtures

> Optional step

You can fixtures from the [fixtures](backend/internal/adapter/db_adapter/fixtures) directory if you desire.
The dynamic migrations insert a minimal set of data required to make the app work properly. You can check out the
minimal inserted data in [10000_apply_minimal_fixtures.sql](backend/internal/db/migrations/dynamic/10000_apply_minimal_fixtures.sql)

## Backend Tests

> Make sure you are in the [backend](backend) directory

> Make sure you spin up the test database with `docker compose up`

1. Install [Mockgen](https://github.com/uber-go/mock) with `go install go.uber.org/mock/mockgen@latest` to generate
   mocks
    - There are `go generate` commands already in the files so you can simply do `go generate ./...`
2. You can run all tests with `go test ./...` locally
3. Locally the [.env.local.testing](backend/.env.local.testing) is used
4. For the Github Action the [.env.github.testing](backend/.env.github.testing) is used
    - Check out the [ci.yml](.github/workflows/ci.yml) and check for the service used in the **test_backend** job
    - The environment variable `TESTING_ENVIRONMENT` determines which .env file to use

## E2E Tests (Playwright)

> Make sure you are in the [frontend](frontend) directory

### Setup

1. Install dependencies: `npm install`
2. Install Playwright browsers: `npm run test:e2e:install`

### Running Tests

```bash
npm run test:e2e          # Run all tests headless
npm run test:e2e:ui       # Open interactive UI mode
npm run test:e2e:headed   # Run tests with visible browser
npm run test:e2e:debug    # Run tests in debug mode
```

### Requirements

- Frontend dev server running (`npm run dev`)
- Backend server running (`air` or `go run .`)
- Test user credentials configured via environment variables:
  - `E2E_TEST_EMAIL` - Test user email
  - `E2E_TEST_PASSWORD` - Test user password

# Production

For production make sure you define the proper values for your envs (no matter in which way you provide them)

- `WEB_HOST` - Reflects your Frontend URL (eg. https://yourdomain.com)
- `JWT_KEY` - Should be a long and secure password
- `NUXT_API_HOST` - Reflects your Backend URL (eg https://api.yourdomain.com)
## MCP Server (AI Access)

The backend exposes an MCP (Model Context Protocol) server at `/api/mcp` (Streamable HTTP), protected by an embedded OAuth 2.1 authorization server (dynamic client registration, PKCE, rotating refresh tokens). AI clients like Claude can manage bank accounts, transactions, employees, salaries and read liquidity forecasts on behalf of a user.

Connect locally:

```bash
claude mcp add --transport http liquiswiss-local http://localhost:8087/api/mcp
```

Authentication runs through the browser: login, consent page (with an explicit notice that data flows to the chosen LLM provider), redirect back. Sessions match web logins (20 min access tokens, 90 day rotating refresh tokens). Users can revoke access anytime under Profile settings, "Verbundene Anwendungen". Public base URL of the backend is configured via `BACKEND_PUBLIC_URL`.
