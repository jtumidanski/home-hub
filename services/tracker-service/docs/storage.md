# Storage

All tables are created in the PostgreSQL `tracker` schema (set via the connection's `search_path`). Schema management is handled by GORM AutoMigrate on startup. GORM table names are unprefixed; the `tracker` schema is selected by the connection.

## Tables

### tracking_items

| Column       | Type         | Constraints                                                |
|--------------|--------------|------------------------------------------------------------|
| id           | UUID         | PRIMARY KEY                                                |
| tenant_id    | UUID         | NOT NULL, INDEX                                            |
| user_id      | UUID         | NOT NULL, INDEX                                            |
| name         | VARCHAR(100) | NOT NULL                                                   |
| scale_type   | VARCHAR(20)  | NOT NULL                                                   |
| scale_config | JSONB        |                                                            |
| color        | VARCHAR(20)  | NOT NULL                                                   |
| sort_order   | INT          | NOT NULL, DEFAULT 0                                        |
| created_at   | TIMESTAMPTZ  | NOT NULL                                                   |
| updated_at   | TIMESTAMPTZ  | NOT NULL                                                   |
| deleted_at   | TIMESTAMPTZ  | INDEX                                                      |

### schedule_snapshots

| Column           | Type        | Constraints                                                  |
|------------------|-------------|--------------------------------------------------------------|
| id               | UUID        | PRIMARY KEY                                                  |
| tracking_item_id | UUID        | NOT NULL, INDEX                                              |
| schedule         | JSONB       | NOT NULL (array of day-of-week integers, 0=Sun … 6=Sat)      |
| effective_date   | DATE        | NOT NULL                                                     |
| created_at       | TIMESTAMPTZ | NOT NULL                                                     |

### tracking_entries

| Column           | Type         | Constraints                                                |
|------------------|--------------|------------------------------------------------------------|
| id               | UUID         | PRIMARY KEY                                                |
| tenant_id        | UUID         | NOT NULL                                                   |
| user_id          | UUID         | NOT NULL                                                   |
| tracking_item_id | UUID         | NOT NULL                                                   |
| date             | DATE         | NOT NULL                                                   |
| value            | JSONB        | NULL when `skipped` is true                                |
| skipped          | BOOLEAN      | NOT NULL, DEFAULT false                                    |
| note             | VARCHAR(500) |                                                            |
| created_at       | TIMESTAMPTZ  | NOT NULL                                                   |
| updated_at       | TIMESTAMPTZ  | NOT NULL                                                   |

## Indexes

| Table              | Index Name                          | Columns                                | Type   |
|--------------------|-------------------------------------|----------------------------------------|--------|
| tracking_items     | idx_tracking_item_tenant_user_name  | (tenant_id, user_id, name) WHERE deleted_at IS NULL | UNIQUE |
| tracking_items     | idx_tracking_item_user              | (user_id)                              | INDEX  |
| schedule_snapshots | idx_schedule_item                   | (tracking_item_id)                     | INDEX  |
| schedule_snapshots | idx_schedule_item_date              | (tracking_item_id, effective_date)     | UNIQUE |
| tracking_entries   | idx_entry_item_date                 | (tracking_item_id, date)               | UNIQUE |
| tracking_entries   | idx_entry_tenant_user_date          | (tenant_id, user_id, date)             | INDEX  |

## Value Encoding

The `tracking_entries.value` JSONB column stores scale-appropriate payloads:

- sentiment: `{"rating": "positive" | "neutral" | "negative"}`
- numeric: `{"count": <int>}` (non-negative)
- range: `{"value": <int>}` (within `tracking_items.scale_config.min`/`max`)
- skipped entries: `value` is `NULL`

The `tracking_items.scale_config` JSONB column is `NULL` for `sentiment` and `numeric` items, and `{"min": <int>, "max": <int>}` for `range` items.

The `schedule_snapshots.schedule` JSONB column is a JSON array of day-of-week integers (e.g. `[1,2,4,5]`). An empty array means "every day".

## Migration Rules

- Migrations run automatically on service startup via GORM AutoMigrate.
- Migration order: `tracking_items`, `schedule_snapshots`, `tracking_entries`.
- Soft delete is implemented via the `deleted_at` column on `tracking_items`; the unique constraint on `(tenant_id, user_id, name)` is partial (`WHERE deleted_at IS NULL`) so a deleted item's name can be reused.
