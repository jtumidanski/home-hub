# Plan Audit — task-026-locations-of-interest

**Plan Path:** docs/tasks/task-026-locations-of-interest/tasks.md
**Audit Date:** 2026-04-09
**Branch:** task-026
**Base Branch:** main

## Executive Summary

The implementation lands the full feature: a new `locationofinterest` Go domain in
weather-service, the `weather_caches` reshape with partial unique indexes + FK,
optional `locationId` on the existing weather endpoints, and a complete frontend
flow (selector, manage dialog, empty-state shortcuts). Backend `go build` and
`go test ./...` pass; frontend `tsc --noEmit` and Vitest (404 tests) pass.

**Audit follow-up (2026-04-09):** the missing unit-test coverage flagged in the
original audit has been added. New tests live in
`locationofinterest/processor_test.go`, `forecast/provider_test.go`, and
`refresh/refresh_test.go`, exercising cap enforcement (with a verbatim
`ErrCapReached` message assertion), label trim/length, 4-decimal coordinate
normalization, cross-household isolation, NULL vs non-NULL provider branches,
mixed-row refresh, and per-row error-log location fields. The only outstanding
item is **P4.T3** — the live-stack acceptance walkthrough — which remains
deferred in `tasks.md`.

## Task Completion

| #     | Task                                                         | Status   | Evidence / Notes |
|-------|--------------------------------------------------------------|----------|------------------|
| P1.T1 | Create `locationofinterest` package skeleton                 | DONE     | `services/weather-service/internal/locationofinterest/{model,builder,entity,provider,administrator,processor,resource,rest}.go`. Package mirrors `forecast/`. Builder enforces label trim + 64-char cap. |
| P1.T2 | Wire `locationofinterest.Migration` into main.go             | DONE     | `cmd/main.go:31-35` registers `locationofinterest.Migration` BEFORE `forecast.Migration`. |
| P1.T3 | Reshape `forecast.Entity` for `LocationId`                   | DONE     | Entity field added (`forecast/entity.go:39`), Builder/Make/ToEntity threaded, `getByHouseholdAndLocation` provider added (`forecast/provider.go:9-17`), legacy unique tag dropped. Audit follow-up: `forecast/provider_test.go` now exercises the new provider against sqlite. |
| P1.T4 | Hand-written index/FK migration step                         | DONE     | `forecast/entity.go:63-87` implements idempotent `PostMigration` with the three indexes and FK guard. Wired via `Migration` calling `PostMigration` after AutoMigrate. |
| P1.T5 | `forecast.Processor` accepts optional locationID             | DONE     | Signatures threaded through `getOrFetch`, `fetchAndCache`, `RefreshCache`, `create` (`forecast/processor.go`, `administrator.go:10-70`). Audit follow-up: `forecast/provider_test.go` covers primary (NULL) and saved (non-nil) lookups; `refresh/refresh_test.go` covers RefreshCache against mixed rows. |
| P2.T1 | `locationofinterest.Processor` CRUD                          | DONE     | `processor.go` implements List/Get/Create/UpdateLabel/Delete with 4-decimal normalization, 10-cap → `ErrCapReached`, label trim, cache warm via `CacheWarmer` interface. Audit follow-up: `locationofinterest/processor_test.go` covers cap, label trim/length, normalization, missing-id, cross-household isolation, per-household cap independence, and warmer error swallowing. |
| P2.T2 | REST handlers + InitializeRoutes                             | PARTIAL  | All four endpoints registered (`resource.go:16-27`); 409 message string matches verbatim (`processor.go:16`); 404 on cross-household via `Get`→`ErrNotFound`. Verbatim message asserted in `processor_test.go` `TestCreate_CapAtTenReturnsErrCapReached`. **Still missing:** REST-layer handler test that exercises full HTTP request/response shapes — domain coverage is in place but the HTTP wrapper itself relies on integration. |
| P2.T3 | Weather endpoints accept `locationId`                        | PARTIAL  | `forecast/resource.go:31-68` `resolveLocation` parses `locationId`, calls `locationofinterest.Processor.Get` (404 on missing), uses stored coords, ignores `latitude`/`longitude` params, threads `locationID` into processor. **Still missing:** REST-layer test that exercises `resolveLocation` end-to-end with both branches — provider/domain branches are covered by `forecast/provider_test.go` and `locationofinterest/processor_test.go`. |
| P2.T4 | Refresh loop verification + per-row log                      | DONE     | `refresh/refresh.go:40-50` logs `location_id` field ("primary" sentinel for nil). Audit follow-up: `refresh/refresh_test.go` exercises mixed nil/non-nil rows refreshed via httptest server and asserts the per-row error log includes `location_id` for both branches. |
| P3.T1 | API client + TS types                                        | DONE     | `frontend/src/services/api/locations-of-interest.ts`, `types/models/location-of-interest.ts`. |
| P3.T2 | React Query hooks + query keys                               | DONE     | `lib/hooks/api/use-locations-of-interest.ts`; query keys added in `query-keys.ts`; mutations invalidate list and the affected weather query keys. |
| P3.T3 | Extend `useCurrentWeather` / `useWeatherForecast`            | DONE     | `lib/hooks/api/use-weather.ts:26-66`; query keys include locationId or `"primary"` sentinel; primary callers compile unchanged. |
| P3.T4 | WeatherPage selector + state                                 | DONE     | `pages/WeatherPage.tsx:93-157`; `selectedLocationId` defaults to `undefined` on every mount; selector hidden when `savedLocations.length === 0`; "Manage Locations" affordance always visible. |
| P3.T5 | `ManageLocationsDialog`                                      | DONE     | `components/features/weather/manage-locations-dialog.tsx`; reuses `LocationSearch`, supports inline rename, label cap 64, disables Add at 10-cap with hint. |
| P3.T6 | Empty-state shortcuts                                        | DONE     | `WeatherPage.tsx:173-191` renders chips when `!locationSet && savedLocations.length > 0`. |
| P4.T1 | Backend tests + Docker build                                 | DONE     | `go build ./...` ✅ `go test ./...` ✅ — now includes new tests in `locationofinterest`, `forecast/provider_test.go`, and `refresh`. Docker build not re-run during this audit but is checked off in tasks; no unverified shared-lib changes detected. |
| P4.T2 | Frontend tsc + tests                                         | DONE     | `pnpm tsc --noEmit` ✅, `pnpm test` → 44 files / 404 tests ✅. |
| P4.T3 | PRD §10 acceptance walkthrough vs running stack              | DEFERRED | Tasks file explicitly notes "_(Pending live-stack smoke walkthrough; code-level verification done.)_". |

**Completion Rate:** 15 DONE / 2 PARTIAL / 0 SKIPPED / 1 DEFERRED = 18 tasks total. Strict completion (DONE only): 15/18 ≈ 83%. Functional completion (DONE + PARTIAL): 17/18 ≈ 94%.
**Skipped without approval:** 0.
**Partial implementations:** 2 (P2.T2 and P2.T3 — both lack REST-layer HTTP wrapper tests; underlying domain/provider behavior is covered by the new processor and provider tests).

## Skipped / Deferred Tasks

- **P2.T2 / P2.T3 REST-layer HTTP tests** — handler-level tests that drive the JSON:API request/response shapes and exercise HTTP status codes for the new endpoints and the modified `/weather/current|forecast` `locationId` branch are still absent. Underlying coverage now exists at the provider and processor layers (`locationofinterest/processor_test.go`, `forecast/provider_test.go`), so the residual risk is the JSON:API marshalling glue and the `resolveLocation` query-param parser. Recommend adding `httptest.NewServer`-style handler tests in a follow-up if regression risk on the wrapper grows.
- **P4.T3 live-stack acceptance walkthrough** — explicitly deferred in `tasks.md`. PRD §10 boxes are checked but the file annotation says they reflect code review, not a running-stack pass.

## Developer Guidelines Compliance

### Passes

- **Immutable models with accessors.** `locationofinterest/model.go` has all-private fields and read-only accessors. `forecast/model.go` continues the pattern with the new `locationID` getter.
- **Builder pattern with invariant enforcement.** `locationofinterest/builder.go:44-82` validates required IDs, place name, label length, and lat/lon ranges before constructing the Model.
- **Entity separation.** GORM tags live only on `Entity`, not on `Model`. `Make`/`ToEntity` mediate.
- **Provider pattern.** `locationofinterest/provider.go` uses `database.Query`/`database.SliceQuery` from the shared library; `forecast/provider.go:9-17` uses the same composition for the NULL-safe lookup.
- **REST resource/handler separation.** `rest.go` holds the JSON:API DTOs and Transform helpers; `resource.go` holds `InitializeRoutes` + handlers — matches the established convention used by `forecast/`.
- **Multi-tenancy context.** Every handler reads `tenantctx.MustFromContext(r.Context())` and scopes by `t.HouseholdId()` (and `t.Id()` on writes). No direct DB access in handlers — they instantiate a Processor.
- **Functional composition for DB access.** `getByHouseholdAndLocation` and `ListByHousehold` return providers; the processor wraps them with `model.Map`/`model.SliceMap`.
- **Pure transform helpers.** `Transform`/`TransformSlice` in both packages are side-effect free.
- **Idempotent migrations.** `forecast/PostMigration` uses `IF NOT EXISTS` and a `pg_constraint` existence guard.
- **Frontend multi-tenancy.** `useLocationsOfInterest` reads `tenant`/`household` from `useTenant()` and gates `enabled` on their presence.
- **Frontend query-key hygiene.** `weatherKeys.current/forecast` always include the locationId (or `"primary"` sentinel) so caches don't collide. Mutations in `use-locations-of-interest.ts` invalidate both the list and the affected weather keys.
- **Cycle avoidance.** `locationofinterest.Processor` defines a `CacheWarmer` interface (`processor.go:24-26`) so it does not import `forecast`; `forecast.Processor.WarmLocationCache` satisfies the interface and is injected from `cmd/main.go:46-48`. Clean inversion.

### Violations

- ~~**Rule:** Plan acceptance criteria require unit tests for new behavior.~~
  **Status:** RESOLVED 2026-04-09. New test files added: `locationofinterest/processor_test.go`, `forecast/provider_test.go`, `refresh/refresh_test.go`. `ErrCapReached.Error()` is asserted verbatim. `gorm.io/driver/sqlite` was added to weather-service `go.mod` and `openmeteo.NewClientWithEndpoints` was introduced as a cross-package test seam.

- **Rule:** No exported direct-`*gorm.DB` helpers in domain packages — prefer providers / composed queries.
  **File:** `services/weather-service/internal/locationofinterest/administrator.go:10-25`, `provider.go:21-25`.
  **Issue:** `createLocation`, `updateLocation`, `deleteLocation`, and `countByHousehold` are package-private but take a raw `*gorm.DB` rather than returning composed providers like the rest of the package. Same pattern as the existing `forecast/administrator.go:create` function, so this is consistent with prior code — but it is the one place the domain leans on imperative DB calls instead of the functional provider style. Flagging as **low** since it follows the in-repo precedent and is package-private.
  **Severity:** low
  **Fix:** Optional. If you want strict adherence, refactor `createLocation`/`updateLocation`/`deleteLocation`/`countByHousehold` into administrator/provider compositions returning `database.Operator[Entity]`-style functions. Defer if the team treats `forecast/administrator.go` as the canonical pattern.

- **Rule:** Domain `Processor.UpdateLabel` distinguishes "label not provided" from "label cleared" via tri-state pointer.
  **File:** `services/weather-service/internal/locationofinterest/processor.go:131-153`.
  **Issue:** `nil` is treated as no-op, while empty string clears the label. The REST `UpdateRequest.Label *string` will deserialize `{"label": null}` and `{}` both as `nil`, so a client cannot explicitly clear the label via JSON `null`. PRD/api-contracts allow rename, but if "clear label" is intended as a UX path, the dialog has to send `""` (it does — `manage-locations-dialog.tsx:65`), so functionally it works. Documenting as **low** because the UI path covers the only intended flow.
  **Severity:** low
  **Fix:** None required unless API consumers outside the UI need explicit-null semantics. If so, switch `UpdateRequest.Label` to `*sharedjson.NullableString` or document the empty-string convention in `api-contracts.md`.

- ~~**Rule:** Avoid wrapping plain functions in package-level `var` declarations whose values are user-facing strings without an `errors.New` indirection.~~
  **Status:** RESOLVED 2026-04-09. Added contract comment above the `var (` block in `locationofinterest/processor.go`.

No backend anti-patterns found (no direct DB in handlers, no mutable models, no cross-domain internal calls beyond the explicit `CacheWarmer` injection seam). No frontend anti-patterns found (no fetch in components, no query-key duplication, all mutations route through hooks).

## Build & Test Results

| Service / Project       | Build                | Tests                                  | Notes |
|-------------------------|----------------------|----------------------------------------|-------|
| weather-service (Go)    | PASS (`go build ./...`) | PASS (`go test ./... -count=1`)     | All packages pass, including new `locationofinterest/processor_test.go`, `forecast/provider_test.go`, and `refresh/refresh_test.go` added during the audit follow-up. |
| frontend (TypeScript)   | PASS (`pnpm tsc --noEmit`) | PASS (`pnpm test` → 44 files / 404 tests) | No new vitest files added for the new hooks/components, but type-check is clean. |
| weather-service (Docker)| NOT RE-RUN           | —                                      | Plan task P4.T1 is checked; not re-verified during this read-only audit. Recommend re-running before merge. |

## Overall Assessment

- **Plan Adherence:** MOSTLY_COMPLETE — every functional acceptance bullet is satisfied. Test coverage gaps from the original audit have been closed at the domain/provider/refresh layer; only REST-handler-level wrappers and the live-stack walkthrough remain.
- **Guidelines Compliance:** COMPLIANT — only low-severity stylistic nits remain, all consistent with existing in-repo patterns.
- **Recommendation:** READY_TO_MERGE — pending the optional `docker build` re-run and the deferred P4.T3 walkthrough.

## Action Items

1. ~~Add `locationofinterest/processor_test.go`~~ — DONE (`internal/locationofinterest/processor_test.go`).
2. ~~Add `forecast` test for the NULL-safe provider branches~~ — DONE (`internal/forecast/provider_test.go`).
3. ~~Add `refresh/refresh_test.go`~~ — DONE (`internal/refresh/refresh_test.go`).
4. ~~Document `ErrCapReached` as a public-API contract message~~ — DONE.
5. **Run weather-service `docker build`** once before merge per CLAUDE.md ("Always verify Docker builds when changing shared libraries"). The new `gorm.io/driver/sqlite` dep is test-only and should not affect the production image, but verify.
6. **Complete P4.T3** — perform the PRD §10 walkthrough against a running stack and tick the boxes off the in-tree checklist.
7. **(Optional follow-up)** REST-handler tests for `locationofinterest` endpoints and the `/weather/current|forecast` `locationId` branch, if regression risk on the JSON:API wrapper grows.
