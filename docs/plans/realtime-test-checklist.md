# Real-time (SSE) Test-Checkliste, offene Entities

Temporary file. Delete after testing session.

## Already verified

- [x] transaction: create / update / delete (list + grid, per-card blink, red delete blink)
- [x] bank_account: create / update / delete (list + grid)
- [x] employee: create / rename / delete (list + grid, detail page name input blink)
- [x] salary: create / update / delete / duplicate (detail page cards, red delete blink)
- [x] salary_cost: create / update / delete / bulk copy / single-id copy (dialog cards + toast, parent salary card blink)
- [x] duplicate_salary MCP tool (new, integration-tested)
- [x] forecast: refresh via dedicated event (dashboard)
- [x] Dialog survives reload via `?costs=<salaryID>` query param

## Open to test

| Entity | Where visible | Blink target |
|---|---|---|
| category | Organisation page + Transaktionen | container |
| vat | Organisation page + Transaktionen | container |
| vat_setting | Organisation page | container |
| organisation (rename) | Header dropdown + Organisation page | container |
| salary_cost_label (rename/delete) | Organisation page | container |
| forecast_exclusion | Prognose | via forecast event |
| fiat rates | Prognose amounts | none (no event, cron-synced) |

## Open behavior tests (security/lifecycle)

- [ ] Org switch kills stream instantly (no cross-org events)
- [ ] Logout kills stream instantly
- [ ] Token expiry after 20 min: stream closes, EventSource reconnects silently
- [ ] Second member of same org receives events
- [ ] Read-only member receives refresh triggers (data via REST unchanged)
