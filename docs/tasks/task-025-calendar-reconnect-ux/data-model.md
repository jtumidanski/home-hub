# Data Model — Calendar Reconnect

## `calendar_connections` (modified)

### New columns

| Column | Type | Nullable | Default | Notes |
|---|---|---|---|---|
| `error_code` | `varchar(40)` | yes | NULL | Enum value, see below. NULL when healthy. |
| `error_message` | `text` | yes | NULL | Raw debug string from the underlying error. Never user-facing. |
| `last_error_at` | `timestamptz` | yes | NULL | Wall-clock time of the most recent failure record. |
| `last_sync_attempt_at` | `timestamptz` | yes | NULL | Wall-clock time of the most recent sync attempt regardless of outcome. |
| `consecutive_failures` | `integer` | no | `0` | Count of consecutive transient failures. |

### `error_code` enum values

| Value | Severity | Status transition |
|---|---|---|
| `token_revoked` | hard | → `disconnected` immediately |
| `refresh_unauthorized` | hard | → `disconnected` immediately |
| `token_decrypt_failed` | hard | → `disconnected` immediately |
| `refresh_http_error` | transient | counter → `error` after 3 |
| `unknown` | transient | counter → `error` after 3 |

The enum is enforced in code, not as a DB constraint, to keep migrations simple and let new codes ship without schema changes.

### Field interaction rules

| Event | `status` | `error_code` | `error_message` | `last_error_at` | `last_sync_attempt_at` | `last_sync_at` | `consecutive_failures` |
|---|---|---|---|---|---|---|---|
| Sync attempt starts | unchanged | unchanged | unchanged | unchanged | now | unchanged | unchanged |
| Sync attempt succeeds | `connected` | NULL | NULL | NULL | now | now | 0 |
| Transient failure (counter < 3) | unchanged | set | set | now | now | unchanged | +1 |
| Transient failure (counter reaches 3) | `error` | set | set | now | now | unchanged | +1 |
| Hard failure | `disconnected` | set | set | now | now | unchanged | max(current, 3) |
| Successful reauthorize callback | `connected` | NULL | NULL | NULL | unchanged | unchanged | 0 |

### Migration

- Use GORM `AutoMigrate` to add the columns. This is consistent with the existing `connection.Migration` function.
- No backfill SQL required. Columns get their declared defaults on existing rows:
  - `consecutive_failures` → `0` (column default).
  - All other new columns → `NULL`.
- Existing rows in any status are unaffected by the migration; they begin populating the new columns on their next sync attempt.

### Indexes

No new indexes. The new columns are not used in WHERE clauses by any new queries — they are only read alongside the row when listing a household's connections, which already uses the existing `idx_connections_tenant_household` index.

## Model invariants

`connection.Model` continues to be immutable post-construction. The new fields are added as private fields with getters and corresponding builder setters, matching the existing pattern in `connection/model.go`.

## Processor surface (new methods)

| Method | Purpose |
|---|---|
| `RecordSyncAttempt(id, at)` | Sets `last_sync_attempt_at = at`. Called at the top of `syncOne`. |
| `RecordSyncSuccess(id, eventCount, at)` | Replaces `UpdateSyncInfo`. Sets `last_sync_at`, `last_sync_attempt_at`, `last_sync_event_count`; clears error fields and counter; sets status to `connected` if it was `error`. |
| `RecordSyncFailure(id, code, message, at)` | Sets error fields, increments counter, transitions status per the table above. |
| `ClearErrorState(id)` | Used by the OAuth callback handler on successful reauthorize. Clears error fields, resets counter, sets status to `connected`. |

`UpdateStatus(id, status)` is retained for compatibility but new code paths use the more specific methods above.
