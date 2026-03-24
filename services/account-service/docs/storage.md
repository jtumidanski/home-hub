# Storage

All tables are created in the PostgreSQL `account` schema. Schema management is handled by GORM AutoMigrate on startup.

## Tables

### tenants

| Column     | Type      | Constraints |
|------------|-----------|-------------|
| id         | UUID      | PRIMARY KEY |
| name       | TEXT      | NOT NULL    |
| created_at | TIMESTAMP | NOT NULL    |
| updated_at | TIMESTAMP | NOT NULL    |

### households

| Column     | Type      | Constraints |
|------------|-----------|-------------|
| id         | UUID      | PRIMARY KEY |
| tenant_id  | UUID      | NOT NULL    |
| name       | TEXT      | NOT NULL    |
| timezone   | TEXT      | NOT NULL    |
| units      | TEXT      | NOT NULL    |
| created_at | TIMESTAMP | NOT NULL    |
| updated_at | TIMESTAMP | NOT NULL    |

### memberships

| Column       | Type      | Constraints |
|--------------|-----------|-------------|
| id           | UUID      | PRIMARY KEY |
| tenant_id    | UUID      | NOT NULL    |
| household_id | UUID      | NOT NULL    |
| user_id      | UUID      | NOT NULL    |
| role         | TEXT      | NOT NULL    |
| created_at   | TIMESTAMP | NOT NULL    |
| updated_at   | TIMESTAMP | NOT NULL    |

### preferences

| Column              | Type      | Constraints        |
|---------------------|-----------|--------------------|
| id                  | UUID      | PRIMARY KEY        |
| tenant_id           | UUID      | NOT NULL           |
| user_id             | UUID      | NOT NULL           |
| theme               | TEXT      | NOT NULL, DEFAULT 'light' |
| active_household_id | UUID      | NULLABLE           |
| created_at          | TIMESTAMP | NOT NULL           |
| updated_at          | TIMESTAMP | NOT NULL           |

## Relationships

- `households.tenant_id` references a tenant.
- `memberships.tenant_id` references a tenant.
- `memberships.household_id` references a household.
- `memberships.user_id` references a user (external).
- `preferences.tenant_id` references a tenant.
- `preferences.user_id` references a user (external).
- `preferences.active_household_id` references a household (nullable).

## Indexes

| Table       | Index Name          | Columns                  | Type   |
|-------------|---------------------|--------------------------|--------|
| households  | (auto)              | tenant_id                | INDEX  |
| memberships | (auto)              | tenant_id                | INDEX  |
| memberships | (auto)              | household_id             | INDEX  |
| memberships | (auto)              | user_id                  | INDEX  |
| memberships | idx_household_user  | household_id, user_id    | UNIQUE |
| preferences | idx_tenant_user     | tenant_id, user_id       | UNIQUE |

## Migration Rules

- Migrations run automatically on service startup via GORM AutoMigrate.
- Migration order: tenants, households, memberships, preferences.
