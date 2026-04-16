# Dashboard Widget Skip/Partial Status — Implementation Plan

Last Updated: 2026-04-16

---

## Executive Summary

The dashboard Habits and Workout widgets currently show only two visual states (done/pending), but the backend already returns skip and partial data. This is a frontend-only change to two widget components, adding distinct visual indicators for skipped items (both widgets) and partially completed items (workout widget). No backend, API, or data model changes are needed.

## Current State Analysis

### Habits Widget (`habits-widget.tsx`)
- Builds a `completedItemIds` set by filtering entries where `!skipped && value !== null`
- Renders each item as either **done** (green `Check` + strikethrough) or **pending** (hollow `Circle`)
- Skipped entries are excluded from the completed set, so they render identically to pending items
- The `skipped` field is already present in the entry data from `/api/v1/trackers/today`

### Workout Widget (`workout-widget.tsx`)
- Checks `item.performance?.status === "done"` for each exercise
- Renders each item as either **done** (green `Check` + strikethrough) or **pending** (hollow `Circle`)
- `"skipped"` and `"partial"` statuses both render as pending since only `"done"` is checked
- All four status values (`pending`, `done`, `skipped`, `partial`) are already in the API response from `/api/v1/workouts/today`

## Proposed Future State

### Habits Widget — 3 visual states
| State | Icon | Icon Color | Text |
|-------|------|------------|------|
| Done | `Check` | `text-green-600` | Strikethrough, muted |
| Skipped | `SkipForward` | `text-muted-foreground` | Strikethrough, muted |
| Pending | `Circle` | `text-muted-foreground` | Normal |

### Workout Widget — 4 visual states
| State | Icon | Icon Color | Text |
|-------|------|------------|------|
| Done | `Check` | `text-green-600` | Strikethrough, muted |
| Skipped | `SkipForward` | `text-muted-foreground` | Strikethrough, muted |
| Partial | `CircleDot` | `text-yellow-600` | Normal |
| Pending | `Circle` | `text-muted-foreground` | Normal |

## Implementation Phases

### Phase 1: Habits Widget Update

**Goal:** Add skipped state detection and rendering to the habits widget.

**Task 1.1 — Build `skippedItemIds` set** (Effort: S)
- Add a second `Set` alongside `completedItemIds` that collects tracking item IDs from entries where `skipped === true`
- Acceptance: Skipped entries are identified separately from completed and pending

**Task 1.2 — Render skipped state** (Effort: S)
- Import `SkipForward` from lucide-react
- Add a third branch to the item rendering: if `skippedItemIds.has(item.id)`, render `SkipForward` icon with muted strikethrough text
- Add `aria-label="Skipped"` to the icon for accessibility
- Acceptance: Skipped habits show `SkipForward` icon with strikethrough, visually distinct from both done (green check) and pending (hollow circle)

### Phase 2: Workout Widget Update

**Goal:** Add skipped and partial state rendering to the workout widget.

**Task 2.1 — Expand status rendering logic** (Effort: S)
- Import `SkipForward` and `CircleDot` from lucide-react
- Replace the binary `done` boolean with a status variable derived from `item.performance?.status`
- Render four states: done (existing), skipped (`SkipForward` + strikethrough), partial (`CircleDot` yellow + normal text), pending (existing)
- Add `aria-label` attributes to each icon
- Acceptance: All four workout performance statuses render with distinct icons and text treatments

### Phase 3: Verification

**Task 3.1 — Visual verification** (Effort: S)
- Start the dev server and navigate to the dashboard
- Verify all states render correctly with test data (or by skipping/completing items via the habits and workout pages, then returning to the dashboard)
- Confirm existing behaviors are unaffected: loading skeletons, error states, empty states, navigation links, pull-to-refresh

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `SkipForward`/`CircleDot` not in project's lucide-react version | Low | Low | Both icons exist in lucide-react since early versions; verify import works |
| Visual inconsistency between widgets | Low | Low | Use identical icon/color/text classes for "skipped" in both widgets |

## Success Metrics

- Skipped habits and workouts are visually distinguishable from pending items on the dashboard
- Partially completed workouts are visually distinguishable from pending and done items
- No regressions in existing widget functionality

## Required Resources and Dependencies

- **Files to modify:** 2 frontend component files
- **Dependencies:** lucide-react (already installed, just importing additional icons)
- **Backend changes:** None
- **API changes:** None

## Timeline Estimate

This is a small, well-scoped change: ~30 minutes of implementation + verification.
