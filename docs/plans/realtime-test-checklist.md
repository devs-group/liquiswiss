# Real-time (SSE) Test-Checkliste

Temporary file. Delete after testing session.

## Verified

- [x] transaction: create / update / delete (list + grid, per-card blink, red delete blink)
- [x] bank_account: create / update / delete (list + grid)
- [x] employee: create / rename / delete (list + grid, detail page name input blink)
- [x] salary: create / update / delete / duplicate (detail page cards, red delete blink)
- [x] salary_cost: create / update / delete / bulk copy / single-id copy (dialog cards + toast, parent salary card blink)
- [x] organisation: rename + currency via MCP (field-level blink per changed input)
- [x] vat_setting: all 4 fields via MCP, single + combined changes (field-level blink, date compare fix)
- [x] invitation: create (admin/editor/read-only) + delete via REST (per-card blink, mails in Mailpit)
- [x] member: remove via REST (per-card red blink, streams killed)
- [x] forecast: silent refresh via dedicated event (dashboard)
- [x] Dialog survives reload via `?costs=<salaryID>` query param
- [x] Per-element-only rule enforced: container blink removed, toast fallback

## New MCP tools (integration-tested)

duplicate_salary, update_organisation, update_vat_setting,
list/create/delete/resend_invitation, remove_organisation_member

## Open to test

| Entity | Where visible | Feedback |
|---|---|---|
| category | Transaktionen (no mgmt UI yet, issue #26) | toast |
| vat | Transaktionen | toast |
| salary_cost_label | Label dropdown in salary cost dialog (editable there); label names on cost cards | toast; note: open cost lists show old label name until costs refetch (only `salary_cost` events trigger that) |
| forecast_exclusion | Prognose | via forecast event |
| fiat rates | Prognose amounts | none (no event, cron-synced) |

## Open behavior tests (security/lifecycle)

- [ ] Org switch kills stream instantly (no cross-org events)
- [ ] Logout kills stream instantly
- [ ] Token expiry after 20 min: stream closes, EventSource reconnects silently
- [ ] Second member of same org receives events
- [ ] Read-only member receives refresh triggers (data via REST unchanged)
- [ ] Invitation flow via new MCP tools (created after last reconnect)
