# Recurring Tasks & Reminders — Task Checklist

Last Updated: 2026-03-31

---

## Phase 1: Recurrence Computation Package [S]

- [ ] **1.1** Add RRULE Go library dependency (`github.com/teambition/rrule-go` or equivalent) to `go.mod`
- [ ] **1.2** Create `internal/recurrence/recurrence.go` — `NextOccurrence(rrule string, anchor time.Time, current time.Time) (time.Time, error)` function that computes the next occurrence date from an RRULE string anchored to the original start date
- [ ] **1.3** Create `internal/recurrence/recurrence.go` — `Validate(rrule string) error` function that validates an RRULE string against supported patterns
- [ ] **1.4** Create `internal/recurrence/recurrence.go` — `IsExpired(rrule string, anchor time.Time, endDate *time.Time, current time.Time) bool` function that checks if a series has passed its end date
- [ ] **1.5** Create `internal/recurrence/recurrence_test.go` — unit tests covering: daily, weekly single-day, weekly multi-day, every-N-days, every-N-weeks, monthly, yearly, with end date, past end date, schedule anchoring (completed late still anchors to original), DST boundary, month-end edge case (e.g., Jan 31 → Feb 28)

**Acceptance**: All recurrence tests pass. Package has no dependencies on task/reminder code.

---

## Phase 2: Backend Domain & Storage [L]

### 2A: Schema & Entity Changes

- [ ] **2A.1** Add columns to `task.Entity`: `SeriesId *uuid.UUID`, `RecurrenceRule *string`, `RecurrenceStart *time.Time` (type:date), `RecurrenceEndDate *time.Time` (type:date), `OccurrenceIndex *int` — all nullable with appropriate GORM tags and indexes on `SeriesId`
- [ ] **2A.2** Add columns to `reminder.Entity`: `SeriesId *uuid.UUID`, `RecurrenceRule *string`, `RecurrenceStart *time.Time`, `RecurrenceEndDate *time.Time` (type:date), `OccurrenceIndex *int` — all nullable with index on `SeriesId`
- [ ] **2A.3** Verify GORM AutoMigrate adds columns cleanly on startup (test with existing DB)

**Acceptance**: Service starts, auto-migration adds columns without data loss. Existing queries unaffected.

### 2B: Model & Builder Changes

- [ ] **2B.1** Add fields to `task.Model`: `seriesID *uuid.UUID`, `recurrenceRule *string`, `recurrenceStart *time.Time`, `recurrenceEndDate *time.Time`, `occurrenceIndex *int` — with getter methods and `IsRecurring() bool`
- [ ] **2B.2** Update `task.Builder` with setter methods for all new fields; update `Build()` to copy them
- [ ] **2B.3** Update `task.Make()` and `task.ToEntity()` to map new fields between entity and model
- [ ] **2B.4** Add fields to `reminder.Model`: same series fields with getter methods and `IsRecurring() bool`
- [ ] **2B.5** Update `reminder.Builder` with setter methods; update `Build()`
- [ ] **2B.6** Update `reminder.Make()` and `reminder.ToEntity()` to map new fields

**Acceptance**: Existing builder tests still pass. New fields round-trip through Make/ToEntity.

### 2C: Administrator & Provider Changes

- [ ] **2C.1** Extend `task.create()` to accept series parameters (seriesID, recurrenceRule, recurrenceStart, recurrenceEndDate, occurrenceIndex)
- [ ] **2C.2** Add `task.createNextOccurrence()` — creates a new pending task in the same series with incremented occurrence index and computed next due date
- [ ] **2C.3** Add `task.getBySeriesID(seriesID)` provider — returns all tasks with matching series_id, ordered by occurrence_index
- [ ] **2C.4** Add `task.getActiveBySeriesID(seriesID)` provider — returns non-completed, non-deleted task for a series (for uncomplete cleanup)
- [ ] **2C.5** Add `task.deleteBySeriesIDFromIndex(seriesID, fromIndex)` — soft-deletes all tasks in series with occurrence_index >= given index (for future-scope delete)
- [ ] **2C.6** Extend `reminder.create()` to accept series parameters
- [ ] **2C.7** Add `reminder.createNextOccurrence()` — creates a new reminder in the same series with incremented index and computed next scheduled_for
- [ ] **2C.8** Add `reminder.getBySeriesID(seriesID)` provider

**Acceptance**: New DB functions create/query series data correctly. Existing create/query functions unchanged for non-recurring items.

### 2D: Processor Logic Changes

- [ ] **2D.1** Modify `task.Processor.Update()` — when status transitions to `completed` and task `IsRecurring()`: compute next due date via recurrence package, call `createNextOccurrence()` within same transaction. If past end date, skip.
- [ ] **2D.2** Modify `task.Processor.Update()` — when status transitions from `completed` to `pending` and task `IsRecurring()`: find and delete the auto-generated next occurrence via `getActiveBySeriesID()`
- [ ] **2D.3** Modify `task.Processor.Create()` — when recurrenceRule is provided: generate series_id, set recurrence_start from dueOn, set occurrence_index to 0
- [ ] **2D.4** Add scope-aware `task.Processor.UpdateWithScope(id, scope, ...)` — `occurrence`: detach from series (clear series fields); `future`: update current + series metadata
- [ ] **2D.5** Add scope-aware `task.Processor.DeleteWithScope(id, scope)` — `occurrence`: soft-delete + generate next; `future`: soft-delete + end series
- [ ] **2D.6** Modify `dismissal.Processor.Create()` — after dismissing a recurring reminder: compute next scheduled_for, call `reminder.createNextOccurrence()`. If past end date, skip.
- [ ] **2D.7** Modify `reminder.Processor.Create()` — when recurrenceRule is provided: generate series_id, set recurrence_start from scheduledFor, set occurrence_index to 0
- [ ] **2D.8** Add scope-aware `reminder.Processor.UpdateWithScope()` and `DeleteWithScope()` — same semantics as tasks
- [ ] **2D.9** Write processor unit tests for: recurring task creation, completion triggering next occurrence, uncomplete cleaning up next occurrence, end date stopping generation, scope-aware update/delete
- [ ] **2D.10** Write processor unit tests for: recurring reminder creation, dismissal triggering next occurrence, scope-aware update/delete

**Acceptance**: All new and existing processor tests pass. Non-recurring paths unchanged.

---

## Phase 3: Backend REST API [M]

- [ ] **3.1** Extend `task.CreateRequest` with `RecurrenceRule *string` and `RecurrenceEndDate *string` fields
- [ ] **3.2** Extend `task.UpdateRequest` with `EditScope *string`, `RecurrenceRule *string`, `RecurrenceEndDate *string` fields
- [ ] **3.3** Extend `task.RestModel` with `SeriesId *string`, `RecurrenceRule *string`, `RecurrenceStart *string`, `RecurrenceEndDate *string`, `OccurrenceIndex *int` fields
- [ ] **3.4** Update `task.Transform()` to map new model fields to RestModel
- [ ] **3.5** Update task `CreateHandler` to parse recurrence fields and pass to processor
- [ ] **3.6** Update task `UpdateHandler` to parse editScope and route to `UpdateWithScope()` for recurring tasks
- [ ] **3.7** Update task `DeleteHandler` to parse `scope` query parameter and route to `DeleteWithScope()` for recurring tasks
- [ ] **3.8** Add `GET /api/v1/tasks/series/{seriesId}` handler returning completed occurrence history
- [ ] **3.9** Extend `reminder.CreateRequest`, `UpdateRequest`, `RestModel` with recurrence fields (mirror task changes)
- [ ] **3.10** Update `reminder.Transform()` for new fields
- [ ] **3.11** Update reminder Create/Update/Delete handlers with recurrence and scope handling
- [ ] **3.12** Add validation: 400 if recurring task created without dueOn; 400 if recurring item updated/deleted without scope; 400 if invalid RRULE
- [ ] **3.13** Update service docs: `docs/rest.md`, `docs/domain.md`, `docs/storage.md`

**Acceptance**: API accepts and returns recurrence fields. Scope enforcement works. Existing non-recurring API calls unaffected. Docker build passes.

---

## Phase 4: Frontend [L]

### 4A: Types, Schemas & API Client

- [ ] **4A.1** Update `types/models/task.ts` — add `seriesId`, `recurrenceRule`, `recurrenceStart`, `recurrenceEndDate`, `occurrenceIndex` to `TaskAttributes`
- [ ] **4A.2** Update `types/models/reminder.ts` — add same series attributes to `ReminderAttributes`
- [ ] **4A.3** Update `lib/schemas/task.schema.ts` — add `recurrenceRule` (optional string) and `recurrenceEndDate` (optional date string)
- [ ] **4A.4** Update `lib/schemas/reminder.schema.ts` — add `recurrenceRule` and `recurrenceEndDate`
- [ ] **4A.5** Update `services/api/productivity.ts` — extend task/reminder create/update methods to include recurrence fields; add `scope` query param to update/delete; add `getTaskSeriesHistory(seriesId)` method
- [ ] **4A.6** Update `lib/hooks/api/use-tasks.ts` — pass recurrence fields through create/update mutations; add `useTaskSeriesHistory(seriesId)` hook; pass scope to update/delete mutations
- [ ] **4A.7** Update `lib/hooks/api/use-reminders.ts` — pass recurrence fields and scope through mutations

**Acceptance**: TypeScript types match API response. Hooks pass recurrence data correctly.

### 4B: Shared Components

- [ ] **4B.1** Create `components/common/recurrence-picker.tsx` — preset buttons (Does not repeat, Daily, Weekly, Monthly, Yearly, Custom) + custom interval builder (every N days/weeks with day-of-week checkboxes) + optional end date. Outputs RRULE string. Shows human-readable summary.
- [ ] **4B.2** Create `components/common/recurring-scope-dialog.tsx` — dialog with two radio options: "This occurrence only" / "This and all future occurrences". Adapted from calendar's `recurring-scope-dialog.tsx`. Returns scope string (`occurrence` | `future`).

**Acceptance**: Recurrence picker produces valid RRULE strings. Scope dialog returns correct scope value.

### 4C: Form Integration

- [ ] **4C.1** Integrate recurrence picker into `create-task-dialog.tsx` — show picker below due date field; when recurrence is set, make due date required
- [ ] **4C.2** Integrate recurrence picker into `create-reminder-dialog.tsx` — show picker below scheduled time field
- [ ] **4C.3** Add edit-task dialog (or modify existing) to show recurrence picker + trigger scope dialog when editing a recurring task
- [ ] **4C.4** Add scope dialog to task delete flow — when deleting a recurring task, show scope dialog before proceeding
- [ ] **4C.5** Add scope dialog to reminder edit/delete flows — same pattern as tasks

**Acceptance**: Users can set recurrence on create. Edit/delete of recurring items prompts for scope.

### 4D: List & Detail UI

- [ ] **4D.1** Update `task-card.tsx` — show recurring indicator icon (e.g., `Repeat` from lucide-react) when `seriesId` is present
- [ ] **4D.2** Update `reminder-card.tsx` — show recurring indicator icon when `seriesId` is present
- [ ] **4D.3** Show "Next due: [date]" toast/confirmation after completing a recurring task
- [ ] **4D.4** Add series history view — accessible from recurring task card, shows list of completed occurrences with dates using `useTaskSeriesHistory` hook

**Acceptance**: Recurring items are visually distinguishable. Completion shows next date. History is viewable.

---

## Phase 5: Verification & Polish [M]

- [ ] **5.1** Run all existing backend tests — confirm zero regressions
- [ ] **5.2** Run all existing frontend tests — confirm zero regressions
- [ ] **5.3** Docker build and start productivity-service — verify migration runs cleanly
- [ ] **5.4** Manual end-to-end test: create recurring task (weekly), complete it, verify next occurrence
- [ ] **5.5** Manual end-to-end test: create recurring reminder (daily), dismiss it, verify next occurrence
- [ ] **5.6** Manual end-to-end test: edit "this occurrence only", verify detached; edit "this and all future", verify series updated
- [ ] **5.7** Manual end-to-end test: delete "this and all future", verify series ends
- [ ] **5.8** Manual end-to-end test: verify non-recurring tasks and reminders are completely unaffected
- [ ] **5.9** Lint check across all changed files
