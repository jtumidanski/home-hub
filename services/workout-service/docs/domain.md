# workout-service — Domain

A single-user weekly workout planner and logger. Owns the `workout.*` Postgres
schema and serves `/api/v1/workouts/...`. There is no household scope —
every row is keyed by `(tenant_id, user_id)`.

## Domain packages

- `theme` — taxonomy bucket (e.g., Muscle, Cardio). CRUD + soft delete + first-request seeding.
- `region` — body region (e.g., Chest, Legs). Same shape as `theme`. Default seed list installed on first request.
- `exercise` — the user's personal exercise catalog. Three `kind`s: `strength`, `isometric`, `cardio`. `kind` and `weightType` are immutable. Each exercise has a primary `regionId` plus optional `secondaryRegionIds` (used for filtering only — never volume math).
- `week` — one row per (user, ISO Monday). Lazy create on first mutation. `restDayFlags` is the only patchable field today.
- `planneditem` — one row per planned exercise on a given day, ordered by `position`. Single-add, bulk-add (transactional), update, delete, reorder.
- `performance` — 1:1 with planned item. Two modes: `summary` (default) and `per_set` (strength only). Switches between modes through explicit `PUT`/`DELETE .../performance/sets` calls.
- `weekview` — composite REST projection. Owns `GET/PATCH /weeks/{weekStart}`, `POST /weeks/{weekStart}/copy`, and the planned-item HTTP handlers. Lives outside `week`/`planneditem` to break what would otherwise be an import cycle between them.
- `today` — the mobile-default landing endpoint. Resolves the current week + day (UTC) and projects only that day's items, including the rest-day flag.
- `summary` — the per-week reporting projection. On-demand totals per day, per theme, and per primary region.

## Key rules

- ISO Monday normalization is the only source of truth for week starts. Clients can submit any day; the server snaps it to Monday inside `week.NormalizeToMonday`.
- Soft-deleted catalog rows (themes, regions, exercises) remain visible in historical projections via the read-through joins in `weekview` and `summary`.
- The performance state machine implements PRD §4.4.1 in two helpers: `applyExplicitStatus` for explicit requests and `deriveStatusFromActuals` for the auto path.
- Per-set guardrails:
  - summary actuals are rejected while `mode = per_set` (409 — caller must collapse first via `DELETE .../performance/sets`)
  - `weightUnit` cannot change while per-set rows exist (409)
- Volume math:
  - `strengthVolume = Σ sets × reps × weight` after unit conversion
  - bodyweight strength items contribute only to `itemCount`
  - isometric items contribute only to `itemCount`
  - region totals use the primary region only — secondary regions never contribute (prevents double-counting)
  - unit selection picks the user's most-used unit, ties broken to `lb` for weight and `mi` for distance
