# Recurring Tasks & Reminders — Context & Dependencies

Last Updated: 2026-03-31

---

## Key Files — Backend (productivity-service)

### Task Domain
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `internal/task/model.go` | Immutable domain model with getters | Add series fields: seriesID, recurrenceRule, recurrenceStart, recurrenceEndDate, occurrenceIndex + `IsRecurring()` method |
| `internal/task/entity.go` | GORM entity, `Make()`, `ToEntity()`, `Migration()` | Add series columns, update Make/ToEntity/Migration |
| `internal/task/builder.go` | Builder pattern with validation | Add setters for series fields |
| `internal/task/administrator.go` | DB create/update/softDelete/restore | Extend `create` with series params; add `createNextOccurrence`; add `deleteBySeries` for future-scope delete |
| `internal/task/provider.go` | DB queries (getByID, getAll, counts) | Add `getBySeriesID` for history; add `getActiveBySeriesID` for uncomplete cleanup |
| `internal/task/processor.go` | Business logic (Create, Update, Delete) | Trigger next occurrence on completion; scope-aware Update/Delete; uncomplete cleanup |
| `internal/task/rest.go` | REST models, CreateRequest, UpdateRequest, Transform | Add recurrence fields to all structs |
| `internal/task/resource.go` | HTTP handlers (routes, request parsing) | Parse recurrence fields on create/update; parse scope on update/delete; add series history route |

### Reminder Domain
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `internal/reminder/model.go` | Immutable domain model | Add series fields + `IsRecurring()` |
| `internal/reminder/entity.go` | GORM entity | Add series columns |
| `internal/reminder/builder.go` | Builder with validation | Add setters for series fields |
| `internal/reminder/administrator.go` | DB create/update/dismiss/snooze/delete | Extend `create` with series params; add `createNextOccurrence` |
| `internal/reminder/provider.go` | DB queries | Add `getBySeriesID` |
| `internal/reminder/processor.go` | Business logic | Trigger next occurrence on dismiss |
| `internal/reminder/rest.go` | REST models | Add recurrence fields |
| `internal/reminder/resource.go` | HTTP handlers | Parse recurrence fields; scope parameter |

### Reminder Dismissal Subdomain
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `internal/reminder/dismissal/processor.go` | Creates dismissal audit record, calls `remProc.Dismiss()` | After dismiss, check if reminder is recurring and generate next occurrence |

### New Package
| File | Purpose |
|------|---------|
| `internal/recurrence/recurrence.go` | RRULE parsing, next occurrence computation |
| `internal/recurrence/recurrence_test.go` | Unit tests for all recurrence patterns |

### Service Bootstrap
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `cmd/main.go` | Route init, migrations | No structural changes — GORM AutoMigrate picks up entity changes automatically |

---

## Key Files — Frontend

### Types & Schemas
| File | Changes Needed |
|------|----------------|
| `frontend/src/types/models/task.ts` | Add series attributes: seriesId, recurrenceRule, recurrenceStart, recurrenceEndDate, occurrenceIndex |
| `frontend/src/types/models/reminder.ts` | Add series attributes |
| `frontend/src/lib/schemas/task.schema.ts` | Add recurrenceRule, recurrenceEndDate to Zod schema |
| `frontend/src/lib/schemas/reminder.schema.ts` | Add recurrenceRule, recurrenceEndDate to Zod schema |

### API & Hooks
| File | Changes Needed |
|------|----------------|
| `frontend/src/services/api/productivity.ts` | Add series history method; extend create/update/delete to pass recurrence/scope params |
| `frontend/src/lib/hooks/api/use-tasks.ts` | Add `useTaskSeriesHistory()` hook; pass recurrence fields through mutations |
| `frontend/src/lib/hooks/api/use-reminders.ts` | Pass recurrence fields through mutations |

### Components
| File | Changes Needed |
|------|----------------|
| `frontend/src/components/features/tasks/create-task-dialog.tsx` | Integrate recurrence picker |
| `frontend/src/components/features/reminders/create-reminder-dialog.tsx` | Integrate recurrence picker |
| `frontend/src/components/features/tasks/task-card.tsx` | Add recurring indicator icon |
| `frontend/src/components/features/reminders/reminder-card.tsx` | Add recurring indicator icon |
| `frontend/src/pages/TasksPage.tsx` | Scope dialog on edit/delete of recurring items |
| `frontend/src/pages/RemindersPage.tsx` | Scope dialog on edit/delete of recurring items |

### New Components
| File | Purpose |
|------|---------|
| `frontend/src/components/common/recurrence-picker.tsx` | Shared recurrence pattern selector (presets + custom interval) |
| `frontend/src/components/common/recurring-scope-dialog.tsx` | "This occurrence" vs "This and all future" dialog (adapted from calendar's `recurring-scope-dialog.tsx`) |

---

## Reference Implementations

### Calendar Service — RRULE Handling
- `frontend/src/lib/schemas/calendar-event.schema.ts` — `RECURRENCE_OPTIONS` array with RRULE preset strings
- `frontend/src/components/features/calendar/recurring-scope-dialog.tsx` — Scope choice dialog pattern

### Calendar Service — Backend Recurrence
- Calendar events store recurrence as RRULE and expand via Google's `singleEvents=true`
- Our approach differs: generate-on-trigger, no pre-expansion

---

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Storage strategy | Generate-on-trigger (single active occurrence per series) | Avoids background jobs and storage bloat; tasks/reminders only need the next occurrence, not a month view |
| RRULE storage | Store RFC 5545 RRULE string in column | Future-proof for richer patterns; can parse with standard libraries |
| Schedule anchoring | Anchor to original schedule, not completion date | Prevents schedule drift; matches user mental model ("trash day is Tuesday") |
| Edit/delete scope | Two choices: "this occurrence" / "this and all future" | Simpler than calendar's three-way; "entire series including past" is rarely useful for tasks |
| Completion history | Keep completed instances as separate rows | Enables streak tracking; completed tasks already persist via soft-delete model |
| Reminder next occurrence | Create immediately on dismiss (dormant until scheduled_for) | Keeps series visible in reminder list; consistent with generate-on-trigger model |
| End conditions | Optional end date only, no max count | Simplest UX; users can manually stop a series |
| Recurrence presets | Simple presets backed by RRULE internally | Covers 95% of household use cases; defer full RRULE editor |

---

## Invariants to Preserve

1. **One active occurrence per series**: At any time, only one non-completed, non-deleted task (or one non-dismissed reminder) exists per series_id
2. **Non-recurring behavior unchanged**: All existing code paths must work identically when series fields are nil
3. **Tenant isolation**: series_id is a UUID (globally unique), but all queries still filter by tenant_id and household_id
4. **Audit trail intact**: Dismissal and snooze audit tables continue to record per-occurrence; restoration audit works for recurring tasks
5. **Transaction safety**: Completion/dismissal + next occurrence creation happen in a single DB transaction
