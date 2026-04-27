# Storage

All tables are created in the PostgreSQL `dashboard` schema
(`internal/config/config.go:Load`, `DB.Schema = "dashboard"`). Schema management
is handled by GORM AutoMigrate on startup, invoked from
`internal/dashboard/entity.go:Migration`.

## Tables

### dashboards

Defined by `internal/dashboard/entity.go:Entity`.

| Column          | Type         | Constraints                                        |
|-----------------|--------------|----------------------------------------------------|
| id              | UUID         | PRIMARY KEY                                        |
| tenant_id       | UUID         | NOT NULL, INDEX (`idx_dashboards_scope`)           |
| household_id    | UUID         | NOT NULL, INDEX (`idx_dashboards_scope`)           |
| user_id         | UUID         | NULLABLE, INDEX (`idx_dashboards_scope`)           |
| name            | VARCHAR(80)  | NOT NULL                                           |
| sort_order      | INT          | NOT NULL, DEFAULT 0                                |
| layout          | JSONB        | NOT NULL                                           |
| schema_version  | INT          | NOT NULL, DEFAULT 1                                |
| created_at      | TIMESTAMPTZ  | NOT NULL                                           |
| updated_at      | TIMESTAMPTZ  | NOT NULL                                           |

`user_id IS NULL` identifies a household-scoped row; a non-null `user_id`
identifies a user-scoped row owned by that user. See `docs/domain.md` for the
scope-derivation rules.

## Indexes

| Index Name                        | Columns                                   | Type     | Condition              |
|-----------------------------------|-------------------------------------------|----------|------------------------|
| `idx_dashboards_scope`            | `(tenant_id, household_id, user_id)`      | INDEX    |                        |
| `idx_dashboards_household_partial`| `(tenant_id, household_id)`               | INDEX    | `WHERE user_id IS NULL` |

- `idx_dashboards_scope` is created via the GORM struct tags on `Entity`.
- `idx_dashboards_household_partial` is created by `Migration` via raw
  `CREATE INDEX IF NOT EXISTS ...`. It accelerates the household-only paths
  used by `provider.go:visibleToCaller` and the seed-count query
  `provider.go:countHouseholdScoped`.

## Seed Advisory Lock

`Seed` serializes concurrent callers via a Postgres transaction-scoped advisory
lock so two replicas cannot both insert the first household-scoped row.

- Key derivation (`processor.go:seedLockKey`): SHA-256 of the 32-byte
  concatenation `tenantID[:] || householdID[:]`; the low 8 bytes of the digest
  are reinterpreted as a big-endian `int64`. Every replica hashes to the same
  advisory-lock slot for a given (tenant, household).
- Acquisition (`processor.go:acquireSeedLock`): inside the `Seed` transaction,
  `SELECT pg_advisory_xact_lock(?)` with the derived key. The lock releases
  automatically at transaction end.
- Dialect guard: non-Postgres dialects (SQLite in tests) fall through as a
  no-op; production always runs Postgres, and tests rely on the surrounding
  transaction for serialization.

## JSON Payload Caps

Enforced by `internal/layout/validator.go` (see `docs/domain.md` for the full
rule list):

| Cap                                   | Constant                              | Value    |
|---------------------------------------|---------------------------------------|----------|
| Full layout document                  | `shared.MaxLayoutBytes`               | 64 KiB   |
| Per-widget `config` object            | `shared.MaxWidgetConfigBytes`         | 4 KiB    |
| Per-widget `config` nesting depth     | `shared.MaxWidgetConfigDepth`         | 5        |
| Widget count                          | `shared.MaxWidgets`                   | 40       |
| Grid width                            | `shared.GridColumns`                  | 12       |

Values live in `shared/go/dashboard/types.go`.

## Migration Rules

- GORM AutoMigrate runs on startup (`cmd/main.go` → `database.Connect(…,
  database.SetMigrations(dashboard.Migration))`).
- `Migration` runs AutoMigrate for `Entity` then issues the
  `idx_dashboards_household_partial` raw SQL. Both steps are idempotent.
- The shared retention framework migrates its own `retention_runs` table via
  `sr.MigrateRuns` inside `internal/retention/wire.go:Setup`.
