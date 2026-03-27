# Storage

All tables are created in the PostgreSQL `auth` schema. Schema management is handled by GORM AutoMigrate on startup.

## Tables

### users

| Column              | Type      | Constraints     |
|---------------------|-----------|-----------------|
| id                  | UUID      | PRIMARY KEY     |
| email               | TEXT      | NOT NULL        |
| display_name        | TEXT      |                 |
| given_name          | TEXT      |                 |
| family_name         | TEXT      |                 |
| avatar_url          | TEXT      |                 |
| provider_avatar_url | TEXT      |                 |
| created_at          | TIMESTAMP | NOT NULL        |
| updated_at          | TIMESTAMP | NOT NULL        |

### external_identities

| Column           | Type      | Constraints |
|------------------|-----------|-------------|
| id               | UUID      | PRIMARY KEY |
| user_id          | UUID      | NOT NULL    |
| provider         | TEXT      | NOT NULL    |
| provider_subject | TEXT      | NOT NULL    |
| created_at       | TIMESTAMP | NOT NULL    |
| updated_at       | TIMESTAMP | NOT NULL    |

### oidc_providers

| Column     | Type      | Constraints |
|------------|-----------|-------------|
| id         | UUID      | PRIMARY KEY |
| name       | TEXT      | NOT NULL    |
| issuer_url | TEXT      | NOT NULL    |
| client_id  | TEXT      | NOT NULL    |
| enabled    | BOOLEAN   | NOT NULL    |
| created_at | TIMESTAMP | NOT NULL    |
| updated_at | TIMESTAMP | NOT NULL    |

### refresh_tokens

| Column     | Type      | Constraints          |
|------------|-----------|----------------------|
| id         | UUID      | PRIMARY KEY          |
| user_id    | UUID      | NOT NULL             |
| token_hash | TEXT      | NOT NULL             |
| expires_at | TIMESTAMP | NOT NULL             |
| revoked    | BOOLEAN   | NOT NULL, DEFAULT false |
| created_at | TIMESTAMP | NOT NULL             |
| updated_at | TIMESTAMP | NOT NULL             |

## Relationships

- `external_identities.user_id` references a user.
- `refresh_tokens.user_id` references a user.

## Indexes

| Table               | Index Name            | Columns                      | Type   |
|---------------------|-----------------------|------------------------------|--------|
| users               | (auto)                | email                        | UNIQUE |
| external_identities | idx_provider_subject  | provider_subject             | UNIQUE |
| refresh_tokens      | (auto)                | user_id                      | INDEX  |

## Migration Rules

- Migrations run automatically on service startup via GORM AutoMigrate.
- Migration order: users, external_identities, oidc_providers, refresh_tokens.
- After AutoMigrate, `MigrateAvatarData` runs: copies `avatar_url` to `provider_avatar_url` where `provider_avatar_url` is empty and `avatar_url` is not a `dicebear:` descriptor, then clears `avatar_url`. Idempotent.
