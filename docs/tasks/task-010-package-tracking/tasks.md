# Package Tracking — Task Checklist

Last Updated: 2026-03-26

---

## Phase 1: Backend Foundation

- [ ] **1.1** Create service scaffold (`cmd/main.go`, `internal/config/`, Dockerfile, `go.mod`) — Effort: M
- [ ] **1.2** Package domain entity & model (GORM entity, immutable model, all columns/indexes) — Effort: M
- [ ] **1.3** Package domain builder (fluent builder, carrier validation, status transitions) — Effort: S
- [ ] **1.4** Package domain processor & provider (CRUD, archive, unarchive, duplicate check, household limit) — Effort: L
- [ ] **1.5** Package domain resource & REST (JSON:API mapping, all HTTP handlers) — Effort: L
- [ ] **1.6** Tracking event domain (entity, model, builder, provider — append-only events) — Effort: M
- [ ] **1.7** Carrier detection endpoint (`GET /carriers/detect`, regex matching, confidence levels) — Effort: S

**Acceptance**: Service starts, connects to DB, migrates schema, serves CRUD endpoints, detects carriers from tracking numbers.

---

## Phase 2: Carrier API Integration

- [ ] **2.1** Define `CarrierClient` interface and `TrackingResult` types — Effort: S
- [ ] **2.2** OAuth token management (cache, refresh, encrypted persistence) — Effort: M
- [ ] **2.3** USPS carrier client (Tracking API v3, response normalization) — Effort: L
- [ ] **2.4** UPS carrier client (Tracking API v1, response normalization, 250/day budget) — Effort: L
- [ ] **2.5** FedEx carrier client (Track API v1, response normalization, 500/day budget) — Effort: L
- [ ] **2.6** Initial poll on package create (call carrier, store events, handle "not found") — Effort: M
- [ ] **2.7** Manual refresh endpoint (`POST /{id}/refresh`, 5-min rate limit) — Effort: S

**Acceptance**: Creating a package triggers a carrier API call, tracking events are stored, status and ETA are populated.

---

## Phase 3: Background Workers

- [ ] **3.1** Polling scheduler (background goroutine, adaptive intervals by status, carrier budgets) — Effort: L
- [ ] **3.2** Poll execution (update status/ETA, append events, exponential backoff on failure) — Effort: M
- [ ] **3.3** Stale detection (14 days no status change → mark stale, stop polling) — Effort: S
- [ ] **3.4** Archive/cleanup job (daily: delivered → archived after N days, delete archived after M days) — Effort: M

**Acceptance**: Packages are automatically polled, status updates flow in, stale packages stop polling, delivered packages auto-archive and eventually purge.

---

## Phase 4: Infrastructure

- [ ] **4.1** Add `package-service` to `docker-compose.yml` with env vars — Effort: S
- [ ] **4.2** Add `/api/v1/packages` nginx route — Effort: S
- [ ] **4.3** Add `./services/package-service` to `go.work` — Effort: S
- [ ] **4.4** Document env vars in service README — Effort: S

**Acceptance**: Service runs in Docker Compose, API is accessible through nginx at `/api/v1/packages`.

---

## Phase 5: Frontend — Package List

- [ ] **5.1** TypeScript types (`Package`, `TrackingEvent`, `PackageSummary`, `CarrierDetection`) — Effort: S
- [ ] **5.2** API service class (`services/api/package.ts`) — Effort: M
- [ ] **5.3** React Query hooks (`use-packages.ts` — all CRUD + summary + detect) — Effort: M
- [ ] **5.4** Zod validation schemas for create/update forms — Effort: S
- [ ] **5.5** Package list page (`/app/packages`, sorted by ETA, archived toggle) — Effort: L
- [ ] **5.6** Package card component (carrier icon, status badge, ETA, privacy redaction) — Effort: M
- [ ] **5.7** Package detail/expand (tracking event history, carrier website link) — Effort: M
- [ ] **5.8** Create package dialog (tracking input, carrier auto-detect, label, notes, privacy) — Effort: L
- [ ] **5.9** Package quick actions (archive, delete, toggle privacy, edit label/notes) — Effort: M
- [ ] **5.10** Add "Packages" to sidebar nav with in-transit badge — Effort: S
- [ ] **5.11** Register `/app/packages` route in `App.tsx` — Effort: S

**Acceptance**: Users can add, view, edit, archive, and delete packages from the UI. Private packages are redacted. Carrier is auto-detected on input.

---

## Phase 6: Frontend — Calendar & Dashboard

- [ ] **6.1** Calendar overlay (packages with ETAs as styled all-day events, distinct visual) — Effort: L
- [ ] **6.2** Dashboard summary widget (arriving today, in transit, exceptions) — Effort: M

**Acceptance**: Package ETAs appear on the calendar with distinct styling. Dashboard shows delivery counts.

---

## Phase 7: Testing & Documentation

- [ ] **7.1** Backend unit tests (carrier detection, status transitions, builder, processor, privacy) — Effort: L
- [ ] **7.2** Frontend component tests (card, dialog, privacy, calendar overlay) — Effort: M
- [ ] **7.3** Service documentation (`docs/domain.md`, `docs/rest.md`, `docs/storage.md`) — Effort: M
- [ ] **7.4** Bruno API collection (`bruno/packages/`) — Effort: S

**Acceptance**: Tests pass, documentation follows DOCS.md contract, Bruno collection covers all endpoints.

---

## Cross-Cutting Concerns

- [ ] Update `docs/architecture.md` with package-service routing and description
- [ ] Verify Docker build for package-service
- [ ] Verify all existing services still build after `go.work` change
- [ ] End-to-end smoke test: add package → see on list → see on calendar → see on dashboard
