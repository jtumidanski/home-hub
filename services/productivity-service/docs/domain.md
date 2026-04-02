# Domain

## Task

### Responsibility

Represents a household task with status tracking, optional due dates, optional owner assignment, and soft-delete support.

### Core Models

**Model** (`task.Model`)

| Field           | Type       |
|-----------------|------------|
| id              | uuid.UUID  |
| tenantID        | uuid.UUID  |
| householdID     | uuid.UUID  |
| title           | string     |
| notes           | string     |
| status          | string     |
| dueOn           | *time.Time |
| rolloverEnabled | bool       |
| ownerUserID     | *uuid.UUID |
| completedAt     | *time.Time |
| completedByUID  | *uuid.UUID |
| deletedAt       | *time.Time |
| createdAt       | time.Time  |
| updatedAt       | time.Time  |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- Tasks are always created with status "pending".
- Title is required; building a Model with an empty title returns `ErrTitleRequired`.
- Transitioning status from "pending" to "completed" sets completedAt and completedByUID.
- Transitioning status from "completed" to "pending" clears completedAt and completedByUID.
- Deletion is soft: sets deletedAt rather than removing the row.
- Soft-deleted tasks can be restored within a 3-day window.
- `IsDeleted()` returns true when deletedAt is non-nil.
- `IsCompleted()` returns true when status is "completed".

### Processors

**Processor** (`task.Processor`)

| Method                                                                              | Description                                     |
|-------------------------------------------------------------------------------------|-------------------------------------------------|
| `ByIDProvider(id)`                                                                  | Lazy lookup by ID                               |
| `AllProvider(includeDeleted)`                                                       | Lazy list, optionally including soft-deleted     |
| `ByStatusProvider(status)`                                                          | Lazy list filtered by status, excluding deleted  |
| `Create(tenantID, hhID, title, notes, dueOn, rolloverEnabled, ownerUserID)`         | Creates a task with status "pending"             |
| `Update(id, title, notes, status, dueOn, rolloverEnabled, ownerUserID, userID)`     | Updates a task, handles status transitions       |
| `Delete(id)`                                                                        | Soft-deletes a task                             |
| `Restore(id)`                                                                       | Restores if deleted within 3-day window          |
| `PendingCount()`                                                                    | Count of non-deleted pending tasks               |
| `CompletedTodayCount()`                                                             | Count of tasks completed today                   |
| `OverdueCount()`                                                                    | Count of pending tasks past due date             |

**Errors:**

| Error            | Condition                               |
|------------------|-----------------------------------------|
| `ErrNotFound`    | Task does not exist                     |
| `ErrNotDeleted`  | Restore attempted on non-deleted task   |
| `ErrRestoreWindow` | Restore attempted after 3-day window |

---

## Reminder

### Responsibility

Represents a scheduled reminder for a household, with optional owner assignment, dismissal, and snoozing.

### Core Models

**Model** (`reminder.Model`)

| Field            | Type       |
|------------------|------------|
| id               | uuid.UUID  |
| tenantID         | uuid.UUID  |
| householdID      | uuid.UUID  |
| title            | string     |
| notes            | string     |
| scheduledFor     | time.Time  |
| ownerUserID      | *uuid.UUID |
| lastDismissedAt  | *time.Time |
| lastSnoozedUntil | *time.Time |
| createdAt        | time.Time  |
| updatedAt        | time.Time  |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- Title is required; building a Model with an empty title returns `ErrTitleRequired`.
- ScheduledFor is required; building a Model with a zero scheduledFor returns `ErrScheduledForRequired`.
- A reminder is active when: scheduledFor is at or before now, AND not dismissed, AND not snoozed into the future.
- Dismissal sets lastDismissedAt; dismissed reminders are never active.
- Snooze sets lastSnoozedUntil; snoozed reminders are inactive until that time passes.
- Allowed snooze durations: 10, 30, or 60 minutes.
- Reminders use hard deletes.

### Processors

**Processor** (`reminder.Processor`)

| Method                                                               | Description                                        |
|----------------------------------------------------------------------|----------------------------------------------------|
| `ByIDProvider(id)`                                                   | Lazy lookup by ID                                  |
| `AllProvider()`                                                      | Lazy list of all reminders                         |
| `Create(tenantID, hhID, title, notes, scheduledFor, ownerUserID)`    | Creates a reminder                                 |
| `Update(id, title, notes, scheduledFor, ownerUserID)`                | Updates a reminder                                 |
| `Delete(id)`                                                         | Hard-deletes a reminder                            |
| `Dismiss(id)`                                                        | Sets lastDismissedAt to now                        |
| `Snooze(id, durationMinutes)`                                        | Validates duration, sets lastSnoozedUntil          |
| `DueNowCount()`                                                      | Count of due, undismissed, unsnoozed reminders     |
| `UpcomingCount()`                                                     | Count of future, undismissed reminders             |
| `SnoozedCount()`                                                      | Count of currently snoozed, undismissed reminders  |

---

## Reminder Dismissal

### Responsibility

Audit record for when a reminder is dismissed.

### Core Models

**Model** (`dismissal.Model`)

| Field           | Type      |
|-----------------|-----------|
| id              | uuid.UUID |
| tenantID        | uuid.UUID |
| householdID     | uuid.UUID |
| reminderID      | uuid.UUID |
| createdByUserID | uuid.UUID |
| createdAt       | time.Time |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- ReminderID is required; building a Model with a nil reminderID returns `ErrReminderIDRequired`.
- CreatedByUserID is required; building a Model with a nil createdByUserID returns `ErrCreatedByRequired`.

### Processors

**Processor** (`dismissal.Processor`)

| Method                                          | Description                                                         |
|-------------------------------------------------|---------------------------------------------------------------------|
| `Create(tenantID, householdID, reminderID, userID)` | Dismisses the reminder and creates an audit record              |

---

## Reminder Snooze

### Responsibility

Audit record for when a reminder is snoozed.

### Core Models

**Model** (`snooze.Model`)

| Field           | Type      |
|-----------------|-----------|
| id              | uuid.UUID |
| tenantID        | uuid.UUID |
| householdID     | uuid.UUID |
| reminderID      | uuid.UUID |
| durationMinutes | int       |
| snoozedUntil    | time.Time |
| createdByUserID | uuid.UUID |
| createdAt       | time.Time |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- ReminderID is required; building a Model with a nil reminderID returns `ErrReminderIDRequired`.
- CreatedByUserID is required; building a Model with a nil createdByUserID returns `ErrCreatedByRequired`.
- DurationMinutes must be positive; building a Model with a non-positive value returns `ErrDurationMinutesRequired`.
- SnoozedUntil is required; building a Model with a zero snoozedUntil returns `ErrSnoozedUntilRequired`.

### Processors

**Processor** (`snooze.Processor`)

| Method                                                              | Description                                                     |
|---------------------------------------------------------------------|-----------------------------------------------------------------|
| `Create(tenantID, householdID, reminderID, userID, durationMinutes)` | Snoozes the reminder and creates an audit record               |

---

## Task Restoration

### Responsibility

Audit record for when a soft-deleted task is restored.

### Core Models

**Model** (`restoration.Model`)

| Field           | Type      |
|-----------------|-----------|
| id              | uuid.UUID |
| tenantID        | uuid.UUID |
| householdID     | uuid.UUID |
| taskID          | uuid.UUID |
| createdByUserID | uuid.UUID |
| createdAt       | time.Time |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- TaskID is required; building a Model with a nil taskID returns `ErrTaskIDRequired`.
- CreatedByUserID is required; building a Model with a nil createdByUserID returns `ErrCreatedByRequired`.

### Processors

**Processor** (`restoration.Processor`)

| Method                                          | Description                                                    |
|-------------------------------------------------|----------------------------------------------------------------|
| `Create(tenantID, householdID, taskID, userID)` | Restores the task and creates an audit record                  |

---

## Summary

### Responsibility

Provides aggregated counts for tasks and reminders, used by dashboard views.

### Core Models

**TaskSummary** (`summary.TaskSummary`)

| Field               | Type  |
|---------------------|-------|
| PendingCount        | int64 |
| CompletedTodayCount | int64 |
| OverdueCount        | int64 |

**ReminderSummary** (`summary.ReminderSummary`)

| Field         | Type  |
|---------------|-------|
| DueNowCount   | int64 |
| UpcomingCount | int64 |
| SnoozedCount  | int64 |

**DashboardSummary** (`summary.DashboardSummary`)

| Field            | Type      |
|------------------|-----------|
| PendingTaskCount | int64     |
| DueReminderCount | int64     |
| GeneratedAt      | time.Time |
