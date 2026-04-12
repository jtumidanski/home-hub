# Sort Tasks & Reminders by Due Date — Context

Last Updated: 2026-04-12

---

## Key Files

| File | Role |
|------|------|
| `services/productivity-service/internal/task/provider.go` | Task query providers — `getAll` on line 15 needs ORDER BY |
| `services/productivity-service/internal/reminder/provider.go` | Reminder query providers — `getAll` on line 15 needs ORDER BY |
| `services/productivity-service/internal/task/entity.go` | Task entity — has `Status` (string), `DueOn` (*time.Time) |
| `services/productivity-service/internal/reminder/entity.go` | Reminder entity — has `ScheduledFor` (time.Time), `LastDismissedAt` (*time.Time) |

## Key Decisions

1. **Server-side only** — Sorting is baked into the GORM query, no API sort parameters, no frontend changes
2. **Pending before completed** — Tasks use `status ASC` as primary sort (pending < completed alphabetically)
3. **Nulls last** — Tasks with no `due_on` sink to the bottom via `NULLS LAST`
4. **Active before dismissed** — Reminders use `last_dismissed_at IS NULL DESC` to surface active reminders first
5. **No indexes** — Household data volumes are small enough that existing indexes suffice

## Dependencies

- None. This change is self-contained within the productivity-service provider layer.

## Database Considerations

- PostgreSQL supports `NULLS LAST` and boolean expressions in ORDER BY natively
- GORM passes `.Order()` string arguments through to SQL directly
- No schema migration needed
