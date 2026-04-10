# Overdue Task Display on Mobile — Task Checklist

Last Updated: 2026-04-09

---

## Phase 1: Fix TaskCard Component

- [ ] Add `isTaskOverdue` to import in `task-card.tsx`
- [ ] Compute overdue status: `const overdue = isTaskOverdue(task)`
- [ ] Update Badge variant to use three-way logic: `isCompleted ? "secondary" : overdue ? "destructive" : "default"`
- [ ] Update Badge text: `overdue ? "overdue" : attributes.status`

## Phase 2: Verify

- [ ] Frontend builds without errors
- [ ] Manual test: overdue pending task shows red "overdue" badge on mobile
- [ ] Manual test: non-overdue pending task shows default "pending" badge on mobile
- [ ] Manual test: completed task shows secondary "completed" badge on mobile
- [ ] Manual test: task with no due date shows default "pending" badge on mobile
