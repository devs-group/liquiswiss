---
title: "feat: Replace SendGrid with SMTP + add Mailpit and currency fallback"
type: feat
status: active
date: 2026-04-26
origin: https://github.com/devs-group/liquiswiss/issues/24
---

# feat: Replace SendGrid with SMTP + add Mailpit and currency fallback

## Overview

Remove `sendgrid-go` SDK lock-in. Switch outbound mail to plain SMTP with locally rendered Go HTML templates. Add Mailpit to local compose so dev mail flows are inspectable. Add a static JSON fallback for fiat rates when no Fixer.io key is configured. Drop `SEND_GRID_TOKEN` and `SEND_GRID_TEMPLATE_ID` env vars.

## Problem Frame

Current setup binds the app to SendGrid (API + dynamic templates) and Fixer.io (API key required). Onboarding new contributors requires either provider accounts or accepting broken mail flows in dev (only a log line shows the action URL). Recent commits (`e5d12d2`, `cde7d45`) added graceful degradation but left the dev mail experience unusable.

This plan supersedes two earlier sketches: `docs/plans/smtp-email-adapter.md` and `docs/plans/mock-fixer-io-service.md`. Differences from those sketches:
- No SendGrid dual-provider factory. SendGrid SDK removed entirely. Production uses SMTP relay (SendGrid SMTP, Brevo, SES, etc.) — provider chosen at deploy time, not at code time.
- Currency fallback is implicit (`FIXER_IO_KEY` unset → static JSON), no explicit `USE_MOCK_FIXER_IO` flag.

## Requirements Trace

- R1. `git clone && make up` (no BWSM, no external accounts) → invitation flow shows full HTML email in Mailpit UI at `http://localhost:8025`
- R2. Same fresh clone → fiat rates load from local fallback, currency conversions work in UI
- R3. Existing prod deploy continues to work after migration (SMTP creds via BWSM, optional Fixer.io key still used)
- R4. No code reference to `sendgrid-go`; templates live in repo
- R5. BWSM `liquiswiss-prod` updated with SMTP creds; old SendGrid secrets removed
- R6. `CLAUDE.md` mentions Mailpit + local rates fallback

## Scope Boundaries

- DKIM/SPF tuning for production SMTP relay (handled per-deployment ops)
- Migrating existing template content edits made in SendGrid GUI (re-author from scratch in Go templates — current production templates are German invitation/registration/reset, content already mirrored in current Go code)
- Adding new email types beyond the existing three (registration, password reset, invitation)
- Localization framework for templates (German only, matches current behavior)

### Deferred to Separate Tasks

- BWSM secret rotation (`liquiswiss-prod`): add `SMTP_*` keys, remove `SEND_GRID_*` keys — done as ops step alongside deploy, not in this PR
- Production deploy: choosing actual SMTP relay provider and credentials

## Context & Research

### Relevant Code and Patterns

- `backend/internal/adapter/sendgrid_adapter/sendgrid_adapter.go` — current adapter; defines `ISendgridAdapter` with `SendRegistrationMail`, `SendPasswordResetMail`, `SendInvitationMail`. Templates assembled inline as `models.SendgridMail` struct (Subject, PreHeader, Hello, Content, ButtonText, ButtonUrl, Greetings).
- `backend/internal/service/fixer_io_service/fixer_io_service.go` — current Fixer.io fetch. `FetchFiatRates()` already early-returns when `FIXER_IO_KEY == ""` (commit `e5d12d2`). `RequiresInitialFetch()` checks DB.
- `backend/main.go:86-92` — wiring point; `NewSendgridAdapter`, `NewAPIService`, `NewAPI`, `NewFixerIOService`.
- `backend/config/config.go` — flat `Config` struct, `getEnv` helper. Add SMTP fields here, remove SendGrid fields.
- `backend/pkg/models/mail.go` — `SendgridMail` struct (rename to neutral name).
- `backend/internal/mocks/sendgrid_adapter.go` — generated mock. After interface rename + package move, regenerate via `go generate ./...`.
- Mock regeneration directive: `//go:generate mockgen -package=mocks -destination ../../mocks/...`
- Test files referencing `ISendgridAdapter` mock: `backend/internal/api/handlers/{users,member,transactions,employees,bank_account,forecast_integration,invitation,currencies,salary_costs,salaries,main}_test.go` (11 files).
- `docker-compose.yml` — root local compose; add `mailpit` service.
- `_deployment/docker-compose.yml` — prod-mirror compose; add SMTP env vars to backend, do **not** add Mailpit.
- No `.env.example` exists — env vars come from compose defaults + BWSM. Document new vars in CLAUDE.md instead.
- BWSM project `liquiswiss-dev` — local injection. Currently empty for `SEND_GRID_*` and `FIXER_IO_KEY` per CLAUDE.md.

### Institutional Learnings

- From CLAUDE.md: `air` runs with `--no-migrate`. Not relevant here (no schema changes).
- From memory: when adding params to `NewAPIService`, all 13+ test files must be updated. Same care needed when renaming the email adapter param.
- From memory: always `git push` after commit per CLAUDE.md.

### External References

- Mailpit (`axllent/mailpit`) — drop-in SMTP server with web UI; no auth needed for local. SMTP `1025`, web `8025`.
- Go `html/template` — stdlib. Sufficient for three transactional templates.
- Go `net/smtp` — stdlib but RFC-strict. `gomail.v2` (or `wneessen/go-mail`, more current) handles MIME, attachments, TLS, modern auth — preferred. Decide in Phase 2.

## Key Technical Decisions

- **D1. Single SMTP adapter, no dual-provider factory.** Production reaches SendGrid (or any other provider) via SMTP relay. Fewer code paths, one less dependency, identical behavior locally and in prod. Removes `sendgrid-go` from `go.mod`.
- **D2. Templates as Go `html/template` files embedded via `//go:embed`.** Keeps templates version-controlled, reviewable, testable. One base layout (`base.tmpl`) + per-type partial (`registration.tmpl`, `password_reset.tmpl`, `invitation.tmpl`). Inline CSS, table-based, 600px max-width for client compatibility.
- **D3. Implicit Fixer.io fallback.** When `FIXER_IO_KEY` is set, fetch from API. When unset, load `fallback_rates.json` and upsert via existing `apiService.UpsertFiatRate`. No new env flag; matches existing graceful-degradation pattern (`e5d12d2`).
- **D4. Use `wneessen/go-mail` library.** Modern, well-maintained, simpler MIME handling than stdlib `net/smtp`, supports STARTTLS + implicit TLS + plain, optional auth. Confirm in Phase 2 if a simpler stdlib-only path is preferred.
- **D5. Adapter rename: `sendgrid_adapter` → `email_adapter`, `ISendgridAdapter` → `IEmailAdapter`.** Reflects provider neutrality. Method signatures unchanged (`SendRegistrationMail`, `SendPasswordResetMail`, `SendInvitationMail`) so callers in `api_service`/handlers need only import + type rename.
- **D6. Environment variable contract.**
  - Add: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM_ADDRESS`, `SMTP_FROM_NAME`, `SMTP_TLS` (`off|starttls|implicit`).
  - Remove: `SEND_GRID_TOKEN`, `SEND_GRID_TEMPLATE_ID`.
  - Local defaults in compose: `SMTP_HOST=mailpit`, `SMTP_PORT=1025`, no auth, `SMTP_TLS=off`, `SMTP_FROM_ADDRESS=no-reply@liquiswiss.local`.
- **D7. Mailpit not added to `_deployment/docker-compose.yml`.** Local-only service.
- **D8. Fallback rates source.** Bundle a static JSON snapshot of ~30 currencies covering currencies seeded in DB. Manual refresh policy (drift acceptable for local dev). Commit script optional — out of scope this plan.

## Open Questions

### Resolved During Planning

- Q: Keep dual SendGrid+SMTP factory? → No (D1).
- Q: Explicit `USE_MOCK_FIXER_IO` flag? → No, key-presence drives behavior (D3).
- Q: Templates in DB or files? → Files, embedded via `//go:embed` (D2).
- Q: stdlib `net/smtp` or library? → `wneessen/go-mail` (D4) pending Phase 2 confirmation.
- Q: Add Mailpit to prod-mirror compose? → No (D7).

### Deferred to Implementation

- Exact module path for embedded templates (likely `backend/internal/adapter/email_adapter/templates/`).
- Test strategy for SMTP send: stand up an in-process SMTP listener (e.g. `github.com/mhale/smtpd`) inside Go tests vs. relying on the compose Mailpit during integration tests. Pick during Unit 4.
- Whether to keep `models.SendgridMail` struct under a renamed name or split per-template structs. Resolve when authoring templates.
- Final fallback rate values + currency list. Snapshot from a one-off Fixer.io call or hand-curated. Resolve in Unit 6.

## High-Level Technical Design

> *Directional guidance for review, not implementation specification.*

```
┌─────────────────────────────────────────────────────────┐
│ caller (api_service / handler)                          │
│    .SendInvitationMail(email, token, org, invitedBy)    │
└─────────────────┬───────────────────────────────────────┘
                  │ IEmailAdapter
                  ▼
        ┌─────────────────────┐
        │ smtpEmailAdapter    │   templates/  (embed.FS)
        │   - render(tmpl,    │ ◄──── base.tmpl
        │     data) → []byte  │       registration.tmpl
        │   - send(to, body)  │       password_reset.tmpl
        └─────────┬───────────┘       invitation.tmpl
                  │ go-mail
                  ▼
        ┌─────────────────────┐
        │ SMTP host:port      │
        │  local: mailpit:1025│
        │  prod : <relay>:587 │
        └─────────────────────┘

graceful skip: SMTP_HOST=="" → log warn, return nil (mirror current sendgrid behavior)
```

```
┌──────────────────────────────────────────────┐
│ FixerIOService.FetchFiatRates()              │
│                                              │
│   if FIXER_IO_KEY == "":                     │
│       loadFallbackJSON() → upsertAll()       │
│       return                                 │
│                                              │
│   <existing Fixer.io HTTP path>              │
└──────────────────────────────────────────────┘
```

## Implementation Units

- [ ] **Unit 1: Add Mailpit to local compose**

**Goal:** Local SMTP capture target running on `make up`.

**Requirements:** R1

**Dependencies:** None

**Files:**
- Modify: `docker-compose.yml`

**Approach:**
- Add `mailpit` service with image `axllent/mailpit:latest` (pin minor in PR).
- Expose `1025:1025` (SMTP) and `8025:8025` (web UI).
- Add `MP_MAX_MESSAGES=500` to keep memory bounded.
- Backend service depends on `mailpit` (no healthcheck needed; SMTP send tolerates startup race).
- Add SMTP env vars to backend service block with local defaults pointing at `mailpit:1025`.

**Patterns to follow:**
- Existing `database` / `phpmyadmin` service blocks in `docker-compose.yml`.

**Test scenarios:**
- Integration: `make up` → `curl -fsS http://localhost:8025` returns Mailpit UI HTML.
- Integration: `nc -zv localhost 1025` succeeds after compose up.

**Verification:**
- Mailpit UI reachable, SMTP port open, backend container has `SMTP_HOST=mailpit` in environment.

---

- [ ] **Unit 2: Config and env contract**

**Goal:** Add SMTP config fields, remove SendGrid fields.

**Requirements:** R3, R4

**Dependencies:** None

**Files:**
- Modify: `backend/config/config.go`
- Modify: `_deployment/docker-compose.yml` (add SMTP env vars to backend service)
- Modify: `docker-compose.yml` (add SMTP env vars to backend service if not already in Unit 1)

**Approach:**
- `Config` adds: `SMTPHost`, `SMTPPort` (int), `SMTPUser`, `SMTPPass`, `SMTPFromAddress`, `SMTPFromName`, `SMTPTLS` (string: `off|starttls|implicit`, default `off`).
- Remove `SendgridToken`, `SendgridTemplateID`.
- Keep `FixerIOURl`, `FixerIOKey` unchanged.
- Document new env vars in CLAUDE.md (Unit 7) since `.env.example` no longer exists; compose provides local defaults.

**Patterns to follow:**
- Existing `getEnv` helper, flat struct shape.

**Test scenarios:**
- Test expectation: none — pure config plumbing, exercised via integration in later units.

**Verification:**
- `go build ./...` passes after this unit alone fails because callers still reference removed fields — that is expected; downstream units fix callers. Land Unit 2 + 3 together if breakage is undesirable.

---

- [ ] **Unit 3: Email adapter package + SMTP implementation**

**Goal:** Replace `sendgrid_adapter` with `email_adapter` (SMTP-backed).

**Requirements:** R1, R3, R4

**Dependencies:** Unit 2

**Files:**
- Create: `backend/internal/adapter/email_adapter/email_adapter.go` (interface + constructor)
- Create: `backend/internal/adapter/email_adapter/smtp.go` (SMTP send logic via `go-mail`)
- Create: `backend/internal/adapter/email_adapter/templates.go` (embed + render helpers)
- Create: `backend/internal/adapter/email_adapter/templates/base.tmpl`
- Create: `backend/internal/adapter/email_adapter/templates/registration.tmpl`
- Create: `backend/internal/adapter/email_adapter/templates/password_reset.tmpl`
- Create: `backend/internal/adapter/email_adapter/templates/invitation.tmpl`
- Create: `backend/internal/adapter/email_adapter/email_adapter_test.go`
- Modify: `backend/pkg/models/mail.go` (rename `SendgridMail` → `EmailContent`, or keep as legacy alias)
- Delete: `backend/internal/adapter/sendgrid_adapter/sendgrid_adapter.go` (and the directory)
- Modify: `backend/go.mod`, `backend/go.sum` (add `go-mail`, remove `sendgrid-go`)

**Approach:**
- Define `IEmailAdapter` with the existing three method signatures verbatim: `SendRegistrationMail(email, code string) error`, `SendPasswordResetMail(email, code string) error`, `SendInvitationMail(email, token, organisationName, invitedByName string) error`. Keeping signatures stable means callers and mocks need only import-path + type-name updates.
- `NewEmailAdapter(cfg config.Config) IEmailAdapter` — embedded templates parsed once at construction.
- Graceful skip: if `cfg.SMTPHost == ""`, log warn (mirror current behavior at `sendgrid_adapter.go:34-38`) and return nil from each Send method.
- Render flow: `executeTemplate("invitation", data)` → produces full HTML body → SMTP send with `Content-Type: text/html; charset=utf-8`.
- `SMTP_TLS` mapping: `off` → plain, `starttls` → STARTTLS upgrade, `implicit` → TLS from connect.
- Auth: skip when `SMTP_USER == ""` (Mailpit).
- Templates carry the same German copy as current Go code; reuse strings verbatim from `sendgrid_adapter.go:88-99`, `124-135`, `159-171`.

**Execution note:** Test-first for the rendering and graceful-skip paths — they are pure functions easy to TDD. SMTP wire path covered separately by integration test in Unit 4.

**Patterns to follow:**
- Mock generation directive at top of `sendgrid_adapter.go` — replicate for `email_adapter`.
- Logger usage pattern: `logger.Logger.Warnw(...)`, `logger.Logger.Error(...)`.
- `models.SendgridMail` template variable shape (Subject, PreHeader, Hello, Content, ButtonText, ButtonUrl, Greetings) — preserve, just rename type.

**Test scenarios:**
- Happy path: `SendInvitationMail` produces HTML containing the invited org name, invited-by name, and the constructed `auth/invitation?token=...` URL.
- Happy path: `SendRegistrationMail` URL contains `/auth/validate?email=...&code=...` and validity-hours phrasing.
- Happy path: `SendPasswordResetMail` URL contains `/auth/reset-password?email=...&code=...`.
- Edge case: `SMTPHost == ""` → all three Send methods return `nil` and emit a warn log; no SMTP connection attempted.
- Edge case: template rendering of HTML-special chars in `invitedByName` (e.g. `O'Brien`) does not produce malformed HTML (relies on `html/template` auto-escaping).
- Error path: SMTP server unreachable → `Send*` returns wrapped error; caller can inspect.
- Error path: missing template name → constructor fails, surfaces clear error.

**Verification:**
- `go test ./internal/adapter/email_adapter/...` passes.
- `grep -ri "sendgrid" backend/` returns no matches outside this plan.
- `go.mod` no longer imports `github.com/sendgrid/sendgrid-go`.

---

- [ ] **Unit 4: Wire adapter into main.go + regenerate mocks + fix tests**

**Goal:** App boots with new adapter; all tests pass.

**Requirements:** R1, R3, R4

**Dependencies:** Unit 3

**Files:**
- Modify: `backend/main.go` (replace `sendgrid_adapter.NewSendgridAdapter(cfg.SendgridToken)` with `email_adapter.NewEmailAdapter(cfg)`)
- Modify: `backend/internal/service/api_service/api_service.go` (rename param type `ISendgridAdapter` → `IEmailAdapter`, update import)
- Modify: `backend/internal/service/api_service/auth.go`, `backend/internal/service/api_service/invitation.go` (adjust references)
- Modify: `backend/internal/api/router.go` (adjust imports/types)
- Regenerate: `backend/internal/mocks/sendgrid_adapter.go` → `backend/internal/mocks/email_adapter.go` (delete old, generate new via `go generate ./...`)
- Modify: `backend/internal/api/handlers/users_test.go`, `member_test.go`, `transactions_test.go`, `employees_test.go`, `bank_account_test.go`, `forecast_integration_test.go`, `invitation_test.go`, `currencies_test.go`, `salary_costs_test.go`, `salaries_test.go`, `main_test.go` (rename mock references — `MockISendgridAdapter` → `MockIEmailAdapter`)

**Approach:**
- Wholesale find-and-replace across 11 test files for the mock type name and import path.
- Regenerate mocks: `cd backend && go generate ./...`.
- `NewAPIService` and `NewAPI` parameter names/positions stay identical — only the type changes. No call-site arity changes.

**Patterns to follow:**
- Memory: `NewAPIService` signature changes propagate to 13+ files. Same care here for mock-type rename.

**Test scenarios:**
- Integration: full backend test suite (`go test -count=1 ./...`) passes against MariaDB.
- Integration: end-to-end invitation flow test sends mail via SMTP test listener and asserts on captured message body.

**Verification:**
- `go vet ./...` clean.
- `go test -count=1 ./...` green.
- Manual: `make up` then trigger an invitation; confirm email visible in `http://localhost:8025`.

---

- [ ] **Unit 5: Fixer.io fallback path**

**Goal:** When `FIXER_IO_KEY` unset, seed fiat rates from bundled JSON.

**Requirements:** R2, R3

**Dependencies:** None (parallel with Units 1–4)

**Files:**
- Create: `backend/internal/service/fixer_io_service/fallback_rates.json`
- Create: `backend/internal/service/fixer_io_service/fallback.go` (embed + load + parse)
- Modify: `backend/internal/service/fixer_io_service/fixer_io_service.go` (replace early-return on missing key with fallback path)
- Create: `backend/internal/service/fixer_io_service/fallback_test.go`

**Approach:**
- Embed JSON via `//go:embed fallback_rates.json`.
- JSON shape: `{"rates": [{"base":"CHF","target":"EUR","rate":1.04}, ...]}` (matches old plan, accepted shape).
- In `FetchFiatRates()`: if `cfg.FixerIOKey == ""`, parse fallback JSON, iterate, call `f.apiService.UpsertFiatRate(...)`. Log info that fallback path is in use. Return.
- Existing non-prod skip-when-rates-exist behavior remains intact (lines 44-50 of `fixer_io_service.go`); fallback only runs when DB is empty *and* no key.
- Currency coverage: include all currencies seeded by static migrations (verify list against `currencies` table seed). Aim for ~30 pairs (subset of full N×N matrix — only pairs where base is in seeded currencies).

**Patterns to follow:**
- Existing `UpsertFiatRate` call shape at `fixer_io_service.go:103-107`.
- `models.CreateFiatRate` struct.

**Test scenarios:**
- Happy path: with `FIXER_IO_KEY=""`, `FetchFiatRates()` upserts every pair from `fallback_rates.json` into DB.
- Happy path: with `FIXER_IO_KEY=set`, `FetchFiatRates()` ignores fallback and hits HTTP path (existing behavior).
- Edge case: malformed `fallback_rates.json` → `FetchFiatRates()` logs error and returns without panic.
- Edge case: `RequiresInitialFetch()` reports true on empty DB; fallback path satisfies it on next call.
- Integration: fresh `make up` with `FIXER_IO_KEY=""` → DB has at least N fiat rates, currency conversion API endpoint returns valid result for `CHF→EUR`.

**Verification:**
- `go test ./internal/service/fixer_io_service/...` passes.
- Fresh DB after `make up` (no key): `SELECT COUNT(*) FROM fiat_rates` > 0.

---

- [ ] **Unit 6: Curate fallback rate snapshot**

**Goal:** Realistic, dated rates committed.

**Requirements:** R2

**Dependencies:** Unit 5

**Files:**
- Modify: `backend/internal/service/fixer_io_service/fallback_rates.json`

**Approach:**
- Source: one-off `curl` against Fixer.io with the prod key (run manually by maintainer), or hand-curate against published reference rates as of plan date.
- Include a `"snapshot_date"` field at JSON top level for future drift tracking.
- Cover the seeded currencies in the `currencies` migration table; leave less-common targets out.

**Patterns to follow:**
- Static fixture style — keep file small and human-readable.

**Test scenarios:**
- Test expectation: none — pure data file. Covered via Unit 5 happy-path test asserting count > 0 and `CHF→EUR` rate parses as float.

**Verification:**
- File parses as valid JSON, contains `snapshot_date`, ≥20 rate entries.

---

- [ ] **Unit 7: Documentation updates**

**Goal:** CLAUDE.md reflects new contributor workflow.

**Requirements:** R6

**Dependencies:** Units 1–5

**Files:**
- Modify: `CLAUDE.md` (add Mailpit section, update "No required BWSM secrets locally" note to mention SMTP defaults + currency fallback, remove `SEND_GRID_*` references, list new `SMTP_*` env vars)
- Modify: `docs/ai/architecture.md` if it mentions SendGrid (search first)
- Modify: `docs/ai/authentication.md` if it mentions email flow (search first)

**Approach:**
- Add a "Local mail capture" subsection: `http://localhost:8025` → Mailpit UI; SMTP host `mailpit:1025` no auth.
- Add a "Local currency rates" line: when `FIXER_IO_KEY` unset, rates load from `fallback_rates.json`.
- Update External Services list: SendGrid → "SMTP relay (any provider)"; Fixer.io note unchanged but mention fallback.

**Test scenarios:**
- Test expectation: none — docs.

**Verification:**
- `grep -i sendgrid CLAUDE.md docs/` returns no stale refs.
- New contributor following CLAUDE.md from clone reaches working app + visible mail without external accounts.

## System-Wide Impact

- **Interaction graph:** Email send is called from `api_service.{auth,invitation}.go` during registration, password reset, invitation accept. No callbacks/observers.
- **Error propagation:** Send methods returning error must continue to propagate to handler → HTTP 500 path; graceful-skip (no SMTP host) returns nil to preserve current "feature disabled" semantics.
- **State lifecycle risks:** Registration creates a DB user before sending mail (existing behavior). If SMTP send fails non-gracefully in prod, user record exists without delivery — same as current SendGrid behavior, no regression. Out-of-scope to add transactional outbox.
- **API surface parity:** Public HTTP API unchanged. Internal `IEmailAdapter` is the single interface; no parallel mixins.
- **Integration coverage:** End-to-end mail capture via Mailpit during integration test for invitation flow recommended.
- **Unchanged invariants:** Caller signatures (`SendRegistrationMail`, `SendPasswordResetMail`, `SendInvitationMail`) preserved verbatim. `Config` getter shape preserved (only field set differs). `IFixerIOService` interface unchanged. `fiat_rates` table schema unchanged. Cron schedule (every 12h) unchanged.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Production SMTP relay misconfigured at deploy → mail delivery silently fails | Add startup log line stating SMTP target host + auth-on/off; wire BWSM rotation in same release window; smoke-test invitation in staging post-deploy. |
| Re-authored Go templates render differently from existing SendGrid GUI templates | Pixel-compare against current production templates in Mailpit before cutover; copy German strings verbatim from existing `sendgrid_adapter.go`. |
| `wneessen/go-mail` adds dependency surface | Audit for transitive deps; alternative is stdlib `net/smtp` + manual MIME — fall back if go-mail surface is unjustified. |
| Static fallback rates drift from reality | Acceptable for local dev; document refresh procedure in CLAUDE.md; not used in prod. |
| BWSM rotation order mistake → prod down | Sequence: add `SMTP_*` to BWSM **first**, deploy app reading new vars, **then** remove `SEND_GRID_*` after one healthy deploy cycle. |
| 11 test files mass-renamed → merge conflicts with in-flight branches | Land in a single PR; communicate before merge; rename is mechanical. |

## Documentation / Operational Notes

- BWSM `liquiswiss-prod`: add `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM_ADDRESS`, `SMTP_FROM_NAME`, `SMTP_TLS`. Remove `SEND_GRID_TOKEN`, `SEND_GRID_TEMPLATE_ID` after one healthy deploy.
- Deploy sequence: (1) BWSM add SMTP keys → (2) deploy app version with this PR → (3) verify staging invitation flow → (4) deploy prod → (5) BWSM remove SendGrid keys.
- Monitoring: ensure backend logs surface SMTP-send failures; existing Sentry-equivalent (if any) auto-captures returned errors from handlers.
- Dev onboarding: `make up` and visit `http://localhost:8025` — invitation/registration mail visible, no provider accounts needed.

## Sources & References

- **Origin issue:** https://github.com/devs-group/liquiswiss/issues/24
- Predecessor sketches (superseded): `docs/plans/smtp-email-adapter.md`, `docs/plans/mock-fixer-io-service.md` — delete after this plan ships.
- Related commits: `e5d12d2` (graceful key skip), `cde7d45` (BWSM optional locally), `46345c5` (bws-wrapped Makefile).
- Current adapter: `backend/internal/adapter/sendgrid_adapter/sendgrid_adapter.go`
- Current Fixer service: `backend/internal/service/fixer_io_service/fixer_io_service.go`
- Mailpit: https://mailpit.axllent.org/
- go-mail: https://github.com/wneessen/go-mail
