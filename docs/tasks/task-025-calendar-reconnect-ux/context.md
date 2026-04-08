# Context — Task 025: Calendar Reconnect UX

Last Updated: 2026-04-08

This file captures the codebase context that informed the PRD and the plan, so the implementer doesn't have to re-derive it.

## Symptom

When the calendar background sync fails to refresh a user's Google OAuth tokens, `sync.Engine.syncOne` (`services/calendar-service/internal/sync/sync.go:91-97`) silently flips the connection's `status` to `disconnected` and logs the error server-side. The frontend `ConnectionStatus` component (`frontend/src/components/features/calendar/connection-status.tsx:30-34`) renders only a bare red `disconnected` badge — no reason text, no recovery action. The user's only option is to delete the connection and re-add it.

The existing `ReauthorizeBanner` (`frontend/src/components/features/calendar/reauthorize-banner.tsx`) only fires for write-access upgrades on still-connected calendars. It does not surface for broken connections.

## Pipeline today (where the bug lives)

### Backend

- **`services/calendar-service/internal/sync/sync.go:82-117`** — `syncOne`. On `getValidAccessToken` error, calls `connProc.UpdateStatus(conn.Id(), "disconnected")` and returns. On success, calls `connProc.UpdateSyncInfo(conn.Id(), totalEvents)`.
- **`services/calendar-service/internal/sync/sync.go:119-149`** — `getValidAccessToken`. Decrypt access → check expiry → decrypt refresh → `gcClient.RefreshToken` → encrypt → `UpdateTokens`. Each step returns an unwrapped `error` that the caller cannot classify.
- **`services/calendar-service/internal/googlecal/client.go:88-112`** — `RefreshToken`. On non-200, returns `fmt.Errorf("token refresh returned %d: %s", ...)`. No typed error. **This is the classification gap.**
- **`services/calendar-service/internal/googlecal/client.go:337-353`** — `SyncTokenInvalidError` is the existing typed-error pattern; the new `TokenRefreshError` should mirror it.
- **`services/calendar-service/internal/connection/processor.go`** — `Processor` exposes `UpdateStatus`, `UpdateTokens`, `UpdateSyncInfo`, `UpdateTokensAndWriteAccess`. All sync-engine writes go through `noTenantDB()` (line 137-139) because the engine runs without tenant context.
- **`services/calendar-service/internal/connection/administrator.go`** — small `db.Model(&Entity{}).Where("id = ?", id).Updates(map)` helpers. New write helpers follow this pattern.
- **`services/calendar-service/internal/connection/entity.go:10-37`** — `Entity` struct. Migration via `AutoMigrate` (lines 32-37). New columns are added by extending the struct.
- **`services/calendar-service/internal/connection/model.go:14-50`** — `Model` is immutable; getters only.
- **`services/calendar-service/internal/connection/builder.go:18-93`** — `Builder` validates `provider`, `email`, `accessToken`, `refreshToken`, `userDisplayName` in `Build()`. New optional fields skip validation.
- **`services/calendar-service/internal/connection/rest.go:9-43`** — `RestModel` and `Transform`. JSON-tagged flat struct.
- **`services/calendar-service/internal/connection/resource.go:96-192`** — `callbackHandler`. The reauthorize branch (lines 152-173) currently calls `UpdateTokensAndWriteAccess`, which writes `status = 'connected'` (`administrator.go:65`) but does NOT clear error fields because they don't exist yet.

### Frontend

- **`frontend/src/types/models/calendar.ts:1-11`** — `CalendarConnectionAttributes`. Already declares `status` as `"connected" | "disconnected" | "syncing" | "error"`. The `error` value is in the type but no backend code currently writes it.
- **`frontend/src/components/features/calendar/connection-status.tsx`** — single horizontal flex row. Status badge only renders when `!isHealthy`. Sync button is disabled unless `attrs.status === "connected"` (line 44). No human messaging, no Reconnect button.
- **`frontend/src/lib/hooks/api/use-calendar.ts:163-174`** — `useReauthorizeCalendar`. Already POSTs to the reauthorize endpoint and redirects to the returned auth URL. **Reusable as-is for the new Reconnect button.**
- **`frontend/src/components/features/calendar/reauthorize-banner.tsx`** — separate component for write-access upgrades. Out of scope; remains unchanged.

## Auth boundary verification

`callbackHandler` is registered as a public route (no JWT required) — `connection/resource.go:39-44`. Tenant context for the reauthorize branch is recovered from the OAuth state row, not from the request, so adding `ClearErrorState` calls or extending `UpdateTokensAndWriteAccess` does not introduce a new auth surface.

The sync engine runs from a background goroutine and uses `connection.NewProcessor(l, ctx, e.db)` with a tenant-less context. All new processor methods must use `noTenantDB()` for the same reason `UpdateStatus` already does.

## Decisions locked from the PRD

| # | Question | Answer |
|---|---|---|
| 1 | Failure representation? | Structured `error_code` enum + raw `error_message` (PRD §4.1). |
| 2 | Transient retry policy? | Counter; escalate to `error` after 3 consecutive transient failures (PRD §4.2). |
| 3 | Status semantics? | `error` = retrying transparently; `disconnected` = needs user action (PRD §4.3). |
| 4 | Reconnect surface? | Inline button in `ConnectionStatus`, not a banner (PRD §4.5). |
| 5 | Reuse existing OAuth endpoint? | Yes — `POST /calendar/connections/google/authorize` with `reauthorize: true` (PRD §4.5). |
| 6 | Detect `invalid_grant` specifically? | Yes — produces `error_code = token_revoked` distinct from generic refresh errors (PRD §4.1). |
| 7 | Backfill existing rows? | No — defaults handle it (PRD §6 migration notes). |
| 8 | Store failure history? | No — only the latest failure is retained (PRD §2 non-goals). |
| 9 | Notifications? | No — failure surfacing is in-app only (PRD §2 non-goals). |
| 10 | Task number? | task-025. |

## Decisions made during plan creation (not in PRD)

| # | Question | Decision | Rationale |
|---|---|---|---|
| P1 | How to clear error state on successful reauthorize: extend `UpdateTokensAndWriteAccess` or add a `ClearErrorState` call after it? | Extend `UpdateTokensAndWriteAccess` to also clear error fields in the same write. Still add `ClearErrorState` as a primitive for completeness. | One fewer DB round trip; avoids an "intermediate" state where tokens are updated but error fields linger. See plan §3.6. |
| P2 | How to detect `invalid_grant` — status code or body? | Parse the JSON body's `error` field. | Google returns HTTP **400** with `{"error":"invalid_grant",...}` for revoked refresh tokens, not 401. Status code alone is insufficient. |
| P3 | Counter increment: read-then-write or atomic SQL? | Single atomic `UPDATE` with `CASE`/`GREATEST` expressions. | Eliminates the read-modify-write race entirely instead of relying on the sync loop's serial behavior. Future-proof against concurrent failure writers. See plan §3.4. |
| P7 | `error_message` length cap? | Truncate to 500 chars in `RecordSyncFailure` before storing. | Bounds row size and prevents any chance of a token leaking from a malformed upstream body. |
| P4 | Sync button on `error` rows: enabled or disabled? | Disabled (preserve existing `attrs.status !== "connected"` check). | Safer default — user must Reconnect to manually retry, avoiding rate-limit toast spam. The system is still retrying on its schedule. See plan §6.3 / risk R8. |
| P5 | Use `gorm.Expr("NULL")` or nil pointers in the `Updates(map)` to clear columns? | `gorm.Expr("NULL")`. | Version-stable across GORM v1/v2; explicitly nullifies regardless of map type quirks. |
| P6 | Where does `errorCodeToMessage` live? | New file `frontend/src/components/features/calendar/error-code-message.ts`. | PRD §7 explicitly suggests this; small enough to be inline but a dedicated file is greppable and unit-testable. |

## Useful files for the implementer

### Backend
- `services/calendar-service/internal/connection/entity.go` — schema additions.
- `services/calendar-service/internal/connection/model.go` + `builder.go` — model fields and getters.
- `services/calendar-service/internal/connection/administrator.go` — new write helpers.
- `services/calendar-service/internal/connection/processor.go` — new processor methods + threshold constant.
- `services/calendar-service/internal/connection/rest.go` — JSON:API attribute additions.
- `services/calendar-service/internal/connection/resource.go:96-192` — callback handler (verify, no expected change after Task 3.6).
- `services/calendar-service/internal/sync/sync.go:82-149` — primary classification + write site.
- `services/calendar-service/internal/googlecal/client.go:88-112,337-353` — `TokenRefreshError` addition + `RefreshToken` restructure.
- `services/calendar-service/internal/crypto/` — `ErrDecryptFailed` sentinel.
- `services/calendar-service/internal/connection/processor_test.go` — pattern for new processor tests.

### Frontend
- `frontend/src/types/models/calendar.ts` — type extension.
- `frontend/src/components/features/calendar/connection-status.tsx` — primary UI change.
- `frontend/src/components/features/calendar/error-code-message.ts` — new helper file.
- `frontend/src/lib/hooks/api/use-calendar.ts:163-174` — verify `useReauthorizeCalendar` is reusable as-is.

### Reference
- `docs/tasks/task-025-calendar-reconnect-ux/prd.md` — full requirements.
- `docs/tasks/task-025-calendar-reconnect-ux/data-model.md` — schema + state transitions table.
- `docs/tasks/task-025-calendar-reconnect-ux/ux-flow.md` — UI states and reconnect flow.
- `docs/tasks/task-009-calendar-sync/` — original calendar sync work, for historical context.

## Known unknowns

- Exact HTTP status code Google uses for non-`invalid_grant` refresh failures (401 vs 400 vs 5xx). The plan's classification rules cover all three branches; if Google ever returns a 4xx other than 400/401 with no recognized OAuth error, it falls into `unknown` (transient) which is the safe default.
- Whether `crypto.Encryptor` already exposes a sentinel error today. Plan §2.3 adds one if missing.
- Whether `frontend/src/lib/` has an existing relative-time formatter to reuse. Worth a quick grep before writing one inline in `connection-status.tsx`.
