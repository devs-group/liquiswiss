# Authentication

## JWT Dual-Token Flow

LiquiSwiss uses two JWT tokens stored in HTTP-only cookies:

| Token | Lifetime | Purpose |
|-------|----------|---------|
| Access Token | 15 minutes | API authentication |
| Refresh Token | 3 months | Session persistence, tracked in database for revocation |

## Flow

1. **Login**: User provides email/password → Backend issues both tokens as HTTP-only cookies
2. **API Requests**: Access token validates each request
3. **Auto-refresh**: Backend middleware automatically refreshes expired access tokens if refresh token is valid
4. **Logout**: Refresh token is blacklisted in `refresh_tokens` database table

## Key Files

| Purpose | File |
|---------|------|
| JWT generation & verification | [backend/pkg/auth/auth.go](../../backend/pkg/auth/auth.go) |
| Auth middleware | [backend/internal/middleware/auth.go](../../backend/internal/middleware/auth.go) |
| Auth handlers | [backend/internal/api/handlers/auth.go](../../backend/internal/api/handlers/auth.go) |
| Frontend auth composable | [frontend/app/composables/useAuth.ts](../../frontend/app/composables/useAuth.ts) |
| Frontend auth middleware | [frontend/app/middleware/auth.global.ts](../../frontend/app/middleware/auth.global.ts) |

## Debugging Auth Issues

1. Check if refresh token exists in `refresh_tokens` table
2. Verify `JWT_KEY` is consistent between environments
3. Check cookie settings (SameSite, Secure flags)
