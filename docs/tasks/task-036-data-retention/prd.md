# Data Retention & Purging — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-11
---

## 1. Overview

Home Hub services today purge data inconsistently. Package-service has a sophisticated three-stage cleanup loop (stale → archive → hard-delete), calendar-service expires OAuth state tokens, and weather-service refreshes its cache — but every other service either soft-deletes forever (productivity, recipe, tracker, workout) or never deletes at all (account, shopping, category). Soft-deleted rows accumulate indefinitely, time-series data (tracker entries, workout performances, past calendar events) grows unbounded, and there is no story for households who want control over how long their data is kept.

This feature introduces a unified data-retention framework. Account-service becomes the source of truth for retention policies, exposing a JSON:API surface so households can configure granular retention windows per data category. Each affected service runs a periodic reaper that consults policy and applies it: hard-deleting expired soft-deleted records, aging out historical/time-series rows, and emitting audit entries so households can see what was removed. Sensible defaults ship out of the box so users get good behavior without touching settings.

The framework also closes the orphan-record risk: when a soft-deleted parent (e.g., a workout exercise) is reaped, dependent rows (performances, sets) are cascade-deleted in the same transaction so we never leave dangling foreign keys.

## 2. Goals

Primary goals:
- Centralize retention configuration in account-service with per-tenant defaults and per-household / per-user overrides where appropriate.
- Ship reapers in productivity, recipe, tracker, workout, calendar, and package services that honor the configured policy.
- Provide a small shared scheduler/reaper helper in `shared/go` so each service is not reinventing the ticker + observability boilerplate.
- Publish opinionated defaults that give a reasonable experience without configuration.
- Emit a per-run audit trail so households can see what was deleted and operators can verify reapers are doing their job.
- Cascade-delete dependent rows alongside their parents so orphans cannot accumulate.

Non-goals:
- GDPR-style "delete my entire account" cascade flow (related but tracked separately).
- Retention for short-TTL infrastructure data: auth refresh tokens, calendar OAuth state, weather cache. These already self-manage and are not user-meaningful.
- Backup / restore from cold storage. Reaped data is gone.
- Cross-service event-driven reaping. Each service polls policy on its own schedule.
- A general-purpose "settings service" — policy lives in account-service alongside existing preferences.

## 3. User Stories

- As a household admin, I want to configure how long completed tasks, old workout performances, and past calendar events are kept, so my household's data stays manageable and private.
- As a user of personal-domain features (tracker, workouts), I want to set retention for *my* data independently of household-wide retention, so I am not forced into someone else's policy.
- As a household admin, I want sensible defaults so I do not have to touch settings to get good behavior.
- As a household admin, I want to see an audit log of what was purged in the last few runs, so I can trust the system is doing the right thing and notice if something looks wrong.
- As a power user, I want to manually trigger a purge for a specific category, so I can clean things up immediately rather than waiting for the next scheduled run.
- As an operator, I want each reaper run to emit structured logs and metrics, so I can alert on failures and understand purge volumes.
- As a household admin, I want soft-deleted items (tasks, recipes) to remain restorable for a configurable window before they are permanently removed, so accidental deletes are recoverable.

## 4. Functional Requirements

### 4.1 Retention Policy Model

- A `retention_policies` resource lives in account-service. A policy is a set of category-keyed retention windows scoped to either a household (`tenant_id + household_id`) or a user (`tenant_id + user_id`).
- Categories are an enumerated, versioned list maintained in account-service. v1 categories:
  - `productivity.completed_tasks` (household)
  - `productivity.deleted_tasks_restore_window` (household)
  - `recipe.deleted_recipes_restore_window` (household)
  - `recipe.restoration_audit` (household)
  - `tracker.entries` (user)
  - `tracker.deleted_items_restore_window` (user)
  - `workout.performances` (user)
  - `workout.deleted_catalog_restore_window` (user)
  - `calendar.past_events` (household)
  - `package.archive_window` (household)
  - `package.archived_delete_window` (household)
  - `system.retention_audit` (household)
- Each window is stored as an integer number of days plus a unit hint (`days`). Minimum: 1 day. Maximum: 3650 days (~10 years). No "never" / "forever" value is permitted.
- Soft-delete restore windows have a separate minimum of 1 day and maximum of 365 days; these protect accidental deletion and longer is not useful.
- A policy resource always returns a fully-resolved view (effective values), with each category annotated by its source: `default`, `tenant`, `household`, or `user`.
- When no override is set, the system default (compiled into account-service, see §6 for table) is returned.

### 4.2 Defaults

The system ships these defaults (tunable later):

| Category | Default | Notes |
|---|---|---|
| `productivity.completed_tasks` | 365 days | Counted from completion timestamp |
| `productivity.deleted_tasks_restore_window` | 30 days | Was 3 days |
| `recipe.deleted_recipes_restore_window` | 30 days | Was 3 days |
| `recipe.restoration_audit` | 90 days | |
| `tracker.entries` | 730 days | Two years of personal history |
| `tracker.deleted_items_restore_window` | 30 days | |
| `workout.performances` | 1825 days | Five years; long-term trend visibility |
| `workout.deleted_catalog_restore_window` | 30 days | Themes / regions / exercises |
| `calendar.past_events` | 365 days | Past events only; future events untouched |
| `package.archive_window` | 7 days | Delivered → archived |
| `package.archived_delete_window` | 30 days | Archived → hard-deleted |
| `system.retention_audit` | 180 days | Audit table self-cleanup |

### 4.3 Reaper Behavior

- Each affected service runs a `retention reaper` background loop on a configurable interval (default: 6 hours, jittered ±10% to avoid thundering herd).
- On each tick, the reaper:
  1. Loads applicable policies from account-service (see §4.4 for caching).
  2. For each category it owns, runs a scoped query selecting candidate rows whose age exceeds the configured window.
  3. Deletes in batches (default: 500 rows) inside a single transaction per batch, so a long-running reaper does not hold locks indefinitely.
  4. Cascades dependent rows in the same transaction (e.g., reaping a workout exercise also reaps its `performances` and `performance_sets`).
  5. Emits one `retention_run` audit row per (service, category) per tick, summarizing scanned/deleted/duration/error.
- Reapers must be idempotent and safe to run concurrently across replicas. Use a row-level advisory lock keyed on `(tenant_id, category)` so a second replica skips a category that another replica is currently reaping.
- Reapers honor the request context for graceful shutdown.
- A reaper that fails on one tenant must continue with the next tenant; failures are logged and recorded in the audit row but do not abort the run.

### 4.4 Policy Distribution

- Account-service is the source of truth. Other services do not store policy locally except as a short-lived cache.
- Each service caches resolved policies in memory with a TTL of 5 minutes per (tenant, household-or-user) key.
- If account-service is unreachable when the reaper needs policy:
  - If a cached value exists (even if past TTL), use it.
  - Otherwise, **skip the run for that tenant** and log a warning. The reaper must never fall back to "0 days" or otherwise delete based on unknown policy.
- Policy changes take effect on the next reaper tick (within ~6 hours by default), or immediately if the user triggers a manual purge (§4.6).

### 4.5 Cascade Delete Rules

Cascade rules are explicit per service:

- **workout-service**: deleting an `exercise` cascades to `performance_sets` referencing it via performances; deleting a `region` cascades to its exercises (and their performances/sets); deleting a `theme` cascades to regions. All cascades happen in one transaction.
- **recipe-service**: deleting a `recipe` cascades to its ingredients, instructions, restoration audit rows, and meal-plan references. If a meal plan would be left empty, the meal plan itself is preserved (it can hold historical context); only the slot is cleared.
- **productivity-service**: deleting a `task` cascades to subtasks, reminders, and `task_restorations`.
- **tracker-service**: deleting a `tracking_item` cascades to `tracking_entries`. Reaping `tracking_entries` does not cascade upward (the item is preserved).
- **calendar-service**: deleting a `calendar_event` is leaf-level; no children.
- **package-service**: existing behavior — deleting a `package` cascades to its `tracking_events`.

### 4.6 Manual Purge

- Account-service exposes a manual purge endpoint that emits a fan-out request to the relevant service. The request includes `(tenant_id, scope, category)`.
- Each service exposes an internal endpoint `POST /retention/purge` that runs a single reaper pass for the given category and scope, returning a count.
- Manual purges must be authorized to a household admin (for household-scoped categories) or to the user (for user-scoped categories).
- Manual purges are rate-limited to one per (tenant, category) per 60 seconds.
- Manual purges emit a `retention_run` audit row with `trigger = 'manual'`.

### 4.7 Audit Trail

- A `retention_runs` table lives in each service that operates a reaper. Schema in §6.
- The audit trail records: service, category, tenant, scope (household/user id), trigger (`scheduled` | `manual`), scanned count, deleted count, started_at, finished_at, error (nullable).
- Account-service exposes a read-only aggregated audit view across services via a fan-out query (or, simpler v1: each service's audit is queried via JSON:API and aggregated client-side in the UI).
- The `system.retention_audit` reaper category trims `retention_runs` rows older than the configured window.

### 4.8 Settings UI

- A new "Data Retention" section appears in household settings.
- Household-scoped categories are grouped under "Household data"; user-scoped categories under "My personal data".
- Each row shows the category label, current value, source badge (`default` / `household` / `user`), and an editor (number input + days unit, bounded by min/max).
- A "View recent purges" panel shows the last 20 entries from the aggregated audit, grouped by category.
- A "Purge now" button is available per category for power users; confirmation modal required.

## 5. API Surface

All endpoints follow existing JSON:API conventions and the `/api/v1` prefix used elsewhere. Detailed contracts in `api-contracts.md`.

### Account-service (new)

- `GET /api/v1/retention-policies` — returns the resolved policy for the caller's active household and personal scope. Response includes per-category effective value and source.
- `PATCH /api/v1/retention-policies/household/:household_id` — set or clear household-scoped overrides. Body is a sparse map of `category → days | null` (null clears the override and reverts to default).
- `PATCH /api/v1/retention-policies/user` — same shape, scoped to the calling user.
- `POST /api/v1/retention-policies/purge` — manual purge trigger. Body: `{ category, scope: "household" | "user" }`. Returns 202 with a correlation id; account-service fans out to the owning service.
- `GET /api/v1/retention-runs` — paginated audit feed across services for the caller's household/user. Query params: `category`, `limit`, `cursor`.

### Each reaper-owning service (internal, not exposed via gateway)

- `POST /internal/retention/purge` — manual purge trigger. Request: `{ tenant_id, scope, category }`. Response: `{ scanned, deleted, run_id }`. Authorized via internal service token.

### Error cases

- `400` on out-of-range values (below minimum or above 3650 days).
- `403` if a non-admin user tries to write a household-scoped policy.
- `404` on unknown category.
- `429` on manual purge rate-limit hit.
- `503` if a reaper cannot reach account-service for policy and has no usable cache (the request itself succeeds; the affected category is skipped and logged).

## 6. Data Model

### 6.1 New table: `retention_policy_overrides` (account-service)

```
retention_policy_overrides
  id              uuid pk
  tenant_id       uuid not null
  scope_kind      text not null check (scope_kind in ('household','user'))
  scope_id        uuid not null   -- household_id or user_id
  category        text not null
  retention_days  int not null check (retention_days between 1 and 3650)
  created_at      timestamptz not null default now()
  updated_at      timestamptz not null default now()
  unique (tenant_id, scope_kind, scope_id, category)
  index on (tenant_id, scope_kind, scope_id)
```

System defaults are compiled into account-service code (not stored in a table) so they are versioned with the deployment.

### 6.2 New table: `retention_runs` (one per reaper-owning service)

```
retention_runs
  id            uuid pk
  tenant_id     uuid not null
  scope_kind    text not null
  scope_id      uuid not null
  category      text not null
  trigger       text not null check (trigger in ('scheduled','manual'))
  dry_run       bool not null default false
  scanned       int not null default 0
  deleted       int not null default 0
  started_at    timestamptz not null
  finished_at   timestamptz
  error         text
  index on (tenant_id, started_at desc)
  index on (category, started_at desc)
```

### 6.3 Existing tables — new behavior, no schema change

These tables already have `deleted_at` (GORM `gorm.DeletedAt`) or timestamp columns the reaper will use:

- `productivity.tasks` — reaper hard-deletes where `deleted_at < now() - restore_window` and where `completed_at < now() - completed_window` (if completed).
- `recipe.recipes`, `recipe.recipe_restorations` — same pattern.
- `tracker.tracking_items`, `tracker.tracking_entries` — entries reaped by `entry_date`.
- `workout.exercises|regions|themes`, `workout.performances`, `workout.performance_sets` — performances reaped by `performed_at`; catalog by `deleted_at`.
- `calendar.calendar_events` — reaped by `end_time < now() - past_events_window`.
- `package.packages`, `package.tracking_events` — already implemented; refactored to read policy from account-service.

### 6.4 Migration notes

- All new tables include `tenant_id` per project standards.
- Cascade deletes are implemented in application code inside a single transaction, **not** via DB-level `ON DELETE CASCADE`, so the reaper has full control and audit visibility into what was removed.
- The package-service refactor must preserve current production behavior; default values match current env-var values exactly.

See `data-model.md` for ER diagram and `migration-plan.md` for the rollout sequence.

## 7. Service Impact

- **account-service** — owns `retention_policy_overrides`, system-default constants, the public retention API (§5), the manual-purge fan-out, and authorization checks. New `retention` package.
- **productivity-service** — adds reaper for soft-deleted tasks past restore window and completed tasks past completion window. Adds `retention_runs` table. Cascade rules per §4.5.
- **recipe-service** — same pattern; reaps recipes and `recipe_restorations`.
- **tracker-service** — reaps `tracking_entries` by date; reaps soft-deleted `tracking_items` past restore window.
- **workout-service** — reaps `performances` + `performance_sets` by date; reaps soft-deleted catalog (themes/regions/exercises) with cascade.
- **calendar-service** — reaps `calendar_events` whose `end_time` is older than the configured window.
- **package-service** — refactor existing cleanup loop to load policy from account-service instead of env vars; emit `retention_runs` rows; otherwise behavior unchanged.
- **shared/go** — new `retention` helper package providing: ticker loop with jitter, advisory-lock helper, audit-row writer, account-service policy client with TTL cache, and the fallback-on-cache-miss safety logic.
- **ui** — new "Data Retention" settings page; "View recent purges" panel; "Purge now" affordance per category.

## 8. Non-Functional Requirements

### Performance
- Reapers must batch deletes (default 500 rows) and yield between batches so they cannot dominate DB connections.
- Each reaper run for a single (tenant, category) must complete in under 60 seconds at the 95th percentile under steady-state load. If a category exceeds this, batch size or run frequency is adjusted.
- Policy lookups must be cached in-process for 5 minutes; cold lookups must add no more than 50ms to a reaper tick.

### Security & Authorization
- Household-scoped policy writes require household admin role.
- User-scoped policy writes require the authenticated user to match the scope.
- Internal `/internal/retention/purge` endpoints require an internal service token; they are not exposed through the public gateway.
- Manual purge writes are rate-limited to one per (tenant, category) per 60 seconds.

### Observability
- Every reaper run emits structured logs: `service`, `category`, `tenant_id`, `scope`, `scanned`, `deleted`, `duration_ms`, `trigger`, `error`.
- Each service exposes Prometheus metrics: `retention_scanned_total`, `retention_deleted_total`, `retention_run_duration_seconds`, `retention_run_failures_total`, all labeled by `service` and `category`.
- A daily summary log per service reports total deleted across all categories.

### Multi-tenancy
- Every query scopes by `tenant_id`. The reaper iterates tenants explicitly rather than running global queries; this also lets it skip a tenant cleanly if its policy is unavailable.
- The `retention_runs` audit is tenant-scoped and not visible across tenants.

### Reliability
- Reapers must never delete based on a missing or stale "0 days" policy. The fallback-on-cache-miss logic in `shared/go/retention` is mandatory.
- Cascade deletes happen inside a single transaction per parent so partial cascades cannot leave orphans.
- Advisory locks prevent two replicas from reaping the same `(tenant, category)` simultaneously.

## 9. Resolved Decisions

- **Aggregated audit feed:** `GET /api/v1/retention-runs` fans out at request time to each reaper-owning service in v1. Revisit moving to a central event log if request latency becomes a problem.
- **Dry-run mode:** v1 supports `POST /internal/retention/purge?dry_run=true` (and a corresponding flag on the public `POST /api/v1/retention-policies/purge` endpoint). A dry run executes the same scan and cascade walk inside a transaction that is rolled back, returning the counts that *would* have been deleted. The audit row is written with `trigger = 'manual'` and a `dry_run = true` column for visibility.
- **Shrink-warning UX:** when a user edits a category to a smaller window than the current effective value, the settings UI must call the owning service for a row-count preview (reusing the dry-run path) and show a confirmation modal: "this will permanently delete approximately N rows on the next reaper run". The modal requires explicit confirmation before the `PATCH` is sent.

## 9a. Forward Note: Account-Deletion Coexistence

A future GDPR / "delete my account" or "delete my household" task will need to remove all of a tenant's data immediately rather than waiting for reapers. That task should:

- Reuse the per-service cascade implementations from §4.5 rather than reinventing them, ideally by exposing a `POST /internal/retention/purge-tenant` variant that drops the policy-driven age check and removes everything for the given scope.
- Write `retention_runs` rows with a new `trigger = 'tenant_delete'` value so the audit trail captures the event.
- Coordinate with account-service so the policy and override rows for the deleted scope are removed last.

The future task's PRD must reference this task (task-036) to keep the two flows aligned.

## 10. Acceptance Criteria

- [ ] `retention_policy_overrides` table exists in account-service with the schema in §6.1.
- [ ] System defaults from §4.2 are compiled into account-service and returned when no override exists.
- [ ] `GET /api/v1/retention-policies` returns the fully-resolved policy with per-category source annotations.
- [ ] `PATCH` endpoints for household and user scopes accept partial updates and validate min/max bounds.
- [ ] `POST /api/v1/retention-policies/purge` fans out to the correct service and returns 202 with a correlation id.
- [ ] `GET /api/v1/retention-runs` returns a paginated, tenant-scoped audit feed.
- [ ] `shared/go/retention` package exists and provides: ticker-with-jitter loop, advisory lock helper, policy client with TTL cache, audit writer, and the cache-miss safety guard.
- [ ] productivity, recipe, tracker, workout, calendar, and package services each run a reaper that consults policy and writes `retention_runs` rows.
- [ ] Cascade delete rules in §4.5 are implemented and verified by tests for each service.
- [ ] Reapers skip tenants with unavailable policy and log a warning rather than deleting.
- [ ] Reapers honor advisory locks and do not double-process across replicas (verified by integration test).
- [ ] Manual purge endpoints exist on each reaper-owning service, are rate-limited, and emit audit rows with `trigger = 'manual'`.
- [ ] Manual purge endpoints (internal and public) accept a `dry_run` flag that performs the scan and cascade walk inside a rolled-back transaction and returns the would-have-deleted counts.
- [ ] Settings UI calls the dry-run path to preview row counts before sending a `PATCH` that shrinks any retention window, and shows a confirmation modal with the previewed count.
- [ ] Prometheus metrics listed in §8 are exposed by each reaper-owning service.
- [ ] The `system.retention_audit` reaper trims old `retention_runs` rows per its configured window.
- [ ] UI "Data Retention" settings page exists, shows household and user categories, supports edits within bounds, shows source badges, and includes the recent-purges panel and per-category "Purge now" button.
- [ ] Package-service env-var-based config is replaced by account-service policy lookup with no behavioral regression for existing households.
- [ ] Documentation in `docs/architecture.md` is updated with a new section describing the retention framework.
