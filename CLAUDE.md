# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LiquiSwiss is a liquidity planning application with a Go backend and Nuxt 4 frontend. It helps organizations forecast cash flow based on employee salaries, salary costs, transactions, and multi-currency exchange rates.

## Quick Commands

### Frontend (`frontend/`)
```bash
nvm use              # Use correct Node version (required first)
npm install          # Install dependencies
npm run dev          # Dev server at http://localhost:3000
npm run lint:fix     # Lint and fix
npm run build        # Production build
```

### Backend (`backend/`)
```bash
go mod tidy          # Install dependencies
go run .             # Run dev server (or use `air` for hot-reloading)
go test -count=1 ./...  # Run all tests (requires docker compose up)
go test -count=1 ./internal/service/api_service -run TestName  # Run specific test
go generate ./...    # Regenerate mocks
make modernize       # Apply Go modernize suggestions
```

**Important**:
- Always run `go test -count=1 ./...` after backend changes and check if new tests are required.
- Tests require MariaDB running (`docker compose up`) and `.env.local.testing` configured.

### Database
```bash
docker compose up    # Start MariaDB
make goose-static-create <name>   # Create schema migration (from backend/)
make goose-dynamic-create <name>  # Create function/view migration (from backend/)
```

### Running Both Servers in Background

Run these commands in background to have both servers running with real-time logs:

```bash
# Backend (run in background from backend/)
air

# Frontend (run in background from frontend/, requires nvm)
source ~/.nvm/nvm.sh && nvm use && npm run dev
```

**Note**: During hot-reloading, refreshing, or in-between states, you may see transient errors in the logs until the code changes are complete or fixed.

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

**Examples:**
- `Add link field to transactions to enter url`
- `Fix calculation of VAT for quarterly transactions`
- `Update employee form validation`

## Configuration

See `.env.example`, `backend/.env.example`, and `frontend/.env.example` for required environment variables.

## Plans

- Store implementation plans in `docs/plans/`
- Delete plan files once fully implemented

## Context from Previous Sessions

- Check for `claude_chat_history.txt` in root for context from the previous session

## General Guidelines

- **Current year is 2026**: Always search for up-to-date methods and documentation (2025-2026) to prevent outdated implementations. Libraries and frameworks evolve rapidly.
