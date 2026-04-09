# Workout Tracker — Context

Last Updated: 2026-04-09

Companion to `plan.md`. Captures the files, decisions, and dependencies a future session needs to pick this task up cold.

---

## Source-of-truth documents (in this folder)

- `prd.md` — product requirements, user stories, acceptance criteria. **Authoritative for scope.**
- `data-model.md` — exact column types, indexes, FK behavior, cascade rules. **Authoritative for schema.**
- `api-contracts.md` — request/response shapes, status codes, error semantics. **Authoritative for API.**
- `plan.md` — phased implementation breakdown.
- `tasks.md` — checklist tracker.

When PRD/data-model/api-contracts and plan.md disagree, the PRD wins; update `plan.md` to match.

## Key project conventions to follow

- Service layout matches `services/tracker-service/internal/<domain>/{model,entity,builder,processor,provider,resource,rest}.go`. Mirror it.
- Shared libs only: `shared/{auth,database,server,tenant}`. Do not call into another service's internals.
- Partial unique indexes (`WHERE deleted_at IS NULL`) are created via `db.Exec` after `AutoMigrate`. Pattern lives in `category-service` and `tracker-service`.
- Default seeding on first request: pattern in `category-service`. Idempotent — only seed when both target tables are empty for the `(tenant, user)`.
- Time zone for "today" comes from account-service preferences. `tracker-service/internal/today` is the existing example to mirror exactly.
- JSON:API conventions per project standard; tenant + user from JWT, never from path/query.
- Per `CLAUDE.md`: planning is separate from implementation — wait for explicit approval before editing code.
- Per `CLAUDE.md`: after multi-service changes, build and test ALL affected services. Verify Docker builds when shared libs change (none expected here).

## Critical files to read before starting each phase

### Backend foundation (Phases A–F)
- `services/tracker-service/cmd/main.go` — service bootstrap pattern
- `services/tracker-service/internal/config/` — env-var + config layout
- `services/tracker-service/internal/trackingitem/` — full domain package reference (model/entity/builder/processor/provider/resource/rest)
- `services/tracker-service/internal/today/` — TZ-aware "today" endpoint reference
- `services/category-service/...` — default-seeding-on-first-request pattern, partial unique index pattern
- `shared/auth/`, `shared/tenant/`, `shared/database/`, `shared/server/` — middleware + DB + server bootstrap
- `go.work` — must add `services/workout-service` here

### Infra (Phase H)
- `deploy/compose/docker-compose.yml`
- `deploy/compose/nginx.conf`
- `deploy/k8s/tracker-service.yaml` — closest analog manifest
- `deploy/k8s/ingress.yaml`
- `.github/workflows/` — confirm where the per-service image matrix lives

### Frontend (Phase G)
- `frontend/src/App.tsx` — routing
- `frontend/src/components/` — sidebar, layout
- `frontend/src/services/` — existing API client patterns; tracker-service client is the closest analog
- `frontend/src/pages/` — existing page structure
- `frontend/src/lib/` — Zod helpers, query client, TZ utilities

### Documentation (Phase H)
- `docs/architecture.md` — add §3.12 entry
- `services/tracker-service/docs/{domain,rest,storage}.md` — shape to match for `services/workout-service/docs/`

## Architectural decisions (from PRD, repeated for quick recall)

- **Schema**: `workout.*`, owned exclusively by `workout-service`.
- **Multi-tenancy**: every row has `(tenant_id, user_id)`. No `household_id` — this feature is personal-scoped.
- **Lazy weeks**: `workout.weeks` row is created on first mutation; `GET` returns 404 when absent (frontend interprets as empty-week prompt trigger).
- **Week start**: ISO Monday. Server normalizes any day-of-week URL value to the Monday of its ISO week. Single helper, single source of truth.
- **Soft delete read-through**: planned items always join through to exercises (incl. soft-deleted) via `ON DELETE RESTRICT`. Display the *current* name with a `(deleted)` indicator. No name snapshotting.
- **Volume math**: only the primary `region_id` counts. `secondary_region_ids` exist for filtering only.
- **Bodyweight items**: contribute to `itemCount`, never to `strengthVolume` (unknown absolute load). `weight` field on a bodyweight item represents *added* weight.
- **Isometric items**: contribute to `itemCount`, never to `strengthVolume`. Use `sets × duration_seconds`.
- **Per-set mode**: strength only. Isometric/cardio reject with `422`. Mode switching has fixed semantics (max-reps × max-weight × count on collapse).
- **Performance weight unit**: lives at the performance level, applies to summary AND all per-set rows. Mixing units within a performance is illegal. Switching unit while per-set rows exist → `409`.
- **Status state machine**: PRD §4.4.1 is canonical. Server derives `partial`/`done` when client sends actuals without a status.
- **Copy-from-previous**: explicit user action only; never background. Rejects on non-empty target with `409`. "Most recent prior week" = most recent ISO week before target with at least one planned item for the same user.
- **Today TZ**: from account-service user preferences; mirror tracker-service.
- **No pagination, no full-text search, no CSV export, no charts in v1.**

## Open questions (from PRD §9, all deferred to v1.1+)

- Default unit per user (global pref).
- Confirm Monday-start weeks (locked for v1).
- Per-set mode for isometric/cardio.
- `assisted` weight type.
- Cross-week progression view.

These are explicitly out of scope. If a future session wants to pull any of them in, that needs PRD-level approval first.

## Dependencies

- **Upstream services**: account-service (user TZ preference). No other inter-service dependency.
- **Downstream consumers**: frontend only. No other service consumes `workout-service`.
- **Shared libs**: `shared/{auth,database,server,tenant}` — read-only consumers.
- **Database**: Postgres (existing instance in compose + k3s).

## Out-of-scope reminders

Do not, under any circumstances, in this task:

- Add a workout dashboard tile to productivity-service
- Add household sharing
- Add cross-week analytics, PR detection, or charts
- Add CSV export
- Add equipment metadata, body measurements, calorie tracking
- Reuse `category-service` types for themes/regions — they are independent user-owned lists with their own schema
