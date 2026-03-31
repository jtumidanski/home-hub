# Recipe Cook Mode — UX Flow

## Entry Point

```
RecipeDetailPage
  Header: [Back] [Title] [Planner Status]
  Actions: [Cook Mode] [Edit] [Delete]    <-- new button
```

The "Cook Mode" button uses a distinct icon (e.g., `Maximize2` or `Monitor` from lucide) to visually differentiate it from edit/delete actions.

## Cook Mode Overlay

### All Steps View (Default)

```
+--------------------------------------------------+
| [All Steps | Single Step]         [eye] [X Close] |
+--------------------------------------------------+
|                                                   |
|  = Section Name (if present)                      |
|                                                   |
|  1  Preheat the oven to 350F(10 minutes).         |
|                                                   |
|  2  Dice the onion(1, medium) and saut in a       |
|     skillet(large) for 5 minutes.                  |
|                                                   |
|  3  Mix flour(2 cups) with sugar(1/2 cup)         |
|     and baking powder(1 tsp).                      |
|                                                   |
|  ...scrollable...                                 |
+--------------------------------------------------+
```

- Colored segments rendered inline (orange ingredients, green timers, etc.)
- Text scaled via `clamp()` to fill width
- Vertical scroll for overflow

### Single Step View

```
+--------------------------------------------------+
| [All Steps | Single Step]              [X Close]  |
+--------------------------------------------------+
|                                                   |
|  = Section Name (if applicable)                   |
|                                                   |
|                                                   |
|     Dice the onion(1, medium) and                 |
|     saut in a skillet for                          |
|     5 minutes.                                     |
|                                                   |
|                                                   |
+--------------------------------------------------+
| [< Prev]         2 / 12            [Next >]       |
+--------------------------------------------------+
```

- Single step fills the available space with maximum font size
- Navigation bar pinned to bottom
- Arrow keys (left/right) navigate between steps
- Step counter centered

## Interaction Summary

| Action | Result |
|--------|--------|
| Click "Cook Mode" button | Open overlay, acquire wake lock, show all steps |
| Toggle to "Single Step" | Switch to single-step view starting at step 1 |
| Toggle to "All Steps" | Switch back to scrollable all-steps view |
| Click Next / Right Arrow | Advance to next step (single-step mode) |
| Click Prev / Left Arrow | Go to previous step (single-step mode) |
| Swipe Left | Next step (single-step mode, touch devices) |
| Swipe Right | Previous step (single-step mode, touch devices) |
| Click X / Press Escape | Close overlay, release wake lock |
| Tab loses focus | Release wake lock |
| Tab regains focus | Re-acquire wake lock |
