# Task 017 — UI Polish & Cleanup: Implementation Plan

Last Updated: 2026-03-31

## Executive Summary

Nine quality-of-life fixes across the Home Hub frontend plus one backend bug fix in the cooklang time parser. Most items are small, isolated changes. The two medium-effort items are adding calendar popovers (items 4 & 5, which share a pattern) and ingredients pagination (item 3). Total effort is estimated at ~1 day of focused work.

## Current State

- Meal planner grid shows redundant meal-type badges on recipe cards
- `ParseMinutes("1h 20m")` returns `1` instead of `80` due to naive regex
- Ingredients page loads all results with no pagination or count display
- Week navigation on Meal Planner and Calendar is limited to prev/next arrows
- Status filter shows raw "all" value instead of "All statuses" label
- Recipe tags render in raw lowercase
- Dashboard shows unnecessary role subtitle
- Household members page uses different header pattern than other detail pages

## Proposed Future State

All nine items resolved. Consistent UX patterns across pages, correct time parsing, proper pagination with counts, and calendar-based week selection.

## Implementation Phases

### Phase 1: Quick Wins (Items 1, 6, 7, 8, 9)

Five small, independent changes with no new dependencies. Can be done in parallel.

### Phase 2: Backend Bug Fix (Item 2)

Fix `ParseMinutes` to handle compound durations. Independent of frontend work.

### Phase 3: New UI Components (Items 4, 5)

Install shadcn Calendar + Popover components, then implement calendar popovers in both the week selector and calendar page. Items 4 and 5 share the same interaction pattern.

### Phase 4: Ingredients Pagination (Item 3)

Wire up existing backend pagination to the frontend with controls and count display.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `react-day-picker` conflicts with existing deps | Low | Medium | Check peer deps before installing |
| Calendar popover mobile responsiveness | Medium | Low | Test on small viewports, use responsive classes |
| ParseMinutes edge cases (fractional hours, unusual formats) | Medium | Low | Comprehensive test suite covering all known formats |
| Status filter label fix breaking other select components | Low | Low | Change is scoped to the specific select instance |

## Success Metrics

- All 9 acceptance criteria from the PRD pass
- No regressions in existing functionality
- `recipe-service` tests pass including new time-parsing cases
- Frontend builds cleanly with no type errors
