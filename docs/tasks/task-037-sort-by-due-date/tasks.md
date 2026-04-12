# Sort Tasks & Reminders by Due Date — Task Checklist

Last Updated: 2026-04-12

---

## Implementation

- [x] Add ORDER BY to `getAll` in `task/provider.go` — sort by `status ASC`, `due_on ASC NULLS LAST`
- [x] Add ORDER BY to `getAll` in `reminder/provider.go` — sort by `last_dismissed_at IS NULL DESC`, `scheduled_for ASC`

## Verification

- [x] `go build ./...` passes for productivity-service
- [x] `go test ./...` passes for productivity-service
- [x] Docker build passes for productivity-service
- [ ] Manual test: `GET /api/v1/tasks` returns pending before completed, sorted by due_on ascending, nulls last
- [ ] Manual test: `GET /api/v1/reminders` returns active before dismissed, sorted by scheduled_for ascending
