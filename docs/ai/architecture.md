# Architecture Overview

## System Design

LiquiSwiss uses a layered architecture with clear separation of concerns:

**HTTP Handlers → API Service → Database Adapter**

The backend exposes a REST API consumed by the Nuxt frontend. Both communicate via HTTP-only cookies for authentication.

## Key Entry Points

| Purpose | File |
|---------|------|
| Backend entry point | [backend/main.go](../../backend/main.go) |
| API routes definition | [backend/internal/api/router.go](../../backend/internal/api/router.go) |
| Frontend entry point | [frontend/app/app.vue](../../frontend/app/app.vue) |
| Frontend routing | [frontend/app/pages/](../../frontend/app/pages/) |

## Interface-Based Design

All major components use interfaces for testability:

| Interface | Purpose | Location |
|-----------|---------|----------|
| `IAPIService` | Business logic layer | [backend/internal/service/api_service/api_service.go](../../backend/internal/service/api_service/api_service.go) |
| `IDatabaseAdapter` | Database operations | [backend/internal/adapter/db_adapter/db_adapter.go](../../backend/internal/adapter/db_adapter/db_adapter.go) |
| `ISendgridAdapter` | Email service | [backend/internal/adapter/sendgrid_adapter/sendgrid_adapter.go](../../backend/internal/adapter/sendgrid_adapter/sendgrid_adapter.go) |

## Multi-Tenancy

Users can belong to multiple organizations via the `users_2_organisations` join table. All queries filter by `organisation_id` derived from user context.

## External Services

- **Fixer.io**: Currency exchange rates (synced every 12 hours via cronjob in `main.go`)
- **SendGrid**: Transactional emails
