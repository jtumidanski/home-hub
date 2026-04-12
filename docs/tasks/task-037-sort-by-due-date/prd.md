# Sort Tasks & Reminders by Due Date — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-12
---

## 1. Overview

The Tasks and Reminders list pages currently return data in undefined order — no ORDER BY clause exists in the database queries and no client-side sorting is applied. Items appear in whatever order the database returns them, which is effectively random from the user's perspective.

This feature adds fixed server-side sorting so that both pages display items in a predictable, useful order: pending/active items first, then sorted by due date ascending. This ensures the most urgent, actionable items are always visible at the top of the list.

## 2. Goals

Primary goals:
- Tasks are sorted by status (pending first), then by `due_on` ascending with nulls last
- Reminders are sorted by dismissal state (active first), then by `scheduled_for` ascending

Non-goals:
- User-configurable sort columns or direction toggles in the UI
- Sort query parameters on the API
- Pagination or infinite scroll
- Client-side sorting logic

## 3. User Stories

- As a household member, I want tasks sorted by due date so that I see the most urgent tasks first
- As a household member, I want completed tasks pushed below pending ones so that my active work is front and center
- As a household member, I want reminders sorted by scheduled time so that upcoming reminders appear at the top
- As a household member, I want dismissed reminders pushed below active ones so that resolved items don't clutter the view
- As a household member, I want tasks without a due date to appear at the bottom so that dated tasks with real deadlines are prioritized

## 4. Functional Requirements

### 4.1 Task Sorting

- Tasks MUST be sorted by the following criteria in order:
  1. `status` ascending — `"pending"` before `"completed"`
  2. `due_on` ascending — earliest due date first
  3. `NULL` due dates sort after all non-null due dates (NULLS LAST)
- This sort order MUST be applied at the database query level in the task provider
- The sort order applies to all task list queries (the `getAll` provider)

### 4.2 Reminder Sorting

- Reminders MUST be sorted by the following criteria in order:
  1. Dismissal state — reminders where `last_dismissed_at IS NULL` sort first (active before dismissed)
  2. `scheduled_for` ascending — earliest scheduled time first
- This sort order MUST be applied at the database query level in the reminder provider
- The sort order applies to all reminder list queries (the `getAll` provider)

### 4.3 No Frontend Changes

- The frontend already renders items in the order received from the API
- No frontend code changes are required

## 5. API Surface

No API changes. The existing endpoints return the same response shape — only the ordering of items in the list changes.

- `GET /api/v1/tasks` — response items now sorted by status then due_on ASC
- `GET /api/v1/reminders` — response items now sorted by dismissal state then scheduled_for ASC

## 6. Data Model

No schema changes. Sorting uses existing columns:

**Tasks:**
- `status` (string, not null) — "pending" or "completed"
- `due_on` (*time.Time, nullable) — the task's due date

**Reminders:**
- `last_dismissed_at` (*time.Time, nullable) — null means active/undismissed
- `scheduled_for` (time.Time, not null) — when the reminder is scheduled

### Index Considerations

If list performance degrades on large datasets, composite indexes could be added:
- Tasks: `(tenant_id, household_id, status, due_on)`
- Reminders: `(tenant_id, household_id, last_dismissed_at, scheduled_for)`

These are optional and only needed if query plans show sequential scans.

## 7. Service Impact

| Service | Impact |
|---------|--------|
| **productivity-service** | Modify task and reminder provider `getAll` functions to include ORDER BY clauses |
| **frontend** | No changes |
| **All other services** | No impact |

### Specific File Changes

- `services/productivity-service/internal/task/provider.go` — Add `.Order()` clauses to the `getAll` query
- `services/productivity-service/internal/reminder/provider.go` — Add `.Order()` clauses to the `getAll` query

## 8. Non-Functional Requirements

- Sorting must not degrade list query performance for typical household sizes (< 1,000 items)
- No new dependencies or configuration required
- Multi-tenancy scoping (`tenant_id`, `household_id`) is unchanged — sorting is applied after tenant filtering

## 9. Open Questions

None — all decisions resolved during scoping.

## 10. Acceptance Criteria

- [ ] `GET /api/v1/tasks` returns pending tasks before completed tasks
- [ ] Within each status group, tasks are sorted by `due_on` ascending
- [ ] Tasks with no `due_on` appear after tasks with a due date within their status group
- [ ] `GET /api/v1/reminders` returns active (undismissed) reminders before dismissed reminders
- [ ] Within each dismissal group, reminders are sorted by `scheduled_for` ascending
- [ ] Existing task and reminder filtering (by status, owner, search) continues to work
- [ ] Docker build passes for productivity-service
