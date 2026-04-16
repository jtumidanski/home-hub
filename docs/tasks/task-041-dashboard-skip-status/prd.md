# Dashboard Widget Skip/Partial Status — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-16
---

## 1. Overview

The dashboard Habits and Workout widgets (introduced in task-034) currently display two visual states: **done** (green checkmark + strikethrough) and **pending** (hollow circle). Both backend APIs already return skip data — `skipped: true` on tracker entries and `status: "skipped"` on workout performance — but the widgets ignore it, rendering skipped items identically to pending ones.

This creates a misleading view of the user's day: items they have consciously decided to skip appear as if they still need attention. The Workout widget also has a `"partial"` performance status that is rendered the same as pending.

This feature adds distinct visual states for "skipped" (both widgets) and "partial" (workout widget only) so the dashboard accurately reflects the user's progress through their day.

## 2. Goals

Primary goals:
- Skipped habits and workouts are visually distinct from pending items on the dashboard
- Partially completed workouts are visually distinct from pending items on the dashboard
- Skipped/partial items read as "resolved" (no action needed), not "outstanding"
- Visual treatment is consistent between the two widgets and fits the compact dashboard context

Non-goals:
- No skip/unskip actions from the dashboard (widgets remain read-only)
- No changes to the full habit or workout pages
- No progress bars or summary counts on the widgets
- No backend or API changes

## 3. User Stories

- As a user, I want to see which habits I've skipped today on the dashboard so I can distinguish them from habits I still need to log
- As a user, I want to see which workout exercises I've skipped today on the dashboard so I can distinguish them from exercises I still need to do
- As a user, I want to see which workout exercises I've partially completed so I can distinguish them from ones not yet started

## 4. Functional Requirements

### 4.1 Habits Widget — Skipped State

- Detect skipped habits by checking for entries where `skipped === true` for a given tracking item
- Render skipped habits with:
  - A `SkipForward` icon (from lucide-react) in `text-muted-foreground` color
  - Strikethrough text in `text-muted-foreground` (matching the "done" strikethrough pattern)
- Skipped items appear as "resolved" — visually similar to done but distinguishable by icon

### 4.2 Workout Widget — Skipped State

- Detect skipped exercises by checking `item.performance?.status === "skipped"`
- Render skipped exercises with:
  - A `SkipForward` icon (from lucide-react) in `text-muted-foreground` color
  - Strikethrough text in `text-muted-foreground`
- Consistent with the habits widget skipped treatment

### 4.3 Workout Widget — Partial State

- Detect partially completed exercises by checking `item.performance?.status === "partial"`
- Render partial exercises with:
  - A `CircleDot` icon (from lucide-react) in `text-yellow-600` color (or similar amber/yellow to convey "in progress")
  - Normal text (no strikethrough — the item is started but not resolved)
- Visually distinct from pending (hollow circle), done (green check), and skipped (skip icon)

### 4.4 State Priority Summary

**Habits widget states (3 total):**

| State | Condition | Icon | Icon Color | Text Style |
|-------|-----------|------|------------|------------|
| Done | Entry exists, `skipped=false`, `value` is set | `Check` | `text-green-600` | Strikethrough, muted |
| Skipped | Entry exists, `skipped=true` | `SkipForward` | `text-muted-foreground` | Strikethrough, muted |
| Pending | No entry for this item | `Circle` | `text-muted-foreground` | Normal |

**Workout widget states (4 total):**

| State | Condition | Icon | Icon Color | Text Style |
|-------|-----------|------|------------|------------|
| Done | `performance.status === "done"` | `Check` | `text-green-600` | Strikethrough, muted |
| Skipped | `performance.status === "skipped"` | `SkipForward` | `text-muted-foreground` | Strikethrough, muted |
| Partial | `performance.status === "partial"` | `CircleDot` | `text-yellow-600` | Normal |
| Pending | `performance.status === "pending"` or undefined | `Circle` | `text-muted-foreground` | Normal |

## 5. API Surface

No new or modified endpoints. Existing endpoints already return all required data:

| Endpoint | Relevant Field | Values |
|----------|---------------|--------|
| `GET /api/v1/trackers/today` | `entries[].attributes.skipped` | `true` / `false` |
| `GET /api/v1/workouts/today` | `items[].performance.status` | `"pending"` / `"done"` / `"skipped"` / `"partial"` |

## 6. Data Model

No data model changes. All required fields already exist.

## 7. Service Impact

| Service | Change |
|---------|--------|
| **Frontend** | Update `habits-widget.tsx`: build a `skippedItemIds` set alongside `completedItemIds`, render skipped items with `SkipForward` icon and strikethrough |
| **Frontend** | Update `workout-widget.tsx`: expand status check from binary (done vs not) to handle all four statuses with distinct icon/color/text treatment |

## 8. Non-Functional Requirements

- **Performance**: No impact — no additional API calls; just different rendering logic on data already fetched
- **Accessibility**: Icon-only status indicators should be supplemented with `aria-label` attributes (e.g., "skipped", "partially completed") for screen readers
- **Consistency**: The `SkipForward` icon and muted strikethrough pattern should be identical between both widgets

## 9. Open Questions

None — all questions resolved during scoping.

## 10. Acceptance Criteria

- [ ] Habits widget renders skipped habits with a `SkipForward` icon and strikethrough muted text
- [ ] Habits widget renders completed habits with a green `Check` icon and strikethrough muted text (existing, unchanged)
- [ ] Habits widget renders pending habits with a hollow `Circle` icon and normal text (existing, unchanged)
- [ ] Workout widget renders skipped exercises with a `SkipForward` icon and strikethrough muted text
- [ ] Workout widget renders partially completed exercises with a `CircleDot` icon in yellow and normal text
- [ ] Workout widget renders completed exercises with a green `Check` icon and strikethrough muted text (existing, unchanged)
- [ ] Workout widget renders pending exercises with a hollow `Circle` icon and normal text (existing, unchanged)
- [ ] Skipped items visually read as "resolved" — clearly not pending
- [ ] No changes to backend services or API contracts
- [ ] Existing widget behavior (loading, error, empty states, linking) is unaffected
