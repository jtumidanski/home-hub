# Tracker Emoji Ratings — Implementation Plan

Last Updated: 2026-04-09

---

## Executive Summary

Replace text/character-based sentiment indicators (`+`, `~`, `-`, "Good", "OK", "Bad") with face emoji (😊, 😐, 😞) across three frontend components. This is a small, low-risk, frontend-only change touching three files with no backend or data model modifications.

## Current State

Sentiment ratings are displayed in three locations:

1. **Today View** (`today-view.tsx:110-114`) — Buttons show `+ Good`, `~ OK`, `- Bad` using a `ratings` array with `emoji` and `label` fields
2. **Calendar Grid** (`calendar-grid.tsx:309-311`) — Cells show single `+`/`~`/`-` characters via inline ternary
3. **Calendar Grid Cell Editor** (`calendar-grid.tsx:379-381`) — Popover buttons show `+`/`~`/`-` via inline ternary
4. **Month Report** (`month-report.tsx:94-96`) — Stats show `+ N`, `~ N`, `- N` as text spans

## Proposed Future State

All four locations display face emoji instead:

| Rating   | Before      | After |
|----------|-------------|-------|
| positive | `+` / "Good" | 😊   |
| neutral  | `~` / "OK"   | 😐   |
| negative | `-` / "Bad"  | 😞   |

- Today view buttons show emoji only (no text label)
- Calendar cells show emoji (sized via existing `text-[10px]` class)
- Month report stats show `😊 N`, `😐 N`, `😞 N`

## Implementation Phases

### Phase 1: Update Today View

**File:** `frontend/src/components/features/tracker/today-view.tsx`

**Change:** Update the `ratings` array (lines 110-114) to use emoji and remove the label from the button render (line 120).

- Change `emoji` values: `"+"` → `"😊"`, `"~"` → `"😐"`, `"-"` → `"😞"`
- Remove `label` field (or keep for accessibility but don't render)
- Update button content from `{r.emoji} {r.label}` to just `{r.emoji}`

**Effort:** S

### Phase 2: Update Calendar Grid

**File:** `frontend/src/components/features/tracker/calendar-grid.tsx`

Two locations need changes:

1. **CellContent display** (line 311): Change the ternary mapping from `+`/`-`/`~` to `😊`/`😞`/`😐`
2. **CellEditor buttons** (line 381): Change the button text from `+`/`-`/`~` to `😊`/`😞`/`😐`

**Effort:** S

### Phase 3: Update Month Report

**File:** `frontend/src/components/features/tracker/month-report.tsx`

**Change:** Update the stat display spans (lines 95-97) from `+ {N}` / `~ {N}` / `- {N}` to `😊 {N}` / `😐 {N}` / `😞 {N}`.

**Effort:** S

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Emoji renders inconsistently across OS/browsers | Low | Low | Modern browsers handle emoji natively; the chosen emoji are in the basic set |
| Calendar cell overflow from emoji width | Low | Low | Emoji at `text-[10px]` should fit within existing `w-6 h-6` cells; verify visually |
| Confusion with numeric `+`/`-` buttons | N/A | N/A | Numeric scale uses separate component with different context — no overlap |

## Success Metrics

- All sentiment displays use face emoji across today view, calendar grid, and month report
- No layout shifts or overflow in calendar cells
- Existing selection behavior and color-coded bars unchanged
- Frontend builds without errors

## Required Resources and Dependencies

- No new packages or dependencies
- No backend changes
- No database migrations

## Timeline

**Estimated total effort:** ~15 minutes of implementation + visual verification. Single-commit change.
