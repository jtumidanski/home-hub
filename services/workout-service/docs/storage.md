# workout-service — Storage

Schema: `workout.*`. GORM `AutoMigrate` runs on startup; partial unique indexes
and explicit FK clauses are created via `db.Exec` in each domain package's
`Migration` function.

## Tables

| Table | Purpose |
| --- | --- |
| `workout.themes` | Per-user taxonomy bucket. Unique `(tenant_id, user_id, name)` for active rows. |
| `workout.regions` | Per-user body region. Same shape as themes. |
| `workout.exercises` | Personal exercise catalog. `kind` and `weight_type` immutable; `secondary_region_ids` is jsonb (UUID strings). |
| `workout.weeks` | One row per (user, ISO Monday). `rest_day_flags` is jsonb (int array). |
| `workout.planned_items` | Planned exercises within a week. `(week_id, day_of_week, position)` is the primary read pattern. |
| `workout.performances` | 1:1 with planned items. `mode` ∈ {`summary`, `per_set`}. |
| `workout.performance_sets` | Per-set rows for `mode = per_set`. Strength items only. |

## Storage notes

- `secondary_region_ids` and `rest_day_flags` are both stored as jsonb instead of native postgres array types. The jsonb representation lets us reuse the existing tracker-service jsonb pattern and gives us the `@>` containment operator for the `regionId` filter on `GET /exercises`.
- Partial unique indexes (`WHERE deleted_at IS NULL`) on themes, regions, and exercises are created in `Migration` via `db.Exec` because GORM does not generate them natively.
- Foreign-key cascades are explicitly re-declared after `AutoMigrate`:
  - `planned_items.week_id` → `weeks(id)` `ON DELETE CASCADE`
  - `planned_items.exercise_id` → `exercises(id)` `ON DELETE RESTRICT`
  - `performances.planned_item_id` → `planned_items(id)` `ON DELETE CASCADE`
  - `performance_sets.performance_id` → `performances(id)` `ON DELETE CASCADE`
- Soft delete on themes, regions, and exercises is implemented via `deleted_at` columns. Read-through joins in `weekview` and `summary` use `Where("id IN ?", ids)` (no `deleted_at IS NULL` filter) so historical references continue to render the original name.
- Tenant scoping is automatic via `shared/go/database`'s tenant callbacks. Reads, updates, and deletes that touch tables with a `tenant_id` column have `WHERE tenant_id = ?` injected from the request context.

## Default seeding

On the first request from a `(tenant_id, user_id)` pair, the service inserts:

- Themes: `Muscle, Cardio`
- Regions: `Chest, Shoulders, Back, Biceps, Triceps, Core, Legs, Glutes, Full Body, Other`

Seeding is wrapped in a transaction and is idempotent — if any row exists for the user, no rows are inserted.
