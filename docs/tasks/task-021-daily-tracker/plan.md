# Daily Tracker — Implementation Plan

Last Updated: 2026-04-02

---

## Executive Summary

Build a new `tracker-service` Go microservice and supporting frontend pages to deliver a personal daily habit/wellness tracking feature. Users define custom tracking items with flexible rating scales (sentiment, numeric, range) and per-day-of-week schedules, then log entries day-by-day on a monthly calendar grid. Completed months display an aggregate dashboard report.

The service is fully self-contained with no cross-service dependencies. Data is user-scoped (tenant_id + user_id), not household-scoped. The implementation follows established Home Hub patterns: DDD domain layers, GORM entities, JSON:API transport, shared module integration.

---

## Current State Analysis

- **No existing tracker infrastructure** — this is a greenfield service
- **Established service pattern** — 9 existing Go services provide a clear template (category-service for simple reference, recipe-service for complex reference)
- **Shared modules ready** — auth, database, server, tenant, model, logging modules all available
- **Frontend framework ready** — React + Vite + shadcn/ui + TanStack Query + react-hook-form + Zod
- **Infrastructure patterns established** — Docker Compose, nginx, k8s manifests, CI/CD all have consistent patterns to replicate

### Key Differentiator: User-Scoped (Not Household-Scoped)

Most existing services are household-scoped. Tracker data is personal — scoped by `tenant_id` + `user_id` only. This means:
- No `household_id` in entities or queries
- Tenant callbacks need to filter on user_id instead of household_id
- Auth middleware still provides tenant context, but household_id is ignored for data access

---

## Proposed Future State

```
services/tracker-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── trackingitem/
│   │   ├── model.go
│   │   ├── builder.go
│   │   ├── entity.go
│   │   ├── processor.go
│   │   ├── provider.go
│   │   ├── administrator.go
│   │   ├── resource.go
│   │   └── rest.go
│   ├── schedule/
│   │   ├── model.go
│   │   ├── builder.go
│   │   ├── entity.go
│   │   ├── provider.go
│   │   └── administrator.go
│   ├── entry/
│   │   ├── model.go
│   │   ├── builder.go
│   │   ├── entity.go
│   │   ├── processor.go
│   │   ├── provider.go
│   │   ├── administrator.go
│   │   ├── resource.go
│   │   └── rest.go
│   └── month/
│       ├── model.go
│       ├── processor.go
│       ├── resource.go
│       └── rest.go
├── docs/
│   ├── domain.md
│   ├── rest.md
│   └── storage.md
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

### Domain Breakdown

| Domain | Responsibility |
|--------|---------------|
| `trackingitem` | CRUD for tracking items, soft delete, name uniqueness |
| `schedule` | Schedule snapshot management, effective date lookups |
| `entry` | Entry logging, skip management, value validation |
| `month` | Month summary computation, completion calculation, report generation |

The `schedule` domain is a supporting domain without its own REST endpoints — snapshots are managed as a side-effect of tracking item creation/update. The `month` domain has no entities — it computes results from trackingitem, schedule, and entry data.

---

## Implementation Phases

### Phase 1: Service Scaffold & Core Domain (Backend)

Set up the service skeleton and implement the tracking item CRUD with schedule snapshots. This phase delivers the foundational data model and basic API endpoints.

**Goal:** `POST/GET/PATCH/DELETE /api/v1/trackers` working end-to-end with schedule snapshot versioning.

### Phase 2: Entry Logging (Backend)

Implement entry creation, updating, skipping, and deletion. Includes value validation against scale types.

**Goal:** `PUT/DELETE /api/v1/trackers/{id}/entries/{date}` and skip endpoints working. `GET /api/v1/trackers/entries?month=` returning entries.

### Phase 3: Month Computation & Reports (Backend)

Implement month summary with dynamic completion calculation and monthly dashboard report generation.

**Goal:** `GET /api/v1/trackers/months/{YYYY-MM}` and `/report` endpoints working with correct completion logic, schedule snapshot awareness, and aggregate statistics.

### Phase 4: Today Quick-Entry (Backend)

Implement the today endpoint that returns only items scheduled for today with their current entries.

**Goal:** `GET /api/v1/trackers/today` working correctly with schedule-aware filtering.

### Phase 5: Infrastructure & Deployment

Docker, docker-compose, nginx routing, k8s manifests, CI/CD pipeline, go.work integration.

**Goal:** Service builds, deploys, and routes correctly in both local and k8s environments.

### Phase 6: Frontend — Tracking Item Management

Sidebar entry, item setup page with CRUD, color palette picker, schedule day toggles.

**Goal:** Users can create, edit, reorder, and delete tracking items from the UI.

### Phase 7: Frontend — Monthly Calendar Grid

Calendar grid view with month navigation, cell rendering by scale type, inline entry editor.

**Goal:** Users can view the monthly grid and log/edit/skip entries by clicking cells.

### Phase 8: Frontend — Today Quick-Entry View

Quick-entry page showing today's scheduled items with inline editing.

**Goal:** Users can log today's entries quickly from a focused view.

### Phase 9: Frontend — Monthly Dashboard Report

Dashboard view for completed months with aggregate stats and visualizations.

**Goal:** Completed months display report with per-item statistics and mini charts.

### Phase 10: Testing, Documentation & Polish

Integration tests, service documentation, API contract verification, edge case handling.

**Goal:** Comprehensive test coverage, complete service docs, all acceptance criteria verified.

---

## Risk Assessment and Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Schedule snapshot complexity — off-by-one errors in effective date lookups | High | Medium | Table-driven tests covering edge cases: mid-month schedule changes, items created/deleted mid-month, month boundaries |
| Completion calculation correctness with mixed schedule snapshots | Medium | High | Unit test every combination: items created mid-month, deleted mid-month, schedule changed mid-month, all entries filled, partial fills, all skipped |
| User-scoped vs household-scoped tenant callbacks | Medium | Medium | Verify tenant callback filtering works with user_id; may need custom callback registration if shared module assumes household_id |
| Report statistics accuracy (averages, std dev, ratios) | Low | Medium | Golden-file tests with known inputs and expected outputs |
| Calendar grid performance with many items x 31 days | Low | Low | Month summary endpoint returns all data in one call; frontend renders from cached data |

---

## Success Metrics

- All 20 acceptance criteria from the PRD pass
- Tracker-service builds and passes CI
- Month summary endpoint returns in < 200ms for 20 tracking items
- All data correctly scoped by tenant_id + user_id (no cross-user access)
- Service integrates cleanly into existing infrastructure (compose, nginx, k8s)

---

## Required Resources and Dependencies

### Shared Modules (existing, no changes expected)

- `shared/go/auth` — JWT validation, auth middleware
- `shared/go/database` — GORM connection, migrations, tenant callbacks
- `shared/go/server` — HTTP server, handler wrappers, JSON:API helpers
- `shared/go/tenant` — Tenant context extraction
- `shared/go/model` — Provider pattern
- `shared/go/logging` — Structured logging, tracing

### External Dependencies

- PostgreSQL — new `tracker` schema
- No external API calls (self-contained service)

### Frontend Dependencies (existing)

- React, Vite, shadcn/ui, TanStack Query, react-hook-form, Zod, Tailwind CSS

---

## Timeline Estimates

| Phase | Effort | Dependencies |
|-------|--------|-------------|
| Phase 1: Service Scaffold & Tracking Items | L | None |
| Phase 2: Entry Logging | M | Phase 1 |
| Phase 3: Month Computation & Reports | L | Phase 1, 2 |
| Phase 4: Today Quick-Entry | S | Phase 1, 2 |
| Phase 5: Infrastructure | M | Phase 1 |
| Phase 6: Frontend — Item Management | M | Phase 1 |
| Phase 7: Frontend — Calendar Grid | XL | Phase 2, 3 |
| Phase 8: Frontend — Today Quick-Entry | M | Phase 4 |
| Phase 9: Frontend — Dashboard Report | L | Phase 3 |
| Phase 10: Testing & Documentation | M | All phases |

**Parallelization:** Phases 5-6 can start as soon as Phase 1 is complete. Phase 4 can run alongside Phase 3. Frontend phases 7-9 can be parallelized if API contracts are stable.
