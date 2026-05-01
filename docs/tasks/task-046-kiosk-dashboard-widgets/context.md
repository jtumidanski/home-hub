# Context — Kiosk Dashboard Widgets (task-046)

This is a quick-reference companion to `plan.md`. It lists the files touched, the conventions to mirror, and the small set of design deviations the plan adopts after reading the actual code.

## Source-of-truth documents

- `prd.md` — feature requirements
- `design.md` — architecture decisions on top of the PRD
- `api-contracts.md` — seed endpoint contract
- `data-model.md` — widget allowlist + Zod schemas + optional schema changes

## Deviations from `design.md`

These are corrections the plan adopts after reading the code; they do not change feature behavior, only mechanism.

1. **`household_preferences` is NOT a key/value bag.** It is a fixed-schema GORM `Entity` with one mutable field today (`default_dashboard_id`). Adding `kiosk_dashboard_seeded` requires a real boolean column on `household_preferences`, plus a new dedicated sub-route to flip it. (The shared `/api/v1/preferences` path the design referenced is a *user-only* prefs row with `theme` + `active_household_id`; it has no per-household scoping.)
2. **Dedicated PATCH sub-route instead of extending the existing UpdateRequest.** The existing `PATCH /household-preferences/{id}` body has a documented FIXME: nil pointers cannot distinguish "absent" from "explicit null" so they always clear the field. Rather than retrofit a custom JSON unmarshaler in this task, we add a dedicated sub-route `PATCH /api/v1/household-preferences/{id}/kiosk-seeded` (mirrors the pattern used by `dashboards/order` for non-JSON:API actions). Body: `{"value": true}`. The flag is write-once-true; the frontend never writes false.
3. **No new query hook for tomorrow's calendar.** The existing `useCalendarEvents(start, end)` is already date-keyed and serves both "today" and "tomorrow" calls without refactoring `CalendarWidget`. Design §5.2's "extract `useCalendarEventsForDate`" refactor is dropped — the existing hook is the shared plumbing.
4. **Tasks/reminders share queries via the existing list hooks.** `useTasks()` and `useReminders()` already return the full household list keyed once. The today/tomorrow adapters slice client-side; React Query dedupes naturally.

## Widgets — read-only convention

Per design §4.2, every new adapter:
- Imports only query hooks (`useTasks`, `useReminders`, `useCurrentWeather`, `useWeatherForecast`, `useCalendarEvents`). No mutation hooks.
- Header rows are `<Link>` components, not `<button onClick={...}>`.
- Empty / error states are presentational.
- Begins with the comment `// Read-only widget — no mutations. See PRD §4.x.`

## Files — backend

| File | Touch |
|---|---|
| `shared/go/dashboard/types.go` | Add 5 entries to `WidgetTypes` map |
| `shared/go/dashboard/fixtures/widget-types.json` | Add same 5 entries |
| `shared/go/dashboard/types_test.go` | Update `TestIsKnownWidgetType` to assert the 5 new types |
| `services/dashboard-service/internal/dashboard/entity.go` | Add `SeedKey *string` field; migration adds column + partial unique index + brownfield backfill |
| `services/dashboard-service/internal/dashboard/processor.go` | `Processor.Seed` gains a `seedKey *string` arg; per-key advisory lock helper |
| `services/dashboard-service/internal/dashboard/processor_test.go` | New tests covering seedKey branches (nil, fresh, idempotent, two distinct keys) and brownfield backfill |
| `services/dashboard-service/internal/dashboard/rest.go` | `SeedRequest.Key *string` field |
| `services/dashboard-service/internal/dashboard/resource.go` | Validate `key` regex; pass to processor |
| `services/dashboard-service/internal/dashboard/rest_test.go` | Tests for seed-with-key endpoint |
| `services/account-service/internal/householdpreference/entity.go` | Add `KioskDashboardSeeded bool` field |
| `services/account-service/internal/householdpreference/builder.go` | Builder method for new field |
| `services/account-service/internal/householdpreference/model.go` | Getter for new field |
| `services/account-service/internal/householdpreference/rest.go` | RestModel exposes new field |
| `services/account-service/internal/householdpreference/administrator.go` | New `markKioskSeeded(id)` SQL helper |
| `services/account-service/internal/householdpreference/processor.go` | `MarkKioskSeeded(id)` processor method |
| `services/account-service/internal/householdpreference/resource.go` | New PATCH sub-route + handler |
| `services/account-service/internal/householdpreference/rest_test.go` | Tests for the new endpoint and field exposure |

## Files — frontend

| File | Touch |
|---|---|
| `frontend/src/lib/dashboard/widget-types.ts` | Add 5 strings to `WIDGET_TYPES` |
| `frontend/src/lib/dashboard/fixtures/widget-types.json` | Add same 5 strings |
| `frontend/src/lib/dashboard/widgets/tasks-today.ts` | New `WidgetDefinition` |
| `frontend/src/lib/dashboard/widgets/reminders-today.ts` | New `WidgetDefinition` |
| `frontend/src/lib/dashboard/widgets/weather-tomorrow.ts` | New `WidgetDefinition` |
| `frontend/src/lib/dashboard/widgets/calendar-tomorrow.ts` | New `WidgetDefinition` |
| `frontend/src/lib/dashboard/widgets/tasks-tomorrow.ts` | New `WidgetDefinition` |
| `frontend/src/lib/dashboard/widgets/meal-plan-today.ts` | Extend Zod schema with `view` |
| `frontend/src/lib/dashboard/widget-registry.ts` | Register the 5 new widgets |
| `frontend/src/lib/dashboard/kiosk-seed-layout.ts` | New seed layout |
| `frontend/src/lib/dashboard/__tests__/kiosk-seed-layout.test.ts` | New |
| `frontend/src/lib/dashboard/__tests__/widget-registry.test.ts` | Extend to verify 5 new widgets |
| `frontend/src/components/features/dashboard-widgets/tasks-today-adapter.tsx` | New |
| `frontend/src/components/features/dashboard-widgets/reminders-today-adapter.tsx` | New |
| `frontend/src/components/features/dashboard-widgets/weather-tomorrow-adapter.tsx` | New |
| `frontend/src/components/features/dashboard-widgets/calendar-tomorrow-adapter.tsx` | New |
| `frontend/src/components/features/dashboard-widgets/tasks-tomorrow-adapter.tsx` | New |
| `frontend/src/components/features/dashboard-widgets/meal-plan-adapter.tsx` | Branch on `view` |
| `frontend/src/components/features/meals/meal-plan-today-detail.tsx` | New component |
| `frontend/src/components/features/dashboard-widgets/__tests__/*.test.tsx` | Adapter tests for the 5 new + meal-plan view |
| `frontend/src/lib/hooks/use-local-date-offset.ts` | New |
| `frontend/src/lib/date-utils.ts` | Add `getLocalDateStrOffset` |
| `frontend/src/lib/hooks/api/use-dashboards.ts` | `useSeedDashboard` accepts optional `key` |
| `frontend/src/services/api/dashboard.ts` | `seedDashboard` accepts optional `key` |
| `frontend/src/lib/hooks/api/use-household-preferences.ts` | `useMarkKioskSeeded` mutation |
| `frontend/src/services/api/household-preferences.ts` | `markKioskSeeded` method |
| `frontend/src/types/models/dashboard.ts` | `HouseholdPreferencesAttributes` adds `kioskDashboardSeeded` |
| `frontend/src/pages/DashboardRedirect.tsx` | Two-seed orchestration with kiosk gate |
| `frontend/src/pages/__tests__/DashboardRedirect.test.tsx` | Cover all 4 seeding scenarios |
| `frontend/src/pages/__tests__/dashboard-renderer-parity.test.tsx` | Union of seed layouts |

## Conventions

- **Backend tests** use sqlite in-memory (`gorm.io/driver/sqlite`) via `setupTestDB`. Postgres-only paths (advisory locks) are skipped on sqlite.
- **Frontend tests** use Vitest + React Testing Library. Existing adapter tests in `frontend/src/components/features/dashboard-widgets/__tests__/` show the mocking pattern.
- **TDD discipline**: every task has a failing test first. Frequent commits — usually one commit per task. Bigger refactors get a "wire-up" commit at the end.
- **Read-only**: adapters never import mutation hooks. Reviewer enforcement only.

## Hooks already available

| Hook | Location | Notes |
|---|---|---|
| `useTasks` | `frontend/src/lib/hooks/api/use-tasks.ts` | Returns full household list, 5-min stale; share between tasks-today and tasks-tomorrow |
| `useReminders` | `frontend/src/lib/hooks/api/use-reminders.ts` | Returns full list; filter `active && !isReminderSnoozed(r)` for "active" set |
| `useCurrentWeather`, `useWeatherForecast` | `frontend/src/lib/hooks/api/use-weather.ts` | `WeatherDailyAttributes.date` is `YYYY-MM-DD`; pick entry where `date == tomorrow` |
| `useCalendarEvents(start, end)` | `frontend/src/lib/hooks/api/use-calendar.ts` | Already date-keyed; cache shape works for tomorrow without refactor |
| `useLocalDate(tz)` | `frontend/src/lib/hooks/use-local-date.ts` | Pattern to mirror for `useLocalDateOffset` |

## Models referenced

- `Task.attributes`: `title`, `status` (`pending`/`completed`), `dueOn` (YYYY-MM-DD), `ownerUserId`, `completedAt`. Helper `isTaskOverdue(task)`.
- `Reminder.attributes`: `title`, `notes`, `scheduledFor` (ISO timestamp), `active`, `lastSnoozedUntil`.
- `WeatherDaily.attributes`: `date` (YYYY-MM-DD), `highTemperature`, `lowTemperature`, `temperatureUnit`, `summary`, `icon`.
- `CalendarEvent.attributes`: `startTime`, `endTime` (ISO), `allDay` (bool), `title`, `userColor`, `userDisplayName`.
- `HouseholdPreferencesAttributes`: gains `kioskDashboardSeeded: boolean` after this task.

## Acceptance gates per phase

- **Phase A** (allowlist) — `go test ./shared/go/dashboard/...` and `pnpm --filter frontend test src/lib/dashboard/__tests__/widget-types.test.ts` pass with 5 new entries asserted.
- **Phase B** (per-key seed) — `go test ./services/dashboard-service/...` passes; processor tests cover nil/key/idempotent/two-keys/brownfield branches.
- **Phase C** (kiosk flag column) — `go test ./services/account-service/...` passes; new endpoint tests cover happy path + 404 + idempotent flip.
- **Phase D** (frontend wiring) — type-check (`pnpm --filter frontend tsc --noEmit`) clean; service unit tests pass.
- **Phase E** (new widgets) — adapter tests pass; widget-registry test passes; type-check clean.
- **Phase F** (meal-plan extension) — meal-plan-today schema test + adapter branch test pass; existing meal-plan adapter tests still green.
- **Phase G** (kiosk seed layout) — `kiosk-seed-layout.test.ts` passes; updated parity test passes.
- **Phase H** (DashboardRedirect) — all four scenarios in the redirect test pass.
- **Phase I** (full sweep) — `pnpm --filter frontend test`, `pnpm --filter frontend tsc --noEmit`, `pnpm --filter frontend lint`, `go test ./...` from repo root all pass; `scripts/local-up.sh` smoke-renders both dashboards.
