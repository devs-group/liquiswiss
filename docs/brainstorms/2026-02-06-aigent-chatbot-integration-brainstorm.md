# AIgent Chatbot Integration

**Date:** 2026-02-06

## What We're Building

Integrate AIgent (VectorChat API) as inline AI chat assistants on LiquiSwiss pages. Each page gets its own dedicated chatbot that can analyze data AND help users create/edit records (employees, transactions, bank accounts, etc.).

The Go backend proxies all requests to the AIgent API using OAuth authentication, keeping API credentials server-side.

## Why This Approach

- **Backend proxy** over widget embed: Keeps AIgent credentials secure, allows injecting LiquiSwiss-specific context (user data, org data) into requests, and maintains control over the data flow.
- **Inline on page** over floating widget: Each page's chatbot is purpose-built for that domain. Inline placement makes the assistant feel like a native part of the page, not a generic overlay.
- **One chatbot per page**: Specialized chatbots configured in AIgent with domain-specific knowledge (e.g. employee management, transaction analysis, forecast reporting) rather than one generic bot.
- **Simple request/response first**: Avoids SSE/streaming complexity. Can add streaming later.
- **Environment variables for config**: Chatbot IDs configured at deploy time via env vars. Simple and sufficient for v1.

## Key Decisions

1. **Integration style**: Backend proxy (Go backend -> AIgent API via OAuth)
2. **Auth flow**: `POST /public/oauth/token` with client_id/client_secret to get Bearer token
3. **Chat endpoint**: `POST /chat/{chatID}/message` (non-streaming)
4. **UI placement**: Inline chat component on each page
5. **Chatbot mapping**: One AIgent chatbot ID per page, configured via env vars
6. **Capabilities**: Full read/write - chatbot can both analyze and help create/edit data
7. **Session management**: Use AIgent's `session_id` to maintain conversation continuity per user per page

## AIgent API Endpoints Used

| Endpoint | Method | Auth | Purpose |
|---|---|---|---|
| `/public/oauth/token` | POST | None | Get Bearer token (client credentials) |
| `/chat/{chatID}/message` | POST | Bearer | Send message, get response |
| `/chat/{chatID}/stream-message` | POST | Bearer | Future: streaming responses |
| `/api/conversations` | GET | Bearer | List conversation history |
| `/api/chatbots/{chatbotId}/conversations` | POST | Bearer | Create new conversation |

## Key Models

**ChatMessageRequest**: `{ query, session_id, model_override?, attachments?, voice_mode? }`
**ChatResponse**: `{ chat_id, message, context }`
**TokenRequest**: `{ client_id, client_secret }`
**TokenResponse**: `{ access_token, expires_in, scope, token_type }`

## Pages & Chatbot Mapping

| Page | Env Var | Chatbot Purpose |
|---|---|---|
| Forecast (index) | `AIGENT_CHATBOT_FORECAST` | Cash flow analysis, projections |
| Employees | `AIGENT_CHATBOT_EMPLOYEES` | Employee data entry, salary analysis |
| Transactions | `AIGENT_CHATBOT_TRANSACTIONS` | Transaction entry, categorization |
| Bank Accounts | `AIGENT_CHATBOT_BANK_ACCOUNTS` | Account management, balance analysis |
| Settings | `AIGENT_CHATBOT_SETTINGS` | General help, configuration guidance |

## Open Questions

- How should conversation history persist? Per-user per-page session stored in LiquiSwiss DB, or rely entirely on AIgent's session management?
- Should there be a "clear conversation" button?
- What LiquiSwiss context should be sent with each message? (e.g. current org data, page-specific data)
