# LiquiSwiss

Liquidity Planning

# Frontend (Nuxt 3)

Look at the [Nuxt 3 documentation](https://nuxt.com/docs/getting-started/introduction) to learn more.

## Setup

```bash
cd frontend
npm install

or

docker compose build (if you have docker)
```

## Development Server

```bash
cd frontend
npm run dev

or

cd frontend
npm run dev-host (to expose host and be able to connect from another device)

or

docker compose up -d (if you have docker)
```

# Backend (Golang)

## Setup

```bash
cd backend
go mod tidy

or

docker compose build (if you have docker)
```

## Development Server

1. Install [Air](https://github.com/cosmtrek/air): `go install github.com/cosmtrek/air@latest`

```bash
cd backend
air

or

docker compose up -d (if you have docker)
```

## Migrations

- Create Migration: `goose --dir internal/db/migrations create <name-of-migration> sql`
- Apply Migration: `goose --dir internal/db/migrations mysql liquiswiss:password@/liquiswiss up`
- Rollback Migration: `goose --dir internal/db/migrations mysql liquiswiss:password@/liquiswiss down`

## Tests

1. Install [Mockgen](https://github.com/uber-go/mock) with `go install go.uber.org/mock/mockgen@latest` to generate
   mocks
    - There are `go generate` commands already in the files so you can simply do `go generate ./...`
1. You can run all tests with `go test ./...` in the root directory

## Build

1. `cd cmd/liquiswiss && go build` or
2. `make gobuild`

# Production

Build the application for production:

Every push or merge to master will trigger an automatic deployment via Gitlab Pipeline
