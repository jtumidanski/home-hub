# Context — Locations of Interest

Last Updated: 2026-04-09

## Spec Documents (this folder)

- `prd.md` — Product requirements (the source of truth for scope and acceptance criteria).
- `api-contracts.md` — JSON:API request/response shapes for new + modified endpoints.
- `migration-plan.md` — Step-by-step `weather_caches` reshape, including the partial-unique-index SQL.
- `plan.md` — Implementation phases and tasks (this plan).
- `tasks.md` — Checklist mirroring `plan.md`.

## Key Backend Files

| File | Why it matters |
|---|---|
| `services/weather-service/cmd/main.go` | Wires migrations, REST routes, refresh loop. Add new domain package here. |
| `services/weather-service/internal/forecast/entity.go` | Holds `Entity` with `uniqueIndex:idx_weather_household` — must be reshaped (drop unique tag, add `LocationId *uuid.UUID`). |
| `services/weather-service/internal/forecast/model.go` | Immutable Model — must carry optional `locationID *uuid.UUID`. |
| `services/weather-service/internal/forecast/builder.go` | Mirror new field. |
| `services/weather-service/internal/forecast/processor.go` | `getOrFetch`, `fetchAndCache`, `RefreshCache` signatures change to thread `locationID`. |
| `services/weather-service/internal/forecast/provider.go` | `getByHouseholdID` → `getByHouseholdAndLocation` (NULL-safe). |
| `services/weather-service/internal/forecast/rest.go` | `/weather/current` and `/weather/forecast` handlers gain optional `locationId` query param. |
| `services/weather-service/internal/refresh/refresh.go` | Iterates all rows; should work unchanged after data shape — verify and improve log fields. |
| `services/weather-service/internal/geocoding/` | Existing client used by frontend for place search; unchanged. |

## Key Frontend Files

| File | Why it matters |
|---|---|
| `frontend/src/pages/WeatherPage.tsx` | Add location selector and empty-state shortcuts. |
| `frontend/src/services/api/weather.ts` | Add optional `locationId` to current/forecast service calls. |
| `frontend/src/lib/hooks/api/use-weather.ts` | Extend hooks to accept optional `locationId`; include in query keys. |
| `frontend/src/lib/hooks/api/query-keys.ts` | Add new key family for locations-of-interest. |
| `frontend/src/components/dashboard/weather-widget.tsx` | **OUT OF SCOPE** — explicitly unchanged. |
| (new) `frontend/src/services/api/locations-of-interest.ts` | New CRUD client. |
| (new) `frontend/src/lib/hooks/api/use-locations-of-interest.ts` | New TanStack Query hooks. |
| (new) `frontend/src/components/weather/manage-locations-dialog.tsx` | New management UI. |
| (new) `frontend/src/types/models/location-of-interest.ts` | Shared TS types. |

## Key Decisions Locked In (from PRD)

- **Cap = 10** locations per household; enforced in processor, not DB. 11th create returns `409 Conflict` with the **exact** message: `Households can save up to 10 locations of interest. Remove one to add another.`
- **Coordinate precision:** normalize to 4 decimal places on create.
- **No deduplication** by coordinates — same place may appear twice with different labels.
- **No `created_by`** audit field (matches existing weather-service entities).
- **Cache key:** `(household_id, location_id)` with `location_id IS NULL` representing the primary location. Implemented via two **partial unique indexes** in PostgreSQL.
- **Selection state:** lives in component state only — resets to primary on every Weather page mount. No URL param, no localStorage, no server persistence.
- **Synchronous cache warm on create:** failures are logged but do not fail the POST (still returns 201).
- **Authorization:** any active member of the household, not admin-only.
- **Empty-state behavior:** when primary is unset but saved locations exist, render shortcut chips below the existing empty state.
- **Dashboard widget unchanged.**
- **Cross-service cleanup on household delete:** out of scope (matches existing weather-service orphaning behavior).

## Open Items to Verify During Implementation

1. **GORM AutoMigrate vs partial unique indexes** — confirmed in `migration-plan.md` to require a hand-written post-step. No struct-tag-only solution.
2. **Geocoding-search component reusability** — audit during P3.T5; refactor in place if needed.
3. **`weather-service/cmd/main.go` migration ordering** — ensure `locationofinterest.Migration` registers before any FK-adding step.

## Patterns to Mirror

- The `forecast` package is the canonical example for the new `locationofinterest` package: model + builder + entity + provider + administrator + processor + rest + resource. Follow the same file split and immutability conventions.
- JSON:API request/response handling — copy the patterns used by existing `forecast` REST handlers.
- Multi-tenancy: all queries filter by `tenant_id` from request context. Cross-household lookups must always join on `household_id = activeHousehold` and return 404 (not 403) on mismatch.
