# Task 038: Key Context

Last Updated: 2026-04-12

---

## Key Files

### To Create
| File | Purpose |
|------|---------|
| `frontend/src/lib/date-utils.ts` | Shared timezone-aware date utility (new) |
| `frontend/src/lib/__tests__/date-utils.test.ts` | Unit tests for date utility (new) |
| `services/calendar-service/internal/tz/resolver.go` | Timezone resolver for calendar-service (new, copied from pattern) |

### To Modify — Frontend
| File | Lines | Change |
|------|-------|--------|
| `frontend/src/components/features/calendar/calendar-widget.tsx` | 9-14, 30 | Replace `getTodayRange()` with shared `getLocalTodayRange()` |
| `frontend/src/components/features/calendar/calendar-utils.ts` | 187-217 | Wire up `_timezone` param in `getEventsForDay()` |
| `frontend/src/components/features/meals/meal-plan-widget.tsx` | 16-34, 37-38 | Replace `getMonday()`/`getTodayStr()` with shared utils |
| `frontend/src/pages/MealsPage.tsx` | 32-47 | Replace `getMonday()`/`formatDateStr()` with shared utils |
| `frontend/src/pages/TrackerPage.tsx` | 11-14, 18 | Replace `getCurrentMonth()` with `getLocalMonth()` |
| `frontend/src/components/features/tracker/calendar-grid.tsx` | 91, 95, 209 | Replace naive today/month computations |
| `frontend/src/types/models/task.ts` | 45-59 | Simplify `isTaskOverdue()` to string comparison |

### To Modify — Backend
| File | Lines | Change |
|------|-------|--------|
| `services/calendar-service/internal/event/resource.go` | 43, 57 | Use tz resolver for default start/end computation |

### Reference (Do Not Modify)
| File | Purpose |
|------|---------|
| `services/workout-service/internal/tz/resolver.go` | Reference implementation for tz resolver pattern |
| `services/tracker-service/internal/tz/resolver.go` | Second copy of tz resolver pattern |
| `frontend/src/lib/api/client.ts` (line 142-143) | X-Timezone header already sent on all requests |
| `frontend/src/components/features/calendar/calendar-utils.ts` (lines 12-52) | Existing `getTimeInZone()`, `getDateInZone()` helpers |

---

## Key Decisions

1. **No external date library** — `Intl.DateTimeFormat` with `timeZone` option is sufficient and avoids adding bundle size.
2. **Browser timezone as fallback** — When no explicit timezone is provided, `Intl.DateTimeFormat().resolvedOptions().timeZone` is used. This is acceptable because users are expected to be in the same timezone as their household.
3. **Copy tz resolver pattern** — The calendar-service gets its own copy of the timezone resolver (same as workout-service and tracker-service) rather than extracting to a shared package. Extraction can be a future improvement.
4. **String comparison for overdue tasks** — `isTaskOverdue()` switches from Date object comparison to `YYYY-MM-DD` string comparison, which is simpler and avoids timezone ambiguity.

---

## Dependencies

- **Task 035 (completed)**: Added `X-Timezone` header to frontend API client and timezone resolution to workout-service and tracker-service. This task builds on that foundation.
- **account-service**: Provides household timezone field (already exists, no changes needed).
- **Frontend API client**: Already sends `X-Timezone` header on every request (no changes needed).

---

## Patterns to Follow

### Frontend Date Utility Pattern
```typescript
// Use Intl.DateTimeFormat to extract date parts in the target timezone
function getDateParts(tz?: string): { year: number; month: number; day: number } {
  const formatter = new Intl.DateTimeFormat("en-CA", {
    timeZone: tz || Intl.DateTimeFormat().resolvedOptions().timeZone,
    year: "numeric", month: "2-digit", day: "2-digit",
  });
  // Parse formatted parts to extract year/month/day
}
```

### Backend Timezone Resolver Pattern
```go
// Resolution order:
// 1. X-Timezone header (parsed via time.LoadLocation)
// 2. Household timezone via account-service lookup
// 3. UTC fallback with warn-level log
func Resolve(ctx context.Context, l logrus.FieldLogger, headers http.Header, householdID uuid.UUID, lookup HouseholdLookup) *time.Location
```
