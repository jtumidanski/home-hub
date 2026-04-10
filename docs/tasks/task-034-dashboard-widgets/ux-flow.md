# Dashboard Widgets — UX Flow

## Dashboard Layout (Top to Bottom)

```
┌──────────────────────────────────────────────────┐
│  Dashboard                                        │
├──────────────────────────────────────────────────┤
│  [Weather Widget]                    (full width) │
├──────────────────────────────────────────────────┤
│  [Package Summary Widget]            (full width) │
│   - 3 cards or empty state                        │
├──────────────────────────────────────────────────┤
│  [Tasks / Reminders Grid]    (3-col on desktop)   │
├────────────────────────┬─────────────────────────┤
│  Meal Plan Today       │  Calendar Today          │
│  ──────────────────    │  ──────────────────      │
│  Breakfast: Recipe A → │  ● All Day: Event X      │
│  Lunch: Recipe B →     │  ● 9:00 AM: Meeting (J)  │
│  Dinner: Recipe C →    │  ● 2:00 PM: Call (S)     │
│                        │                          │
│  → View Meals          │  → View Calendar         │
├────────────────────────┼─────────────────────────┤
│  Habits Today          │  Workout Today           │
│  ──────────────────    │  ──────────────────      │
│  ✓ Meditation          │  Bench Press             │
│  ✗ Read 30 min         │  Squats                  │
│  ✓ Journal             │  Pull-ups                │
│                        │                          │
│  → View Habits         │  → View Workouts         │
└────────────────────────┴─────────────────────────┘
```

## Desktop Layout Details

- Weather widget: full width (existing)
- Package summary: full width, 3-col grid (existing, updated with empty state)
- Tasks/reminders: full width, 3-col grid (existing)
- New widgets: 2-column grid below existing content
  - Left column: Meal Plan, Habits (list-style content)
  - Right column: Calendar, Workout (list-style content)

## Mobile Layout

All widgets stack vertically in the same order:
1. Weather
2. Package Summary
3. Tasks / Reminders
4. Meal Plan Today
5. Calendar Today
6. Habits Today
7. Workout Today

## Widget Card Pattern

Each new widget follows a consistent card structure:

```
┌────────────────────────────────┐
│  Widget Title              Icon │  ← CardHeader, clickable link to full page
│  ────────────────────────────  │
│  Content rows                  │  ← CardContent
│  ...                           │
│                                │
│  View [Page] →                 │  ← Footer link to full page
└────────────────────────────────┘
```

## Empty States

Each widget shows a centered empty state within the card:

```
┌────────────────────────────────┐
│  Widget Title              Icon │
│  ────────────────────────────  │
│                                │
│       [muted icon]             │
│    No [items] for today        │
│                                │
└────────────────────────────────┘
```

## Meal Plan Widget Detail

- Slot labels shown in title case: "Breakfast", "Lunch", "Dinner", "Snack", "Side"
- Each item shows: `Slot: Recipe Name` as a clickable link
- Multiple items in the same slot each get their own row
- Links navigate to `/app/recipes/{recipeId}`

## Calendar Widget Detail

- All-day events listed first with "All Day" label
- Timed events sorted by start time, displayed as `HH:MM AM/PM`
- Each event shows a colored dot (using `userColor`) or user initial + event title
- Truncate long event titles with ellipsis

## Habits Widget Detail

- Each row: completion indicator (checkmark or empty circle) + habit name
- Completed habits use a checkmark icon with muted/success styling
- Incomplete habits use an empty circle icon

## Workout Widget Detail

- Rest day: centered "Rest Day" message with a relaxation icon
- Active day: list of exercise names (no sets/reps detail on dashboard)
- Exercises with logged performance show a subtle "done" indicator (e.g., checkmark or muted styling similar to habits)
