# Daily Tracker — UX Flow

## 1. Sidebar Entry

New sidebar item: **"Tracker"** under the personal section (not household-scoped).

## 2. Main Views

### 2.1 Tracker Setup (Item Management)

Accessible via a settings/gear icon on the tracker page.

```
+------------------------------------------+
|  My Tracking Items              [+ Add]  |
+------------------------------------------+
|  1. ● Food          Sentiment   Every day  |
|  2. ● Drinks        Numeric     Every day  |
|  3. ● Sleep Quality Range 0-100 Every day  |
|  4. ● Running       Sentiment   MTThFS     |
|  5. ● Lifting       Sentiment   MTThF      |
|  6. ● Volleyball    Sentiment   W          |
+------------------------------------------+
```

- "Add" opens a creation form: name, color (palette picker), scale type (dropdown), scale config (conditional on type), day-of-week toggles, sort order
- Each row has edit/delete actions
- Reorder via sort order field or drag-and-drop

### 2.2 Today Quick-Entry

Accessible via a "Log Today" button on the tracker landing page. Designed for fast daily logging.

```
+------------------------------------------+
|  Today — Thursday, Apr 2       [Calendar] |
+------------------------------------------+
|                                           |
|  Food                          😊 😐 😞  |
|  [note: ________________________]         |
|                                           |
|  Drinks                        [-] 0 [+]  |
|  [note: ________________________]         |
|                                           |
|  Sleep Quality          ◀━━━●━━━━━▶  72   |
|  [note: ________________________]         |
|                                           |
|  Running                       😊 😐 😞  |
|  [note: ________________________]         |
|                                           |
|  Lifting                       😊 😐 😞  |
|  [note: ________________________]         |
|                                           |
|  ✓ 2/5 logged today                      |
+------------------------------------------+
```

- Only shows items scheduled for today (Thursday = Food, Drinks, Sleep, Running, Lifting — not Volleyball)
- Already-logged items show their current value, editable inline
- Each item has a collapsible note field
- Progress counter at bottom
- "[Calendar]" navigates to the full monthly grid

### 2.3 Monthly Calendar Grid (Incomplete Month)

Default landing view. Shows current month by default.

```
         April 2026
         1    2    3    4    5    ...  30
         Tue  Wed  Thu  Fri  Sat
+--------+----+----+----+----+----+---+----+
| Food   | 😊 | 😐 |    | 😊 | 😞 |   |    |
| Drinks |  2 |  0 |    |  1 |  3 |   |    |
| Sleep  | 72 | 68 |    | 80 | 55 |   |    |
| Running| 😊 |  · | 😊 | 😊 |    |   |    |
| Lifting| 😊 |  · |    | 😊 |  · |   |    |
| Volley |  · | 😊 |  · |  · |  · |   |    |
+--------+----+----+----+----+----+---+----+

Legend:  · = not scheduled   [empty] = scheduled, not yet filled
        ⊘ = skipped

Progress: 23/52 entries  ████░░░░░░ 44%

◀ March 2026          April 2026          May 2026 ▶
```

- Dimmed cells (`·`) = not on schedule for that day
- Empty highlighted cells = scheduled but unfilled (action needed)
- Clicking any editable cell opens entry editor
- Month navigation arrows; cannot navigate to future months
- Progress bar shows completion status

### 2.4 Cell Entry Editor

Inline popover or small modal when tapping a cell.

**Sentiment:**
```
+-------------------+
| Running — Apr 1   |
|                   |
|  😊    😐    😞   |
|                   |
| [Skip] [Clear]    |
+-------------------+
```

**Numeric:**
```
+-------------------+
| Drinks — Apr 1    |
|                   |
|  [-]    3    [+]  |
|                   |
| [Skip] [Clear]    |
+-------------------+
```

**Range:**
```
+-------------------+
| Sleep — Apr 1     |
|                   |
|  ◀━━━━━●━━━━━▶   |
|       72          |
|                   |
| [Skip] [Clear]    |
+-------------------+
```

- "Skip" only available on scheduled days
- "Clear" removes the entry entirely
- Optional note field (text input, collapsed by default, expands on tap)

### 2.5 Monthly Dashboard (Complete Month)

Replaces the calendar grid when all entries are filled/skipped.

```
+================================================+
|  March 2026 — Monthly Report        [Calendar] |
+================================================+
|                                                 |
|  ┌─────────────────┐  ┌──────────────────────┐ |
|  │ Completion       │  │ Skip Rate            │ |
|  │ 98% (49/50)      │  │ 4% (2/50)            │ |
|  └─────────────────┘  └──────────────────────┘ |
|                                                 |
|  ── Sentiment Items ──────────────────────────  |
|                                                 |
|  Running           8/10 positive (80%)          |
|  ██████████░░░░  😊 8  😐 1  😞 1              |
|                                                 |
|  Lifting           7/9 positive (78%)           |
|  █████████░░░░░  😊 7  😐 1  😞 1              |
|                                                 |
|  Food              22/31 positive (71%)         |
|  ████████░░░░░░  😊 22  😐 5  😞 4             |
|                                                 |
|  Volleyball        4/4 positive (100%)          |
|  ██████████████  😊 4  😐 0  😞 0              |
|                                                 |
|  ── Numeric Items ────────────────────────────  |
|                                                 |
|  Drinks                                         |
|  Total: 23  |  Avg: 0.7/day  |  50% dry days   |
|  ▁▃▁▁▅▁▁▃▁▁▁▇▁▁▃▁▁▁▁▅▁▁▃▁▁▁▁▁▁▁              |
|  Worst: Apr 12 (4)  Best: 16 dry days           |
|                                                 |
|  ── Range Items ──────────────────────────────  |
|                                                 |
|  Sleep Quality                                  |
|  Avg: 72  |  Min: 35 (Apr 8)  |  Max: 95       |
|  ━━━╲━━━━╱━━━━━━━━━╲━━╱━━━━━━━━━━              |
|  Std Dev: 14.2                                  |
|                                                 |
+================================================+
```

- "[Calendar]" button toggles back to the grid view for editing
- Grouped by scale type for visual coherence
- Sparklines/mini charts rendered client-side from `daily_values` in the report API

## 3. Navigation Flow

```
Sidebar "Tracker"
    │
    ├─→ Today Quick-Entry ([Log Today] button)
    │       │
    │       └─→ [Calendar] → Monthly Calendar Grid
    │
    ├─→ Monthly Calendar Grid (default: current month)
    │       │
    │       ├─→ Cell tap → Entry Editor (inline)
    │       ├─→ Month nav → Navigate to other months
    │       └─→ Auto-switch to Dashboard when complete
    │
    ├─→ Monthly Dashboard (auto for complete months)
    │       │
    │       └─→ [Calendar] → Back to grid view
    │
    └─→ Settings (gear icon)
            │
            └─→ Tracking Item Setup (CRUD)
```

## 4. Month Switcher Behavior

- Incomplete months: show calendar grid
- Complete months: show dashboard (with toggle to grid)
- Current month: calendar grid by default; switches to dashboard if all remaining scheduled days are in the past and all entries are filled/skipped
- Future months: not navigable
