# Daily Tracker — Context

Last Updated: 2026-04-02

---

## Key Files

### Reference Services (for pattern replication)

| File | Purpose |
|------|---------|
| `services/category-service/` | Simple service reference (single domain, minimal dependencies) |
| `services/recipe-service/` | Complex service reference (multiple domains, relationships) |
| `services/category-service/cmd/main.go` | Service entry point pattern |
| `services/category-service/internal/config/config.go` | Configuration loading pattern |
| `services/category-service/internal/category/model.go` | Immutable domain model pattern |
| `services/category-service/internal/category/builder.go` | Fluent builder pattern |
| `services/category-service/internal/category/entity.go` | GORM entity + migration pattern |
| `services/category-service/internal/category/processor.go` | Business logic pattern |
| `services/category-service/internal/category/provider.go` | Lazy data access pattern |
| `services/category-service/internal/category/administrator.go` | Low-level DB ops pattern |
| `services/category-service/internal/category/resource.go` | HTTP handler pattern |
| `services/category-service/internal/category/rest.go` | JSON:API rest model pattern |
| `services/category-service/Dockerfile` | Dockerfile template |

### Shared Modules

| File | Purpose |
|------|---------|
| `shared/go/auth/` | JWT validation, auth middleware |
| `shared/go/database/` | GORM connection, migrations, tenant callbacks |
| `shared/go/server/` | HTTP server, handler wrappers, JSON:API helpers |
| `shared/go/tenant/` | Tenant context extraction |
| `shared/go/model/` | Provider pattern (functional composition) |
| `shared/go/logging/` | Structured logging, OpenTelemetry tracing |

### Infrastructure

| File | Purpose |
|------|---------|
| `deploy/compose/docker-compose.yml` | Docker Compose service definitions |
| `deploy/compose/nginx.conf` | Nginx reverse proxy routing |
| `deploy/k8s/ingress.yaml` | Kubernetes ingress routing |
| `deploy/k8s/category-service.yaml` | K8s deployment template |
| `go.work` | Go workspace file (add new service) |
| `.github/workflows/` | CI/CD pipeline definitions |

### Task Documentation

| File | Purpose |
|------|---------|
| `docs/tasks/task-021-daily-tracker/prd.md` | Product requirements (authoritative) |
| `docs/tasks/task-021-daily-tracker/api-contracts.md` | API request/response contracts |
| `docs/tasks/task-021-daily-tracker/ux-flow.md` | UX wireframes and navigation flow |

---

## Key Decisions

### 1. User-Scoped, Not Household-Scoped

Tracker data is personal. All queries filter on `tenant_id` + `user_id`. The `household_id` from JWT claims is not used for data access. This differs from most other services (productivity, recipe, shopping) which are household-scoped.

### 2. Four Backend Domains

| Domain | Has REST Endpoints | Has Entities |
|--------|-------------------|-------------|
| `trackingitem` | Yes (CRUD) | Yes (tracking_items table) |
| `schedule` | No (managed via trackingitem) | Yes (schedule_snapshots table) |
| `entry` | Yes (entries, skip) | Yes (tracking_entries table) |
| `month` | Yes (summary, report) | No (computed from other domains) |

The `schedule` domain is a supporting module — no direct API surface. The `month` domain is computation-only — no persistence.

### 3. Schedule Snapshots for Historical Accuracy

Schedules are versioned. When a user changes their schedule, a new snapshot is created with `effective_date = today`. Month completion and reports use the snapshot active during each day, not the current schedule. This prevents schedule changes from retroactively altering completion status of past months.

### 4. Dynamic Completion Computation

Month completion is not stored — it's recalculated on each request. This avoids stale state and simplifies the data model. Expected entries = days matching the snapshotted schedule for each active item during the month.

### 5. Soft Delete Preserves History

Deleted tracking items are soft-deleted. Their entries remain, and they appear in historical reports for months where they were active. The `active_until` date constrains which days count as expected.

### 6. Value Stored as JSONB

Entry values use JSONB with scale-type-specific shapes:
- sentiment: `{"rating": "positive"}`
- numeric: `{"count": 3}`
- range: `{"value": 72}`

This avoids polymorphic columns while keeping the schema flexible.

---

## Dependencies

### Service Dependencies

None. Tracker-service is fully self-contained — no cross-service API calls.

### Shared Module Dependencies

All standard shared modules (auth, database, server, tenant, model, logging). No modifications expected.

### Frontend Dependencies

Standard frontend stack. No new npm packages expected beyond what's already in the frontend.

---

## Potential Gotchas

1. **Tenant callback registration** — existing `RegisterTenantCallbacks` may assume household_id. Verify it supports user_id-only scoping, or filter manually in providers.

2. **Schedule snapshot lookups** — "latest effective_date <= target_date" queries need a composite index on `(tracking_item_id, effective_date)` and correct `ORDER BY effective_date DESC LIMIT 1` logic.

3. **Month boundary handling** — items created on Apr 15 only count days 15-30 as expected. Items deleted on Apr 20 only count days 1-20. Both must intersect with schedule snapshots correctly.

4. **Today endpoint day-of-week** — Go's `time.Weekday()` returns 0=Sunday through 6=Saturday, matching the schedule format in the API contract.

5. **JSONB value validation** — must validate at the processor level that the value shape matches the tracking item's scale_type before persisting.

6. **Report statistics** — standard deviation calculation must handle edge cases (0 or 1 data points).
