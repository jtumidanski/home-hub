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

| Column        | Type             | Constraints |
|---------------|------------------|-------------|
| id            | UUID             | PRIMARY KEY |
| tenant_id     | UUID             | NOT NULL    |
| name          | TEXT             | NOT NULL    |
| timezone      | TEXT             | NOT NULL    |
| units         | TEXT             | NOT NULL    |
| latitude      | DOUBLE PRECISION | NULLABLE    |
| longitude     | DOUBLE PRECISION | NULLABLE    |
| location_name | TEXT             | NULLABLE    |
| created_at    | TIMESTAMP        | NOT NULL    |
| updated_at    | TIMESTAMP        | NOT NULL    |

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

| Column              | Type      | Constraints           |
|---------------------|-----------|-----------------------|
| id                  | UUID      | PRIMARY KEY           |
| tenant_id           | UUID      | NOT NULL              |
| user_id             | UUID      | NOT NULL              |
| theme               | TEXT      | NOT NULL, DEFAULT 'light' |
| active_household_id | UUID      | NULLABLE              |
| created_at          | TIMESTAMP | NOT NULL              |
| updated_at          | TIMESTAMP | NOT NULL              |

### invitations

| Column       | Type      | Constraints              |
|--------------|-----------|--------------------------|
| id           | UUID      | PRIMARY KEY              |
| tenant_id    | UUID      | NOT NULL                 |
| household_id | UUID      | NOT NULL                 |
| email        | TEXT      | NOT NULL                 |
| role         | TEXT      | NOT NULL, DEFAULT 'viewer' |
| status       | TEXT      | NOT NULL, DEFAULT 'pending' |
| invited_by   | UUID      | NOT NULL                 |
| expires_at   | TIMESTAMP | NOT NULL                 |
| created_at   | TIMESTAMP | NOT NULL                 |
| updated_at   | TIMESTAMP | NOT NULL                 |

## Relationships

- `households.tenant_id` references a tenant.
- `memberships.tenant_id` references a tenant.
- `memberships.household_id` references a household.
- `memberships.user_id` references a user (external).
- `preferences.tenant_id` references a tenant.
- `preferences.user_id` references a user (external).
- `preferences.active_household_id` references a household (nullable).
- `invitations.tenant_id` references a tenant.
- `invitations.household_id` references a household.
- `invitations.invited_by` references a user (external).

## Indexes

| Table       | Index Name                      | Columns                  | Type           |
|-------------|---------------------------------|--------------------------|----------------|
| households  | (auto)                          | tenant_id                | INDEX          |
| memberships | (auto)                          | tenant_id                | INDEX          |
| memberships | (auto)                          | household_id             | INDEX          |
| memberships | (auto)                          | user_id                  | INDEX          |
| memberships | idx_household_user              | household_id, user_id    | UNIQUE         |
| preferences | idx_tenant_user                 | tenant_id, user_id       | UNIQUE         |
| invitations | (auto)                          | tenant_id                | INDEX          |
| invitations | (auto)                          | household_id             | INDEX          |
| invitations | idx_invitations_email_status    | email, status            | INDEX          |
| invitations | idx_invitations_unique_pending  | household_id, email      | UNIQUE PARTIAL |

The partial unique index `idx_invitations_unique_pending` applies only to rows where `status = 'pending'`.

## Migration Rules

- Migrations run automatically on service startup via GORM AutoMigrate.
- Migration order: tenants, households, memberships, preferences, invitations.
- The invitations partial unique index is created via raw SQL after AutoMigrate.
