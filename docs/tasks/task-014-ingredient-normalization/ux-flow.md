# Ingredient Normalization — UX Flow

## Recipe Create/Edit Flow

```
User writes Cooklang in editor
        │
        ▼
Live preview panel (existing)          Planner Settings (new, collapsible)
  ├─ Ingredients list                    ├─ Classification dropdown
  ├─ Steps rendering                     ├─ Servings yield override
  ├─ Syntax errors                       ├─ Eat-within days
  └─ Normalization preview (new)         ├─ Min gap days
     ├─ Per-ingredient match status      └─ Max consecutive days
     └─ ✓/⚠ indicators before save
        │
        ▼  (on save)
Server validates Cooklang + normalizes ingredients
        │
        ▼
Recipe detail page
  ├─ Existing: title, metadata, steps
  ├─ New: Ingredient Normalization Panel (with Re-normalize button)
  └─ New: Planner Readiness Badge
```

## Ingredient Normalization Panel

Shown on recipe detail page and recipe edit page, below the ingredients list.

### States

**All Resolved:**
```
┌─────────────────────────────────────────────────┐
│ Ingredients                          8/8 ✓      │
├─────────────────────────────────────────────────┤
│ ✓  400g spaghetti          → Spaghetti          │
│ ✓  200g guanciale          → Guanciale          │
│ ✓  100g pecorino romano    → Pecorino Romano    │
│ ✓  4 egg yolks             → Egg Yolk           │
│ ✓  black pepper            → Black Pepper       │
│ ✓  2l water                → Water              │
│ ✓  1 tbsp salt             → Salt               │
│ ✓  1 tbsp olive oil        → Olive Oil          │
└─────────────────────────────────────────────────┘
```

**Some Unresolved:**
```
┌─────────────────────────────────────────────────┐
│ Ingredients              5/8 resolved  [Re-normalize]
├─────────────────────────────────────────────────┤
│ ✓  400g spaghetti          → Spaghetti          │
│ ✓  200g guanciale          → Guanciale          │
│ ⚠  100g pecorino romano    [Resolve ▾]          │
│ ✓  4 egg yolks             → Egg Yolk           │
│ ⚠  black pepper            [Resolve ▾]          │
│ ✓  2l water                → Water              │
│ ⚠  1 tbsp kosher salt      [Resolve ▾]          │
│ ✓  1 tbsp olive oil        → Olive Oil          │
└─────────────────────────────────────────────────┘
```

The "Re-normalize" button re-runs the normalization pipeline for all unresolved ingredients against the current canonical registry. Useful after adding new canonical ingredients or aliases. `manually_confirmed` statuses are preserved.

### Resolve Dropdown

When user clicks "Resolve" on an unresolved ingredient:

```
┌─────────────────────────────────────────────────┐
│ ⚠  pecorino romano         [Resolve ▾]          │
│  ┌───────────────────────────────────────────┐  │
│  │ 🔍 Search ingredients...                  │  │
│  │                                           │  │
│  │ Suggested:                                │  │
│  │   Pecorino                                │  │
│  │   Parmesan                                │  │
│  │   Romano Cheese                           │  │
│  │                                           │  │
│  │ ─────────────────────────────────────     │  │
│  │ + Create "pecorino romano" as new         │  │
│  │                                           │  │
│  │ ☑ Save as alias for future recipes        │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

**Interaction flow:**
1. Dropdown opens with search field and suggestions
2. Suggestions are based on text similarity to the raw ingredient name
3. User can type to filter the canonical ingredient list
4. Selecting a canonical ingredient immediately submits the correction
5. "Save as alias" checkbox (default checked) creates an alias mapping
6. "Create new" option opens an inline form for name + unit family, then assigns it

### Inline Create

```
┌───────────────────────────────────────────┐
│ Create New Ingredient                     │
│                                           │
│ Name: [pecorino romano          ]         │
│ Display Name: [Pecorino Romano  ]         │
│ Unit Family: [Weight ▾]                   │
│                                           │
│           [Cancel]  [Create & Assign]     │
└───────────────────────────────────────────┘
```

## Live Preview Normalization

The existing Cooklang live preview (in the recipe form editor) is extended to show normalization status per ingredient. As the user types, the debounced parse request now returns normalization data alongside the parsed ingredients.

```
┌─────────────────────────────────────────────────┐
│ Preview                                         │
├─────────────────────────────────────────────────┤
│ Ingredients                                     │
│                                                 │
│ ✓  400g spaghetti                               │
│ ✓  200g guanciale                               │
│ ⚠  100g pecorino romano                         │
│ ✓  4 egg yolks                                  │
│                                                 │
│ Steps                                           │
│ 1. Boil spaghetti in salted water...            │
│ 2. ...                                          │
└─────────────────────────────────────────────────┘
```

Status indicators are informational only — resolution actions are not available in preview (only after save). This lets the user know before saving which ingredients will need manual attention.

## Recipe List Filters

The recipe list page gains new filter controls alongside existing search and tag filters:

```
┌─────────────────────────────────────────────────┐
│ Recipes                                         │
│ 🔍 [Search...        ]                          │
│                                                 │
│ Tags: [italian ✕] [+]                           │
│ Classification: [All ▾]  Planner: [All ▾]       │
│ Ingredients: [All ▾]                            │
├─────────────────────────────────────────────────┤
│ ...recipe cards...                              │
└─────────────────────────────────────────────────┘
```

**Filter options:**
- **Classification**: All, Breakfast, Lunch, Dinner, Snack, Side
- **Planner**: All, Ready, Not Ready
- **Ingredients**: All, Fully Resolved, Has Unresolved

Filters combine with AND semantics. All filters are query-param backed for URL shareability.

## Planner Readiness Badge

**On recipe cards (list page):**
```
┌─────────────────────────────┐
│ Pasta Carbonara             │
│ Classic Roman pasta dish    │
│ 🍽 dinner  ⏱ 30min         │
│ italian, pasta              │
│ 8/8 ingredients  [Ready ✓] │   ← green badges
└─────────────────────────────┘

┌─────────────────────────────┐
│ Grilled Salmon              │
│ Simple weeknight dinner     │
│ ⏱ 25min                    │
│ seafood                     │
│ 3/5 ingredients [Not Ready] │   ← muted/warning
└─────────────────────────────┘
```

**On recipe detail page:**
```
┌─────────────────────────────────────────────────┐
│ Planner Status: Ready ✓                         │
│ Classification: Dinner                          │
│ Eat within: 3 days                              │
│ Min gap: 7 days │ Max consecutive: 1 day        │
│ Servings yield: 4                               │
└─────────────────────────────────────────────────┘
```

**When not ready:**
```
┌─────────────────────────────────────────────────┐
│ Planner Status: Not Ready                       │
│ ⚠ Missing planner configuration                │
│ ⚠ Classification is required                   │
│                                                 │
│ [Configure Planner Settings]                    │
└─────────────────────────────────────────────────┘
```

## Planner Settings (Recipe Form)

Collapsible section below the existing recipe form fields, above the Cooklang editor:

```
┌─────────────────────────────────────────────────┐
│ Planner Settings                        [▾]     │
├─────────────────────────────────────────────────┤
│ Classification:  [Dinner ▾]                     │
│                                                 │
│ Servings Yield:  [4      ]  (override recipe)   │
│ Eat Within:      [3      ]  days                │
│ Min Gap:         [7      ]  days between repeats│
│ Max Consecutive: [1      ]  days in a row       │
└─────────────────────────────────────────────────┘
```

## Canonical Ingredient Management

### Ingredient List Page (`/ingredients`)

```
┌─────────────────────────────────────────────────┐
│ Ingredients                                     │
│ 🔍 [Search ingredients...           ]           │
├─────────────────────────────────────────────────┤
│ Name              │ Unit    │ Aliases │ Used In  │
│───────────────────│─────────│─────────│──────────│
│ Chicken Breast    │ weight  │ 3       │ 12       │
│ Spaghetti         │ weight  │ 1       │ 8        │
│ Olive Oil         │ volume  │ 2       │ 15       │
│ Salt              │ weight  │ 4       │ 22       │
│ Garlic            │ count   │ 1       │ 18       │
│                                                 │
│ Page 1 of 5              [< Prev] [Next >]      │
└─────────────────────────────────────────────────┘
│                                                 │
│ [+ New Ingredient]                              │
└─────────────────────────────────────────────────┘
```

Clicking a row navigates to the ingredient detail page.

### Ingredient Detail Page (`/ingredients/:id`)

```
┌─────────────────────────────────────────────────┐
│ ← Back to Ingredients                           │
│                                                 │
│ Chicken Breast                         [Edit]   │
│ Unit Family: Weight                             │
│ Used in 12 recipes                              │
├─────────────────────────────────────────────────┤
│ Aliases                                         │
│                                                 │
│ chicken breasts                         [✕]     │
│ boneless chicken breast                 [✕]     │
│ skinless chicken breast                 [✕]     │
│                                                 │
│ [+ Add alias...                        ]        │
├─────────────────────────────────────────────────┤
│ Used In Recipes                                 │
│                                                 │
│ Pasta Carbonara                          →      │
│ Chicken Stir Fry                         →      │
│ Grilled Chicken Salad                    →      │
│ ...                                             │
│                                                 │
│ Page 1 of 2              [< Prev] [Next >]      │
└─────────────────────────────────────────────────┘
```

Clicking a recipe row navigates to that recipe's detail page. The list is paginated and fetched from `/ingredients/:id/recipes`.

### Delete Canonical Ingredient

When the user clicks delete on a canonical ingredient that has recipe references:

```
┌─────────────────────────────────────────────────┐
│ Delete "Chicken Breast"?                        │
│                                                 │
│ This ingredient is used by 12 recipes.          │
│ Choose how to handle existing references:       │
│                                                 │
│ ○ Reassign references to:                       │
│   [Search ingredient...              ▾]         │
│                                                 │
│ ○ Remove references (ingredients become         │
│   unresolved)                                   │
│                                                 │
│              [Cancel]  [Delete]                  │
└─────────────────────────────────────────────────┘
```

- "Reassign" calls `POST /ingredients/:id/reassign` with the target, then deletes
- "Remove references" calls `DELETE /ingredients/:id` after confirming (ON DELETE SET NULL handles the rest)

### Empty State

When no canonical ingredients exist yet:

```
┌─────────────────────────────────────────────────┐
│ Ingredients                                     │
├─────────────────────────────────────────────────┤
│                                                 │
│         No ingredients yet                      │
│                                                 │
│  Ingredients are added automatically when you   │
│  create recipes. You can also add them manually │
│  to build your registry ahead of time.          │
│                                                 │
│           [+ Add Ingredient]                    │
│                                                 │
└─────────────────────────────────────────────────┘
```

## Responsive Behavior

**Desktop (>= 1024px):**
- Normalization panel appears as a right sidebar section alongside the existing ingredient list
- Planner settings are inline in the recipe form

**Mobile (< 1024px):**
- Normalization panel stacks below the ingredients list
- Planner settings in a collapsible accordion
- Resolve dropdown is full-width
- Canonical ingredient management is a standard mobile list/detail pattern
