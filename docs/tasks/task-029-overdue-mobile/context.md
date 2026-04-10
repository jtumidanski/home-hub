# Overdue Task Display on Mobile — Context

Last Updated: 2026-04-09

---

## Key Files

| File | Role |
|------|------|
| `frontend/src/components/features/tasks/task-card.tsx` | Mobile task card component — **file to change** |
| `frontend/src/types/models/task.ts` | `isTaskOverdue()` utility and `Task` type |
| `frontend/src/pages/TasksPage.tsx` | Desktop reference implementation (lines 138-149) |
| `frontend/src/types/models/__tests__/task.test.ts` | Existing unit tests for `isTaskOverdue()` |

## Key Decisions

1. **Reuse existing `isTaskOverdue()` utility** — no new logic needed, just wire it into `TaskCard`
2. **Match desktop exactly** — same badge variants (`destructive`/`secondary`/`default`) and same text values
3. **No due date text styling changes** — desktop doesn't style the date differently for overdue tasks, so mobile won't either

## Dependencies

- None. All required code already exists.
