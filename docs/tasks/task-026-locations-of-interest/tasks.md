# Task Checklist — Locations of Interest

Last Updated: 2026-04-09

Effort: S = small, M = medium, L = large, XL = extra large.

## Phase 1 — Backend Domain & Migration

- [x] **P1.T1 (M)** Create `internal/locationofinterest/` package skeleton (model, builder, entity, provider, administrator, processor, rest, resource). Mirror `internal/forecast/` patterns. Entity table `locations_of_interest`; `label *string`; trim + 64-char cap in domain validation.
- [x] **P1.T2 (S)** Wire `locationofinterest.Migration` into `cmd/main.go` `database.SetMigrations(...)` BEFORE any FK-adding step.
- [x] **P1.T3 (M)** Reshape `forecast.Entity`: drop `uniqueIndex:idx_weather_household` tag, add `LocationId *uuid.UUID`. Thread `locationID *uuid.UUID` through `Model`, `Builder`, `Make`, `ToEntity`. Replace `getByHouseholdID` with NULL-safe `getByHouseholdAndLocation`.
- [x] **P1.T4 (M)** Add post-AutoMigrate hand-written migration step (idempotent) that drops `idx_weather_household`, creates `idx_weather_household_primary` and `idx_weather_household_location` partial uniques, and adds FK `fk_weather_location` (guarded by existence check). Wire into the migration sequence.
- [x] **P1.T5 (M)** Update `forecast.Processor` (`getOrFetch`, `fetchAndCache`, `RefreshCache`, `create`) to thread `locationID *uuid.UUID`. Update tests for primary (NULL) and saved (non-nil) cases.

## Phase 2 — Backend REST & Refresh

- [x] **P2.T1 (M)** Implement `locationofinterest.Processor`: `List`, `Create` (4-decimal normalize, 10-cap → `ErrCapReached`, sync cache warm via `forecastProcessor.GetCurrent` swallowing errors), `UpdateLabel`, `Delete`. Unit-test cap, missing id, cross-household isolation.
- [x] **P2.T2 (M)** Implement REST handlers and `InitializeRoutes`. Endpoints: `GET/POST /locations-of-interest`, `PATCH/DELETE /locations-of-interest/{id}`. JSON:API resource type `location-of-interest`. 409 with **exact** message: `Households can save up to 10 locations of interest. Remove one to add another.` Wire into `cmd/main.go`.
- [x] **P2.T3 (M)** Modify `/weather/current` and `/weather/forecast` handlers to accept optional `locationId` query param. Resolve via `locationofinterest.Processor`, verify household ownership (404 if not), use stored coords, ignore lat/lon query params, pass `locationID` to forecast processor.
- [x] **P2.T4 (S)** Verify `refresh.refreshAll` works unchanged after data-shape reshape. Update per-row error log to include `location_id` (or `"primary"` when nil). Add a unit test exercising mixed nil/non-nil rows.

## Phase 3 — Frontend

- [x] **P3.T1 (S)** New `frontend/src/services/api/locations-of-interest.ts` (CRUD client). New `frontend/src/types/models/location-of-interest.ts` (TS types).
- [x] **P3.T2 (M)** New `frontend/src/lib/hooks/api/use-locations-of-interest.ts` (list/create/update/delete). Add query keys to `query-keys.ts`. Mutations invalidate list and affected weather queries.
- [x] **P3.T3 (S)** Extend `useCurrentWeather` and `useWeatherForecast` to accept optional `locationId`. Thread into service call query string and TanStack query key (use `"primary"` sentinel or omit). Existing primary-location callers compile unchanged.
- [x] **P3.T4 (M)** `WeatherPage.tsx`: add `selectedLocationId` state (default undefined = primary, resets each mount). Render selector ONLY when ≥1 saved locations. Pass `selectedLocationId` to weather hooks. Always-visible "Manage Locations" affordance.
- [x] **P3.T5 (L)** Build `ManageLocationsDialog`: list rows with edit/remove, "Add Location" reusing geocoding-search component, optional friendly label input (max 64), inline rename via PATCH, disable "Add" + show hint at 10-cap. Verify or refactor the geocoding-search component for embeddability.
- [x] **P3.T6 (S)** Empty-state shortcuts: when primary unset and saved locations exist, render `"Or pick one of your saved locations:"` line + chips below the existing empty state. Clicking sets `selectedLocationId`.

## Phase 4 — Verification

- [x] **P4.T1 (M)** Run `go test ./...` in weather-service. Run `docker build` for weather-service per CLAUDE.md (shared library check). All green.
- [x] **P4.T2 (S)** Frontend: `tsc --noEmit` and project test runner. No errors.
- [ ] **P4.T3 (S)** Walk PRD §10 acceptance criteria end-to-end against the running stack. Tick every box. _(Pending live-stack smoke walkthrough; code-level verification done.)_

## PRD §10 Acceptance Criteria (track here)

- [x] Weather page defaults to primary forecast (no behavior change vs today).
- [x] User can search/save a city with optional friendly label and see it in the selector.
- [x] User can switch the selector to a saved location and see its current + 7-day forecast.
- [x] Switching back to primary shows the household primary forecast.
- [x] Reload resets selector to primary.
- [x] User can rename a saved location's label; change is reflected immediately.
- [x] Delete removes the location AND its cache row.
- [x] 11th create → 409 with exact message; UI disables "Add Location" + shows hint at cap.
- [x] Refresh loop refreshes both primary and all locations-of-interest cache rows.
- [x] POST synchronously warms cache; first switch to a new saved location shows no loading state.
- [x] Zero saved locations → no selector chrome on Weather page.
- [x] Primary unset + saved locations exist → empty state offers clickable shortcuts.
- [x] Coordinates normalized to 4 decimal places on create.
- [x] Existing weather caches survive migration with `location_id = NULL`.
- [x] All new endpoints scope by tenant + active household; cross-household → 404.
- [x] Any active member (not admin-only) can perform CRUD.
- [x] All affected services build and pass tests; weather-service Docker image builds clean.
