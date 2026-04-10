# Tracker Emoji Ratings — Context

Last Updated: 2026-04-09

---

## Key Files

| File | Purpose | Lines of Interest |
|------|---------|-------------------|
| `frontend/src/components/features/tracker/today-view.tsx` | Today's logging interface | 110-114 (ratings array), 120 (button render) |
| `frontend/src/components/features/tracker/calendar-grid.tsx` | Month calendar view | 311 (cell display ternary), 381 (editor button text) |
| `frontend/src/components/features/tracker/month-report.tsx` | Monthly report | 95-97 (stat prefix spans) |

## Decisions

1. **Emoji choice:** Face emoji (😊, 😐, 😞) — decided during scoping
2. **Labels removed:** Today view buttons show emoji only, no text labels
3. **All views updated:** Calendar cells and month report stats also use emoji
4. **No backend changes:** Display-only modification

## Dependencies

- None — purely frontend display changes
- No new packages required
- No API or data model changes

## Mapping Reference

```
positive → 😊  (was: + / "Good")
neutral  → 😐  (was: ~ / "OK")
negative → 😞  (was: - / "Bad")
```
