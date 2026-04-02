# Storage

All tables are created in the PostgreSQL `calendar` schema. Schema management is handled by GORM AutoMigrate on startup.

## Tables

### calendar_connections

| Column              | Type         | Constraints                              |
|---------------------|--------------|------------------------------------------|
| id                  | UUID         | PRIMARY KEY                              |
| tenant_id           | UUID         | NOT NULL, INDEX                          |
| household_id        | UUID         | NOT NULL, INDEX                          |
| user_id             | UUID         | NOT NULL, INDEX                          |
| provider            | VARCHAR(50)  | NOT NULL                                 |
| status              | VARCHAR(20)  | NOT NULL, DEFAULT 'connected'            |
| email               | VARCHAR(255) | NOT NULL                                 |
| access_token        | TEXT         | NOT NULL (encrypted AES-256-GCM)         |
| refresh_token       | TEXT         | NOT NULL (encrypted AES-256-GCM)         |
| token_expiry        | TIMESTAMPTZ  | NOT NULL                                 |
| user_display_name   | VARCHAR(255) | NOT NULL                                 |
| user_color          | VARCHAR(7)   | NOT NULL                                 |
| write_access        | BOOLEAN      | NOT NULL, DEFAULT false                  |
| last_sync_at        | TIMESTAMPTZ  |                                          |
| last_sync_event_count | INT        | DEFAULT 0                                |
| created_at          | TIMESTAMPTZ  | NOT NULL                                 |
| updated_at          | TIMESTAMPTZ  | NOT NULL                                 |

### calendar_sources

| Column        | Type         | Constraints                              |
|---------------|--------------|------------------------------------------|
| id            | UUID         | PRIMARY KEY                              |
| tenant_id     | UUID         | NOT NULL, INDEX                          |
| household_id  | UUID         | NOT NULL, INDEX                          |
| connection_id | UUID         | NOT NULL, INDEX                          |
| external_id   | VARCHAR(255) | NOT NULL                                 |
| name          | VARCHAR(255) | NOT NULL                                 |
| primary       | BOOLEAN      | NOT NULL, DEFAULT false                  |
| visible       | BOOLEAN      | NOT NULL, DEFAULT true                   |
| color         | VARCHAR(7)   |                                          |
| sync_token    | TEXT         |                                          |
| created_at    | TIMESTAMPTZ  | NOT NULL                                 |
| updated_at    | TIMESTAMPTZ  | NOT NULL                                 |

### calendar_events

| Column            | Type         | Constraints                              |
|-------------------|--------------|------------------------------------------|
| id                | UUID         | PRIMARY KEY                              |
| tenant_id         | UUID         | NOT NULL, INDEX                          |
| household_id      | UUID         | NOT NULL, INDEX                          |
| connection_id     | UUID         | NOT NULL, INDEX                          |
| source_id         | UUID         | NOT NULL                                 |
| user_id           | UUID         | NOT NULL, INDEX                          |
| external_id       | VARCHAR(255) | NOT NULL                                 |
| google_calendar_id | VARCHAR(255) |                                         |
| title             | VARCHAR(500) | NOT NULL                                 |
| description       | TEXT         |                                          |
| start_time        | TIMESTAMPTZ  | NOT NULL, INDEX                          |
| end_time          | TIMESTAMPTZ  | NOT NULL, INDEX                          |
| all_day           | BOOLEAN      | NOT NULL, DEFAULT false                  |
| location          | VARCHAR(500) |                                          |
| visibility        | VARCHAR(20)  | NOT NULL, DEFAULT 'default'              |
| user_display_name | VARCHAR(255) | NOT NULL                                 |
| user_color        | VARCHAR(7)   | NOT NULL                                 |
| created_at        | TIMESTAMPTZ  | NOT NULL                                 |
| updated_at        | TIMESTAMPTZ  | NOT NULL                                 |

### calendar_oauth_states

| Column       | Type         | Constraints                              |
|--------------|--------------|------------------------------------------|
| id           | UUID         | PRIMARY KEY (= state parameter value)    |
| tenant_id    | UUID         | NOT NULL                                 |
| household_id | UUID         | NOT NULL                                 |
| user_id      | UUID         | NOT NULL                                 |
| redirect_uri | VARCHAR(500) | NOT NULL                                 |
| reauthorize  | BOOLEAN      | NOT NULL, DEFAULT false                  |
| expires_at   | TIMESTAMPTZ  | NOT NULL                                 |
| created_at   | TIMESTAMPTZ  | NOT NULL                                 |

## Indexes

| Table                  | Index Name                         | Columns                                    | Type   |
|------------------------|------------------------------------|--------------------------------------------|--------|
| calendar_connections   | idx_connections_tenant_household    | (tenant_id, household_id)                  | INDEX  |
| calendar_connections   | idx_connections_user                | (user_id)                                  | INDEX  |
| calendar_connections   | idx_connections_unique_provider     | (tenant_id, household_id, user_id, provider) | UNIQUE |
| calendar_sources       | idx_sources_connection              | (connection_id)                            | INDEX  |
| calendar_sources       | idx_sources_connection_external     | (connection_id, external_id)               | UNIQUE |
| calendar_events        | idx_events_tenant_household_time    | (tenant_id, household_id, start_time, end_time) | INDEX |
| calendar_events        | idx_events_connection               | (connection_id)                            | INDEX  |
| calendar_events        | idx_events_source_external          | (source_id, external_id)                   | UNIQUE |

## Migration Rules

- Migrations run automatically on service startup via GORM AutoMigrate.
- Migration order: calendar_oauth_states, calendar_connections, calendar_sources, calendar_events.
- Token fields (access_token, refresh_token) contain AES-256-GCM ciphertext, base64 encoded.
