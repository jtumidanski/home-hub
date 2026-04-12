# Sort Tasks & Reminders by Due Date — Implementation Plan

Last Updated: 2026-04-12

---

## Executive Summary

Add fixed server-side sorting to the Tasks and Reminders list endpoints in the productivity-service. This is a minimal backend-only change — two GORM `.Order()` calls in two provider files. No API contract changes, no frontend changes, no schema migrations.

## Current State Analysis

- **Task list** (`GET /api/v1/tasks`): The `getAll` provider in `task/provider.go:15-22` queries with a `deleted_at IS NULL` filter but no ORDER BY. Results are returned in undefined database order.
- **Reminder list** (`GET /api/v1/reminders`): The `getAll` provider in `reminder/provider.go:15-18` is a bare query with no filtering or ordering. Results are returned in undefined database order.
- **Frontend**: Both `TasksPage.tsx` and `RemindersPage.tsx` render items in the order received from the API. No client-side sorting exists.

## Proposed Future State

- **Tasks**: Sorted by `status ASC` (pending before completed), then `due_on ASC NULLS LAST` (earliest due date first, undated tasks at bottom).
- **Reminders**: Sorted by `last_dismissed_at IS NULL DESC` (active before dismissed), then `scheduled_for ASC` (earliest scheduled time first).

## Implementation

### Phase 1: Task Sorting

**File**: `services/productivity-service/internal/task/provider.go`

Modify the `getAll` function to add ORDER BY clauses to the GORM query:

```go
func getAll(includeDeleted bool) database.EntityProvider[[]Entity] {
    return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
        if !includeDeleted {
            db = db.Where("deleted_at IS NULL")
        }
        return db.Order("status ASC").Order("due_on ASC NULLS LAST")
    })
}
```

**Acceptance criteria**:
- Pending tasks appear before completed tasks
- Within each status group, tasks with earlier `due_on` appear first
- Tasks with `NULL` `due_on` appear after dated tasks within their group

### Phase 2: Reminder Sorting

**File**: `services/productivity-service/internal/reminder/provider.go`

Modify the `getAll` function to add ORDER BY clauses:

```go
func getAll() database.EntityProvider[[]Entity] {
    return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
        return db.Order("last_dismissed_at IS NULL DESC").Order("scheduled_for ASC")
    })
}
```

The expression `last_dismissed_at IS NULL DESC` evaluates to `true` (1) for active reminders and `false` (0) for dismissed ones. DESC puts `true` first — active reminders on top.

**Acceptance criteria**:
- Active (undismissed) reminders appear before dismissed reminders
- Within each group, reminders with earlier `scheduled_for` appear first

### Phase 3: Verification

- Run `go build ./...` for productivity-service
- Run `go test ./...` for productivity-service
- Run Docker build via `scripts/local-up.sh` or targeted Docker build
- Manual verification: call both endpoints and confirm sort order

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `NULLS LAST` syntax unsupported by DB | Low | Medium | PostgreSQL supports it natively; verify GORM passes it through raw |
| Boolean expression in ORDER BY unsupported | Low | Medium | PostgreSQL supports it; fallback is `CASE WHEN` expression |
| Sort degrades query performance | Very Low | Low | Household datasets are small (< 1,000 items); no index needed |

## Success Metrics

- Both list endpoints return deterministic, correctly sorted results
- No performance regression on list queries
- Docker build passes cleanly

## Required Resources and Dependencies

- **Services**: productivity-service only
- **Dependencies**: None new — uses existing GORM and PostgreSQL capabilities
- **Frontend**: No changes required

## Effort Estimate

**Size: S** — Two functions, two lines each. Estimated implementation time is minimal.
