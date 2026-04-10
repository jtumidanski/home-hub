# Overdue Task Display on Mobile — Implementation Plan

Last Updated: 2026-04-09

---

## Executive Summary

The mobile `TaskCard` component does not use `isTaskOverdue()` to indicate overdue status, causing overdue tasks to appear as "pending" with default styling. The desktop `TasksPage` already implements the correct behavior. This is a minimal frontend-only change: import the existing utility into `TaskCard` and apply the same badge logic.

## Current State Analysis

### Desktop (correct behavior)
- `TasksPage.tsx:142` calls `isTaskOverdue(row.original)` per row
- Badge renders with variant `"destructive"` and text `"overdue"` when true
- Falls back to `"secondary"` for completed, `"default"` for pending

### Mobile (broken behavior)
- `task-card.tsx:51` renders `<Badge variant={isCompleted ? "secondary" : "default"}>`
- Badge text is always `{attributes.status}` — never `"overdue"`
- `isTaskOverdue` is not imported or called

## Proposed Future State

`TaskCard` will import `isTaskOverdue` and compute overdue status, then use the same three-way badge logic as desktop:

| State | Variant | Text |
|-------|---------|------|
| Completed | `secondary` | `completed` |
| Pending + overdue | `destructive` | `overdue` |
| Pending + not overdue | `default` | `pending` |

## Implementation

### Phase 1: Fix TaskCard Component

**File:** `frontend/src/components/features/tasks/task-card.tsx`

1. Add `isTaskOverdue` to the import from `@/types/models/task`
2. Compute `const overdue = isTaskOverdue(task)` inside the component
3. Update Badge variant: `isCompleted ? "secondary" : overdue ? "destructive" : "default"`
4. Update Badge text: `overdue ? "overdue" : attributes.status`

### Phase 2: Verify

1. Run frontend build to confirm no type errors
2. Manual verification on mobile viewport: overdue task shows red badge, non-overdue shows default, completed shows secondary

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `isTaskOverdue` logic incorrect | Very Low | Low | Already tested with unit tests in `task.test.ts` |
| Import path issue | Very Low | Low | Same import pattern used in `TasksPage.tsx` |

## Success Metrics

- Overdue tasks display red "overdue" badge on mobile
- Badge behavior matches desktop exactly across all status/date combinations
- No regressions in desktop view

## Required Resources and Dependencies

- **Dependencies:** None — `isTaskOverdue()` already exists and is tested
- **Services affected:** Frontend only

## Effort Estimate

**Size: S** — ~5 lines changed in a single file
