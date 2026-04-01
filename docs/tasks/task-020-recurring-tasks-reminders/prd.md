# Recurring Tasks & Reminders — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-31

---

## 1. Overview

Tasks and reminders in Home Hub are currently one-shot: a task is created, completed, and done; a reminder fires once and is dismissed. Many household responsibilities are inherently repeating — take out the trash every Tuesday, pay rent on the 1st, check the smoke detectors every 6 months. Users today must manually re-create these items each time, which is tedious and error-prone.

This feature adds recurrence support to both tasks and reminders. Users define a repeating pattern when creating or editing an item, and the system automatically generates the next occurrence when the current one is completed (tasks) or dismissed (reminders). Recurrence rules are stored as RFC 5545 RRULE strings internally, but exposed through simple UI presets.

The design follows a generate-on-trigger model: only the current occurrence exists as a concrete record. The next occurrence is created when the current one is acted upon. This keeps storage lean and avoids the need for background jobs.

## 2. Goals

Primary goals:
- Allow users to create tasks and reminders that repeat on a defined schedule
- Automatically generate the next occurrence when the current one is completed/dismissed
- Preserve completion history for recurring tasks so users can track streaks
- Provide a simple, intuitive recurrence picker that covers common household patterns
- Support editing and deleting with scope choices (this occurrence vs. this and all future)

Non-goals:
- Background job to pre-generate future occurrences
- Push notifications or alerts (separate concern)
- Syncing recurring tasks to Google Calendar
- Per-occurrence ownership reassignment (occurrences inherit series owner)
- Full RRULE editor UI (store RRULE internally but expose simple presets for v1)

## 3. User Stories

- As a household member, I want to create a task that repeats weekly so that I don't have to manually re-create chores each week
- As a household member, I want to create a reminder that repeats monthly so that I'm reminded of recurring obligations like paying rent
- As a household member, I want to see my completion history for recurring tasks so that I can track my consistency
- As a household member, I want to edit a single occurrence of a recurring task without affecting the rest of the series
- As a household member, I want to stop a recurring task/reminder by deleting "this and all future" occurrences
- As a household member, I want to change the recurrence pattern of a series going forward without losing past history
- As a household member, I want recurring tasks to anchor to the original schedule so that my routine doesn't drift when I complete things late

## 4. Functional Requirements

### 4.1 Recurrence Rules

- Users can attach a recurrence rule when creating or editing a task or reminder
- Supported presets:
  - **Daily** — every day
  - **Weekly** — every week on selected day(s) of the week
  - **Every N days** — custom day interval (e.g., every 3 days)
  - **Every N weeks** — custom week interval on selected day(s)
  - **Monthly** — on a specific day of the month (e.g., the 15th)
  - **Yearly** — on a specific month and day
- Recurrence rules are stored as RFC 5545 RRULE strings (e.g., `FREQ=WEEKLY;BYDAY=TU,TH`)
- Optional end date: recurrence can repeat forever (default) or until a specified date
- A task or reminder with a recurrence rule is part of a **series**, identified by a `series_id`

### 4.2 Series and Occurrences

- A **series** is a logical grouping identified by `series_id` (UUID), shared by all occurrences
- Only one **active** (non-completed, non-dismissed) occurrence exists per series at any time
- Each occurrence is a regular task or reminder row with additional series fields
- The first occurrence is created at series creation time with the user-specified due date / scheduled time
- Subsequent occurrences are generated on trigger (see 4.3, 4.4)

### 4.3 Task Completion — Generate Next Occurrence

- When a recurring task is marked as completed:
  1. The current task record is updated to status `completed` with `completed_at` and `completed_by_user_id` (existing behavior)
  2. A new task record is created for the next occurrence:
     - Same `series_id`, `tenant_id`, `household_id`, `title`, `notes`, `rollover_enabled`, `owner_user_id`
     - `due_on` calculated from the **original schedule anchor**, not the completion date
     - `occurrence_index` incremented by 1
     - Status `pending`, no completion fields
  3. If the recurrence has an `end_date` and the next `due_on` would fall after it, no new occurrence is created — the series is finished
- When a recurring task is uncompleted (completed -> pending), the auto-generated next occurrence (if any) must be deleted to maintain the one-active-occurrence invariant
- Schedule anchoring: the next due date is always computed from the series `recurrence_start` date and the RRULE, not from when the user completed the task

### 4.4 Reminder Dismissal — Generate Next Occurrence

- When a recurring reminder is dismissed:
  1. The current reminder record is hard-deleted (existing behavior) after creating the dismissal audit record
  2. A new reminder record is created for the next occurrence:
     - Same `series_id`, `tenant_id`, `household_id`, `title`, `notes`, `owner_user_id`
     - `scheduled_for` calculated from the original schedule anchor using the RRULE
     - `occurrence_index` incremented by 1
  3. If past the `end_date`, no new occurrence is created
- Snoozing a recurring reminder works on the current occurrence only (existing behavior, unchanged)
- The newly generated reminder sits dormant until its `scheduled_for` time arrives

### 4.5 Edit Scope

When editing a recurring task or reminder, the user chooses a scope:

- **This occurrence only**: Detach from series — clear `series_id`, `recurrence_rule`, and related fields on this record. It becomes a standalone item. The series continues; the next trigger-generated occurrence will use the original series values.
- **This and all future occurrences**: Update the current record's fields AND update the series metadata (`recurrence_rule`, `recurrence_start`, `end_date`, `title`, `notes`, etc.) so future generated occurrences inherit the changes. Past completed occurrences are untouched.

### 4.6 Delete Scope

When deleting a recurring task or reminder, the user chooses a scope:

- **This occurrence only**: Delete/soft-delete the current occurrence. Generate the next occurrence immediately (same logic as completion/dismissal, minus the completion tracking).
- **This and all future occurrences**: Delete/soft-delete the current occurrence. Do NOT generate a next occurrence. The series is effectively ended.

### 4.7 Rollover Interaction

- For recurring tasks with `rollover_enabled`: if a recurring task is overdue and rolls over to today, it keeps its series membership. Completing the rolled-over task still generates the next occurrence based on the original schedule.

### 4.8 Summary Endpoints

- `GET /api/v1/summary/tasks`: `pendingCount` should include the active occurrence of recurring series (no change needed — they're regular pending tasks)
- `GET /api/v1/summary/reminders`: same principle — recurring reminders are regular reminder rows

## 5. API Surface

### 5.1 Task Endpoints — Changes

**POST /api/v1/tasks** (create)
- New optional attributes:
  - `recurrenceRule` (string) — RRULE string (e.g., `FREQ=WEEKLY;BYDAY=TU`)
  - `recurrenceEndDate` (string, YYYY-MM-DD) — optional end date
- When `recurrenceRule` is provided, `dueOn` is required (serves as the first occurrence date and schedule anchor)
- Response includes new attributes (see read below)

**GET /api/v1/tasks/{id}** and **GET /api/v1/tasks** (read)
- New attributes in response:
  - `seriesId` (string, UUID, nullable) — null for non-recurring tasks
  - `recurrenceRule` (string, nullable) — RRULE string
  - `recurrenceStart` (string, YYYY-MM-DD, nullable) — anchor date for schedule computation
  - `recurrenceEndDate` (string, YYYY-MM-DD, nullable)
  - `occurrenceIndex` (integer, nullable) — 0-based index in the series

**PATCH /api/v1/tasks/{id}** (update)
- New optional attribute:
  - `editScope` (string, enum: `occurrence`, `future`) — required when task has a `seriesId`. Defaults to error if omitted on a recurring task.
  - `recurrenceRule` (string) — update the rule (only with `future` scope)
  - `recurrenceEndDate` (string, YYYY-MM-DD, nullable) — update or clear end date
- Scope `occurrence`: detaches from series, updates only this record
- Scope `future`: updates this record and series metadata for future occurrences

**DELETE /api/v1/tasks/{id}** (delete)
- New optional query parameter:
  - `scope` (string, enum: `occurrence`, `future`) — required when task has a `seriesId`
  - `occurrence`: soft-deletes this task, generates next occurrence
  - `future`: soft-deletes this task, ends the series

**PATCH /api/v1/tasks/{id}** (status change to completed)
- Existing behavior plus: if task is recurring, generate next occurrence per section 4.3

### 5.2 Reminder Endpoints — Changes

**POST /api/v1/reminders** (create)
- New optional attributes:
  - `recurrenceRule` (string) — RRULE string
  - `recurrenceEndDate` (string, YYYY-MM-DD) — optional end date
- When `recurrenceRule` is provided, `scheduledFor` serves as the first occurrence time and schedule anchor

**GET /api/v1/reminders/{id}** and **GET /api/v1/reminders** (read)
- New attributes:
  - `seriesId` (string, UUID, nullable)
  - `recurrenceRule` (string, nullable)
  - `recurrenceStart` (string, RFC3339, nullable)
  - `recurrenceEndDate` (string, YYYY-MM-DD, nullable)
  - `occurrenceIndex` (integer, nullable)

**PATCH /api/v1/reminders/{id}** (update)
- New optional attributes:
  - `editScope` (string, enum: `occurrence`, `future`) — required when reminder has a `seriesId`
  - `recurrenceRule`, `recurrenceEndDate` — same semantics as tasks

**DELETE /api/v1/reminders/{id}** (delete)
- New optional query parameter:
  - `scope` (string, enum: `occurrence`, `future`) — required when reminder has a `seriesId`
  - `occurrence`: deletes this reminder, generates next occurrence
  - `future`: deletes this reminder, ends the series

**POST /api/v1/reminders/dismissals** (dismiss)
- Existing behavior plus: if reminder is recurring, generate next occurrence per section 4.4

### 5.3 New Endpoint — Series History

**GET /api/v1/tasks/series/{seriesId}**
- Returns all completed occurrences for a task series (completion history)
- Response: JSON:API array of `tasks` resources filtered by `series_id`, ordered by `occurrence_index`
- Scoped to tenant and household

### 5.4 Error Cases

- Creating a recurring task without `dueOn`: 400 `recurrence requires a due date`
- Updating/deleting a recurring item without `editScope`/`scope`: 400 `scope is required for recurring items`
- Invalid RRULE string: 400 `invalid recurrence rule`
- `editScope=future` on a non-recurring item: 400 `scope not applicable to non-recurring item`

## 6. Data Model

### 6.1 New Columns — tasks Table

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `series_id` | UUID | nullable, indexed | Links occurrences in a series |
| `recurrence_rule` | TEXT | nullable | RFC 5545 RRULE string |
| `recurrence_start` | DATE | nullable | Schedule anchor date |
| `recurrence_end_date` | DATE | nullable | Optional end date for the series |
| `occurrence_index` | INTEGER | nullable, default 0 | 0-based position in series |

### 6.2 New Columns — reminders Table

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `series_id` | UUID | nullable, indexed | Links occurrences in a series |
| `recurrence_rule` | TEXT | nullable | RFC 5545 RRULE string |
| `recurrence_start` | TIMESTAMP | nullable | Schedule anchor time |
| `recurrence_end_date` | DATE | nullable | Optional end date for the series |
| `occurrence_index` | INTEGER | nullable, default 0 | 0-based position in series |

### 6.3 Constraints

- If `series_id` is set, `recurrence_rule` and `recurrence_start` must also be set (enforced at application level)
- `series_id` is generated once at series creation and copied to all occurrences
- `occurrence_index` is monotonically increasing within a series
- All new columns include `tenant_id` scoping via existing row-level constraints
- Migrations are additive (new nullable columns) — no destructive changes

### 6.4 Index Additions

- `idx_tasks_series_id` on `tasks(series_id)` — for series history queries
- `idx_reminders_series_id` on `reminders(series_id)` — for series lookups

## 7. Service Impact

### 7.1 productivity-service (Primary)

**Domain layer:**
- Add series fields to `task.Model` and `reminder.Model`
- Add `IsRecurring() bool` method to both models
- New `recurrence` package: RRULE parsing and next-occurrence date computation
- Processor changes: task completion and reminder dismissal trigger occurrence generation

**Storage layer:**
- Add columns to `task.Entity` and `reminder.Entity`
- New migration file for schema changes
- Query for series history (tasks by `series_id`)

**Transport layer:**
- Extend create/update request parsing for recurrence fields
- Add `editScope` / `scope` parameter handling
- Extend response serialization for new attributes
- New handler for `GET /api/v1/tasks/series/{seriesId}`

### 7.2 Frontend (UI)

**Recurrence picker component:**
- Shared component used in both task and reminder create/edit forms
- Preset buttons: Daily, Weekly, Monthly, Yearly, Custom
- Custom mode: interval + unit (days/weeks) + optional day-of-week checkboxes
- Optional end date picker
- Shows human-readable summary (e.g., "Every Tuesday and Thursday")

**Task/reminder list:**
- Recurring indicator icon on items that belong to a series
- Scope choice dialog when editing or deleting a recurring item

**Task completion:**
- When completing a recurring task, show brief confirmation that the next occurrence was created (e.g., "Next due: Tuesday, Apr 7")

**Series history view:**
- Accessible from a recurring task's detail view
- Shows list of completed occurrences with completion dates

## 8. Non-Functional Requirements

### 8.1 Performance
- Next-occurrence generation must complete within the same request as completion/dismissal (no async processing)
- Series history query should be indexed and paginated for series with many completed occurrences
- RRULE computation is CPU-trivial for the supported preset patterns

### 8.2 Multi-tenancy
- All new columns and queries scoped by `tenant_id` and `household_id`
- `series_id` is unique per tenant (UUID generation ensures this)
- Series history endpoint enforces tenant/household scoping

### 8.3 Security
- No new authentication or authorization concerns — series membership doesn't grant additional access
- RRULE input must be validated and sanitized to prevent malformed rules

### 8.4 Observability
- Log occurrence generation events (series_id, occurrence_index, next due date)
- Existing audit tables (dismissals, snoozes, restorations) continue to work unchanged

## 9. Open Questions

None — all key decisions resolved during scoping.

## 10. Acceptance Criteria

- [ ] User can create a recurring task with any supported preset pattern
- [ ] User can create a recurring reminder with any supported preset pattern
- [ ] Completing a recurring task generates the next occurrence with the correct anchored due date
- [ ] Dismissing a recurring reminder generates the next occurrence with the correct anchored scheduled time
- [ ] Uncompleting a recurring task deletes the auto-generated next occurrence
- [ ] User can edit "this occurrence only" — occurrence is detached from series
- [ ] User can edit "this and all future" — series metadata is updated
- [ ] User can delete "this occurrence only" — next occurrence is generated
- [ ] User can delete "this and all future" — series ends, no new occurrence
- [ ] Recurrence with an end date stops generating occurrences after the end date
- [ ] Series history endpoint returns all completed occurrences in order
- [ ] Frontend recurrence picker supports all preset patterns
- [ ] Frontend shows recurring indicator on series items
- [ ] Frontend shows scope choice dialog on edit/delete of recurring items
- [ ] All queries scoped by tenant_id and household_id
- [ ] RRULE input is validated; invalid rules return 400
- [ ] Existing non-recurring task and reminder behavior is completely unchanged
