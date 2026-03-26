# Package Tracking — Implementation Plan

Last Updated: 2026-03-26

---

## Executive Summary

Introduce a new `package-service` microservice and corresponding frontend views to let household members track packages across USPS, UPS, and FedEx from a single interface. The service handles carrier auto-detection, background status polling with adaptive intervals, package lifecycle management (archive/delete), and privacy controls. The frontend adds a dedicated package list page, a calendar overlay for delivery ETAs, and a dashboard summary widget.

No existing services require modification — calendar integration is frontend-only.

---

## Current State Analysis

- **Backend**: Six Go microservices (auth, account, productivity, recipe, weather, calendar) following a consistent DDD pattern with GORM entities, builder/processor/provider/resource/rest layers.
- **Frontend**: React/TypeScript with Vite, shadcn/ui, TanStack React Query, Zod validation. Pages routed via React Router; sidebar nav configured in `nav-config.ts`.
- **Infrastructure**: Docker Compose with Nginx reverse proxy, shared PostgreSQL, Go workspace (`go.work`).
- **Patterns**: All services use tenant/household scoping, JWT auth via shared `auth` package, GORM AutoMigrate, structured logging, OpenTelemetry tracing.

No package tracking capability exists today.

---

## Proposed Future State

- New `services/package-service/` following the established service code pattern.
- Three domains: `package`, `trackingevent`, `carrier` (carrier detection + client interface).
- Background workers: polling scheduler (adaptive intervals), archive/cleanup job.
- Three carrier clients implementing a common `CarrierClient` interface: USPS, UPS, FedEx.
- Frontend: new `/app/packages` route, calendar overlay, dashboard summary widget, sidebar nav entry with badge.
- Infrastructure: new container in docker-compose, nginx route `/api/v1/packages/*`, Go workspace entry, env vars for carrier API credentials.

---

## Implementation Phases

### Phase 1: Backend Foundation

Set up the service skeleton, database schema, and core CRUD operations without carrier API integration.

1. **Service scaffold** — Create `services/package-service/` with `cmd/main.go`, `internal/config/`, Dockerfile, `go.mod`. Wire into `go.work`. (Effort: M)
2. **Package domain — entity & model** — Define `Entity` (GORM) and immutable domain `Model` for the `packages` table with all columns and indexes per the data model. (Effort: M)
3. **Package domain — builder** — Fluent builder enforcing invariants (required fields, carrier enum validation, status transitions). (Effort: S)
4. **Package domain — processor & provider** — Business logic (create, update, archive, unarchive, delete, duplicate check, household limit check) and database access layer. (Effort: L)
5. **Package domain — resource & rest** — JSON:API request/response mapping and HTTP handlers for all package endpoints (CRUD, archive, unarchive, summary). (Effort: L)
6. **Tracking event domain** — entity, model, builder, provider for `tracking_events` table. Events are append-only, queried by package_id. (Effort: M)
7. **Carrier detection** — `GET /api/v1/packages/carriers/detect` endpoint with regex-based pattern matching for UPS, FedEx, USPS formats and confidence levels. (Effort: S)

### Phase 2: Carrier API Integration

Implement the carrier client interface and three carrier implementations.

8. **CarrierClient interface** — Define `CarrierClient` Go interface: `Track(ctx, trackingNumber) (TrackingResult, error)`, `Name() string`. Define `TrackingResult` with normalized status, events, ETA. (Effort: S)
9. **OAuth token management** — Token cache with automatic refresh for carrier APIs. `carrier_tokens` table or in-memory store with encrypted persistence. (Effort: M)
10. **USPS client** — Implement `CarrierClient` for USPS Tracking API v3. OAuth 2.0 client credentials flow, response normalization. (Effort: L)
11. **UPS client** — Implement `CarrierClient` for UPS Tracking API v1. OAuth 2.0 flow, response normalization, 250/day budget tracking. (Effort: L)
12. **FedEx client** — Implement `CarrierClient` for FedEx Track API v1. OAuth 2.0 flow, response normalization, 500/day budget tracking. (Effort: L)
13. **Initial poll on create** — When a package is created, immediately poll the carrier. If "not found", keep as `pre_transit` with warning. Store tracking events. (Effort: M)
14. **Manual refresh endpoint** — `POST /api/v1/packages/{id}/refresh` with 5-minute rate limit per package. (Effort: S)

### Phase 3: Background Workers

Implement the polling scheduler and cleanup jobs.

15. **Polling scheduler** — Background goroutine that queries packages needing a poll based on status and `last_polled_at`. Adaptive intervals: 30m default, 15m for `out_for_delivery`. Respects carrier rate budgets. (Effort: L)
16. **Poll execution** — For each package due for polling, call the carrier client, update package status/ETA, append tracking events, update `last_polled_at` and `last_status_change_at`. Exponential backoff on failure (max 3 retries). (Effort: M)
17. **Stale detection** — Mark packages as `stale` after 14 consecutive days with no status change. Stop polling. (Effort: S)
18. **Archive/cleanup job** — Daily background job: transition `delivered` packages to `archived` after N days, hard-delete `archived` packages after M days. (Effort: M)

### Phase 4: Infrastructure

Wire the service into the deployment stack.

19. **Docker Compose** — Add `package-service` container to `deploy/compose/docker-compose.yml` with env vars for DB, JWKS, and carrier API credentials. (Effort: S)
20. **Nginx routing** — Add `/api/v1/packages` location block proxying to `package-service:8080`. (Effort: S)
21. **Go workspace** — Add `./services/package-service` to `go.work`. (Effort: S)
22. **Environment template** — Document all required env vars (carrier credentials, polling intervals, retention periods) in service README or `.env.example`. (Effort: S)

### Phase 5: Frontend — Package List

Build the primary package management UI.

23. **TypeScript types** — Define `Package`, `TrackingEvent`, `PackageSummary`, `CarrierDetection` types in `types/models/package.ts`. (Effort: S)
24. **API service** — Create `services/api/package.ts` with methods for all package endpoints. (Effort: M)
25. **React Query hooks** — Create `lib/hooks/api/use-packages.ts` with hooks: `usePackages`, `usePackage`, `useCreatePackage`, `useUpdatePackage`, `useDeletePackage`, `useArchivePackage`, `useUnarchivePackage`, `useRefreshPackage`, `usePackageSummary`, `useDetectCarrier`. (Effort: M)
26. **Zod schemas** — Create `lib/schemas/package.schema.ts` for create/update form validation. (Effort: S)
27. **Package list page** — `/app/packages` route with sortable list of active packages, toggle for archived, status badges, carrier icons, ETA display. (Effort: L)
28. **Package card component** — Card displaying carrier icon, tracking number (truncated), label, notes, status badge, ETA, last updated, added-by user. Privacy redaction for other members' private packages. (Effort: M)
29. **Package detail/expand** — Expandable view showing full tracking event history and carrier website link. (Effort: M)
30. **Create package dialog** — Form with tracking number input (triggers carrier detect on blur/paste), carrier selector (auto-populated), label, notes, private toggle. Shows warning if carrier returns "not found". (Effort: L)
31. **Package quick actions** — Archive, delete, toggle privacy, edit label/notes actions on package cards. (Effort: M)
32. **Sidebar nav entry** — Add "Packages" to the Productivity nav group in `nav-config.ts` with badge showing `inTransitCount`. (Effort: S)
33. **Route registration** — Add `/app/packages` route to `App.tsx`. (Effort: S)

### Phase 6: Frontend — Calendar Overlay & Dashboard

Integrate packages into existing views.

34. **Calendar overlay** — On the calendar page, query packages with ETAs and render as styled all-day events (dashed border, distinct color) at the estimated delivery date. Click navigates to `/app/packages`. (Effort: L)
35. **Dashboard summary widget** — Card showing arriving-today count, in-transit count, exception count with icons. Uses `usePackageSummary` hook. (Effort: M)

### Phase 7: Testing & Documentation

36. **Backend unit tests** — Tests for carrier detection, status transitions, builder invariants, processor logic, privacy redaction. (Effort: L)
37. **Frontend component tests** — Tests for package card, create dialog, privacy redaction, calendar overlay rendering. (Effort: M)
38. **Service documentation** — Create `docs/domain.md`, `docs/rest.md`, `docs/storage.md` per the DOCS.md contract. (Effort: M)
39. **Bruno collection** — API test collection under `bruno/packages/` for manual endpoint testing. (Effort: S)

---

## Risk Assessment and Mitigation

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Carrier API rate limits exceeded | Polling stops, stale data | Medium | Per-carrier request budgets, household package limit (25), adaptive polling intervals |
| Carrier API response format changes | Broken tracking | Medium | Normalize into common model, version-pin API calls, log raw responses for debugging |
| Carrier API downtime | No status updates | Medium | Exponential backoff, retry logic, stale detection prevents infinite retries |
| OAuth token refresh failures | API calls fail | Low | Token cached with expiry buffer, refresh before expiry, fallback to re-auth |
| Tracking number format ambiguity | Wrong carrier detected | Low | Confidence levels (high/medium/low), user can override, initial poll validates |
| Large household with many packages | Rate limit pressure | Low | 25-package active limit per household, configurable |
| Privacy flag bypass | Data leak | Low | Server-side redaction in list responses, 403 on detail endpoint for non-owners |

---

## Success Metrics

- Users can add a tracking number and see status updates within 30 minutes
- Carrier auto-detection correctly identifies carrier with high confidence for standard tracking numbers
- Package ETAs appear on the calendar and update automatically when carriers revise estimates
- Dashboard summary provides accurate at-a-glance delivery overview
- Private packages are fully redacted for non-owners in all views and endpoints
- Background polling stays within carrier API rate budgets under normal household load
- Delivered packages auto-archive and eventually purge without manual intervention

---

## Required Resources and Dependencies

### External APIs (all require registration for credentials)
- **USPS Tracking API v3** — OAuth 2.0 client credentials
- **UPS Tracking API v1** — OAuth 2.0 client credentials, 250 req/day limit
- **FedEx Track API v1** — OAuth 2.0 client credentials, 500 req/day limit

### Internal Dependencies
- Shared Go packages: `auth`, `database`, `logging`, `server`, `tenant`
- Frontend: shadcn/ui components, React Query, Zod, React Router
- Infrastructure: PostgreSQL, Nginx, Docker

### No cross-service dependencies
- Calendar integration is frontend-only (no calendar-service API calls needed)
- Package-service is fully independent

---

## Timeline Estimates

| Phase | Effort |
|-------|--------|
| Phase 1: Backend Foundation | L |
| Phase 2: Carrier API Integration | XL |
| Phase 3: Background Workers | L |
| Phase 4: Infrastructure | S |
| Phase 5: Frontend — Package List | XL |
| Phase 6: Frontend — Calendar & Dashboard | L |
| Phase 7: Testing & Documentation | L |
