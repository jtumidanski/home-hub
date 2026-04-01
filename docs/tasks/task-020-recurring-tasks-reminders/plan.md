# Recurring Tasks & Reminders — Implementation Plan

Last Updated: 2026-03-31

---

## Executive Summary

Add recurrence support to the productivity-service's tasks and reminders. Users define a repeating pattern (daily, weekly, monthly, etc.) and the system generates the next occurrence when the current one is completed (tasks) or dismissed (reminders). The implementation follows a generate-on-trigger model — no background jobs, no pre-expansion of future occurrences.

The work spans four phases: (1) RRULE computation library, (2) backend domain/storage/API changes, (3) frontend recurrence UI, and (4) integration testing and polish. The calendar-service's existing RRULE handling and the frontend's `RecurringScopeDialog` serve as reference implementations.

---

## Current State Analysis

### What exists today
- **Tasks**: CRUD with soft-delete, status transitions (pending/completed), rollover, ownership, 3-day restore window
- **Reminders**: CRUD with hard-delete, dismiss/snooze lifecycle, ownership, audit tables for dismissals and snoozes
- **No recurrence**: Explicitly deferred in task-001; no series, RRULE, or occurrence concepts anywhere
- **Calendar precedent**: The calendar-service already handles RRULE for Google Calendar events, with a `RecurringScopeDialog` in the frontend

### What changes
- New `recurrence` package in productivity-service for RRULE parsing and next-date computation
- New columns on `tasks` and `reminders` tables: `series_id`, `recurrence_rule`, `recurrence_start`, `recurrence_end_date`, `occurrence_index`
- Modified processors: task completion and reminder dismissal trigger next-occurrence creation
- Extended REST API: recurrence fields on create/update, scope parameter on update/delete, series history endpoint
- New frontend components: recurrence picker, scope dialog reuse, recurring indicators

---

## Proposed Future State

After implementation:
- Users create recurring tasks/reminders through simple preset patterns (Daily, Weekly, Monthly, etc.)
- Completing a recurring task auto-creates the next pending occurrence anchored to the original schedule
- Dismissing a recurring reminder auto-creates the next occurrence at the computed future time
- Edit/delete operations offer "this occurrence" vs "this and all future" scope choices
- Completed recurring task history is queryable per series
- All new data is tenant-scoped and multi-tenant safe

---

## Implementation Phases

### Phase 1: Recurrence Computation Package (Backend Foundation)

Create a standalone `recurrence` package that parses RRULE strings and computes next occurrence dates. This has zero dependencies on the rest of the service and can be built and tested in isolation.

**Key decisions:**
- Use a Go RRULE library (e.g., `github.com/teambition/rrule-go`) rather than hand-rolling RFC 5545 parsing
- Support the preset subset: FREQ=DAILY, FREQ=WEEKLY with BYDAY, FREQ=MONTHLY with BYMONTHDAY, FREQ=YEARLY, plus INTERVAL for custom N-day/N-week patterns
- UNTIL clause for end dates
- Computation is pure: takes (RRULE string, anchor date, current occurrence date) and returns the next date

### Phase 2: Backend Domain & Storage (Core Logic)

Extend task and reminder domain models, entities, builders, administrators, providers, and processors to support recurrence fields and the generate-on-trigger lifecycle.

**Sub-phases:**
1. **Schema**: Add columns to entities, run auto-migration
2. **Models & Builders**: Add series fields to immutable models and builders
3. **Administrators**: Extend `create` functions to accept series fields; add `createNextOccurrence` helpers
4. **Providers**: Add `getBySeriesID` query for task history; add `getActiveBySeriesID` for uncomplete cleanup
5. **Processors**: Modify `Update` (status transition to completed) to trigger next occurrence creation; modify dismissal processor similarly; add scope-aware update/delete logic

### Phase 3: Backend REST API (Transport Layer)

Extend REST models, request/response types, and HTTP handlers.

**Changes:**
- `CreateRequest`: add `recurrenceRule`, `recurrenceEndDate` fields
- `UpdateRequest`: add `editScope`, `recurrenceRule`, `recurrenceEndDate` fields
- `RestModel`: add `seriesId`, `recurrenceRule`, `recurrenceStart`, `recurrenceEndDate`, `occurrenceIndex`
- `DeleteTask`/`DeleteReminder`: parse `scope` query parameter
- New `GET /api/v1/tasks/series/{seriesId}` handler
- Error handling for missing scope on recurring items

### Phase 4: Frontend (UI)

Build the recurrence picker, scope dialog, and recurring indicators.

**Components:**
1. **Recurrence picker**: Shared component with preset buttons + custom interval builder. Reuse pattern from calendar's `RECURRENCE_OPTIONS`
2. **Scope dialog**: Adapt calendar's `RecurringScopeDialog` for tasks/reminders (two choices: "this occurrence" / "this and all future")
3. **Form integration**: Add recurrence picker to `create-task-dialog.tsx` and `create-reminder-dialog.tsx`; add scope dialog to edit/delete flows
4. **List indicators**: Show recurring icon on task/reminder cards
5. **Series history**: Add completion history view accessible from recurring task detail
6. **API hooks**: Extend `use-tasks.ts` and `use-reminders.ts` hooks to pass recurrence attributes; add series history hook
7. **Schema updates**: Extend Zod schemas for recurrence fields

---

## Risk Assessment and Mitigation

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| RRULE edge cases (DST, leap years, month-end) | Medium | Medium | Use battle-tested RRULE library; write targeted edge-case tests |
| Uncomplete race condition (two users acting on same series) | Medium | Low | DB transaction wraps completion + next occurrence creation |
| Orphaned occurrences if completion fails mid-transaction | High | Low | Single DB transaction for status change + occurrence creation |
| Breaking existing non-recurring behavior | High | Low | All new columns nullable; no changes to existing code paths when series fields are nil |
| Frontend complexity creep in recurrence picker | Medium | Medium | Start with simple presets only; defer custom RRULE editor |

---

## Success Metrics

- All existing task and reminder tests continue to pass unchanged
- New unit tests for recurrence computation cover: daily, weekly (single/multi day), monthly, yearly, custom intervals, end dates, schedule anchoring
- New integration tests for: task completion generating next occurrence, reminder dismissal generating next occurrence, scope-aware edit/delete, uncomplete cleanup
- Frontend E2E: create recurring task, complete it, verify next occurrence appears

---

## Required Resources and Dependencies

### External Dependencies
- Go RRULE library: `github.com/teambition/rrule-go` (or equivalent)
- No new frontend dependencies — reuse existing shadcn/ui components

### Internal Dependencies
- `shared/go/database` — existing query helpers (EntityProvider, SliceQuery)
- `shared/go/model` — existing Map/SliceMap providers
- Calendar service `RecurringScopeDialog` — reference implementation for frontend scope dialog

---

## Timeline Estimates

| Phase | Effort | Dependencies |
|-------|--------|--------------|
| Phase 1: Recurrence package | S | None |
| Phase 2: Backend domain & storage | L | Phase 1 |
| Phase 3: Backend REST API | M | Phase 2 |
| Phase 4: Frontend | L | Phase 3 |
| **Total** | **XL** | Sequential |

Phases 1-3 are backend and strictly sequential. Phase 4 (frontend) depends on Phase 3 API being complete. Within Phase 4, schema/hooks/API-client work can proceed before component work.
