# Real-time (SSE) Test-Checkliste

Temporary file. Delete after testing session.
Status: feature deployed to production (pipeline green, commit 765b4da).

## Verified

### Entities (create / update / delete via MCP + REST, per-element blink)
- [x] transaction (list + grid, red delete blink, clone dialog unaffected by source changes)
- [x] bank_account (list + grid)
- [x] employee (list + grid, detail page name input blink)
- [x] salary (detail page cards, duplicate via new MCP tool, timeline auto-caps)
- [x] salary_cost (dialog cards + toast, parent salary card blink, bulk copy, single-id copy)
- [x] salary_cost_label (rename updates cost cards + dropdown live, dropdown blink on selected label)
- [x] organisation rename + currency (field-level blink per changed input)
- [x] vat_setting all 4 fields, single + combined changes (field-level blink)
- [x] vat (banner in nested dialog)
- [x] invitation create admin/editor/read-only + delete (per-card blink, mails in Mailpit)
- [x] member remove (per-card red blink, member streams killed)
- [x] forecast (silent refresh)

### Dialog conventions (all dialogs)
- [x] Query-param reload persistence: ?transaction, ?bankAccount, ?salary, ?costs&cost&label (3-level), ?transaction&vat, ?employee=new, ?organisation=new, ?invite=1
- [x] onNuxtReady restore (dialog host mount race fixed, verified on /settings/organisations)
- [x] External-change banner + "Neu laden" in all edit dialogs (via useRealtimeChanges composable)
- [x] Own-action suppression: no toast/banner for changes this tab made itself
      (deterministic: mutation recorded at request start, removed on error; see #28 for server-side successor)

### New MCP tools (integration-tested)
duplicate_salary, update_organisation, update_vat_setting,
list/create/delete/resend_invitation, remove_organisation_member

## Open behavior tests (security/lifecycle)

- [ ] Org switch kills stream instantly (no cross-org events)
- [ ] Logout kills stream instantly
- [ ] Token expiry after 20 min: stream closes, EventSource reconnects silently
- [ ] Second member of same org receives events (and own-suppression stays per-tab)
- [ ] Read-only member receives refresh triggers (data via REST unchanged)

## Follow-ups

- Issue #28: server-computed `own` flag (context plumbing through service layer)
- Issue #27: close after lifecycle tests pass
- category: no management UI yet (issue #26), events + toast wired
