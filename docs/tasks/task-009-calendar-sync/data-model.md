# Household Calendar — Data Model

All entities live in the `calendar` PostgreSQL schema, owned by the calendar-service. Managed via GORM AutoMigrate.

---

## Entity: `calendar_connections`

Stores OAuth connection state and tokens for external calendar providers.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PK | Application-generated |
| `tenant_id` | UUID | NOT NULL, INDEX | Tenant scope |
| `household_id` | UUID | NOT NULL, INDEX | Household scope |
| `user_id` | UUID | NOT NULL, INDEX | Owning user |
| `provider` | VARCHAR(50) | NOT NULL | Calendar provider (`google`) |
| `status` | VARCHAR(20) | NOT NULL, DEFAULT `connected` | Connection status |
| `email` | VARCHAR(255) | NOT NULL | Provider account email |
| `access_token` | TEXT | NOT NULL | Encrypted OAuth access token |
| `refresh_token` | TEXT | NOT NULL | Encrypted OAuth refresh token |
| `token_expiry` | TIMESTAMPTZ | NOT NULL | Access token expiration |
| `user_display_name` | VARCHAR(255) | NOT NULL | Display name from JWT claims at connection time |
| `last_sync_at` | TIMESTAMPTZ | | Last successful sync time |
| `last_sync_event_count` | INT | DEFAULT 0 | Event count from last sync |
| `created_at` | TIMESTAMPTZ | NOT NULL | Record creation |
| `updated_at` | TIMESTAMPTZ | NOT NULL | Last update |

**Indexes:**

- `idx_connections_tenant_household` on `(tenant_id, household_id)`
- `idx_connections_user` on `(user_id)`
- `UNIQUE(tenant_id, household_id, user_id, provider)` — one connection per provider per user per household

**Status values:** `connected`, `disconnected`, `syncing`, `error`

---

## Entity: `calendar_sources`

Stores individual Google Calendars associated with a connection, with visibility toggles.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PK | Application-generated |
| `tenant_id` | UUID | NOT NULL, INDEX | Tenant scope |
| `household_id` | UUID | NOT NULL, INDEX | Household scope |
| `connection_id` | UUID | NOT NULL, FK → calendar_connections(id), INDEX | Parent connection |
| `external_id` | VARCHAR(255) | NOT NULL | Google Calendar ID (e.g., `primary`, email, or hash) |
| `name` | VARCHAR(255) | NOT NULL | Calendar display name |
| `primary` | BOOLEAN | NOT NULL, DEFAULT false | Whether this is the user's primary Google Calendar |
| `visible` | BOOLEAN | NOT NULL, DEFAULT true | Whether events from this calendar appear on the household calendar |
| `color` | VARCHAR(7) | | Google Calendar color (informational) |
| `sync_token` | TEXT | | Google Calendar incremental sync token for this calendar |
| `created_at` | TIMESTAMPTZ | NOT NULL | Record creation |
| `updated_at` | TIMESTAMPTZ | NOT NULL | Last update |

**Indexes:**

- `idx_sources_connection` on `(connection_id)`
- `UNIQUE(connection_id, external_id)` — one entry per Google Calendar per connection

---

## Entity: `calendar_events`

Stores synced calendar events from external providers.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PK | Application-generated |
| `tenant_id` | UUID | NOT NULL, INDEX | Tenant scope |
| `household_id` | UUID | NOT NULL, INDEX | Household scope |
| `connection_id` | UUID | NOT NULL, FK → calendar_connections(id), INDEX | Source connection |
| `source_id` | UUID | NOT NULL, FK → calendar_sources(id), INDEX | Source calendar |
| `user_id` | UUID | NOT NULL, INDEX | Event owner (denormalized for query efficiency) |
| `external_id` | VARCHAR(255) | NOT NULL | Provider event ID (Google Calendar event ID) |
| `title` | VARCHAR(500) | NOT NULL | Event title |
| `description` | TEXT | | Event description |
| `start_time` | TIMESTAMPTZ | NOT NULL, INDEX | Event start |
| `end_time` | TIMESTAMPTZ | NOT NULL | Event end |
| `all_day` | BOOLEAN | NOT NULL, DEFAULT false | Whether this is an all-day event |
| `location` | VARCHAR(500) | | Event location |
| `visibility` | VARCHAR(20) | NOT NULL, DEFAULT `default` | Google Calendar visibility (default, public, private, confidential) |
| `user_display_name` | VARCHAR(255) | NOT NULL | Display name of the event owner |
| `user_color` | VARCHAR(7) | NOT NULL | Hex color assigned to the user |
| `created_at` | TIMESTAMPTZ | NOT NULL | Record creation |
| `updated_at` | TIMESTAMPTZ | NOT NULL | Last update |

**Indexes:**

- `idx_events_tenant_household_time` on `(tenant_id, household_id, start_time, end_time)` — primary query index
- `idx_events_connection` on `(connection_id)`
- `UNIQUE(source_id, external_id)` — prevents duplicate events per source calendar

**Visibility values:** `default`, `public`, `private`, `confidential`

---

## Entity: `calendar_oauth_states`

Temporary storage for OAuth CSRF state validation. Short-lived records.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UUID | PK | The state parameter value |
| `tenant_id` | UUID | NOT NULL | Tenant scope |
| `household_id` | UUID | NOT NULL | Household scope |
| `user_id` | UUID | NOT NULL | Initiating user |
| `redirect_uri` | VARCHAR(500) | NOT NULL | Frontend redirect URI |
| `expires_at` | TIMESTAMPTZ | NOT NULL | State expiration (10 minutes) |
| `created_at` | TIMESTAMPTZ | NOT NULL | Record creation |

**Cleanup:** Expired states are deleted during sync loop or on a separate cleanup tick.

---

## Relationships

```
calendar_connections 1──* calendar_sources      (connection_id FK)
calendar_connections 1──* calendar_events        (connection_id FK)
calendar_sources     1──* calendar_events        (source_id FK)
calendar_connections 1──* calendar_oauth_states  (conceptual, via user_id)
```

Events are filtered by `calendar_sources.visible = true` when querying for the household calendar view. Events from hidden sources remain in the database (to avoid re-syncing) but are excluded from API responses.

## User Color Assignment

Each user in a household is assigned a consistent color for event display. Colors are assigned deterministically based on the order of connection creation within a household, cycling through a predefined palette:

```
#4285F4 (blue), #EA4335 (red), #34A853 (green), #FBBC04 (yellow),
#8E24AA (purple), #00ACC1 (cyan), #FF7043 (orange), #78909C (grey)
```

The color is stored on each event record (denormalized) so that event queries don't require joining to connections.

---

## Migration Notes

- All migrations run via GORM AutoMigrate on service startup
- Schema prefix: `calendar.*`
- Token encryption: access_token and refresh_token are encrypted before storage using an encryption key provided via environment variable (`CALENDAR_TOKEN_ENCRYPTION_KEY`)
- The encryption/decryption logic lives in the connection domain, not in GORM hooks, to keep the entity layer simple
