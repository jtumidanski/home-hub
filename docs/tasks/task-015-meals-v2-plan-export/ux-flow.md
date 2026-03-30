# Meals v2 — UX Flow

## Admin Planner Screen

Located in the existing hidden admin app alongside recipe management.

### Screen Layout

```
┌─────────────────────────────────────────────────────────┐
│ Meal Planner                                            │
│                                                         │
│ ┌─────────────────────┐  ┌────────────────────────────┐ │
│ │ Week: ◄ Apr 6, 2026 ►│  │ 🔓 Locked │ Save │ Export │ │
│ └─────────────────────┘  └────────────────────────────┘ │
│                                                         │
│ ┌───────────────────────────────────────────────────┐   │
│ │                  WEEKLY GRID                       │   │
│ │                                                    │   │
│ │  Monday    │ Breakfast │ Lunch │ Dinner │ Snack    │   │
│ │  Apr 6     │           │       │ Tacos  │          │   │
│ │            │           │       │ (Side) │          │   │
│ │  Tuesday   │           │       │ Pasta  │          │   │
│ │  Apr 7     │           │       │        │          │   │
│ │  ...       │           │       │        │          │   │
│ └───────────────────────────────────────────────────┘   │
│                                                         │
│ ┌────────────────────────┐ ┌──────────────────────────┐ │
│ │   RECIPE SELECTOR      │ │  INGREDIENT PREVIEW      │ │
│ │                        │ │                           │ │
│ │  🔍 Search recipes...  │ │  2 lb chicken breast     │ │
│ │  Filter: [All ▾]       │ │  3 yellow onion          │ │
│ │                        │ │  4 tbsp olive oil        │ │
│ │  □ Chicken Tacos       │ │  2 jar tomato sauce      │ │
│ │    dinner · serves 4   │ │                           │ │
│ │  □ Pasta Carbonara     │ │  ⚠ 1 unresolved          │ │
│ │    dinner · serves 4   │ │                           │ │
│ │  □ Overnight Oats      │ │                           │ │
│ │    breakfast · serves 2 │ │                           │ │
│ └────────────────────────┘ └──────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### User Flow

```
1. Navigate to Meal Planner
         │
         ▼
2. Select week (← ► arrows, step by 7 days; start day configurable in frontend)
         │
    ┌────┴────┐
    │ Existing │──→ Load plan + items
    │  plan?   │
    └────┬────┘
         │ No
         ▼
3. Empty grid shown, plan created on first save
         │
         ▼
4. Search/browse recipes in selector panel
   - Search by name (debounced)
   - Filter by classification (auto-applied when clicking a slot cell)
   - Shows: name, classification, servings_yield, last used date, usage count
         │
         ▼
5. Add recipe to grid
   - Click recipe → opens day/slot picker
   - Or drag recipe to grid cell (stretch goal)
   - Set serving multiplier or planned servings
   - Add optional notes
         │
         ▼
6. Review ingredient preview (auto-updates)
   - Shows consolidated list
   - Flags unresolved ingredients with warning
         │
         ▼
7. Save plan
   - Creates plan_week (if new) and plan_items
   - Shows success toast
         │
         ▼
8. Lock plan (optional)
   - Prevents further edits
   - Lock icon toggles locked state
         │
         ▼
9. Export markdown
   - Opens preview modal
   - Copy to clipboard button
   - Download as .md file button
```

### Interactions

| Action | Behavior |
|--------|----------|
| Click recipe in selector | Opens popover: pick day, slot, servings, notes → adds to grid |
| Click item in grid | Opens edit popover: change day, slot, servings, notes |
| Click × on grid item | Removes item (with confirmation if plan has been saved) |
| Week arrows | Navigate to previous/next week (7-day step); loads existing plan or shows empty grid |
| Save button | Persists plan and all items; disabled when no changes |
| Lock toggle | Locks/unlocks plan; when locked, grid becomes read-only |
| Duplicate button | Opens date picker to choose target week, then copies plan |
| Export button | Generates markdown preview in modal |

### Validation Warnings

The UI shall display inline warnings for:

| Condition | Warning |
|-----------|---------|
| Recipe has been deleted | "⚠ This recipe has been deleted" badge on the item |
| Recipe has unresolved ingredients | "⚠ Some ingredients may not consolidate correctly" on the item |
| Editing a locked plan | Grid is read-only; "Unlock to edit" message shown |
| Plan already exists for week | Automatically loads existing plan instead of creating new |
| Duplicate target week has existing plan | "A plan already exists for that week" error on duplicate |

### Export Modal

```
┌─────────────────────────────────────────┐
│ Export Meal Plan                    [×]  │
│                                         │
│ ┌─────────────────────────────────────┐ │
│ │ # Meal Plan: Week of April 6, 2026 │ │
│ │                                     │ │
│ │ ## Monday (2026-04-06)              │ │
│ │ - **Dinner:** Chicken Tacos         │ │
│ │                                     │ │
│ │ ## Tuesday (2026-04-07)             │ │
│ │ - **Dinner:** Pasta Carbonara       │ │
│ │                                     │ │
│ │ ## Consolidated Ingredients         │ │
│ │ - 2 lb chicken breast              │ │
│ │ - 3 yellow onion                   │ │
│ │ - 4 tbsp olive oil                 │ │
│ └─────────────────────────────────────┘ │
│                                         │
│          [Copy to Clipboard] [Download] │
└─────────────────────────────────────────┘
```

The preview shows rendered markdown. Copy and download provide the raw markdown text.
