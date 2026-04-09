# Workout Tracker — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-09
---

## 1. Overview

Workout Tracker is a personal weekly strength-and-cardio tracker that lets a user define a regimen of exercises for each day of the week, log actual performance as they work through each session, and review weekly summaries to track progression over time. It is scoped to an individual user — there is no household sharing.

Each week is an independent plan: the user opens the current week, fills it with exercises (either fresh, or by copying from the previous week's planned values or actual logged values), and then logs sets/reps/weight as they complete each exercise. Exercises are organized by a two-level taxonomy — **theme** (e.g., muscle, cardio) and **region** (e.g., chest, back, legs) — both of which the user can manage.

The feature serves users who want a structured, lightweight place to plan and log workouts without the overhead of a full fitness app, with enough history to see week-over-week progression. v1 keeps reporting simple (per-week summary); deeper time-series analytics are deferred.

## 2. Goals

Primary goals:
- Let a user define a personal exercise catalog with theme/region categorization
- Let a user plan a weekly workout regimen (which exercises on which day-of-week, in order)
- Let a user log actual performance per exercise — at the per-exercise level by default, with the option to break down per individual set
- Let a user copy a previous week's regimen forward, choosing between planned values and actual logged values
- Let a user review per-week workout summaries
- Persist all weeks indefinitely so historical progression is available

Non-goals (v1):
- Time-series progression charts across weeks (deferred to v2)
- Personal record (PR) detection / celebration
- Supersets, circuits, drop sets, AMRAP, EMOM, RPE, tempo annotations
- Rest timers, workout timers, in-app stopwatches
- Body measurements, bodyweight tracking, calorie tracking
- Equipment metadata (barbell/dumbbell/machine/cable taxonomy)
- Assisted-machine exercises with negative-weight semantics (only `free` and `bodyweight` weight types in v1)
- Pagination on the exercise list endpoint (catalogs are expected to stay under ~200 items)
- Server-side full-text search on exercises
- CSV / data export
- Workout dashboard tile on the productivity-service summary
- Social sharing, leaderboards, public profiles
- Importing exercise libraries from external sources
- Household-shared regimens or team workouts
- Mobile-native features (camera form check, wearables sync, etc.)

## 3. User Stories

- As a user, I want to define exercises with a name, theme, region, and default sets/reps/weight (or duration/distance for cardio) so I have a personal catalog to plan from.
- As a user, I want to add and edit my own theme and region categories so I can organize exercises in a way that matches how I think about my training.
- As a user, I want to assign exercises to specific days of the current week, in the order I'll perform them, so I have a clear plan for each session.
- As a user, I want to log actual sets, reps, and weight as I perform each exercise so I have an accurate record of what I did.
- As a user, I want to break down a single exercise into per-set entries when my reps or weight differ between sets so I can track progressive overload precisely.
- As a user, I want to mark an exercise as skipped or partially done so my history reflects reality.
- As a user, when I open a new empty week, I want the app to ask me whether to copy the previous week's planned values, copy its actual logged values, or start fresh, so I have a fast baseline.
- As a user, I want to see a summary of each completed week (what I did each day, totals per region/theme) so I can review my training at a glance.
- As a user, I want to navigate forward and backward between weeks to plan ahead and review history.
- As a user, I want my old weeks to remain visible even if I later rename or delete an exercise so my history stays accurate.

## 4. Functional Requirements

### 4.1 Exercise Catalog

- A user can create, list, update, and soft-delete exercises.
- Each exercise has:
  - `name` (1–100 chars, unique per user among non-deleted exercises)
  - `themeId` (FK to a theme owned by the user)
  - `regionId` — the **primary** region (FK to a region owned by the user)
  - `secondaryRegionIds` — optional array of region UUIDs that this exercise also targets (e.g., bench press → chest primary, triceps + front delts secondary). Empty by default. The primary region MUST NOT also appear in the secondary list.
  - `kind` — enum, one of `strength`, `isometric`, `cardio`. Determines the value shape used when logging. Immutable.
  - `weightType` — enum, one of `free` or `bodyweight`. Default `free`. Immutable. Only meaningful for `strength` and `isometric` kinds.
  - For `strength`: optional defaults `defaultSets`, `defaultReps`, `defaultWeight`, `defaultWeightUnit` (`lb` | `kg`)
  - For `isometric`: optional defaults `defaultSets`, `defaultDurationSeconds`, optional `defaultWeight` + unit (for weighted planks). No reps.
  - For `cardio`: optional defaults `defaultDurationSeconds`, `defaultDistance`, `defaultDistanceUnit` (`mi` | `km` | `m`)
  - Optional `notes` (≤ 500 chars)
- `kind` and `weightType` are immutable after creation (changing them would invalidate logged data).
- For `weightType = bodyweight`, `weight` represents *added* weight only (e.g., dip belt, weighted vest). May be 0 or null. Volume math contributes `sets × reps` only — no weight-based volume — because we don't know the user's bodyweight.
- Soft delete: `deletedAt` is set; the exercise is excluded from the active catalog and from new-week planning. Historical planned items continue to resolve to it by FK and continue to render its current name. See §4.7.
- Hard delete is not supported.

### 4.2 Theme & Region Taxonomy

- Themes and regions are user-owned, user-editable lists. They are NOT hardcoded enums.
- Each user starts with a seeded default set on first access:
  - Themes: `Muscle`, `Cardio`
  - Regions: `Chest`, `Shoulders`, `Back`, `Biceps`, `Triceps`, `Core`, `Legs`, `Glutes`, `Full Body`, `Other`
- Theme: `id`, `name` (1–50 chars, unique per user), `sortOrder`
- Region: `id`, `name` (1–50 chars, unique per user), `sortOrder`
- Themes and regions can be created, renamed, reordered, and soft-deleted by the user.
- A theme or region cannot be hard-deleted while any exercise references it. Soft delete is allowed; the existing exercises retain their reference and continue to display the (frozen) name in historical contexts.

### 4.3 Weekly Plan

- Weeks are independent and identified by their **ISO week start date** (the Monday of that week, in the user's local time, stored as a `DATE`).
- The server normalizes any date passed in the URL to the Monday of its ISO week. Clients may pass any day-of-week and get back the same week resource.
- A week is created lazily — it does not exist in the database until the user first interacts with it (adds an item, copies a prior week, sets rest day flags, etc.).
- Each week has:
  - `weekStartDate` (the Monday)
  - `restDayFlags` — array of day-of-week ints (0=Mon..6=Sun) explicitly marked as rest days. A flagged day renders "Rest day" in the UI even if no items are planned. A day may legally have both the flag and planned items; no validation prevents this.
  - An ordered list of **planned exercise items**.
- Each planned item has:
  - `id` (uuid)
  - `dayOfWeek` (0=Monday … 6=Sunday) — this feature uses Monday-start weeks consistently
  - `position` (int, ≥ 0) — sort order within that day
  - `exerciseId` (FK to an exercise in the catalog; may reference a soft-deleted exercise after the fact)
  - **Planned target** (copied from the exercise's defaults at insertion time but independently editable):
    - For `strength`: `plannedSets`, `plannedReps`, `plannedWeight`, `plannedWeightUnit`
    - For `isometric`: `plannedSets`, `plannedDurationSeconds`, optional `plannedWeight` + unit
    - For `cardio`: `plannedDurationSeconds`, `plannedDistance`, `plannedDistanceUnit`
  - Optional `notes` (≤ 500 chars)
- The user can add, remove, reorder, and edit planned items on any week (past, present, or future). **Editing planned values after performance is logged is allowed.** The plan is informational, not auditable; the historical record is the performance row, not the plan. There is no lockout, no audit log.
- A planned item can appear on multiple days of a week; each day-of-week assignment is a distinct planned item row.
- Bulk add: clients can submit multiple planned items in one request (see §5.4).

### 4.4 Logging Actuals

- Each planned item has an associated **performance** that records what actually happened.
- A performance has a status: `pending` | `done` | `skipped` | `partial`.
- A performance has a single `weightUnit` (`lb` | `kg`) that applies to summary actuals AND all per-set rows. Mixing units within a single performance is not allowed.
- Performance value modes:
  - **Summary mode** (default): one row of actuals per planned item.
    - Strength: `actualSets`, `actualReps`, `actualWeight`
    - Isometric: `actualSets`, `actualDurationSeconds`, optional `actualWeight`
    - Cardio: `actualDurationSeconds`, `actualDistance`, `actualDistanceUnit`
  - **Per-set mode** (strength only): a list of N **set entries**, each with `setNumber`, `reps`, `weight`. Used when reps/weight vary across sets.
- A user can switch a strength performance from summary to per-set and back at any time.
  - Switching summary → per-set seeds N rows from `actualSets`/`actualReps`/`actualWeight` (or from planned values if no actuals are recorded yet).
  - Switching per-set → summary collapses the per-set rows into a summary row (`actualSets = count`, `actualReps = max reps`, `actualWeight = max weight`) and discards the per-set rows. Frontend warns before this happens.
- Isometric and cardio performances do not support per-set mode.
- Optional per-performance `notes` (≤ 500 chars).
- Logging actuals on a future-dated week is allowed (the user may pre-fill or correct a past day).

#### 4.4.1 Status State Machine

```
States: pending, partial, done, skipped

Transitions:
  pending  --[log actuals]-->        partial   (auto)
  pending  --[mark done]-->          done      (explicit)
  pending  --[skip]-->               skipped   (explicit)
  partial  --[mark done]-->          done      (explicit)
  partial  --[edit actuals]-->       partial   (stays)
  partial  --[skip]-->               skipped   (clears actuals)
  done     --[edit actuals]-->       done      (stays — user already committed)
  done     --[unmark done]-->        partial   (if any actuals exist)
  done     --[unmark done]-->        pending   (if no actuals exist)
  done     --[skip]-->               skipped   (clears actuals)
  skipped  --[unskip]-->             pending   (server clears skip)
```

- `pending` is the implicit/default state when no performance row exists.
- The server derives `partial` vs `done` automatically when the client only sends actuals without a status.
- "Mark day complete" is a frontend convenience that PATCHes each non-skipped item on a day to `done` in sequence. No dedicated API.

### 4.5 Copy-From-Previous Flow

- When the user navigates to a week that has zero planned items, the UI presents three choices:
  1. **Copy planned from week N−1** — copy the planned items (and their planned targets) from the most recent week that has planned items. Does not copy any performance data.
  2. **Copy actual from week N−1** — copy the planned items from the most recent week that has planned items, but use that week's *actual* values (where present) as the new week's *planned* values. A planned item with no actuals falls back to its planned values. Per-set actuals collapse to a summary row (max weight × total reps).
  3. **Start fresh** — leave the week empty.
- "The most recent prior week" is the most recent ISO week before the target week that has at least one planned item belonging to this user. Copying never crosses users.
- Copy is an **explicit user action**. There is no background job that auto-creates weeks.
- Copy is idempotent only in the sense that the UI only offers it on an empty week; the API will reject a copy request if the target week already has planned items, with `409 Conflict`.

### 4.6 Weekly Summary

- For any week (past, present, or future) the API returns a `WeekSummary` resource:
  - `weekStartDate`
  - `restDayFlags`
  - `totalPlannedItems`, `totalPerformedItems` (`done` + `partial`), `totalSkippedItems`
  - Per-day breakdown: for each day-of-week, the list of planned items with status, planned target, and actual summary, plus a `restDay` flag.
  - Per-theme totals: count of items, total strength volume (`Σ sets × reps × weight`, after unit conversion); total cardio duration and distance.
  - Per-region totals: same shape as per-theme. **Only the primary region counts** in volume math — secondary regions are ignored to prevent double-counting.
- v1 displays this in the UI as a per-week summary. There are no cross-week aggregates in v1.

### 4.7 Historical Display of Soft-Deleted Entities

- Exercises, themes, and regions that are soft-deleted remain visible in historical weeks. The UI displays their **current name** (not a snapshot) with a `(deleted)` indicator.
- Soft-deleted entities cannot be added to new planned items.
- Renaming an active exercise/theme/region updates its display in all historical views automatically — there is no name snapshotting anywhere in the schema. Joins from `planned_items` to `exercises` (and from `exercises` to `themes`/`regions`) include soft-deleted rows; the FK guarantees they always resolve.
- Trade-off: if you rename an exercise and then soft-delete it, history shows the *post-rename* name, not what it was called when you logged it. Acceptable for a personal app.

## 5. API Surface

All endpoints live under `/api/v1/workouts/...` and are served by the new `workout-service`. Authentication is JWT-based per the existing pattern; tenant and user are extracted from the JWT.

JSON:API conventions per project standard. Detailed request/response shapes are in `api-contracts.md`.

### 5.1 Themes

- `GET    /api/v1/workouts/themes` — list active themes for the user
- `POST   /api/v1/workouts/themes` — create
- `PATCH  /api/v1/workouts/themes/{id}` — update name / sortOrder
- `DELETE /api/v1/workouts/themes/{id}` — soft delete (rejected if hard-delete is requested and exercises reference it)

### 5.2 Regions

- `GET    /api/v1/workouts/regions`
- `POST   /api/v1/workouts/regions`
- `PATCH  /api/v1/workouts/regions/{id}`
- `DELETE /api/v1/workouts/regions/{id}`

### 5.3 Exercises

- `GET    /api/v1/workouts/exercises` — list active exercises for the user; supports `?themeId=` and `?regionId=` filters
- `GET    /api/v1/workouts/exercises/{id}`
- `POST   /api/v1/workouts/exercises`
- `PATCH  /api/v1/workouts/exercises/{id}` — partial update (everything except `kind`)
- `DELETE /api/v1/workouts/exercises/{id}` — soft delete

### 5.4 Weeks & Planned Items

- `GET    /api/v1/workouts/weeks/{weekStart}` — returns the week resource with embedded planned items and their performances. `weekStart` is any date in `YYYY-MM-DD`; the server normalizes to the Monday of that ISO week. Returns `404` when no week row exists; the frontend interprets this as "show the empty-week prompt."
- `PATCH  /api/v1/workouts/weeks/{weekStart}` — update week-level attributes (currently `restDayFlags`). Lazily creates the week row if absent.
- `POST   /api/v1/workouts/weeks/{weekStart}/copy` — body specifies `mode`: `planned` or `actual`. Creates the week (if absent) by copying from the most recent prior week. Returns `409` if the target week already has planned items. Returns `404` if no source week exists.
- `POST   /api/v1/workouts/weeks/{weekStart}/items` — add a new planned item to the week
- `POST   /api/v1/workouts/weeks/{weekStart}/items/bulk` — add multiple planned items in a single transaction
- `PATCH  /api/v1/workouts/weeks/{weekStart}/items/{itemId}` — update a planned item (day, position, planned target, notes)
- `DELETE /api/v1/workouts/weeks/{weekStart}/items/{itemId}` — remove a planned item (also deletes its performance and per-set rows)
- `POST   /api/v1/workouts/weeks/{weekStart}/items/reorder` — bulk reorder items within or across days; body is `[{itemId, dayOfWeek, position}, ...]`

### 5.4a Today

- `GET    /api/v1/workouts/today` — returns the items planned for the current day in the user's local time zone, with embedded performances. Mirrors `tracker-service`'s `today` endpoint. This is the primary mobile entry point.

### 5.5 Performances

- `PATCH  /api/v1/workouts/weeks/{weekStart}/items/{itemId}/performance` — update summary actuals, status, notes
- `PUT    /api/v1/workouts/weeks/{weekStart}/items/{itemId}/performance/sets` — replace per-set entries (switches the performance into per-set mode)
- `DELETE /api/v1/workouts/weeks/{weekStart}/items/{itemId}/performance/sets` — collapse per-set rows back into a summary row

### 5.6 Summary

- `GET    /api/v1/workouts/weeks/{weekStart}/summary` — computed `WeekSummary` (see §4.6)

### 5.7 Error Cases

- `400` — invalid value: negative sets, name length violation, day-of-week out of range, primary region appears in secondary list, etc.
- `404` — exercise/theme/region/week/item not found; copy source not found; week has no row yet (empty-week signal)
- `409` — copy attempted on a non-empty week; uniqueness violation on theme/region/exercise name; attempt to change `weight_unit` on a performance with existing per-set rows; attempt to set summary actuals on a performance currently in `per_set` mode
- `422` — attempt to mutate `kind` or `weightType` of an existing exercise; attempt to add a soft-deleted exercise to a planned item; per-set mode requested for an isometric/cardio item

## 6. Data Model

New schema: `workout.*`. All tables include `tenant_id` and `user_id`. No `household_id`.

Tables:

| Table                       | Purpose                                                              |
|-----------------------------|----------------------------------------------------------------------|
| `workout.themes`            | User-owned theme list                                                |
| `workout.regions`           | User-owned region list                                               |
| `workout.exercises`         | User exercise catalog                                                |
| `workout.weeks`             | Lazy week container, one row per `(user_id, week_start_date)`        |
| `workout.planned_items`     | One row per planned exercise on a given day of a given week          |
| `workout.performances`      | 1:1 with `planned_items` — summary actuals + status + mode flag      |
| `workout.performance_sets`  | Per-set rows for performances in per-set mode                        |

Detailed columns, indexes, and migration notes in `data-model.md`.

Key constraints:

- `(tenant_id, user_id, name) UNIQUE WHERE deleted_at IS NULL` on themes, regions, exercises
- `(tenant_id, user_id, week_start_date) UNIQUE` on `workout.weeks`
- `exercises.kind` ∈ `{'strength','isometric','cardio'}`, immutable
- `exercises.weight_type` ∈ `{'free','bodyweight'}`, immutable
- `exercises.region_id` MUST NOT also appear in `exercises.secondary_region_ids`
- `planned_items.day_of_week` ∈ `[0,6]`
- `planned_items.position` ≥ 0
- `performances.mode` ∈ `{'summary','per_set'}`; `per_set` only for `strength` items
- `performances.weight_unit` is the single unit for both summary and per-set values; switching it on a performance with existing per-set rows is rejected
- `performance_sets.set_number` ≥ 1, unique per performance
- Foreign keys cascade-delete from `weeks → planned_items → performances → performance_sets`. Soft-deleted exercises/themes/regions are NOT cascade-deleted; FK is `ON DELETE RESTRICT` and the row stays resolvable via join.

## 7. Service Impact

### 7.1 New: `workout-service`

Owns schema `workout.*`. Domains:
- `theme` — CRUD, soft delete, default seeding on first request per user
- `region` — CRUD, soft delete, default seeding on first request per user
- `exercise` — CRUD, soft delete, kind immutability, default value validation
- `week` — lazy week resolution, copy-from-previous orchestration
- `planneditem` — add/update/remove/reorder within a week
- `performance` — summary + per-set logging, mode switching, status derivation
- `summary` — read-only `WeekSummary` projection

Follows the standard service code pattern (`model.go`, `entity.go`, `builder.go`, `processor.go`, `provider.go`, `resource.go`, `rest.go`). Uses the shared `auth`, `database`, `server`, `tenant` modules.

### 7.2 frontend

- New "Workout" section in the sidebar (placement: under the existing Tracker entry, since both are personal-scoped)
- **Today view** — the default landing page on mobile, optimized for one-handed mid-workout use; large tap targets, sticky bottom action bar, no drag-and-drop
- Weekly view — desktop-first; back/forward navigation between weeks
- Empty-week prompt offering Copy Planned / Copy Actual / Start Fresh
- Day columns with drag-and-drop reorder of planned items (desktop)
- Exercise picker modal that filters by theme/region (and includes secondary-region matches), with client-side filtering/search
- Per-item logging UI: summary mode by default, "Track per set" affordance to switch (strength only)
- "Mark day complete" affordance that bulk-marks all non-skipped items on a day as `done`
- Rest day toggle per day on the week view
- Theme/region management screen
- Exercise catalog screen (separate from week view); supports primary + secondary region selection
- Per-week summary view

### 7.3 nginx / ingress

- Add `/api/v1/workouts -> workout-service` to the local docker-compose nginx config and the k3s ingress manifest

### 7.4 docs/architecture.md

- Add `workout-service` to the service list and routing table
- Add a §3.12 entry describing responsibilities, schema, and rules

### 7.5 docker-compose & k8s manifests

- New service entry; postgres database/schema; healthcheck endpoint
- New image: `ghcr.io/<owner>/home-hub-workout`

## 8. Non-Functional Requirements

- **Multi-tenancy**: every query MUST filter by `tenant_id` and `user_id`. No row of any workout-service table is reachable without both. Use the shared tenant middleware.
- **Authorization**: scoped to the JWT subject; no cross-user reads. There is no household ACL because there is no household scope.
- **Mobile-first logging**: the Today view is the primary mobile interaction point. It must work one-handed and tolerate mid-workout taps (sweaty hands, glanceable). The full week view is desktop-first. This drives UX choices: large tap targets, no drag-and-drop on mobile, sticky action bars.
- **Performance**: a typical week has ≤ 30 planned items. The week-fetch and Today endpoints should return in ≤ 100ms server-side at p95 with primary-key + composite-index lookups. Per-week summary is computed on demand from the same data; no caching in v1.
- **Observability**: structured Logrus logs with `request_id`, `user_id`, `tenant_id` per the project standard; OpenTelemetry trace propagation; standard `/health` and `/ready` endpoints.
- **Migrations**: GORM AutoMigrate on startup, per project standard. Initial migration creates all `workout.*` tables.
- **Soft delete**: never expose soft-deleted entities in list endpoints. Always resolvable by ID for historical display.
- **Time zones**: week start date is computed in the user's local time zone (taken from the existing user preferences) and stored as a `DATE` (no time component). The Today endpoint also uses this preference. Documented assumption: a user changing time zones may see their "current week" / "today" boundary shift; this is acceptable for v1.
- **Date normalization**: any `weekStart` value passed in a URL is normalized server-side to the Monday of its ISO week before processing. Clients may pass any day-of-week.
- **Units**: weight unit and distance unit are stored at the appropriate granularity (per-exercise default, per-planned-item planned target, per-performance actuals). A single performance has one weight unit that applies to both summary and per-set values. Unit conversion happens only inside `WeekSummary` totals.

## 9. Open Questions

- **Default unit per user.** Should the user have a global default weight unit (`lb` / `kg`) and distance unit (`mi` / `km`) configured under account preferences, so the exercise catalog defaults to it? Punting to v1.1.
- **Week start day.** Locked to Monday for v1. Confirm this matches your expectation; if you want Sunday-start to match how the rest of Home Hub renders weeks, that's a small change.
- **Per-set mode for isometric/cardio.** Disallowed in v1. Could be extended later (e.g., interval training: 5×400m repeats).
- **Assisted-machine exercises.** `weight_type` enum reserves room for an `assisted` value; deferred to v1.1 if needed.
- **Cross-week progression view.** Per-exercise time-series charts (e.g., "show Machine Chest Press weight over the last 12 weeks"). Deferred to v2.

## 10. Acceptance Criteria

- [ ] A new `workout-service` exists, builds, runs in docker-compose, and passes its own test suite
- [ ] `workout.*` schema is created via GORM AutoMigrate on first startup
- [ ] On first request, a user is auto-seeded with default themes (Muscle, Cardio) and the default region list
- [ ] A user can CRUD themes, regions, and exercises via the API; soft delete works; uniqueness is enforced per user
- [ ] Exercises support all three `kind` values (`strength`, `isometric`, `cardio`) and both `weightType` values (`free`, `bodyweight`); both fields reject mutation after creation
- [ ] Exercises support a primary region plus a `secondaryRegionIds` list; primary cannot also appear in secondary
- [ ] A user can add planned items to a week, bulk-add multiple items, edit them, reorder within and across days, and remove them
- [ ] `weekStart` URL values are normalized to the Monday of the ISO week
- [ ] `GET /weeks/{weekStart}` returns `404` when no row exists; `PATCH /weeks/{weekStart}` lazily creates it
- [ ] A user can mark days as rest days via `restDayFlags`; the summary reflects this per day
- [ ] A user can log performance in summary mode and switch to per-set mode (strength only) and back; collapse uses max-weight / max-reps / count semantics
- [ ] Performance `weight_unit` lives at the performance level; switching it on a per-set performance is rejected
- [ ] Status state machine matches §4.4.1; all transitions are exercised by tests
- [ ] Copy-from-previous works in both `planned` and `actual` modes; rejects on non-empty target; returns 404 with no source
- [ ] Week summary endpoint returns correct totals per theme and per primary region for a populated week (secondary regions are NOT counted in volume math)
- [ ] `GET /api/v1/workouts/today` returns the current day's items in the user's local time zone
- [ ] Soft-deleted exercises/themes/regions continue to render correctly in historical week views, using their current name
- [ ] Sidebar contains a new Workout entry; Today view, weekly view, exercise catalog, and theme/region management screens are reachable
- [ ] Today view is the default landing page on mobile and is usable one-handed
- [ ] Empty-week prompt offers Copy Planned / Copy Actual / Start Fresh
- [ ] All endpoints reject cross-user access (verified by integration test)
- [ ] `docs/architecture.md` and the service's own `docs/{domain,rest,storage}.md` are written
- [ ] nginx local config and k3s ingress route `/api/v1/workouts` to `workout-service`
- [ ] CI builds the new service image and publishes `ghcr.io/<owner>/home-hub-workout`
