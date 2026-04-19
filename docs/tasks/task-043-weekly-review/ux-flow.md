# Weekly Review — UX Flow

Plain-text sketches of the screen at the key states. The component is `WorkoutReviewPage`; it lives inside `WorkoutShell` so the top app tabs (Today / Week / Exercises / Taxonomy / Review) are always visible above this content.

---

## Populated week — desktop (md: and up)

```
┌────────────────────────────────────────────────────────────────────────────┐
│  « Prev     Week of 2026-04-13     Next »                                  │
│  ↞ Previous populated (2026-04-06)        Next populated ↠  (disabled)     │
├────────────────────────────────────────────────────────────────────────────┤
│  ┌────────────────────────────────────────────────────────────────────┐    │
│  │  Week of 2026-04-13                                                │    │
│  │  [ 12 Planned ]   [ 9 Performed ]   [ 2 Pending ]   [ 1 Skipped ]  │    │
│  └────────────────────────────────────────────────────────────────────┘    │
│                                                                            │
│  Per Day                                                                   │
│  ┌──────┬──────┬──────┬──────┬──────┬──────┬──────┐                        │
│  │ Mon  │ Tue  │ Wed  │ Thu  │ Fri  │ Sat  │ Sun  │                        │
│  │      │      │      │      │      │      │ Rest │                        │
│  │ 3 ex │ 3 ex │ — —  │ 3 ex │ 3 ex │ — —  │ —    │                        │
│  │      │      │      │      │      │      │      │                        │
│  │ ┌──┐ │ ┌──┐ │      │ ┌──┐ │ ┌──┐ │      │      │                        │
│  │ │it│ │ │it│ │      │ │it│ │ │it│ │      │      │                        │
│  │ └──┘ │ └──┘ │      │ └──┘ │ └──┘ │      │      │                        │
│  │ ┌──┐ │ ┌──┐ │      │ ┌──┐ │ ┌──┐ │      │      │                        │
│  │ │it│ │ │it│ │      │ │it│ │ │it│ │      │      │                        │
│  │ └──┘ │ └──┘ │      │ └──┘ │ └──┘ │      │      │                        │
│  └──────┴──────┴──────┴──────┴──────┴──────┴──────┘                        │
│                                                                            │
│  By Theme                                                                  │
│  Muscle     4 items · 9,280 lb                                             │
│  Cardio     1 item  · 3.1 mi                                               │
│                                                                            │
│  By Region                                                                 │
│  Chest      2 items · 2,700 lb                                             │
│  Back       1 item  · 1,800 lb                                             │
│  ...                                                                       │
└────────────────────────────────────────────────────────────────────────────┘
```

Per-item card (strength, summary mode, done):

```
┌──────────────────────────────────┐
│ Bench Press              [Done]  │
│ Planned: 3×10 @ 135 lb           │
│ Actual:  3×10 @ 140 lb  ✓        │
└──────────────────────────────────┘
```

Per-item card (strength, per-set mode, partial):

```
┌─────────────────────────────────────────────┐
│ Squat                         [Partial]     │
│ Planned: 3×8 @ 225 lb                       │
│ Actual:  set 1: 8 @ 225 · set 2: 8 @ 235 ·  │
│          set 3: 6 @ 245                     │
└─────────────────────────────────────────────┘
```

Per-item card (cardio, done):

```
┌────────────────────────────────────────┐
│ Easy Run                     [Done]    │
│ Planned: 30:00 · 3.0 mi                │
│ Actual:  28:45 · 3.1 mi  ✓             │
└────────────────────────────────────────┘
```

Per-item card (skipped):

```
┌────────────────────────────────────┐
│ Overhead Press          [Skipped]  │   ← name has strikethrough
│ Planned: 3×8 @ 95 lb               │   ← muted
│ Actual:  Skipped                   │
└────────────────────────────────────┘
```

Per-item card (pending):

```
┌─────────────────────────────────────┐
│ Dumbbell Row           [Pending]    │   ← italic, muted
│ Planned: 3×12 @ 45 lb               │
│ Actual:  —                          │
└─────────────────────────────────────┘
```

---

## Populated week — mobile

Single-column stack, same cards, same totals card pinned top:

```
« Prev     Week of 2026-04-13     Next »
↞ Prev populated          Next populated ↠

┌─────────────────────────────┐
│ Week of 2026-04-13          │
│ 12 Planned  9 Performed     │
│ 2 Pending   1 Skipped       │
└─────────────────────────────┘

Monday — 3 exercises
  [Bench Press     Done]
  [Incline Press   Done]
  [Cable Fly       Partial]

Tuesday — 3 exercises
  ...

Wednesday — Nothing scheduled

...

Sunday — Rest day
```

---

## Empty week state

```
┌────────────────────────────────────────────────────────────────┐
│  « Prev     Week of 2026-05-04     Next »                      │
│  ↞ Previous populated (2026-04-13)    Next populated ↠ (none)  │
├────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  No workouts logged for this week                        │  │
│  │                                                          │  │
│  │  Jump to the most recent populated week, or navigate     │  │
│  │  using Prev / Next above.                                │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────┘
```

When on an empty week, the summary endpoint returns `404`. The page then calls `GET /weeks/nearest?reference=...&direction=prev` and `...&direction=next` independently to determine whether the jump buttons are enabled, and to get their targets for the inline hint (e.g., `Previous populated (2026-04-13)`).

---

## Interaction flow

1. User clicks `Review` in the shell — lands on `/app/workouts/review` → internally resolves to `/app/workouts/review/{currentWeekStart}` (without redirect — the page just uses the fallback).
2. Page fetches `GET /summary` for the current week.
   - `200` → render populated layout, populate jump buttons from `previousPopulatedWeek` / `nextPopulatedWeek`.
   - `404` → render empty-week card, fetch both `nearest?prev` and `nearest?next` in parallel to populate jump buttons.
3. User clicks `« Prev` or `Next »` — navigate replaces URL with `addDays(weekStart, ±7)`. React Router handles the transition; hook re-fetches.
4. User clicks `↞ Previous populated` — navigate replaces URL with the target `weekStartDate`. Same hook path.
5. Browser back returns to the previously displayed week.

---

## Accessibility notes

- Status badges render both the label (`Done`, `Partial`, etc.) and a color; color is decorative, not load-bearing.
- `✓` has `aria-label="Target met"`.
- Strikethrough on skipped item name has a parallel `Skipped` status badge so screen readers are not dependent on visual styling.
- All buttons (prev / next / jump) are focusable and have explicit labels.
- The day grid uses `<section>` per day with an `h2` containing the day name for good document outline.
