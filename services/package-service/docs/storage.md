# Storage

All tables are created in the PostgreSQL `package` schema. Schema management is handled by GORM AutoMigrate on startup.

## Tables

### packages

| Column               | Type         | Constraints                              |
|----------------------|--------------|------------------------------------------|
| id                   | UUID         | PRIMARY KEY                              |
| tenant_id            | UUID         | NOT NULL, INDEX                          |
| household_id         | UUID         | NOT NULL, INDEX                          |
| user_id              | UUID         | NOT NULL                                 |
| tracking_number      | VARCHAR(64)  | NOT NULL                                 |
| carrier              | VARCHAR(16)  | NOT NULL                                 |
| label                | VARCHAR(255) |                                          |
| notes                | TEXT         |                                          |
| status               | VARCHAR(24)  | NOT NULL, DEFAULT 'pre_transit', INDEX   |
| private              | BOOLEAN      | NOT NULL, DEFAULT false                  |
| estimated_delivery   | DATE         |                                          |
| actual_delivery      | TIMESTAMPTZ  |                                          |
| last_polled_at       | TIMESTAMPTZ  |                                          |
| last_status_change_at| TIMESTAMPTZ  |                                          |
| archived_at          | TIMESTAMPTZ  |                                          |
| created_at           | TIMESTAMPTZ  | NOT NULL                                 |
| updated_at           | TIMESTAMPTZ  | NOT NULL                                 |

### tracking_events

| Column      | Type         | Constraints                              |
|-------------|--------------|------------------------------------------|
| id          | UUID         | PRIMARY KEY                              |
| package_id  | UUID         | NOT NULL, INDEX                          |
| timestamp   | TIMESTAMPTZ  | NOT NULL, INDEX (composite with package_id, sort: desc) |
| status      | VARCHAR(24)  | NOT NULL                                 |
| description | VARCHAR(512) | NOT NULL                                 |
| location    | VARCHAR(255) |                                          |
| raw_status  | VARCHAR(128) |                                          |
| created_at  | TIMESTAMPTZ  | NOT NULL                                 |

## Indexes

| Table            | Index Name                | Columns                                    | Type     | Condition |
|------------------|---------------------------|--------------------------------------------|----------|-----------|
| packages         | idx_pkg_household_status  | (tenant_id, household_id, status)          | INDEX    |           |
| packages         | idx_pkg_household_tracking| (tenant_id, household_id, tracking_number) | UNIQUE   |           |
| packages         | idx_pkg_polling           | (status, last_polled_at)                   | INDEX    | WHERE status IN ('pre_transit', 'in_transit', 'out_for_delivery') |
| packages         | idx_pkg_cleanup           | (status, archived_at)                      | INDEX    | WHERE status IN ('delivered', 'archived') |
| tracking_events  | idx_te_package_time       | (package_id, timestamp DESC)               | INDEX    |           |
| tracking_events  | idx_te_dedup              | (package_id, timestamp, description)       | UNIQUE   |           |

## Migration Rules

- Migrations run automatically on service startup via GORM AutoMigrate.
- Migration order: packages, tracking_events.
- Conditional indexes (idx_pkg_polling, idx_pkg_cleanup) created via raw SQL after AutoMigrate.
- Deduplication index (idx_te_dedup) created via raw SQL after removing any existing duplicate rows.
