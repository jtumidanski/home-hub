# Domain

## Task

### Responsibility

Represents a household task with status tracking, optional due dates, and soft-delete support.

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
| completedAt     | *time.Time |
| completedByUID  | *uuid.UUID |
| deletedAt       | *time.Time |
| createdAt       | time.Time  |
| updatedAt       | time.Time  |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- Tasks are always created with status "pending".
- Transitioning status from "pending" to "completed" sets completedAt and completedByUID.
- Transitioning status from "completed" to "pending" clears completedAt and completedByUID.
- Deletion is soft: sets deletedAt rather than removing the row.
- Soft-deleted tasks can be restored within a 3-day window.
- `IsDeleted()` returns true when deletedAt is non-nil.
- `IsCompleted()` returns true when status is "completed".

### Processors

**Processor** (`task.Processor`)

| Method                                                                   | Description                                     |
|--------------------------------------------------------------------------|-------------------------------------------------|
| `ByIDProvider(id)`                                                       | Lazy lookup by ID                               |
| `AllProvider(includeDeleted)`                                            | Lazy list, optionally including soft-deleted     |
| `ByStatusProvider(status)`                                               | Lazy list filtered by status, excluding deleted  |
| `Create(tenantID, hhID, title, notes, dueOn, rolloverEnabled)`          | Creates a task with status "pending"             |
| `Update(id, title, notes, status, dueOn, rolloverEnabled, userID)`      | Updates a task, handles status transitions       |
| `Delete(id)`                                                            | Soft-deletes a task                             |
| `Restore(id)`                                                           | Restores if deleted within 3-day window          |
| `PendingCount()`                                                        | Count of non-deleted pending tasks               |
| `CompletedTodayCount()`                                                 | Count of tasks completed today                   |
| `OverdueCount()`                                                        | Count of pending tasks past due date             |

**Errors:**

| Error            | Condition                               |
|------------------|-----------------------------------------|
| `ErrNotFound`    | Task does not exist                     |
| `ErrNotDeleted`  | Restore attempted on non-deleted task   |
| `ErrRestoreWindow` | Restore attempted after 3-day window |

---

## Reminder

### Responsibility

Represents a scheduled reminder for a household, with support for dismissal and snoozing.

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
| lastDismissedAt  | *time.Time |
| lastSnoozedUntil | *time.Time |
| createdAt        | time.Time  |
| updatedAt        | time.Time  |

All fields are immutable after construction. Access is through getter methods.

### Invariants

- A reminder is active when: scheduledFor is at or before now, AND not dismissed, AND not snoozed into the future.
- Dismissal sets lastDismissedAt; dismissed reminders are never active.
- Snooze sets lastSnoozedUntil; snoozed reminders are inactive until that time passes.
- Allowed snooze durations: 10, 30, or 60 minutes.
- Reminders use hard deletes.

### Processors

**Processor** (`reminder.Processor`)

| Method                                                        | Description                                        |
|---------------------------------------------------------------|----------------------------------------------------|
| `ByIDProvider(id)`                                            | Lazy lookup by ID                                  |
| `AllProvider()`                                               | Lazy list of all reminders                         |
| `Create(tenantID, hhID, title, notes, scheduledFor)`          | Creates a reminder                                 |
| `Update(id, title, notes, scheduledFor)`                      | Updates a reminder                                 |
| `Delete(id)`                                                  | Hard-deletes a reminder                            |
| `Dismiss(id)`                                                 | Sets lastDismissedAt to now                        |
| `Snooze(id, durationMinutes)`                                 | Validates duration, sets lastSnoozedUntil          |
| `DueNowCount()`                                              | Count of due, undismissed, unsnoozed reminders     |
| `UpcomingCount()`                                             | Count of future, undismissed reminders             |
| `SnoozedCount()`                                              | Count of currently snoozed, undismissed reminders  |

---

## Reminder Dismissal

### Responsibility

Audit record for when a reminder is dismissed.

### Core Models

**Entity** (`reminderdismissal.Entity`)

| Field           | Type      |
|-----------------|-----------|
| id              | uuid.UUID |
| tenantId        | uuid.UUID |
| householdId     | uuid.UUID |
| reminderId      | uuid.UUID |
| createdByUserId | uuid.UUID |
| createdAt       | time.Time |

---

## Reminder Snooze

### Responsibility

Audit record for when a reminder is snoozed.

### Core Models

**Entity** (`remindersnooze.Entity`)

| Field           | Type      |
|-----------------|-----------|
| id              | uuid.UUID |
| tenantId        | uuid.UUID |
| householdId     | uuid.UUID |
| reminderId      | uuid.UUID |
| durationMinutes | int       |
| snoozedUntil    | time.Time |
| createdByUserId | uuid.UUID |
| createdAt       | time.Time |

---

## Task Restoration

### Responsibility

Audit record for when a soft-deleted task is restored.

### Core Models

**Entity** (`taskrestoration.Entity`)

| Field           | Type      |
|-----------------|-----------|
| id              | uuid.UUID |
| tenantId        | uuid.UUID |
| householdId     | uuid.UUID |
| taskId          | uuid.UUID |
| createdByUserId | uuid.UUID |
| createdAt       | time.Time |

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
| HouseholdName    | string    |
| Timezone         | string    |
| PendingTaskCount | int64     |
| DueReminderCount | int64     |
| GeneratedAt      | time.Time |
