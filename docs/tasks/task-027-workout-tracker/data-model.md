# Workout Tracker — Data Model

Schema: `workout.*`. Owned by `workout-service`. All tables are scoped by `(tenant_id, user_id)` — no household scope.

GORM `AutoMigrate` runs on startup. Forward-only migrations.

---

## 1. `workout.themes`

| Column      | Type        | Notes                                                |
|-------------|-------------|------------------------------------------------------|
| id          | uuid        | PK                                                   |
| tenant_id   | uuid        | not null                                             |
| user_id     | uuid        | not null                                             |
| name        | varchar(50) | not null                                             |
| sort_order  | int         | not null, default 0, ≥ 0                             |
| created_at  | timestamptz | not null                                             |
| updated_at  | timestamptz | not null                                             |
| deleted_at  | timestamptz | nullable                                             |

Indexes:
- `UNIQUE (tenant_id, user_id, name) WHERE deleted_at IS NULL`
- `INDEX (tenant_id, user_id, deleted_at)`

Default seed on first request per `(tenant_id, user_id)`: `Muscle (sort_order=0)`, `Cardio (sort_order=1)`.

---

## 2. `workout.regions`

Same shape as `workout.themes`.

Default seed on first request per `(tenant_id, user_id)`:
`Chest, Shoulders, Back, Biceps, Triceps, Core, Legs, Glutes, Full Body, Other` (sort_order 0..9).

---

## 3. `workout.exercises`

| Column                   | Type         | Notes                                                                 |
|--------------------------|--------------|-----------------------------------------------------------------------|
| id                       | uuid         | PK                                                                    |
| tenant_id                | uuid         | not null                                                              |
| user_id                  | uuid         | not null                                                              |
| name                     | varchar(100) | not null                                                              |
| kind                     | varchar(16)  | not null, in `('strength','isometric','cardio')`, immutable           |
| weight_type              | varchar(16)  | not null, in `('free','bodyweight')`, default `'free'`, immutable     |
| theme_id                 | uuid         | not null, FK → `workout.themes(id)`                                   |
| region_id                | uuid         | not null, FK → `workout.regions(id)` — primary region                 |
| secondary_region_ids     | jsonb        | not null, default `'[]'` — array of region UUIDs                      |
| default_sets             | int          | nullable, ≥ 0; meaningful for `strength`/`isometric`                  |
| default_reps             | int          | nullable, ≥ 0; meaningful for `strength`                              |
| default_weight           | numeric(7,2) | nullable, ≥ 0; meaningful for `strength`/`isometric`                  |
| default_weight_unit      | varchar(4)   | nullable, in `('lb','kg')`; meaningful for `strength`/`isometric`     |
| default_duration_seconds | int          | nullable, ≥ 0; meaningful for `isometric`/`cardio`                    |
| default_distance         | numeric(8,3) | nullable, ≥ 0; meaningful for `cardio`                                |
| default_distance_unit    | varchar(4)   | nullable, in `('mi','km','m')`; meaningful for `cardio`               |
| notes                    | varchar(500) | nullable                                                              |
| created_at               | timestamptz  | not null                                                              |
| updated_at               | timestamptz  | not null                                                              |
| deleted_at               | timestamptz  | nullable                                                              |

Indexes:
- `UNIQUE (tenant_id, user_id, name) WHERE deleted_at IS NULL`
- `INDEX (tenant_id, user_id, theme_id)`
- `INDEX (tenant_id, user_id, region_id)`
- `INDEX (tenant_id, user_id, deleted_at)` — supports the "include soft-deleted on join" read pattern

FK behavior on `theme_id` / `region_id`: `ON DELETE RESTRICT`. Soft-delete of theme/region is allowed in app logic; hard delete is blocked at the DB layer.

`secondary_region_ids` is a JSON array of region UUIDs. The application enforces that each entry references an existing region row (active or soft-deleted) owned by the same `(tenant_id, user_id)`. The primary `region_id` MUST NOT also appear in `secondary_region_ids`.

`weight_type` semantics:
- `free` — `default_weight` is the load on the bar/machine.
- `bodyweight` — `default_weight` is *added* weight only (e.g., dip belt). May be 0 or null. Volume math contributes `sets × reps` only; no weight-based volume.

`isometric` items use `default_sets` + `default_duration_seconds` (+ optional `default_weight` for weighted planks). They have no `default_reps`.

---

## 4. `workout.weeks`

| Column            | Type        | Notes                                                     |
|-------------------|-------------|-----------------------------------------------------------|
| id                | uuid        | PK                                                        |
| tenant_id         | uuid        | not null                                                  |
| user_id           | uuid        | not null                                                  |
| week_start_date   | date        | not null, always a Monday (server normalizes input)       |
| rest_day_flags    | int[]       | not null, default `'{}'` — day-of-week ints (0=Mon..6=Sun) |
| created_at        | timestamptz | not null                                                  |
| updated_at        | timestamptz | not null                                                  |

Indexes:
- `UNIQUE (tenant_id, user_id, week_start_date)`
- `INDEX (tenant_id, user_id, week_start_date DESC)` — for "most recent prior week" lookups

A row is created lazily on first mutation that targets the week (add planned item, copy from previous, set rest day flags, etc.).

`rest_day_flags` lets the user mark specific days as explicit rest days. A day with the flag set displays "Rest day" in the UI even if no items are planned. A day may legally have both rest_day_flag set and planned items; no validation prevents this.

---

## 5. `workout.planned_items`

| Column                   | Type         | Notes                                                              |
|--------------------------|--------------|--------------------------------------------------------------------|
| id                       | uuid         | PK                                                                 |
| tenant_id                | uuid         | not null                                                           |
| user_id                  | uuid         | not null                                                           |
| week_id                  | uuid         | not null, FK → `workout.weeks(id)` ON DELETE CASCADE               |
| exercise_id              | uuid         | not null, FK → `workout.exercises(id)` ON DELETE RESTRICT          |
| day_of_week              | int          | not null, in `[0,6]` (0=Monday)                                    |
| position                 | int          | not null, ≥ 0                                                      |
| planned_sets             | int          | nullable                                                           |
| planned_reps             | int          | nullable                                                           |
| planned_weight           | numeric(7,2) | nullable                                                           |
| planned_weight_unit      | varchar(4)   | nullable, in `('lb','kg')`                                         |
| planned_duration_seconds | int          | nullable                                                           |
| planned_distance         | numeric(8,3) | nullable                                                           |
| planned_distance_unit    | varchar(4)   | nullable, in `('mi','km','m')`                                     |
| notes                    | varchar(500) | nullable                                                           |
| created_at               | timestamptz  | not null                                                           |
| updated_at               | timestamptz  | not null                                                           |

Indexes:
- `INDEX (week_id, day_of_week, position)` — primary read pattern (week view)
- `INDEX (tenant_id, user_id, exercise_id)` — for "what weeks reference this exercise" if ever needed

Notes:
- The valid set of `planned_*` columns is derived from the joined `exercises.kind`:
  - `strength`: `planned_sets`, `planned_reps`, `planned_weight`, `planned_weight_unit`
  - `isometric`: `planned_sets`, `planned_duration_seconds`, optional `planned_weight` + unit
  - `cardio`: `planned_duration_seconds`, `planned_distance`, `planned_distance_unit`
  Application logic enforces shape; no DB CHECK constraint.
- No name/kind snapshotting. All display fields come from a join to `workout.exercises` that includes soft-deleted rows. The `ON DELETE RESTRICT` FK guarantees the row is always resolvable.

---

## 6. `workout.performances`

1:1 with `planned_items`. A row is created the first time actuals are written (or the user marks the item `done`/`skipped`). Absence of a row is equivalent to `status='pending'`.

| Column                  | Type         | Notes                                                                  |
|-------------------------|--------------|------------------------------------------------------------------------|
| id                      | uuid         | PK                                                                     |
| tenant_id               | uuid         | not null                                                               |
| user_id                 | uuid         | not null                                                               |
| planned_item_id         | uuid         | not null, UNIQUE, FK → `workout.planned_items(id)` ON DELETE CASCADE   |
| status                  | varchar(16)  | not null, in `('pending','done','skipped','partial')`                  |
| mode                    | varchar(16)  | not null, in `('summary','per_set')`, default `'summary'`              |
| weight_unit             | varchar(4)   | nullable, in `('lb','kg')` — applies to summary actuals AND all per-set rows |
| actual_sets             | int          | nullable; valid only when `mode='summary'`                             |
| actual_reps             | int          | nullable; valid only when `mode='summary'`                             |
| actual_weight           | numeric(7,2) | nullable; valid only when `mode='summary'`                             |
| actual_duration_seconds | int          | nullable; isometric/cardio only                                        |
| actual_distance         | numeric(8,3) | nullable; cardio only                                                  |
| actual_distance_unit    | varchar(4)   | nullable, in `('mi','km','m')`                                         |
| notes                   | varchar(500) | nullable                                                               |
| created_at              | timestamptz  | not null                                                               |
| updated_at              | timestamptz  | not null                                                               |

Indexes:
- `UNIQUE (planned_item_id)` — already implied by the unique FK column

`weight_unit` lives at the performance level (not per-set). Switching `weight_unit` on a performance that has existing per-set rows is rejected with `409`.

---

## 7. `workout.performance_sets`

Per-set rows for `mode='per_set'` performances. Only valid for `strength` items — `isometric` and `cardio` performances never have rows here.

| Column          | Type         | Notes                                                                |
|-----------------|--------------|----------------------------------------------------------------------|
| id              | uuid         | PK                                                                   |
| tenant_id       | uuid         | not null                                                             |
| user_id         | uuid         | not null                                                             |
| performance_id  | uuid         | not null, FK → `workout.performances(id)` ON DELETE CASCADE          |
| set_number      | int          | not null, ≥ 1                                                        |
| reps            | int          | not null, ≥ 0                                                        |
| weight          | numeric(7,2) | not null, ≥ 0                                                        |
| created_at      | timestamptz  | not null                                                             |

Indexes:
- `UNIQUE (performance_id, set_number)`

The unit for `weight` comes from the parent `performances.weight_unit`.

---

## 8. Cascade Behavior Summary

| Source delete             | Effect                                                             |
|---------------------------|--------------------------------------------------------------------|
| `weeks` row deleted       | `planned_items` cascade → `performances` cascade → `performance_sets` cascade |
| `planned_items` row deleted | `performances` cascade → `performance_sets` cascade               |
| `exercises` soft delete   | No cascade. `planned_items.exercise_id` continues to resolve. Display falls back to `exercise_name_snapshot`. |
| `exercises` hard delete   | Blocked by FK (`ON DELETE RESTRICT`)                               |
| `themes` / `regions` soft delete | No cascade. Exercises continue to reference the soft-deleted row. Display uses the (now frozen) name. |
| `themes` / `regions` hard delete | Blocked by FK on `exercises` table                          |

Hard deletes are not exposed in the API; this matrix exists as a defense-in-depth statement of the schema's behavior.

---

## 9. Migration Plan

Initial migration (only one needed for v1):

1. Create schema `workout` if not exists.
2. Create tables in this order to satisfy FKs: `themes`, `regions`, `exercises`, `weeks`, `planned_items`, `performances`, `performance_sets`.
3. Create the indexes listed above.
4. No data backfill (greenfield service).

GORM `AutoMigrate` is the mechanism, per project standard. Custom SQL is only required for the partial unique indexes (`WHERE deleted_at IS NULL`), which GORM doesn't generate natively — those are created via `db.Exec(...)` in the service's migration entry point, following the pattern already used in `tracker-service` and `category-service`.

---

## 10. Default Seeding

On the first request from a `(tenant_id, user_id)` pair, the service checks whether `workout.themes` has any rows for that pair. If not, it inserts the default theme list and the default region list in a single transaction. This is the same pattern used by `category-service` for default category seeding.

Seeding is idempotent: if any theme or region already exists for the user, no rows are inserted.
