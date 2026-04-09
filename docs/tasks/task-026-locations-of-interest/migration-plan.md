# Migration Plan

The weather-service uses GORM AutoMigrate on startup. This change has two parts:

1. Create the new `weather.locations_of_interest` table.
2. Reshape `weather.weather_caches` to support multiple cache rows per household.

## Part 1 — Create `locations_of_interest`

Standard GORM AutoMigrate handles this. Add the new entity to the migration list in the weather-service startup sequence; it must run **before** any migration step that adds the FK on `weather_caches.location_id`.

## Part 2 — Reshape `weather_caches`

The current state:

- Column set is fixed (id, tenant_id, household_id, lat, lon, units, current_data, forecast_data, fetched_at, created_at, updated_at).
- Unique index `idx_weather_household` on `household_id`.

Target state:

- New nullable `location_id` UUID column with FK to `locations_of_interest(id) ON DELETE CASCADE`.
- Drop `idx_weather_household`.
- Add **partial unique** index on `(household_id) WHERE location_id IS NULL` (one primary cache per household).
- Add **partial unique** index on `(household_id, location_id) WHERE location_id IS NOT NULL` (one cache per saved location).

GORM AutoMigrate cannot express partial unique indexes via struct tags alone, and it will not drop the existing unique index automatically. The migration must be done in steps:

### Step-by-step

1. **Add column**: GORM AutoMigrate adds `location_id UUID NULL` (no constraint). Existing rows get `NULL`, which correctly represents "primary location cache".
2. **Hand-written migration step (idempotent)**, run after AutoMigrate inside the same startup sequence:
   - `DROP INDEX IF EXISTS idx_weather_household`
   - `CREATE UNIQUE INDEX IF NOT EXISTS idx_weather_household_primary ON weather.weather_caches (household_id) WHERE location_id IS NULL`
   - `CREATE UNIQUE INDEX IF NOT EXISTS idx_weather_household_location ON weather.weather_caches (household_id, location_id) WHERE location_id IS NOT NULL`
   - `ALTER TABLE weather.weather_caches ADD CONSTRAINT fk_weather_location FOREIGN KEY (location_id) REFERENCES weather.locations_of_interest(id) ON DELETE CASCADE` (guarded with a `pg_constraint` existence check or wrapped in `DO $$ ... $$` so re-runs are safe)
3. Verify in tests that:
   - Existing rows survive with `location_id = NULL`.
   - Inserting a second row with the same `household_id` and `location_id = NULL` fails.
   - Inserting two rows with the same `household_id` but different non-null `location_id` succeeds.
   - Deleting a row from `locations_of_interest` cascades to its cache row.

### Why partial indexes?

A standard `UNIQUE (household_id, location_id)` would treat `NULL` as distinct, allowing multiple "primary" cache rows per household — which violates the invariant. Partial indexes split the constraint into the two cases cleanly and are well-supported in PostgreSQL.

### Rollback consideration

If rollback is ever needed (revert this feature):
- Drop the partial indexes and FK.
- Recreate `idx_weather_household` as `UNIQUE (household_id)` — this will fail if any household has more than one cache row, so first delete all rows where `location_id IS NOT NULL`.
- Drop the `location_id` column.
- Drop the `locations_of_interest` table.

This is documented for completeness, not as part of normal deployment.
