# Task 018 — Recipe Cook Mode: Context

Last Updated: 2026-03-31

---

## Key Files

| File | Role |
|------|------|
| `frontend/src/pages/RecipeDetailPage.tsx` | Recipe detail page — add Cook Mode button here |
| `frontend/src/components/features/recipes/recipe-steps.tsx` | Current step renderer — extract segment logic from here |
| `frontend/src/types/models/recipe.ts` | TypeScript types for `Step`, `Segment`, `RecipeDetailAttributes` |
| `frontend/src/lib/hooks/api/use-recipes.ts` | Recipe data fetching hooks (no changes needed) |
| `frontend/src/components/features/recipes/cooklang-preview.tsx` | Cooklang preview (reference for segment rendering patterns) |

## New Files to Create

| File | Purpose |
|------|---------|
| `frontend/src/components/features/recipes/step-segment.tsx` | Shared segment renderer extracted from recipe-steps |
| `frontend/src/components/features/recipes/cook-mode.tsx` | Full-screen cook mode overlay component |
| `frontend/src/lib/hooks/use-wake-lock.ts` | Screen Wake Lock API hook |

## Key Decisions

1. **Shared segment renderer**: Extract from `recipe-steps.tsx` rather than duplicating. Both normal view and cook mode use the same `StepSegment` component with a size variant.

2. **Single component file**: Cook mode lives in one file (`cook-mode.tsx`) with internal sub-components for header, all-steps view, and single-step view. No need for a separate directory — it's a self-contained overlay.

3. **CSS `clamp()` for text scaling**: Use viewport-relative units via Tailwind arbitrary values (e.g., `text-[clamp(1.25rem,3.5vw,3rem)]`) rather than JavaScript-based resize calculations.

4. **Custom `useWakeLock` hook**: Encapsulates the Wake Lock API lifecycle (acquire, release, re-acquire on visibility change) in a reusable hook rather than inlining the logic.

5. **Swipe via touch events**: Use raw `touchstart`/`touchend` event tracking rather than adding a gesture library dependency.

## Dependencies

- No new npm packages
- Screen Wake Lock API: browser-native, Chromium + Safari 16.4+, graceful degradation on Firefox
- All recipe data already available from `RecipeDetailAttributes.steps` — no API calls

## Data Flow

```
RecipeDetailPage
  └── useRecipe(id) → data.data.attributes.steps: Step[]
       └── CookMode(steps, title, open, onClose)
            ├── All Steps View → StepSegment(segment, size="large")
            └── Single Step View → StepSegment(segment, size="large")
```

## Segment Color Conventions

| Type | Light | Dark |
|------|-------|------|
| Ingredient | `text-orange-600` | `text-orange-400` |
| Cookware | `text-blue-600` | `text-blue-400` |
| Timer | `text-green-600` | `text-green-400` |
| Reference | `text-purple-600` | `text-purple-400` |
