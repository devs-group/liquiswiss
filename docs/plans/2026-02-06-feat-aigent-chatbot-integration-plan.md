---
title: AIgent Chatbot Integration
type: feat
date: 2026-02-06
---

# AIgent Chatbot Integration

## Overview

Integrate AIgent (VectorChat API) as inline AI chat assistants on LiquiSwiss pages. The Go backend proxies all chat requests to AIgent using OAuth2, injecting page-relevant data as context. Each page gets its own dedicated chatbot (configured via env vars). The chat UI is an expandable card on each page.

## Problem Statement

LiquiSwiss users currently have no AI assistance for data entry, analysis, or navigation. Adding chatbot assistants reduces friction for common tasks: forecasting questions, employee onboarding, transaction categorization, etc.

## Proposed Solution

**Backend proxy architecture**: Go backend authenticates with AIgent via OAuth2 client credentials, proxies user messages with LiquiSwiss data context, and returns responses. No direct frontend-to-AIgent communication.

**Context per message**: Backend gathers relevant page data (employees, transactions, etc.) for the current user/organisation and sends it alongside each message to AIgent.

**Expandable card UI**: Compact card on each page that expands into a chat interface when clicked. One chatbot ID per page.

## Technical Approach

### Phase 1: Backend - AIgent Adapter & Config

#### 1.1 Configuration

Add to `backend/config/config.go`:

```go
// Config struct additions
AigentAPIURL             string  // AIGENT_API_URL
AigentClientID           string  // AIGENT_CLIENT_ID
AigentClientSecret       string  // AIGENT_CLIENT_SECRET
AigentChatIDForecast     string  // AIGENT_CHAT_ID_FORECAST
AigentChatIDEmployees    string  // AIGENT_CHAT_ID_EMPLOYEES
AigentChatIDTransactions string  // AIGENT_CHAT_ID_TRANSACTIONS
AigentChatIDBankAccounts string  // AIGENT_CHAT_ID_BANK_ACCOUNTS
AigentChatIDSettings     string  // AIGENT_CHAT_ID_SETTINGS
```

Update `backend/.env.example` with the new variables.

**Files:**
- `backend/config/config.go`
- `backend/.env.example`

#### 1.2 AIgent Adapter

Create `backend/internal/adapter/aigent_adapter/` following the `ISendgridAdapter` pattern:

```
backend/internal/adapter/aigent_adapter/
  aigent_adapter.go   // IAigentAdapter interface + AigentAdapter struct
```

**Interface:**

```go
type IAigentAdapter interface {
    SendMessage(chatID string, query string, sessionID string) (*AigentChatResponse, error)
    GetToken() error
}
```

**Implementation details:**
- OAuth2 token stored in-memory on the adapter struct with expiry tracking
- Auto-refresh on 401 (retry once after token refresh)
- Uses `net/http` for API calls (consistent with Fixer.io service)
- `POST /public/oauth/token` for token acquisition
- `POST /chat/{chatID}/message` for sending messages

**Models** (in `backend/pkg/models/aigent.go`):

```go
type AigentTokenRequest struct {
    ClientID     string `json:"client_id"`
    ClientSecret string `json:"client_secret"`
}

type AigentTokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
    Scope       string `json:"scope"`
    TokenType   string `json:"token_type"`
}

type AigentChatRequest struct {
    Query     string `json:"query"`
    SessionID string `json:"session_id"`
}

type AigentChatResponse struct {
    ChatID  string `json:"chat_id"`
    Message string `json:"message"`
    Context string `json:"context"`
}
```

**Files:**
- `backend/internal/adapter/aigent_adapter/aigent_adapter.go` (new)
- `backend/pkg/models/aigent.go` (new)

#### 1.3 Mockgen

Add `//go:generate mockgen` directive for `IAigentAdapter` so tests can mock the adapter.

### Phase 2: Backend - Chat Handler & Routes

#### 2.1 Chat Handler

Create `backend/internal/api/handlers/chat.go`:

- `SendChatMessage(apiService, aigentAdapter, c *gin.Context)` handler
- Accepts: `{ page: string, message: string, session_id?: string }`
- Validates `page` is one of: forecast, employees, transactions, bank-accounts, settings
- Maps `page` → chatbot ID from config
- Gathers page context data via `apiService` (e.g. for employees page, fetches employee list summary)
- Prepends context to user message
- Calls `aigentAdapter.SendMessage(chatID, enrichedMessage, sessionID)`
- Returns: `{ message: string, session_id: string }`

**Context gathering per page:**

| Page | Context Data |
|---|---|
| forecast | Bank account balances, upcoming forecasts summary |
| employees | Employee count, recent hires, salary summary |
| transactions | Transaction summary, category breakdown |
| bank-accounts | Account list with balances, currencies |
| settings | Organisation info, current user role |

#### 2.2 Routes

Add to `backend/internal/api/router.go` in the `protected` group:

```go
protected.POST("/chat", func(ctx *gin.Context) {
    handlers.SendChatMessage(api.APIService, api.AigentService, ctx)
})
```

**Wire in** the aigent adapter:
- Add `AigentService aigent_adapter.IAigentAdapter` to the `API` struct
- Pass it through `NewAPI()`
- Instantiate in `main.go`

**Files:**
- `backend/internal/api/handlers/chat.go` (new)
- `backend/internal/api/router.go` (modify)
- `backend/main.go` (modify)

### Phase 3: Frontend - Chat Composable & Component

#### 3.1 Chat Composable

Create `frontend/app/composables/useChat.ts`:

```typescript
export default function useChat(page: string) {
    const messages = useState<ChatMessage[]>(`chat-messages-${page}`, () => [])
    const sessionId = useState<string | null>(`chat-session-${page}`, () => null)
    const isLoading = useState(`chat-loading-${page}`, () => false)
    const isExpanded = useState(`chat-expanded-${page}`, () => false)

    const sendMessage = async (query: string) => { ... }  // POST /api/chat via $fetch
    const clearChat = () => { ... }

    return { messages, sessionId, isLoading, isExpanded, sendMessage, clearChat }
}
```

**Key rules:**
- Use `$fetch` for `sendMessage` (it's a POST mutation)
- Use `useState` for messages/sessionId (survives client-side navigation, lost on full reload)
- Session resets when organisation changes (app does `reloadNuxtApp({ force: true })`)
- All error messages in German

**Models** in `frontend/app/models/chat.ts`:

```typescript
interface ChatMessage {
    role: 'user' | 'assistant'
    content: string
    timestamp: Date
}

interface ChatRequest {
    page: string
    message: string
    session_id?: string
}

interface ChatResponse {
    message: string
    session_id: string
}
```

#### 3.2 Chat Component

Create `frontend/app/components/ChatCard.vue`:

**Collapsed state:** Compact card with AI icon and "Fragen Sie den Assistenten..." placeholder. Click to expand.

**Expanded state:**
- Message list (scrollable, auto-scroll to bottom)
- Text input + send button
- Clear conversation button
- Close/collapse button
- Loading indicator while waiting for response
- Error display in chat bubble format

**PrimeVue components used:** `Card`, `InputText`, `Button`, `ProgressSpinner`

**Props:** `page: string` (determines which chatbot to use)

**File:** `frontend/app/components/ChatCard.vue` (new)

#### 3.3 Integrate on Pages

Add `<ChatCard page="forecast" />` (etc.) to each page template:

- `frontend/app/pages/index.vue` → `page="forecast"`
- `frontend/app/pages/employees/index.vue` → `page="employees"`
- `frontend/app/pages/transactions/index.vue` → `page="transactions"`
- `frontend/app/pages/bank-accounts/index.vue` → `page="bank-accounts"`
- `frontend/app/pages/settings.vue` → `page="settings"`

Place the `<ChatCard>` at the bottom of each page template, after existing content.

### Phase 4: Testing

#### Backend Tests

Create `backend/internal/api/handlers/chat_test.go`:

- Test message sending with mocked AIgent adapter
- Test invalid page parameter returns 400
- Test missing message returns 400
- Test AIgent adapter failure returns 500
- Test user authentication required (middleware coverage)

**Mock:** Use mockgen-generated mock for `IAigentAdapter`

#### Frontend

Manual testing via dev server. Chat component renders, sends messages, displays responses, handles errors.

## Progress

**Last updated:** 2026-02-06 22:50

**Status:** Blocked - chat endpoint returns 500

**Architecture change:** Moved from global config chat IDs to per-organisation chatbots. Each org gets 5 chatbots auto-provisioned via AIgent API on creation, stored in `organisation_chatbots` DB table. Existing orgs get a "Jetzt KI aktivieren" button on the Organisation page.

**Completed this session:**
- [x] Removed 5 per-page `AigentChatID*` fields from config (keep only API URL, client ID, client secret)
- [x] Created `organisation_chatbots` DB table migration (00038)
- [x] Added `AigentCreateChatbotRequest/Response` and `OrganisationChatbot` models
- [x] Added `CreateChatbot` method to `IAigentAdapter` with 401-retry pattern
- [x] Added `CreateOrganisationChatbot`, `GetOrganisationChatbot`, `HasOrganisationChatbots` to DB adapter
- [x] Created German system prompts per page (`system_prompts.go`)
- [x] Wired `aigentAdapter` into `APIService` struct + constructor (3rd param, `nil` in tests)
- [x] `CreateOrganisation` now auto-provisions 5 chatbots via `provisionChatbots()`
- [x] Added `GetOrganisationChatbot`, `HasOrganisationChatbots`, `ProvisionOrganisationChatbots` to `IAPIService`
- [x] Chat handler uses DB lookup instead of config for chatbot IDs
- [x] Added `GET /api/organisations/chatbots/status` and `POST /api/organisations/:organisationID/chatbots/provision` endpoints
- [x] Added "Jetzt KI aktivieren" button to `organisation.vue` for existing orgs
- [x] Created `ChatWidget.vue` component (inline expandable input, not floating FAB)
- [x] Added `ChatWidget` to forecast page (`index.vue`) above the table
- [x] Updated all 13 test files with `NewAPIService` 3rd param
- [x] Regenerated all mocks (aigent, db, api_service)
- [x] `go vet ./...` passes
- [x] `go test -count=1 ./...` all pass
- [x] Chatbot provisioning tested manually - 5 chatbots stored in DB for org 1

**Blocker: `POST /api/chat` returns 500**
- DB lookup works (chatbot IDs exist in `organisation_chatbots` for org 1)
- The error is in `aigentService.SendMessage()` - the AIgent API call itself fails
- The chat handler at line 67-69 of `chat.go` catches the error but does NOT log it
- **Fix needed:** Add `logger.Logger.Error(err)` before the 500 response in the chat handler so the actual AIgent API error is visible in logs
- Could be: wrong AIgent API URL, token issue, or chatbot ID format mismatch

**Next steps:**
1. Add error logging in chat handler (`chat.go` line 67) to see the actual AIgent error
2. Debug the `SendMessage` call - check AIgent API URL, token acquisition, request format
3. Verify the chatbot IDs from DB match what AIgent expects (UUID format)
4. Once chat works on forecast page, add `ChatWidget` to remaining 4 pages
5. Consider adding a `useChat` composable for shared state across pages (optional)
6. Run lint, tests, commit, push

**Modified files (uncommitted):**
- `backend/config/config.go` - Removed 5 chat ID fields
- `backend/.env.example` - Removed 5 chat ID env vars
- `backend/pkg/models/aigent.go` - Added create chatbot + DB models
- `backend/internal/adapter/aigent_adapter/aigent_adapter.go` - Added CreateChatbot
- `backend/internal/adapter/aigent_adapter/system_prompts.go` - New: system prompts
- `backend/internal/adapter/db_adapter/organisation_chatbot.go` - New: DB methods
- `backend/internal/adapter/db_adapter/db_adapter.go` - Added interface methods
- `backend/internal/adapter/db_adapter/queries/create_organisation_chatbot.sql` - New
- `backend/internal/adapter/db_adapter/queries/get_organisation_chatbot.sql` - New
- `backend/internal/adapter/db_adapter/queries/has_organisation_chatbots.sql` - New
- `backend/internal/db/migrations/static/00038_create-table_organisation_chatbots.sql` - New
- `backend/internal/service/api_service/api_service.go` - Added aigent + chatbot methods
- `backend/internal/service/api_service/organisation.go` - Provisioning + chatbot lookup
- `backend/internal/api/handlers/chat.go` - DB lookup instead of config
- `backend/internal/api/handlers/organisations.go` - Status + provision handlers
- `backend/internal/api/router.go` - New chatbot routes
- `backend/main.go` - Pass aigent to api_service
- `backend/internal/mocks/` - All regenerated
- `backend/internal/api/handlers/*_test.go` - All updated with nil 3rd param
- `backend/internal/service/api_service/forecast_test.go` - Updated
- `frontend/app/components/ChatWidget.vue` - New: inline chat component
- `frontend/app/pages/index.vue` - Added ChatWidget
- `frontend/app/pages/organisation.vue` - Added KI activation button

## Acceptance Criteria

- [ ] Backend proxies chat messages to AIgent via OAuth2
- [ ] OAuth token auto-refreshes on expiry (transparent to user)
- [ ] Each page sends page-specific context data with messages
- [ ] Expandable chat card appears on 5 pages (forecast, employees, transactions, bank-accounts, settings)
- [ ] Messages persist in chat during client-side navigation (useState)
- [ ] Chat resets on organisation switch
- [ ] Loading state shown while waiting for response
- [ ] Error messages shown in German when AIgent is unavailable
- [ ] Clear conversation button resets chat
- [ ] Backend tests pass with mocked AIgent adapter
- [ ] Environment variables documented in `.env.example`
- [ ] Send button disabled while loading (prevent duplicate sends)

## Files Summary

### New Files
- `backend/internal/adapter/aigent_adapter/aigent_adapter.go`
- `backend/pkg/models/aigent.go`
- `backend/internal/api/handlers/chat.go`
- `backend/internal/api/handlers/chat_test.go`
- `frontend/app/composables/useChat.ts`
- `frontend/app/components/ChatCard.vue`
- `frontend/app/models/chat.ts`

### Modified Files
- `backend/config/config.go` - Add AIgent config fields
- `backend/.env.example` - Add AIgent env vars
- `backend/main.go` - Instantiate AIgent adapter, pass to API
- `backend/internal/api/router.go` - Add chat route, accept AIgent adapter
- `frontend/app/pages/index.vue` - Add ChatCard
- `frontend/app/pages/employees/index.vue` - Add ChatCard
- `frontend/app/pages/transactions/index.vue` - Add ChatCard
- `frontend/app/pages/bank-accounts/index.vue` - Add ChatCard
- `frontend/app/pages/settings.vue` - Add ChatCard

## Verification

1. Start backend with AIgent env vars configured → no startup errors
2. Open any page → see collapsed chat card
3. Click card → expands to chat UI
4. Type message → loading state → response appears
5. Navigate to another page → different chatbot, chat preserved per page
6. Switch organisation → all chats reset
7. Stop AIgent service → send message → error shown in German
8. Run `go test -count=1 ./...` → all tests pass
