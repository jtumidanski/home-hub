# Tracker Mobile Visibility — Context

Last Updated: 2026-04-09

## Key Files

| File | Role | Changes Needed |
|------|------|----------------|
| `frontend/src/components/features/tracker/today-view.tsx` | Today view — daily tracker cards | Card state styling, Badge import, progress bar |
| `frontend/src/components/features/tracker/calendar-grid.tsx` | Calendar view — includes `MobileDayView` component | Card state styling in MobileDayView only |
| `frontend/src/components/ui/badge.tsx` | Badge UI component | None — use as-is |

## Key Decisions

1. **Left border approach over background tint for unset items** — A colored left border is more distinctive than a faint background tint and works well in both light/dark mode. It creates a clear "to-do" visual language.

2. **Badge component for logged/skipped** — Using the existing `Badge` component rather than plain text gives proper visual weight and consistent styling. The `secondary` variant works for "skipped", and a custom green-tinted class works for "logged".

3. **No shared file extraction** — The `colorBorderLeft` map is small (16 entries) and only used in two files. Duplicating it is simpler than creating a shared constants file for this scope.

4. **Desktop calendar grid unchanged** — The existing colored fills and dashed borders work well at the small cell scale. This task only targets mobile card-based views.

## Dependencies

- `Badge` component — already exists, uses `class-variance-authority` variants
- `Check` icon — available from `lucide-react` (already a project dependency)
- `colorDot` / `colorBg` maps — existing patterns to follow for the new `colorBorderLeft` map

## Tailwind Considerations

The `colorBorderLeft` map must use full static class names (e.g., `border-l-red-500`) rather than string interpolation (e.g., `` `border-l-${color}-500` ``), because Tailwind's purge step only detects full class strings in source code.
