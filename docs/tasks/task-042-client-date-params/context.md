# Task 042: Key Context

Last Updated: 2026-04-16

---

## Root Cause Recap

The `internal/tz/accountclient.go` household fallback authenticates with `r.Header.Get("Authorization")`, but `shared/go/auth/auth.go:114` only accepts the `access_token` cookie. Outbound lookup is always 401. When `X-Timezone` also fails (observed on mobile), `tz.Resolve` drops to UTC, flipping "today" to "tomorrow" at UTC midnight.

The fix isn't to repair the fallback — it's to stop relying on server-side timezone resolution. The client already knows the right timezone; have it send the date.

---

## Key Files

### To Create

| File | Purpose |
|---|---|
| `shared/go/http/params.go` | `ParseDateParam` helper + tests |
| `shared/go/http/params_test.go` | Unit tests for the parser |
| `frontend/src/lib/hooks/use-local-date.ts` | Poll-based hook that emits re-renders when local date rolls over |
| `frontend/src/lib/hooks/__tests__/use-local-date.test.ts` | Tests for the hook with fake timers |

### To Modify — Backend

| File | Change | PR |
|---|---|---|
| `services/tracker-service/internal/today/resource.go` | Accept `?date=` (PR 1), require it + delete tz (PR 3) | 1, 3 |
| `services/tracker-service/internal/today/processor.go` | Signature cleanup (PR 3) | 3 |
| `services/tracker-service/internal/entry/processor.go:53,123` | Accept `today` param; thread via REST | 3 |
| `services/tracker-service/internal/entry/resource.go` | Parse `?today=` on PUT entry handler | 3 |
| `services/tracker-service/internal/month/processor.go:71` | Accept `today` param | 3 |
| `services/tracker-service/internal/month/resource.go` | Parse `?today=` on month handler | 3 |
| `services/tracker-service/cmd/main.go` | Drop `accountBaseURL` wiring for today routes | 3 |
| `services/workout-service/internal/today/resource.go` | Accept/require `?date=` | 1, 3 |
| `services/workout-service/internal/today/processor.go` | Signature cleanup | 3 |
| `services/workout-service/internal/planneditem/processor.go:251` | Thread date through if on request path | 3 |
| `services/workout-service/cmd/main.go` | Drop `accountBaseURL` wiring | 3 |
| `services/productivity-service/internal/summary/resource.go` | Use `ParseDateParam`; delete `resolveTimezone` | 1, 3 |
| `services/productivity-service/internal/summary/processor.go` | `TaskSummary(date time.Time)` | 3 |
| `services/productivity-service/internal/task/provider.go:38-54` | `countOverdue` / `countCompletedToday` take date | 3 |
| `services/calendar-service/internal/event/resource.go:43-74` | Require `?start`/`?end`; 400 on missing/malformed | 3 |
| `services/calendar-service/cmd/main.go` | Drop `accountBaseURL` wiring | 3 |

### To Delete — Backend (PR 3)

| Path | Notes |
|---|---|
| `services/tracker-service/internal/tz/` | Whole directory (resolver.go, accountclient.go, resolver_test.go) |
| `services/workout-service/internal/tz/` | Whole directory |
| `services/calendar-service/internal/tz/` | Whole directory |
| `resolveTimezone` func in `services/productivity-service/internal/summary/resource.go` | Inline helper — delete fn only |

### To Modify — Frontend

| File | Change | PR |
|---|---|---|
| `frontend/src/lib/api/client.ts:141-146` | Delete `X-Timezone` injection block | 2 |
| `frontend/src/services/api/tracker.ts` | `getToday(tenant, date)` appends `?date=` | 2 |
| `frontend/src/services/api/workout.ts` | `getToday(tenant, date)` appends `?date=` | 2 |
| `frontend/src/lib/hooks/api/use-trackers.ts` | `useTrackerToday(date)`; key includes date; `useMonthSummary(month, today?)` | 2 |
| `frontend/src/lib/hooks/api/use-workouts.ts` | `useWorkoutToday(date)`; key includes date | 2 |
| `frontend/src/lib/hooks/api/use-tasks.ts` | `useTaskSummary(date)`; key includes date | 2 |
| `frontend/src/components/features/trackers/habits-widget.tsx` | Compute date via `useLocalDate`; pass to hook | 2 |
| `frontend/src/components/features/workouts/workout-widget.tsx` | Same pattern | 2 |
| `frontend/src/components/features/tracker/today-view.tsx` | Same pattern | 2 |
| `frontend/src/pages/WorkoutTodayPage.tsx` | Same pattern | 2 |
| `frontend/src/components/features/tracker/calendar-grid.tsx` | Thread `today` into `useMonthSummary` | 2 |
| Dashboard task summary callers | `grep useTaskSummary`; thread date arg | 2 |

### To Modify — Docs (PR 3)

| File | Change |
|---|---|
| `services/tracker-service/docs/rest.md` | Document `?date=`, `?today=`; remove X-Timezone |
| `services/workout-service/docs/rest.md` | Document `?date=`; remove X-Timezone |
| `services/productivity-service/docs/rest.md` | Document `?date=` on summary endpoints |
| `services/calendar-service/docs/rest.md` | Document required `?start`/`?end` |

### Reference (Do Not Modify)

| File | Why reference |
|---|---|
| `frontend/src/lib/date-utils.ts` | Existing `getLocalTodayStr(tz)` etc. — call sites already work in other widgets (meals, calendar) |
| `frontend/src/components/features/meals/meal-plan-widget.tsx` | Template for client-computed-date pattern |
| `frontend/src/components/features/calendar/calendar-widget.tsx` | Second template |
| `shared/go/auth/auth.go:114` | Cookie-based auth (explains why Authorization fallback fails) |
| `shared/go/server/response.go` | Existing error-writing helper (`WriteError`) for 400 responses |

---

## Key Decisions

1. **Three-PR rollout** (not big-bang). Backend accept-both → frontend switch → backend require + cleanup. Avoids any mid-deploy 400s. See `migration-plan.md`.

2. **Shared parser in `shared/go/http`**, not duplicated per service. DRY. `go.work` handles module resolution.

3. **Date anchored to UTC midnight** in backend. `ParseDateParam` returns `time.Date(Y, M, D, 0, 0, 0, 0, time.UTC)`. Calendar day, not instant. DB queries against `type:date` GORM columns compare by date irrespective of location.

4. **Hard-remove `X-Timezone` header** in PR 2. No transitional double-send. Silent fallbacks rot.

5. **Missing param → 400**, not silent default. We own both sides of the wire; loud failures beat silent wrong data.

6. **Secondary `time.Now()` calls get the client date too.** Tracker's entry/month gating, workout's planned-item completion — all switch to client-supplied date. One source of truth per request.

7. **Reminder `CURRENT_TIMESTAMP` stays.** Timestamp comparisons are timezone-invariant; no bug there.

8. **Household timezone field retained** for cron/retention/fanout code that runs without a request context.

9. **Browser tz as fallback in frontend.** If `household.attributes.timezone` is unloaded, fall back to `Intl.DateTimeFormat().resolvedOptions().timeZone`. User is almost always in the household's tz.

10. **`useLocalDate` hook with 60s polling.** Keeps long-open tabs correct across midnight. React Query keys include the date, so the fetch happens automatically on re-render.

---

## Dependencies

- **Task 038 (done):** created `frontend/src/lib/date-utils.ts` with `getLocalTodayStr(tz)` etc. This task consumes those utilities.
- **Task 035 (done):** introduced the `X-Timezone` header + `internal/tz/` pattern. This task undoes that approach.
- **`shared/go/http`:** already in `go.work`; no module setup needed.
- **`shared/go/auth/auth.go`:** cookie-based auth is the root cause of the broken fallback — no changes needed, just understood.

---

## Patterns to Follow

### Backend handler — parse or 400

```go
date, err := httpparams.ParseDateParam(r, "date")
if err != nil {
    server.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
    return
}
// ... use date directly in processor call
```

### Frontend widget — client-computed date

```tsx
import { useTenant } from "@/context/tenant-context";
import { useLocalDate } from "@/lib/hooks/use-local-date";
import { useTrackerToday } from "@/lib/hooks/api/use-trackers";

export function HabitsWidget() {
  const { household } = useTenant();
  const today = useLocalDate(household?.attributes.timezone);
  const { data, isLoading } = useTrackerToday(today);
  // ... render
}
```

### React Query key with date

```ts
export const trackerKeys = {
  // ...
  today: (tenant, household, date: string) =>
    [...trackerKeys.all(tenant, household), "today", date] as const,
};
```

### `useLocalDate` sketch

```ts
import { useEffect, useState } from "react";
import { getLocalTodayStr } from "@/lib/date-utils";

export function useLocalDate(tz?: string): string {
  const [date, setDate] = useState(() => getLocalTodayStr(tz));
  useEffect(() => {
    const id = window.setInterval(() => {
      const next = getLocalTodayStr(tz);
      setDate((prev) => (prev === next ? prev : next));
    }, 60_000);
    return () => window.clearInterval(id);
  }, [tz]);
  return date;
}
```

---

## Verification Checklist (Post-PR 3)

- [ ] `grep -r "tz\.Resolve\|X-Timezone" services/ shared/` returns nothing.
- [ ] `grep -rn "X-Timezone" frontend/src/` returns nothing.
- [ ] Dashboard at device time = 8:30 PM US Eastern shows today's habits/workout, not tomorrow's.
- [ ] Tab left open across local midnight auto-refreshes within 60 seconds.
- [ ] `./scripts/test-all.sh` green.
- [ ] `./scripts/lint-all.sh` green.
- [ ] Docker builds green for tracker, workout, productivity, calendar, frontend.
