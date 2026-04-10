# Overdue Task Display on Mobile — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-09
---

## 1. Overview

The mobile tasks screen does not visually indicate when a task is overdue. On desktop, the `TasksPage` data table uses `isTaskOverdue()` to replace the "pending" badge with a red "overdue" badge (destructive variant). The mobile `TaskCard` component skips this check entirely and always renders the raw `attributes.status` value, so overdue tasks appear as "pending" with default styling.

This is a frontend-only fix. The backend already provides `dueOn` dates and the `isTaskOverdue()` utility already exists — the mobile card simply doesn't use it.

## 2. Goals

Primary goals:
- Overdue tasks display a red "overdue" badge on mobile, matching desktop behavior
- Due date text on overdue tasks is visually consistent with desktop styling

Non-goals:
- Backend changes to overdue counting or timezone handling (separate issue)
- Push notifications or reminders
- Changes to the overdue filter logic (already works)

## 3. User Stories

- As a user viewing tasks on mobile, I want overdue tasks to show a red "overdue" badge so I can immediately see which tasks are past due
- As a user switching between desktop and mobile, I want consistent overdue indicators so I'm not confused by different displays

## 4. Functional Requirements

### 4.1 Overdue Badge on TaskCard

- When `isTaskOverdue(task)` returns `true`, the status Badge must:
  - Display text `"overdue"` instead of `"pending"`
  - Use variant `"destructive"` (red) instead of `"default"`
- When the task is completed, behavior is unchanged: Badge shows `"completed"` with variant `"secondary"`
- When the task is pending and not overdue, behavior is unchanged: Badge shows `"pending"` with variant `"default"`

### 4.2 Due Date Styling

- Due date text styling on mobile should remain consistent with desktop (muted foreground text with calendar icon, no additional overdue-specific styling on the date text itself — desktop does not style the date differently for overdue tasks either)

## 5. API Surface

No API changes required. The existing task response already includes `dueOn` and `status` fields.

## 6. Data Model

No data model changes required.

## 7. Service Impact

| Service | Change |
|---------|--------|
| **frontend** | Update `TaskCard` component to import and use `isTaskOverdue()`, apply destructive badge variant and "overdue" text when applicable |

## 8. Non-Functional Requirements

- No performance impact — `isTaskOverdue()` is a pure function doing a simple date comparison per card
- No accessibility changes needed — the destructive badge variant already has appropriate contrast

## 9. Open Questions

None.

## 10. Acceptance Criteria

- [ ] On mobile, a pending task with a `dueOn` date in the past displays a red "overdue" badge
- [ ] On mobile, a pending task with a `dueOn` date today or in the future displays the default "pending" badge
- [ ] On mobile, a completed task always displays the secondary "completed" badge regardless of due date
- [ ] On mobile, a task with no due date always displays the default "pending" badge (never overdue)
- [ ] Badge variant and text match desktop behavior exactly for all status/date combinations
