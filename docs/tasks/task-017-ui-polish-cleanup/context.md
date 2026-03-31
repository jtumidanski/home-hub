# Task 017 — UI Polish & Cleanup: Context

Last Updated: 2026-03-31

## Key Files

### Backend (Item 2 — Time Parser Fix)

| File | Purpose |
|------|---------|
| `services/recipe-service/internal/recipe/cooklang/parser.go` | `ParseMinutes()` at line 540 — naive regex grabs first digit only |
| `services/recipe-service/internal/recipe/cooklang/parser_test.go` | `TestParseMinutes` at line 320 — only covers simple formats |
| `services/recipe-service/internal/recipe/processor.go` | Lines 86-93, 188-193 — calls `ParseMinutes` for cooklang metadata |

### Frontend — Pages

| File | Items |
|------|-------|
| `frontend/src/components/features/meals/week-grid.tsx` | Item 1 — meal type badge at lines 113-116 |
| `frontend/src/components/features/meals/week-selector.tsx` | Item 4 — week selector, full file (54 lines) |
| `frontend/src/pages/CalendarPage.tsx` | Item 5 — "Today" button at lines 241-251 |
| `frontend/src/pages/IngredientsPage.tsx` | Item 3 — no pagination, full file (174 lines) |
| `frontend/src/pages/RecipesPage.tsx` | Item 7 — tag filter badges |
| `frontend/src/pages/RecipeDetailPage.tsx` | Item 7 — tag badges at lines 140-146 |
| `frontend/src/components/features/recipes/recipe-card.tsx` | Item 7 — tag badges at lines 92-96 |
| `frontend/src/pages/DashboardPage.tsx` | Item 8 — "You are the owner" at lines 52-54 |
| `frontend/src/pages/HouseholdMembersPage.tsx` | Item 9 — header at lines 119-127 |
| `frontend/src/components/common/list-filter-bar.tsx` | Item 6 — status select at lines 62-74 |

### Frontend — Hooks & Services

| File | Purpose |
|------|---------|
| `frontend/src/lib/hooks/api/use-ingredients.ts` | Already supports `page`/`pageSize` params |
| `frontend/src/lib/constants/recipe.ts` | `filterClassificationTags()` helper |

### Reference Pages (for UX alignment — Item 9)

| File | Pattern |
|------|---------|
| `frontend/src/pages/IngredientDetailPage.tsx` | Lines 125-128: `<Button variant="ghost" size="sm">` with arrow + text |
| `frontend/src/pages/RecipeDetailPage.tsx` | Lines 74-77: Same pattern |

## Key Decisions

1. **Parser fix only, no reprocessing** — user will resave recipes to pick up corrected times
2. **Backend pagination already exists** — `GET /ingredients` supports `page[number]`, `page[size]`, returns `meta.total`
3. **shadcn Calendar + Popover not yet installed** — must add via `npx shadcn@latest add calendar popover` (also adds `react-day-picker` and `date-fns` deps)
4. **Title-case is display-only** — stored tag values remain lowercase
5. **Status filter fix** — closed state should show "All statuses" label, not raw "all" value

## Dependencies

- `react-day-picker` — required by shadcn Calendar component (new dependency)
- `date-fns` — required by shadcn Calendar component (new dependency, or may already be present)
- No backend migrations needed
- No new API endpoints needed
