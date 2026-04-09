# Locations of Interest — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-09
---

## 1. Overview

The Weather page currently shows a single forecast tied to the household's primary location (the `latitude`/`longitude` set on the household). Users have asked for the ability to also track the weather for additional places they care about — a vacation home, parents' city, an upcoming trip — without changing the household's primary location.

This feature introduces "locations of interest": named saved places, scoped to a household, that the Weather page can switch to view alongside the primary location. The household's primary location remains the default whenever the Weather page is opened. Locations of interest reuse the existing Open-Meteo geocoding flow for adding new places and inherit the household's unit and timezone preferences.

Storage and caching live in the weather-service (the only consumer), keeping the account-service household model unchanged. The cache key changes from "one row per household" to "one row per (household, location)" so each location of interest gets its own cache and background refresh.

## 2. Goals

Primary goals:
- Allow household members to add named "locations of interest" to a household via city/place search.
- Allow members to edit a location's friendly label and remove locations they no longer care about.
- Enforce a per-household cap on locations of interest to protect upstream rate limits.
- Display the household's primary location forecast by default on the Weather page, with a selector to switch to any saved location of interest.
- Keep all locations' caches warm via the existing background refresh loop.
- Eagerly warm a saved location's cache at creation time so the first switch to it has no extra wait.

Non-goals:
- Per-location unit or timezone overrides (all inherit from the household).
- Persisting the user's last-selected location across page visits or devices — selection always resets to the primary location.
- Removing or moving the household's primary location storage out of account-service.
- Sharing locations of interest across households or between tenants.
- Per-user (rather than per-household) saved locations.
- Weather alerts, push notifications, or automations tied to locations of interest.
- Per-location refresh-interval configuration.
- Promoting a saved location to be the household's primary location ("Use this as primary"), or importing the current primary as a saved location before changing it.
- Drag-to-reorder or search/filter inside the location selector — order is `created_at ASC` and the 10-item cap keeps the list short.
- Exposing a location selector on the dashboard weather widget — the dashboard stays primary-location-only.
- Tracking which household member created a location (no `created_by` audit field, matching existing weather-service entities).
- Cross-service cleanup of locations / cache rows when a household is deleted — orphaning matches existing weather-service behavior.

## 3. User Stories

- As a household member, I want the Weather page to keep defaulting to my household's location so my normal use case is unchanged.
- As a household member, I want to add other places to my household (e.g., "Mom's House", "Beach House") so I can check their weather without losing my primary view.
- As a household member, I want to give each saved place a friendly label so I can recognize it at a glance, separate from the geocoded city name.
- As a household member, I want to switch the Weather page between my primary location and any saved location of interest via a dropdown.
- As a household member, I want to edit or remove locations of interest I no longer need.
- As a household member, I want forecasts for saved locations to be just as fresh as the primary location, with no extra wait when I switch to them.

## 4. Functional Requirements

### 4.1 Locations of Interest Entity

- A `location of interest` belongs to a single household and is scoped by tenant.
- Fields: `id`, `tenant_id`, `household_id`, `label` (optional friendly name), `place_name` (geocoded display name from Open-Meteo, e.g. "Paris, Île-de-France, France"), `latitude`, `longitude`, `created_at`, `updated_at`.
- The user-facing name is `label` if non-empty, otherwise `place_name`.
- `latitude` and `longitude` are required and must both be set (no partial coordinates).
- A household may have at most **10** locations of interest. Attempting to create an 11th returns `409 Conflict` with the exact message: `"Households can save up to 10 locations of interest. Remove one to add another."` (Backend and frontend share this copy.)
- Locations are not deduplicated by coordinates — a household may save the same place twice if they wish (e.g., with different labels).
- On save, `latitude` and `longitude` are normalized to **4 decimal places** (~11 m precision) to keep cache rows tidy when the same place is geocoded twice with slightly different floats.
- Deletion is hard delete; the corresponding cache row in `weather_caches` is also removed.
- On create, the processor synchronously fetches weather for the new coordinates and writes the corresponding `weather_caches` row before returning. If the upstream fetch fails, the location is still created and the cache will be populated by the next refresh tick or first user view; the API response indicates success either way (the failure is logged).

### 4.2 Cache Key Reshape

- The current `weather_caches` table is unique on `household_id`. This must change to support one cache row per (household, location).
- A new nullable `location_id` column references the location of interest. `location_id IS NULL` means "the household's primary location".
- The unique index becomes `(household_id, location_id)` (treating NULL as a distinct value via a partial unique index — see migration plan).
- Cache miss / coordinate-mismatch self-healing logic still applies per row.
- Existing single-row-per-household caches migrate in place: their `location_id` is set to NULL (representing the primary location).

### 4.3 Background Refresh

- The background refresh loop iterates **all** rows in `weather_caches`, not just primary locations. No code change is needed beyond the data-model reshape, since `AllProvider()` already returns every row.
- Open-Meteo rate limit (~1 req/sec) is respected by the existing throttle. With a hard cap of 10 locations per household plus the primary, even 100 households (1100 rows) refresh in ~18 minutes — well within the default 15-minute cadence at typical Home Hub scale. If the loop is still running when the next tick fires, the existing single-flight pattern (or its equivalent) prevents overlap.
- Per-row refresh errors are logged and do not halt the loop, matching current behavior.

### 4.4 REST API

New endpoints under `/api/v1/locations-of-interest`. See `api-contracts.md` for full request/response shapes.

- `GET /api/v1/locations-of-interest` — list all locations of interest for the active household.
- `POST /api/v1/locations-of-interest` — create a new location. Body: `label` (optional, max 64 chars), `placeName`, `latitude`, `longitude`.
- `PATCH /api/v1/locations-of-interest/{id}` — update the friendly label only. Coordinates and place name are immutable; to "move" a saved location, delete and recreate it.
- `DELETE /api/v1/locations-of-interest/{id}` — remove. Also deletes the cache row.

Existing weather endpoints gain an optional query parameter:

- `GET /api/v1/weather/current?locationId={uuid}` — when `locationId` is omitted, behavior is unchanged (uses the lat/lon query params, which today come from the household primary location). When provided, the server resolves the location of interest, verifies it belongs to the active household, and uses its coordinates instead of the lat/lon query params (which become optional in this case).
- `GET /api/v1/weather/forecast?locationId={uuid}` — same.

Geocoding (`GET /api/v1/weather/geocoding`) is unchanged — the frontend reuses it when the user is searching for a place to save.

### 4.5 Authorization

- Any active member of the household may create, list, update, or delete the household's locations of interest. Administrator-only restrictions do not apply.
- All endpoints require JWT auth and use the active household from the request context, matching existing weather endpoints.

### 4.6 Frontend — Weather Page

- The Weather page header gains a location selector (dropdown or segmented control) listing:
  1. The household's primary location (always first, labeled with `household.locationName` or "Primary Location" if unset).
  2. Each saved location of interest, showing the friendly label (or `place_name` fallback).
- The selector is **only rendered when the household has at least one saved location of interest**. With zero saved locations the page looks unchanged from today (no selector chrome).
- On every Weather page mount the selector resets to the primary location. Selection state lives in component state only (no URL param, no localStorage, no server persistence).
- When the user changes the selector, the page swaps to that location's current/forecast queries. The TanStack Query keys must include `locationId` so caches don't collide.
- **Empty-state behavior when household primary location is unset:**
  - The existing "No location set" empty state still renders as the default view.
  - If the household has at least one saved location of interest, a secondary line beneath the empty state reads: `"Or pick one of your saved locations:"` followed by clickable chips/links for each. Selecting one swaps the page into that location's forecast view (the selector then appears in the header so the user can switch back, even though "back" is still the unset primary empty state).
  - This preserves the "always default to primary" rule while still giving users a way through.
- A "Manage Locations" affordance on the Weather page (icon button or link in the header) opens the management UI. It is **always visible**, regardless of whether the household has any saved locations yet.
- The dashboard weather widget (`weather-widget.tsx`) is **not changed** by this feature — it continues to show the household's primary location only.

### 4.7 Frontend — Manage Locations UI

- Accessible from the Weather page header.
- Lists existing locations of interest with friendly label, place name, and a remove button.
- "Add Location" button opens a search dialog reusing the existing geocoding autocomplete component used by the household location picker.
- After picking a place, the user provides an optional friendly label (text input, max 64 chars) and confirms.
- Edit affordance on each row allows renaming the friendly label inline.
- Disables the "Add Location" button and shows a hint when the household is at the 10-location cap.

## 5. API Surface

Detailed in [`api-contracts.md`](./api-contracts.md). Summary:

| Method | Path                                          | Purpose                                  |
|--------|-----------------------------------------------|------------------------------------------|
| GET    | `/api/v1/locations-of-interest`               | List household's locations of interest   |
| POST   | `/api/v1/locations-of-interest`               | Create a new location of interest        |
| PATCH  | `/api/v1/locations-of-interest/{id}`          | Rename a location of interest            |
| DELETE | `/api/v1/locations-of-interest/{id}`          | Delete a location of interest            |
| GET    | `/api/v1/weather/current?locationId=...`      | Current weather for a saved location     |
| GET    | `/api/v1/weather/forecast?locationId=...`     | 7-day forecast for a saved location      |

All bodies follow JSON:API conventions matching existing weather-service endpoints.

## 6. Data Model

New table `weather.locations_of_interest`:

| Column      | Type             | Constraints                                |
|-------------|------------------|--------------------------------------------|
| id          | UUID             | PRIMARY KEY                                |
| tenant_id   | UUID             | NOT NULL, INDEX                            |
| household_id| UUID             | NOT NULL, INDEX                            |
| label       | TEXT             | NULL, max 64 chars enforced in domain      |
| place_name  | TEXT             | NOT NULL                                   |
| latitude    | DOUBLE PRECISION | NOT NULL                                   |
| longitude   | DOUBLE PRECISION | NOT NULL                                   |
| created_at  | TIMESTAMP        | NOT NULL                                   |
| updated_at  | TIMESTAMP        | NOT NULL                                   |

Per-household cap (10 rows) is enforced in the processor on create, not via a DB constraint.

Modifications to `weather.weather_caches`:

| Column        | Change                                                                   |
|---------------|--------------------------------------------------------------------------|
| location_id   | NEW. UUID, NULL. References `locations_of_interest.id` ON DELETE CASCADE |
| (unique idx)  | DROP `idx_weather_household` (UNIQUE on `household_id`)                  |
| (unique idx)  | ADD partial unique on `(household_id)` WHERE `location_id IS NULL`       |
| (unique idx)  | ADD partial unique on `(household_id, location_id)` WHERE `location_id IS NOT NULL` |

See [`migration-plan.md`](./migration-plan.md) for the migration sequence under GORM AutoMigrate, including how to backfill `location_id = NULL` on existing rows safely.

## 7. Service Impact

### weather-service
- New domain package `locationofinterest` with model, entity, builder, processor, REST handlers, mirroring existing patterns from `forecast`.
- Update `forecast.Model` and `forecast.Processor` to carry an optional `locationID *uuid.UUID` and key cache lookups by `(householdID, locationID)`.
- Update REST handlers for `/weather/current` and `/weather/forecast` to accept an optional `locationId` query param, resolve it via the new processor, and pass coordinates through.
- AutoMigrate adds the new table and modifies `weather_caches` (add column + index changes).
- The location-of-interest create handler invokes the forecast processor synchronously after persisting the new row to warm its cache.
- The delete handler removes the cache row in the same transaction (FK `ON DELETE CASCADE` makes this implicit; the handler should still confirm both rows are gone in tests).
- No new external service dependencies.

### account-service
- No changes. Household primary location stays where it is.

### Frontend
- New `services/api/locations-of-interest.ts` API client.
- New `lib/hooks/api/use-locations-of-interest.ts` with list/create/update/delete hooks.
- Update `use-weather.ts` to accept an optional `locationId` and include it in query keys + service calls.
- Update `WeatherPage.tsx` to render a location selector and reset to primary on each mount.
- New `ManageLocationsDialog` component (reusing the existing geocoding search component).

### Other services
- None.

## 8. Non-Functional Requirements

- **Multi-tenancy**: All new tables and endpoints scope by `tenant_id` from request context, matching existing patterns. Queries must always filter by tenant.
- **Authorization**: All endpoints require JWT and operate on the active household in context. Cross-household access is prevented by always joining `household_id = activeHousehold` when fetching by ID.
- **Performance**: Listing locations of interest is bounded at 10 rows per household — no pagination needed. Weather queries with `locationId` add at most one extra DB lookup per request to resolve the location.
- **Rate limits**: With the 10-per-household cap, the existing Open-Meteo throttle and 15-minute refresh cadence are sufficient at expected scale. No new throttling is needed.
- **Observability**: Reuse existing structured logging in the refresh loop. Per-location refresh failures log `householdID` and `locationID` (or "primary").
- **Backward compatibility**: Existing weather endpoints continue to work without `locationId`. Existing cached rows migrate to `location_id = NULL` and remain valid.

## 9. Open Questions

None remaining at spec time. Items to verify during implementation:

- Confirm that GORM AutoMigrate can handle the partial unique index changes, or whether a hand-written migration step is required for the unique index swap (see `migration-plan.md`).
- Confirm the existing geocoding-search frontend component is reusable as-is, or whether it needs a small refactor to be embeddable in the Manage Locations dialog.

## 10. Acceptance Criteria

- [ ] A household member can open the Weather page and see the household's primary-location forecast by default, with no behavior change vs. today.
- [ ] A household member can open Manage Locations, search for a city, save it with an optional friendly label, and see it appear in the Weather page selector.
- [ ] A household member can switch the Weather page selector to a saved location and see that location's current conditions and 7-day forecast.
- [ ] Switching back to the primary location shows the household primary forecast.
- [ ] Reloading the Weather page resets the selector to the primary location regardless of the prior selection.
- [ ] A household member can rename a saved location's friendly label and the change is reflected immediately.
- [ ] A household member can delete a saved location, and its cache row is also deleted.
- [ ] Attempting to create an 11th location returns a 409 with a clear message; the UI disables "Add Location" and shows a hint at the cap.
- [ ] Background refresh loop refreshes both the primary cache row and all locations-of-interest cache rows for each household.
- [ ] Creating a saved location synchronously warms its weather cache row; switching to it on the Weather page does not show a loading state on first view.
- [ ] When the household has zero saved locations, the Weather page renders no selector (matches today's chrome).
- [ ] When the primary location is unset but saved locations exist, the empty state offers clickable shortcuts into each saved location's forecast view.
- [ ] Saved coordinates are normalized to 4 decimal places on create.
- [ ] Existing weather caches survive migration with `location_id = NULL`.
- [ ] All new endpoints scope by tenant and active household; cross-household access attempts return 404.
- [ ] Any active household member (not just administrators) can perform CRUD on locations of interest.
- [ ] All affected services build and pass tests; the weather-service Docker image builds cleanly.
