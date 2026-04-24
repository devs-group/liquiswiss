# AIgent (VectorChat) API Issues

## 1. Generic error response on chat failure

When `POST /chat/{chatID}/message` fails due to an LLM provider error (invalid model, unsupported parameter), the API returns a generic `{"error":"Chat error"}` with status 500. The actual root cause (e.g. invalid model name, unsupported temperature) is only visible in the server logs, not in the API response.

**Improvement:** Return a more descriptive error message or error code so API consumers can diagnose issues without access to server logs.

## 2. Chatbot creation accepts invalid model names silently

`POST /chat/chatbot` accepts any `model_name` string (e.g. `gpt-4o-mini`) without validating it against available models. The chatbot is created successfully but every message sent to it fails with a 500. The error only surfaces at message-send time, not at chatbot-creation time.

**Improvement:** Validate `model_name` against available models (from `GET /models`) during chatbot creation and reject invalid values with a 400.

## 3. Swagger spec has wrong field names for ChatResponse

The swagger definition `ChatResponse` documents:
```
chat_id, message, context
```

But the actual `POST /chat/{chatID}/message` response returns:
```json
{"response": "...", "session_id": "..."}
```

The fields `response` and `session_id` are not in the spec at all. `chat_id`, `message`, and `context` don't exist in the real response.

**Improvement:** Update the swagger spec to match the actual API response.

## 4. Think tool uses `chat-default` model regardless of chatbot's configured model

When `think_tool_enabled: true` on a chatbot configured with `gpt-4o`, certain queries trigger the think tool which internally uses `Model Group=chat-default`. If the chatbot's `temperature_param` is set to anything other than 1.0, this fails because `chat-default` (GPT-5.2) only supports temperature 1.0.

Simple queries like "Hallo" work (no think tool triggered), but complex queries like "Umsatz in 2025?" fail.

**Workaround:** Set `temperature_param` to 1.0 when creating chatbots.

**Improvement:** The think tool should either use the chatbot's configured model, or handle temperature compatibility automatically.

## 5. Agent loop corrupts tool_calls message history

When the agent loop triggers tool_calls (e.g. think tool) and one of those calls fails (e.g. due to issue #4), the session history is left in a corrupted state. Subsequent messages to the same chatbot — even in a fresh session — fail with:

```
An assistant message with 'tool_calls' must be followed by tool messages responding to each 'tool_call_id'
```

This happens even after deleting and re-creating the chatbot with fresh configuration. Certain queries like "Umsatz in 2025?" consistently trigger tool_calls that fail, making the chatbot unusable for those queries.

**Improvement:** Failed tool_calls should be cleaned up from session history so they don't poison subsequent requests. The agent loop should handle tool_call failures gracefully.
