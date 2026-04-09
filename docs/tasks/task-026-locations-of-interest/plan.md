# Implementation Plan — Locations of Interest

Last Updated: 2026-04-09

## Executive Summary

Add per-household "locations of interest" so household members can save extra places (vacation home, parents' city, etc.) and view their forecast on the Weather page without changing the household's primary location. The feature is contained within the **weather-service** (new domain package) and the **frontend**; the account-service and household model are unchanged.

The largest non-trivial change is reshaping `weather.weather_caches` so cache rows are keyed by `(household_id, location_id)` instead of `household_id` alone, with `location_id IS NULL` representing the household's primary location. The migration uses GORM AutoMigrate plus a hand-written step (idempotent SQL) for the partial unique indexes and FK that GORM tags cannot express.

## Current State Analysis

**Backend (`services/weather-service`):**
- `internal/forecast/` — single domain package owning the cache. `Entity` has a unique index on `household_id` (`uniqueIndex:idx_weather_household`). `Processor.getOrFetch` keys solely by `householdID`. `AllProvider` returns every row (the refresh loop already iterates the whole table).
- `internal/refresh/refresh.go` — calls `proc.AllProvider()` then `RefreshCache(m)` per row. No per-row keying logic — refactor reach is minimal.
- `internal/geocoding/` — wraps Open-Meteo geocoding; the frontend already calls it.
- `cmd/main.go` — wires `forecast.Migration` into the DB connect call and registers `forecast.InitializeRoutes` and `geocoding.InitializeRoutes`. Refresh loop is `go refresh.StartRefreshLoop(...)`.
- Multi-tenancy and JWT middleware are already wired (`sharedauth.Middleware`).

**Frontend (`frontend/src`):**
- `pages/WeatherPage.tsx` — uses `useWeatherForecast()` and `useCurrentWeather()` from `lib/hooks/api/use-weather.ts`. Empty state when household has no location.
- `services/api/weather.ts` — current/forecast/geocoding clients.
- `lib/hooks/api/use-weather.ts` — TanStack Query hooks; query keys must be extended.
- Existing geocoding-search component is reused by the household location picker (target for reuse in the Manage Locations dialog).
- `weather-widget.tsx` (dashboard) is explicitly out of scope.

**Account-service:** unchanged. Household primary location stays in `account-service`.

## Proposed Future State

- **New domain package** `internal/locationofinterest/` in weather-service mirroring the `forecast` package layout (model/builder/entity/processor/rest/resource/administrator/provider).
- **Cache reshape:** `weather_caches` gains a nullable `location_id` UUID column (FK to `locations_of_interest.id ON DELETE CASCADE`). Two partial unique indexes replace the old `(household_id)` unique. `forecast.Model`/`Processor` carry an optional `*uuid.UUID` location ID throughout.
- **Existing weather endpoints** (`/weather/current`, `/weather/forecast`) accept an optional `locationId` query parameter that resolves coordinates from `locations_of_interest` (verifying household ownership) and overrides any `lat`/`lon` query params.
- **New endpoints** under `/api/v1/locations-of-interest` for list/create/update/delete. POST synchronously warms the cache; cache failures are logged and do not fail the create.
- **Frontend** Weather page renders a location selector when ≥1 saved locations exist, defaulting to primary on every mount. New `ManageLocationsDialog` is always reachable from the Weather page header. New `use-locations-of-interest.ts` hooks. `use-weather.ts` accepts optional `locationId`, threaded into query keys and service calls.
- **Background refresh** continues to iterate every cache row — no logic change beyond data shape.

## Implementation Phases

### Phase 1 — Backend Domain & Migration

**P1.T1 — Create `locationofinterest` domain package skeleton (M)**
- Create `internal/locationofinterest/{model.go, builder.go, entity.go, provider.go, administrator.go, processor.go, resource.go, rest.go}` mirroring `internal/forecast/` patterns (immutable Model with private fields + Builder + Make/ToEntity).
- Entity table name `locations_of_interest`. Columns per PRD §6. `label` is `*string` to allow NULL.
- Domain validation: trim `label`, reject if length > 64.
- **Acceptance:** package compiles in isolation; entity tags produce the table on `AutoMigrate`.

**P1.T2 — Wire `locationofinterest.Migration` into `cmd/main.go` (S)**
- Add to `database.SetMigrations(...)`. Must run before any forecast cache migration step that adds the FK.
- **Acceptance:** service boot creates the new table with no error.

**P1.T3 — Reshape `forecast.Entity` for `location_id` (M)**
- Add `LocationId *uuid.UUID` field on `forecast.Entity` (no GORM unique tag — drop the existing `uniqueIndex:idx_weather_household`).
- Update `forecast.Model`, `Builder`, `Make`, `ToEntity` to thread `locationID *uuid.UUID`.
- Update all providers (`getByHouseholdID`, `getAll`, etc.) — replace `getByHouseholdID(householdID)` with `getByHouseholdAndLocation(householdID, locationID *uuid.UUID)` (NULL-safe SQL: `location_id IS NULL` when arg is nil, else `location_id = ?`).
- **Acceptance:** `forecast` package builds; existing tests updated and passing.

**P1.T4 — Hand-written index/FK migration step (M)**
- Add a post-AutoMigrate hook (e.g., a `forecast.PostMigration(db)` or new `internal/migration` helper) that runs the idempotent SQL from `migration-plan.md`:
  - `DROP INDEX IF EXISTS idx_weather_household`
  - `CREATE UNIQUE INDEX IF NOT EXISTS idx_weather_household_primary ... WHERE location_id IS NULL`
  - `CREATE UNIQUE INDEX IF NOT EXISTS idx_weather_household_location ... WHERE location_id IS NOT NULL`
  - Add FK `fk_weather_location` guarded by `pg_constraint` existence check.
- Wire into `database.SetMigrations` (or run after the AutoMigrate step on the same DB).
- **Acceptance:** integration test (or manual SQL inspection) confirms both partial uniques and FK exist; existing rows are intact with `location_id = NULL`.

**P1.T5 — `forecast.Processor` accepts optional locationID (M)**
- `getOrFetch`/`fetchAndCache`/`RefreshCache` signatures gain `locationID *uuid.UUID`.
- `create()` writes `location_id` into the cache row.
- `RefreshCache(m)` carries `m.LocationID()` through.
- **Acceptance:** unit tests cover (a) primary cache lookup with NULL location_id; (b) cache lookup with non-nil location_id; (c) refresh loop updates correct row.

### Phase 2 — Backend REST & Refresh

**P2.T1 — `locationofinterest.Processor` CRUD (M)**
- `List(tenantID, householdID)` — ordered `created_at ASC`.
- `Create(tenantID, householdID, label *string, placeName string, lat, lon float64)`:
  - Trim label; reject if >64 chars.
  - Normalize `lat`/`lon` to 4 decimal places via `math.Round(x*10000)/10000`.
  - Count existing rows for the household; if ≥10 return a sentinel `ErrCapReached`.
  - Persist row.
  - Synchronously call `forecastProcessor.GetCurrent(...)` with the new coordinates and `locationID` to warm cache; log and swallow errors.
- `UpdateLabel(tenantID, householdID, id, newLabel *string)` — reject unknown id (404). 64-char cap enforced.
- `Delete(tenantID, householdID, id)` — hard delete; FK cascade removes cache row.
- **Acceptance:** unit tests cover happy paths plus 10-cap, missing id, cross-household access.

**P2.T2 — REST handlers and routes (M)**
- `locationofinterest.InitializeRoutes(db, forecastProc)(l, si, api)` registering:
  - `GET /locations-of-interest`
  - `POST /locations-of-interest`
  - `PATCH /locations-of-interest/{id}`
  - `DELETE /locations-of-interest/{id}`
- JSON:API request/response shapes per `api-contracts.md`. Resource type `location-of-interest`.
- 409 on cap with **exact** message: `Households can save up to 10 locations of interest. Remove one to add another.`
- 404 on cross-household / unknown id.
- Wire into `cmd/main.go`.
- **Acceptance:** REST handler tests cover all endpoints, status codes, and the 409 message verbatim.

**P2.T3 — Modify weather endpoints to accept `locationId` (M)**
- `forecast.InitializeRoutes` (or its handler functions) — read optional `locationId` query param.
- When present: resolve via `locationofinterest.Processor.GetByID(tenantID, householdID, id)` (returns 404 if not owned), use its `lat`/`lon`/normalized coordinates, ignore `latitude`/`longitude` query params, and pass `&id` as `locationID` into `forecastProcessor.GetCurrent`/`GetForecast`.
- When absent: behavior unchanged (locationID = nil).
- **Acceptance:** rest tests cover both branches; cross-household locationId returns 404.

**P2.T4 — Refresh loop (S — verification only)**
- Confirm `refresh.refreshAll` still works without code changes after the data-model reshape (it iterates `proc.AllProvider()` and calls `RefreshCache(m)`; both now naturally include the location_id per row).
- Update log line in `refresh.go` to include `location_id` (or "primary") for per-row error logs.
- **Acceptance:** unit test stubs `forecast.Processor.AllProvider` returning a mix of nil and non-nil location_ids and verifies all are refreshed.

### Phase 3 — Frontend

**P3.T1 — API client + types (S)**
- New `frontend/src/services/api/locations-of-interest.ts` — list/create/update/delete using JSON:API helpers in `services/api/base.ts`.
- New `frontend/src/types/models/location-of-interest.ts` — TS shape mirroring the API contract.

**P3.T2 — React Query hooks (M)**
- New `frontend/src/lib/hooks/api/use-locations-of-interest.ts` — `useLocationsOfInterest()`, `useCreateLocationOfInterest()`, `useUpdateLocationOfInterest()`, `useDeleteLocationOfInterest()`.
- Add query keys to `query-keys.ts`. Mutations invalidate the list and the affected weather queries (current + forecast for that locationId, and primary if relevant).
- **Acceptance:** hooks compile, list query renders, mutation invalidations cover the expected keys.

**P3.T3 — Extend `use-weather.ts` to accept optional `locationId` (S)**
- `useCurrentWeather(locationId?: string)` and `useWeatherForecast(locationId?: string)`.
- Thread `locationId` into the service call query string and TanStack query key (so caches don't collide).
- When `locationId` is present, omit lat/lon from the request.
- **Acceptance:** existing primary-location callers compile unchanged (locationId is optional); switching locationId triggers a new query.

**P3.T4 — Weather page selector + state (M)**
- `WeatherPage.tsx`: `const [selectedLocationId, setSelectedLocationId] = useState<string | undefined>(undefined)` (always undefined on mount = primary).
- Fetch `useLocationsOfInterest()`. Render selector ONLY when `data.length > 0`.
- Selector lists primary first (label = `household.locationName ?? "Primary Location"`), then each saved location's label or place_name fallback.
- Pass `selectedLocationId` into `useWeatherForecast` and `useCurrentWeather`.
- "Manage Locations" affordance always visible — opens dialog from P3.T5.
- **Acceptance:** selector hidden at zero saved, shown at ≥1; selecting a saved location swaps content; reload resets to primary.

**P3.T5 — `ManageLocationsDialog` component (L)**
- New `frontend/src/components/weather/manage-locations-dialog.tsx`.
- Lists existing locations (label, place name, edit/remove buttons).
- "Add Location" button opens the existing geocoding-search component (verify reusability — Open Question §9 in PRD); after pick, optional label input (max 64), confirm → POST.
- Inline rename via PATCH.
- Disables "Add Location" + shows hint when at 10-location cap.
- **Acceptance:** all CRUD flows work end-to-end against the live API; cap UX verified.

**P3.T6 — Empty-state shortcuts (S)**
- When `!locationSet` and `useLocationsOfInterest()` returns ≥1 row, render `"Or pick one of your saved locations:"` line plus chip per saved location below the existing empty state. Clicking a chip sets `selectedLocationId` and the page swaps to that location's view.
- **Acceptance:** unset-primary household with saved locations can drill in via the chips and back via the (now-visible) header selector.

### Phase 4 — Verification

**P4.T1 — Backend tests + Docker build (M)**
- Run `go test ./...` in weather-service.
- Run `docker build` for weather-service per CLAUDE.md instruction (shared libraries may be touched indirectly).
- **Acceptance:** all tests green; image builds clean.

**P4.T2 — Frontend type-check + unit tests (S)**
- `pnpm tsc --noEmit` (or project equivalent), `pnpm test`.
- **Acceptance:** no type errors; existing tests pass; new hook tests added where reasonable.

**P4.T3 — Acceptance criteria walkthrough (S)**
- Walk PRD §10 acceptance list and tick each item against the running stack (manual smoke + automated where possible).
- **Acceptance:** every PRD §10 box ticked.

## Risk Assessment & Mitigation

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Partial unique index migration fails on non-empty `weather_caches` | Med | High | Use `CREATE ... IF NOT EXISTS` and verify idempotence in a local copy of prod data; document rollback per `migration-plan.md`. |
| GORM AutoMigrate ordering puts FK before `locations_of_interest` exists | Med | Med | Register `locationofinterest.Migration` BEFORE the post-step that adds the FK; the FK lives in the hand-written step, not in struct tags. |
| `forecast.Processor` signature change ripples through callers | High | Low | Compiler will catch all sites; update each in one pass. Keep `locationID` as `*uuid.UUID` (nilable) so primary callers pass `nil`. |
| Synchronous cache warming on POST adds latency / fails | Med | Low | Swallow errors and log; client receives 201 either way (per PRD §4.1). |
| Existing geocoding-search component not reusable | Low | Med | Audit during P3.T5; if a small refactor is needed, do it in-place rather than forking. |
| Frontend query-key collisions across primary vs saved location | Med | Med | Always include `locationId` (or sentinel `"primary"`) in TanStack query keys. |
| Refresh loop runtime grows with N households × M locations | Low | Low | Hard cap of 10 + existing throttle keep cadence well within 15-min tick at expected scale (PRD §4.3). Revisit if scale changes. |

## Success Metrics

- All PRD §10 acceptance criteria pass.
- `go test ./...` and `docker build` pass for weather-service.
- Frontend type-check and tests pass.
- No regression in existing primary-location weather behavior (visual + API).
- Zero added external dependencies.

## Required Resources & Dependencies

- **Code:** weather-service (`internal/forecast`, new `internal/locationofinterest`, `cmd/main.go`, `internal/refresh`); frontend (`pages/WeatherPage.tsx`, `services/api/weather.ts`, `services/api/locations-of-interest.ts`, `lib/hooks/api/use-weather.ts`, `lib/hooks/api/use-locations-of-interest.ts`, `components/weather/manage-locations-dialog.tsx`, geocoding search component).
- **Infra:** PostgreSQL feature support for partial unique indexes (already in use). No new services or env vars.
- **External APIs:** Open-Meteo (existing client; no new endpoints called).
- **Docs:** PRD, `api-contracts.md`, `migration-plan.md` (this folder).

## Timeline Estimates

Per CLAUDE.md, no calendar estimates. Effort sizing only:

| Phase | Effort |
|---|---|
| Phase 1 — Backend Domain & Migration | M–L |
| Phase 2 — Backend REST & Refresh | M |
| Phase 3 — Frontend | L |
| Phase 4 — Verification | S–M |
