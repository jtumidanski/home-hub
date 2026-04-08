# Plan — Task 025: Calendar Reconnect UX

Last Updated: 2026-04-08

## Executive Summary

Today, when the calendar background sync can't refresh a Google access token, `sync.Engine.syncOne` flips the connection's `status` to `disconnected` (`services/calendar-service/internal/sync/sync.go:93-97`) and the only on-page artifact is a bare red `disconnected` badge in `ConnectionStatus` (`frontend/src/components/features/calendar/connection-status.tsx:30-34`). There is no reason text, no in-app recovery action, and no way to distinguish a transient network blip from a hard `invalid_grant`. The existing `ReauthorizeBanner` only fires for write-access upgrades and ignores broken connections entirely.

This task closes that gap end-to-end across calendar-service and the frontend:

1. **Schema** — five new columns on `calendar_connections` (`error_code`, `error_message`, `last_error_at`, `last_sync_attempt_at`, `consecutive_failures`), added via `AutoMigrate` to match the existing `connection.Migration` pattern.
2. **Classification** — token refresh failures in `googlecal.Client.RefreshToken` are wrapped in a typed error so `sync.Engine` can distinguish `invalid_grant`, non-`invalid_grant` 401s, transport/5xx errors, and decrypt failures, mapping each to one of five `error_code` enum values.
3. **State machine** — new processor methods (`RecordSyncAttempt`, `RecordSyncSuccess`, `RecordSyncFailure`, `ClearErrorState`) replace `UpdateSyncInfo` / `UpdateStatus` on the sync hot path, implementing the transient counter (3 strikes → `error`) and immediate-disconnect-on-hard-failure rules.
4. **API surface** — additive JSON:API attributes on `calendar-connections` (`errorCode`, `errorMessage`, `lastErrorAt`, `lastSyncAttemptAt`, `consecutiveFailures`).
5. **Callback handler** — extends `callbackHandler` reauthorize branch in `connection/resource.go:152-173` to call `ClearErrorState` so a successful re-OAuth dance fully resets the row.
6. **Frontend** — extends `CalendarConnectionAttributes`, adds an `errorCodeMessage` mapping, renders the failure context inline in `ConnectionStatus`, and exposes a "Reconnect" / "Reconnect anyway" button that reuses `useReauthorizeCalendar`.

The PRD's non-goals are honored: no notifications, no OAuth-flow rework, no historical failure log, no backfill, no automated reconnect.

## Current State Analysis

### Backend — what already exists

- **`connection.Entity`** (`services/calendar-service/internal/connection/entity.go:10-37`) — GORM-managed via `Migration` → `db.AutoMigrate(&Entity{})`. Adding columns is a simple struct extension.
- **`connection.Model`** (`services/calendar-service/internal/connection/model.go:14-50`) — immutable; getters only. Backed by `Builder` (`builder.go`) with chainable setters. New fields follow the same shape.
- **`connection.Processor`** (`services/calendar-service/internal/connection/processor.go:27-139`) — exposes `UpdateStatus`, `UpdateTokens`, `UpdateSyncInfo`, etc. All writes go through small functions in `administrator.go`. The processor uses `noTenantDB()` for sync-engine paths because the engine runs without a tenant context.
- **`connection.administrator.go`** — `updateStatus`, `updateTokens`, `updateSyncInfo`, `updateTokensAndWriteAccess` are tiny `db.Model(&Entity{}).Where(...).Updates(map)` helpers. New write helpers will follow the same pattern.
- **`connection.RestModel` / `Transform`** (`connection/rest.go:9-43`) — flat struct, JSON-tagged. Adding fields is mechanical.
- **`callbackHandler`** (`connection/resource.go:96-192`) — already handles the reauthorize branch (lines 152-173) calling `UpdateTokensAndWriteAccess`, which currently *also* writes `status = 'connected'` (`administrator.go:65`). It does NOT clear error state because no error state exists yet.
- **`sync.Engine.syncOne`** (`sync/sync.go:82-117`) — calls `getValidAccessToken`, on error logs and writes `status = 'disconnected'`. On success calls `UpdateSyncInfo(id, totalEvents)`.
- **`sync.Engine.getValidAccessToken`** (`sync/sync.go:119-149`) — sequential decrypt → expiry check → decrypt refresh → `gcClient.RefreshToken` → encrypt → `UpdateTokens`. Each step returns an unwrapped `error` that the caller cannot classify.
- **`googlecal.Client.RefreshToken`** (`googlecal/client.go:88-112`) — returns `fmt.Errorf("token refresh returned %d: %s", resp.StatusCode, string(body))` for non-200 responses. The HTTP body for revocation is the standard Google `{"error":"invalid_grant", ...}`. The status code on `invalid_grant` is `400 Bad Request`, not 401 — confirmed below.

### Backend — what is missing

- No way to record *why* a connection broke beyond what's in logs.
- No way to count consecutive failures.
- No way to record an attempt timestamp distinct from last success.
- No typed error from `googlecal` to distinguish `invalid_grant` from a transport hiccup.
- No path on the OAuth callback to clear failure context (not needed today because there is none).

### Frontend — what already exists

- **`CalendarConnectionAttributes`** (`frontend/src/types/models/calendar.ts:1-11`) — declares `status` as `"connected" | "disconnected" | "syncing" | "error"`. The `error` value is in the type but no backend code currently writes it.
- **`ConnectionStatus`** (`frontend/src/components/features/calendar/connection-status.tsx`) — renders one row per connection with: color dot, status badge (only when `!isHealthy`), display name, last-sync timestamp, sync button, disconnect button. The status badge uses `attrs.status` as its raw label — no human messaging.
- **`useReauthorizeCalendar`** (`frontend/src/lib/hooks/api/use-calendar.ts:163-174`) — already wired to `calendarService.reauthorizeGoogle` and redirects to the returned auth URL on success. **Reusable as-is for the new Reconnect button.**
- **`ReauthorizeBanner`** (`frontend/src/components/features/calendar/reauthorize-banner.tsx`) — separate component for the write-access upgrade flow. Out of scope for this task and remains unchanged per PRD §4.5.

### Frontend — what is missing

- No `errorCode`/`errorMessage`/`lastErrorAt`/`lastSyncAttemptAt`/`consecutiveFailures` on the type.
- No human-readable mapping of error codes to user messages (PRD §4.6).
- No Reconnect affordance on `disconnected` or `error` rows.
- No surfacing of last sync *attempt* vs last *success* timestamps.

### Verified facts that drive the plan

- `gcClient` has no existing typed error pattern for refresh failures. The closest analog is `googlecal.SyncTokenInvalidError` (`client.go:337-353`), used by `doWithRetry` for HTTP 410. We will mirror this pattern with a `TokenRefreshError` (or similar) carrying the HTTP status and the parsed Google error code.
- Google's `invalid_grant` returns HTTP **400** with body `{"error":"invalid_grant","error_description":"..."}`. Confirmed by Google OAuth docs and consistent with the way `RefreshToken` reads `resp.Body` on non-200. Classification must therefore parse the JSON body, not branch solely on status code.
- `sync.Engine.syncOne` runs from `e.l` (engine logger), which is `noTenantDB`-friendly. All new processor methods must use the same `noTenantDB()` helper that `UpdateStatus`/`UpdateSyncInfo` already use.
- The PRD's `consecutive_failures` semantics — "set to max(current, 3) on hard failure" — keeps the UI state machine simple: any row with `status in ('error','disconnected')` and `consecutive_failures >= 3` is a "first-class failed" row.

## Proposed Future State

### Backend

- `calendar_connections` has the five new columns; `Migration` runs cleanly on empty and populated databases.
- `connection.Model` exposes `ErrorCode() *string`, `ErrorMessage() *string`, `LastErrorAt() *time.Time`, `LastSyncAttemptAt() *time.Time`, `ConsecutiveFailures() int` getters and matching builder setters.
- `connection.Processor` has four new methods replacing the failure-path use of `UpdateStatus`/`UpdateSyncInfo` on the sync hot path:
  - `RecordSyncAttempt(id, at) error` — single column write to `last_sync_attempt_at` + `updated_at`. Called at the top of `syncOne`.
  - `RecordSyncSuccess(id, eventCount, at) error` — sets `last_sync_at`, `last_sync_attempt_at`, `last_sync_event_count`; clears `error_code`, `error_message`, `last_error_at`; resets `consecutive_failures` to 0; sets `status = 'connected'`.
  - `RecordSyncFailure(id, code, message, at) error` — increments `consecutive_failures`; sets `error_code`, `error_message`, `last_error_at`; transitions status per the rules table. Hard codes (`token_revoked`, `refresh_unauthorized`, `token_decrypt_failed`) bypass the counter and force `status = 'disconnected'` + `consecutive_failures = max(current, 3)`. Transient codes (`refresh_http_error`, `unknown`) only update `status` to `'error'` once the counter reaches `failureEscalationThreshold = 3`.
  - `ClearErrorState(id) error` — sets `status = 'connected'`; clears `error_code`, `error_message`, `last_error_at`; resets `consecutive_failures` to 0. Called from the OAuth callback reauthorize branch.
- `UpdateStatus` and `UpdateSyncInfo` are kept for backwards compatibility with non-sync code paths, but the sync engine no longer calls them directly.
- A constant `connection.FailureEscalationThreshold = 3` lives in `processor.go` and is referenced from both the processor and tests.
- `googlecal.TokenRefreshError` is a new typed error (in `googlecal/types.go` or alongside `SyncTokenInvalidError` in `client.go`) carrying:
  ```go
  type TokenRefreshError struct {
      StatusCode int
      OAuthError string // parsed from JSON body's "error" field; "" if unparseable
      Body       string
  }
  ```
  with a helper `func IsInvalidGrant(err error) bool`. `Client.RefreshToken` returns this for any non-200 instead of the current `fmt.Errorf`. Transport errors (the existing `c.httpClient.PostForm` path) remain plain wrapped errors and are classified by the caller as `refresh_http_error`.
- `sync.Engine.syncOne` is restructured:
  1. At entry, calls `RecordSyncAttempt(conn.Id(), now)`.
  2. Calls a new `e.classifyAndRecordTokenError(ctx, conn, err)` helper if `getValidAccessToken` fails. The helper:
     - Decrypt failure → `token_decrypt_failed` (hard).
     - `googlecal.IsInvalidGrant(err)` → `token_revoked` (hard).
     - `*TokenRefreshError` with status 401 → `refresh_unauthorized` (hard).
     - `*TokenRefreshError` with status 5xx or transport error → `refresh_http_error` (transient).
     - Otherwise → `unknown` (transient).
  3. On success, calls `RecordSyncSuccess(conn.Id(), totalEvents, now)` instead of `UpdateSyncInfo`.
  4. The internal `connProc.UpdateTokens` call inside `getValidAccessToken` is unchanged — it still persists the new access token on a successful refresh.
- `getValidAccessToken` itself returns the raw error; classification is the caller's job. This keeps the helper synchronous and avoids leaking processor state into a leaf function.
- `connection.callbackHandler` reauthorize branch calls `proc.ClearErrorState(existing.Id())` after `UpdateTokensAndWriteAccess` succeeds (or, alternatively, `UpdateTokensAndWriteAccess` is extended to clear error state in the same write — see Phase 4 task 4.3 for the chosen approach).
- `RestModel` / `Transform` expose the five new attributes, named per PRD §5 (`errorCode`, `errorMessage`, `lastErrorAt`, `lastSyncAttemptAt`, `consecutiveFailures`).

### Frontend

- `CalendarConnectionAttributes` adds: `errorCode: string | null`, `errorMessage: string | null`, `lastErrorAt: string | null`, `lastSyncAttemptAt: string | null`, `consecutiveFailures: number`.
- New helper `frontend/src/components/features/calendar/error-code-message.ts` exports `errorCodeToMessage(code: string | null): string | null` implementing the PRD §4.6 mapping. Returns `null` for healthy connections so the caller can branch cleanly.
- `ConnectionStatus`:
  - Computes a `failureMessage` from `errorCode` (with a "(no code) → 'This calendar is disconnected.'" fallback for legacy `disconnected` rows that have null `errorCode`).
  - When status is `error` or `disconnected`, renders a second line with the message and (when `lastSyncAttemptAt !== lastSyncAt`) a "Tried X ago · Last success Y ago" subline.
  - Renders a `Reconnect` button (primary style on `disconnected`, tertiary "Reconnect anyway" on `error`) wired to `useReauthorizeCalendar` with `window.location.origin + "/app/calendar"` as the redirect URI.
  - Hides the manual sync button on `disconnected` (it cannot succeed); leaves it enabled on `error` (the system is still retrying and the user may want to nudge it).
  - Status badge color: red for `disconnected`, amber for `error`, no badge for healthy.
- The existing `Disconnect` button remains visible in all non-healthy states.

### Out of scope

- All PRD §2 non-goals: notifications, OAuth flow rework, multi-provider, history log, automated reconnect, backfill.
- No changes to `event` package, `source` package, `googlecal` event/list paths.
- No new endpoints. No changes to URL contracts on `/calendar/connections/google/authorize` or `/callback`.
- No changes to `ReauthorizeBanner` or the write-access upgrade flow.

## Implementation Phases

### Phase 1 — Schema & model

#### Task 1.1 — Add columns to `connection.Entity` (S)

Edit `services/calendar-service/internal/connection/entity.go:10-28` to add:

```go
ErrorCode          *string    `gorm:"type:varchar(40)"`
ErrorMessage       *string    `gorm:"type:text"`
LastErrorAt        *time.Time `gorm:""`
LastSyncAttemptAt  *time.Time `gorm:""`
ConsecutiveFailures int       `gorm:"not null;default:0"`
```

Use pointer types for nullable string/time fields to match the existing `LastSyncAt *time.Time` convention.

- **Acceptance:** struct compiles; `Migration` continues to call `AutoMigrate` and adds the columns on a fresh DB and an existing one.

#### Task 1.2 — Extend `Model`, `Builder`, `ToEntity`, `Make` (S)

Mirror the new fields in:
- `model.go:14-50` — fields + getters: `ErrorCode() *string`, `ErrorMessage() *string`, `LastErrorAt() *time.Time`, `LastSyncAttemptAt() *time.Time`, `ConsecutiveFailures() int`.
- `builder.go:18-93` — fields + chainable setters; include in the `Build()` Model literal. **Do not add validation** for the new fields — they are all optional / default to zero.
- `entity.go:39-81` — `ToEntity` writes them; `Make` reads them through the builder.

- **Acceptance:** `go build ./services/calendar-service/...` passes. `connection/model_test.go` builder coverage extended (Phase 5).

#### Task 1.3 — Verify migration on a populated DB (S)

Run `scripts/local-up.sh` (per memory: always use the local-up/local-down scripts), let calendar-service migrate, confirm via psql that the new columns exist with the documented defaults and existing rows have `consecutive_failures = 0` and NULLs elsewhere.

- **Acceptance:** psql query against `calendar_connections` shows the five new columns; existing seeded connections (if any) load and serialize without errors.

### Phase 2 — Failure classification (`googlecal` + sync engine)

#### Task 2.1 — Add `TokenRefreshError` to `googlecal` (S)

In `services/calendar-service/internal/googlecal/client.go` (next to `SyncTokenInvalidError` at lines 337-353), add:

```go
type TokenRefreshError struct {
    StatusCode int
    OAuthError string
    Body       string
}

func (e *TokenRefreshError) Error() string {
    return fmt.Sprintf("token refresh failed (HTTP %d, oauth_error=%q): %s", e.StatusCode, e.OAuthError, e.Body)
}

func IsInvalidGrant(err error) bool {
    var tre *TokenRefreshError
    if errors.As(err, &tre) {
        return tre.OAuthError == "invalid_grant"
    }
    return false
}
```

Add `errors` to the import block. Replace the non-200 branch in `RefreshToken` (`client.go:102-105`) to:
1. Read the body.
2. Attempt to decode `{"error": "...", "error_description": "..."}` into a tiny anonymous struct. Tolerate decode failure (leave `OAuthError = ""`).
3. Return `&TokenRefreshError{StatusCode: resp.StatusCode, OAuthError: parsed.Error, Body: string(body)}`.

**Security check:** verify the body never contains the refresh token itself before storing it on `error_message`. Google's `invalid_grant` body is the standard `{"error":"invalid_grant","error_description":"Token has been expired or revoked."}` — safe.

- **Acceptance:** `go test ./services/calendar-service/internal/googlecal/...` passes with a new unit test that feeds an `invalid_grant` body and asserts `IsInvalidGrant` returns true. A separate test asserts a 5xx body returns false.

#### Task 2.2 — Wire classification into `sync.Engine.syncOne` (M)

Edit `services/calendar-service/internal/sync/sync.go:82-117`:

1. Add `errors` to imports if not present.
2. Define a new helper:
   ```go
   func (e *Engine) classifyTokenError(err error) (code string, hard bool) {
       if errors.Is(err, crypto.ErrDecryptFailed) { // see Task 2.3
           return "token_decrypt_failed", true
       }
       if googlecal.IsInvalidGrant(err) {
           return "token_revoked", true
       }
       var tre *googlecal.TokenRefreshError
       if errors.As(err, &tre) {
           if tre.StatusCode == http.StatusUnauthorized {
               return "refresh_unauthorized", true
           }
           if tre.StatusCode >= 500 || tre.StatusCode == http.StatusTooManyRequests {
               return "refresh_http_error", false
           }
       }
       // transport / unrecognized
       return "unknown", false
   }
   ```
   Place near the bottom of `sync.go`.
3. At the top of `syncOne` (after the logger setup at line 87), call:
   ```go
   now := time.Now().UTC()
   connProc := connection.NewProcessor(l, ctx, e.db)
   _ = connProc.RecordSyncAttempt(conn.Id(), now)
   ```
4. Replace the `getValidAccessToken` error branch (lines 91-97) with:
   ```go
   accessToken, err := e.getValidAccessToken(ctx, conn)
   if err != nil {
       code, _ := e.classifyTokenError(err)
       l.WithError(err).WithField("error_code", code).Warn("token refresh failed during sync")
       _ = connProc.RecordSyncFailure(conn.Id(), code, err.Error(), now)
       return
   }
   ```
   `RecordSyncFailure` itself encapsulates the hard-vs-transient logic and the counter increment.
5. Replace the success-path call to `UpdateSyncInfo` (line 115) with `RecordSyncSuccess(conn.Id(), totalEvents, now)`.

- **Acceptance:** unit test (Phase 5) feeds a fake `gcClient` that returns an `invalid_grant` `TokenRefreshError` and asserts the connection ends up `disconnected` with `error_code = token_revoked` and `consecutive_failures >= 3` after one cycle. A separate test feeds a transport error twice and asserts `status = connected, consecutive_failures = 2`, then a third call flips to `error`.

#### Task 2.3 — Distinguish decrypt failures from generic errors (S)

`crypto.Encryptor.Decrypt` (in `services/calendar-service/internal/crypto/`) currently returns plain errors. Add a sentinel `var ErrDecryptFailed = errors.New("decrypt failed")` and wrap the existing failure return with `fmt.Errorf("%w: %v", ErrDecryptFailed, err)` so callers can classify via `errors.Is`.

- **Acceptance:** `crypto` unit test asserts `errors.Is(err, crypto.ErrDecryptFailed)` for a corrupted ciphertext input.

### Phase 3 — Processor write methods

#### Task 3.1 — Add `failureEscalationThreshold` constant (S)

In `connection/processor.go`, add:
```go
const FailureEscalationThreshold = 3
```
Export so tests can reference it.

- **Acceptance:** declared in `processor.go`; referenced by `RecordSyncFailure` and tests (Task 5.1).

#### Task 3.2 — `RecordSyncAttempt` (S)

Add to `processor.go`:
```go
func (p *Processor) RecordSyncAttempt(id uuid.UUID, at time.Time) error {
    return updateSyncAttempt(p.noTenantDB(), id, at)
}
```

Add to `administrator.go`:
```go
func updateSyncAttempt(db *gorm.DB, id uuid.UUID, at time.Time) error {
    return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
        "last_sync_attempt_at": at,
        "updated_at":           at,
    }).Error
}
```

- **Acceptance:** processor test seeds a row, calls `RecordSyncAttempt`, reads the row, asserts `LastSyncAttemptAt` is non-nil and ≈ `at`.

#### Task 3.3 — `RecordSyncSuccess` (M)

Add to `processor.go`:
```go
func (p *Processor) RecordSyncSuccess(id uuid.UUID, eventCount int, at time.Time) error {
    return updateSyncSuccess(p.noTenantDB(), id, eventCount, at)
}
```

Add to `administrator.go`:
```go
func updateSyncSuccess(db *gorm.DB, id uuid.UUID, eventCount int, at time.Time) error {
    return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
        "status":                "connected",
        "last_sync_at":          at,
        "last_sync_attempt_at":  at,
        "last_sync_event_count": eventCount,
        "error_code":            gorm.Expr("NULL"),
        "error_message":         gorm.Expr("NULL"),
        "last_error_at":         gorm.Expr("NULL"),
        "consecutive_failures":  0,
        "updated_at":            at,
    }).Error
}
```

(Use `gorm.Expr("NULL")` to force the column to NULL — `nil` in a `map[string]interface{}` works in some GORM versions but `Expr("NULL")` is unambiguous.)

- **Acceptance:** processor test seeds an `error`-state row with non-zero `consecutive_failures` and a populated `error_code`, calls `RecordSyncSuccess`, asserts: status `connected`, all error fields cleared, counter at 0, both timestamps updated, event count set.

#### Task 3.4 — `RecordSyncFailure` (M)

Add to `processor.go`:
```go
const errorMessageMaxLen = 500

func (p *Processor) RecordSyncFailure(id uuid.UUID, code, message string, at time.Time) error {
    if len(message) > errorMessageMaxLen {
        message = message[:errorMessageMaxLen]
    }
    hard := isHardErrorCode(code)
    return updateSyncFailure(p.noTenantDB(), id, code, message, at, hard)
}

func isHardErrorCode(code string) bool {
    switch code {
    case "token_revoked", "refresh_unauthorized", "token_decrypt_failed":
        return true
    }
    return false
}
```

The 500-character cap on `message` is a paranoia bound — Google's `invalid_grant` body is small and well-formed, but 5xx HTML bodies and arbitrary transport errors can be long. Truncating before storage avoids any chance of an oversize row or a token leaking from a malformed upstream response.

Add to `administrator.go`:
```go
func updateSyncFailure(db *gorm.DB, id uuid.UUID, code, message string, at time.Time, hard bool) error {
    // Single atomic UPDATE with CASE expressions so the increment,
    // threshold check, and status transition happen in one SQL statement.
    // No SELECT-then-UPDATE race window.
    var statusExpr clause.Expression
    var counterExpr clause.Expression
    if hard {
        statusExpr = gorm.Expr("'disconnected'")
        counterExpr = gorm.Expr("GREATEST(consecutive_failures + 1, ?)", FailureEscalationThreshold)
    } else {
        statusExpr = gorm.Expr(
            "CASE WHEN consecutive_failures + 1 >= ? THEN 'error' ELSE status END",
            FailureEscalationThreshold,
        )
        counterExpr = gorm.Expr("consecutive_failures + 1")
    }
    return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
        "status":               statusExpr,
        "error_code":           code,
        "error_message":        message,
        "last_error_at":        at,
        "last_sync_attempt_at": at,
        "consecutive_failures": counterExpr,
        "updated_at":           at,
    }).Error
}
```

This relies on PostgreSQL's `GREATEST` and `CASE` (already what `calendar-service` runs on). The single-statement form is race-free even if a future change introduces concurrent failure writers — no read-modify-write window exists. Add `"gorm.io/gorm/clause"` to the import block if not already present.

`FailureEscalationThreshold` is referenced from `administrator.go` (same package as `processor.go`, no import needed).

- **Acceptance:** processor tests cover the matrix from PRD §6 / data-model.md §"Field interaction rules":
  - Transient failure on a healthy row → counter=1, status unchanged.
  - Transient failure on a row with counter=2 → counter=3, status=`error`.
  - Hard failure on a healthy row → counter=3 (escalated), status=`disconnected`.
  - Hard failure on a row with counter=5 → counter=5 (preserved), status=`disconnected`.

#### Task 3.5 — `ClearErrorState` (S)

Add to `processor.go`:
```go
func (p *Processor) ClearErrorState(id uuid.UUID) error {
    return clearErrorState(p.noTenantDB(), id)
}
```

Add to `administrator.go`:
```go
func clearErrorState(db *gorm.DB, id uuid.UUID) error {
    return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
        "status":               "connected",
        "error_code":           gorm.Expr("NULL"),
        "error_message":        gorm.Expr("NULL"),
        "last_error_at":        gorm.Expr("NULL"),
        "consecutive_failures": 0,
        "updated_at":           time.Now().UTC(),
    }).Error
}
```

- **Acceptance:** processor test seeds a `disconnected` row with populated error fields, calls `ClearErrorState`, asserts everything cleared and status is `connected`.

#### Task 3.6 — Decide: extend `UpdateTokensAndWriteAccess` or call `ClearErrorState` from the callback (S)

Two options:
- (A) Extend `updateTokensAndWriteAccess` (`administrator.go:59-68`) to also clear error fields and reset counter in the same write. One DB round trip, one transaction-equivalent.
- (B) Leave `updateTokensAndWriteAccess` alone; have `callbackHandler` call `ClearErrorState` immediately after.

**Choose (A).** It is one fewer round trip and avoids an "intermediate" DB state where tokens are updated but error fields linger. The existing `updateTokensAndWriteAccess` call already sets `status = 'connected'`, so adding the error-field clears is a one-line change and consistent with the function's role of "this connection is now healthy and authorized."

Edit `administrator.go:59-68`:
```go
func updateTokensAndWriteAccess(db *gorm.DB, id uuid.UUID, accessToken, refreshToken string, tokenExpiry time.Time, writeAccess bool) error {
    return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
        "access_token":         accessToken,
        "refresh_token":        refreshToken,
        "token_expiry":         tokenExpiry,
        "write_access":         writeAccess,
        "status":               "connected",
        "error_code":           gorm.Expr("NULL"),
        "error_message":        gorm.Expr("NULL"),
        "last_error_at":        gorm.Expr("NULL"),
        "consecutive_failures": 0,
        "updated_at":           time.Now().UTC(),
    }).Error
}
```

`ClearErrorState` is **still added** (Task 3.5) because it's referenced from the data-model spec and is the right primitive if a future code path needs to clear error state without rotating tokens.

- **Acceptance:** integration test of the callback path confirms a previously `disconnected`-with-error row is fully reset after the reauthorize branch runs.

### Phase 4 — JSON:API surface & callback

#### Task 4.1 — Extend `RestModel` and `Transform` (S)

Edit `connection/rest.go:9-43`:

```go
type RestModel struct {
    Id                  uuid.UUID  `json:"-"`
    Provider            string     `json:"provider"`
    Status              string     `json:"status"`
    Email               string     `json:"email"`
    UserDisplayName     string     `json:"userDisplayName"`
    UserColor           string     `json:"userColor"`
    WriteAccess         bool       `json:"writeAccess"`
    LastSyncAt          *time.Time `json:"lastSyncAt"`
    LastSyncAttemptAt   *time.Time `json:"lastSyncAttemptAt"`
    LastSyncEventCount  int        `json:"lastSyncEventCount"`
    ErrorCode           *string    `json:"errorCode"`
    ErrorMessage        *string    `json:"errorMessage"`
    LastErrorAt         *time.Time `json:"lastErrorAt"`
    ConsecutiveFailures int        `json:"consecutiveFailures"`
    CreatedAt           time.Time  `json:"createdAt"`
}
```

Extend `Transform` to populate the new fields from `m.ErrorCode()`, etc.

- **Acceptance:** JSON:API `GET /calendar/connections` includes the new attributes. Existing frontend consumers ignore unknown attributes — backwards compatible.

#### Task 4.2 — Verify callback handler reset (S)

After Task 3.6, no code change is required in `callbackHandler` — `UpdateTokensAndWriteAccess` already clears the error fields. **Verify** by re-reading `connection/resource.go:152-173` and confirming the reauthorize branch flow:
1. `UpdateTokensAndWriteAccess(...)` (now also clears error state).
2. `ByIDProvider(existing.Id())` reload.
3. `syncTrigger(updatedConn)` to immediately retry.

If anything in this flow needs adjusting (e.g., the immediate `syncTrigger` should also be expected to overwrite `last_sync_attempt_at`), document it. Expected: no change required.

- **Acceptance:** PR description records the verification result. No code diff in `resource.go` for this task.

### Phase 5 — Backend tests

#### Task 5.1 — Processor unit tests (M)

Extend `services/calendar-service/internal/connection/processor_test.go` with:
- `TestRecordSyncAttempt_SetsTimestamp`
- `TestRecordSyncSuccess_ClearsErrorAndResetsCounter`
- `TestRecordSyncFailure_TransientUnderThresholdLeavesStatus`
- `TestRecordSyncFailure_TransientReachesThresholdSetsError`
- `TestRecordSyncFailure_HardFailureSetsDisconnectedAndForcesCounter`
- `TestRecordSyncFailure_HardFailurePreservesHigherCounter`
- `TestClearErrorState_ResetsAllFields`
- `TestUpdateTokensAndWriteAccess_ClearsErrorState` (regression coverage for Task 3.6)

Use the existing test harness pattern in `processor_test.go` (sqlite or test container — match what's there).

- **Acceptance:** `go test ./services/calendar-service/internal/connection/...` passes with all new tests visibly named.

#### Task 5.2 — Classification unit tests (S)

In a new file `services/calendar-service/internal/sync/classify_test.go` (or extend an existing sync test):
- `TestClassifyTokenError_InvalidGrantHard`
- `TestClassifyTokenError_Unauthorized401Hard`
- `TestClassifyTokenError_5xxTransient`
- `TestClassifyTokenError_TransportTransient`
- `TestClassifyTokenError_DecryptFailedHard`

Each constructs the appropriate error and asserts the `(code, hard)` tuple.

- **Acceptance:** `go test ./services/calendar-service/internal/sync/...` passes.

#### Task 5.3 — `googlecal` parsing tests (S)

In `services/calendar-service/internal/googlecal/client_test.go` (create if missing):
- `TestRefreshToken_InvalidGrantReturnsTypedError` — uses `httptest.NewServer` to return a 400 with body `{"error":"invalid_grant","error_description":"Token has been expired or revoked."}`. Asserts `errors.As` finds a `*TokenRefreshError` with `OAuthError == "invalid_grant"` and `IsInvalidGrant(err)` returns true.
- `TestRefreshToken_5xxReturnsTypedErrorWithoutOAuthCode` — server returns 500 with HTML body. Asserts `OAuthError == ""` and `StatusCode == 500`.
- `TestIsInvalidGrant_NilFalse`.

- **Acceptance:** `go test ./services/calendar-service/internal/googlecal/...` passes.

### Phase 6 — Frontend

#### Task 6.1 — Extend the type (S)

Edit `frontend/src/types/models/calendar.ts:1-11`:
```ts
export interface CalendarConnectionAttributes {
  provider: string;
  status: "connected" | "disconnected" | "syncing" | "error";
  email: string;
  userDisplayName: string;
  userColor: string;
  writeAccess: boolean;
  lastSyncAt: string | null;
  lastSyncAttemptAt: string | null;
  lastSyncEventCount: number;
  errorCode: string | null;
  errorMessage: string | null;
  lastErrorAt: string | null;
  consecutiveFailures: number;
  createdAt: string;
}
```

- **Acceptance:** `npm run build` (or the project's tsc check) passes.

#### Task 6.2 — Add `error-code-message.ts` helper (S)

New file `frontend/src/components/features/calendar/error-code-message.ts`:

```ts
const MESSAGES: Record<string, string> = {
  token_revoked:        "Access was revoked from your Google account. Reconnect to resume syncing.",
  refresh_unauthorized: "Google rejected your saved credentials. Reconnect to resume syncing.",
  token_decrypt_failed: "We can't read your stored credentials. Reconnect to resume syncing.",
  refresh_http_error:   "Couldn't reach Google. Retrying automatically.",
  unknown:              "Sync is failing. Retrying automatically.",
};

export function errorCodeToMessage(code: string | null): string | null {
  if (!code) return null;
  return MESSAGES[code] ?? null;
}
```

- **Acceptance:** unit test or visual smoke confirms each documented code maps; unknown codes return `null`.

#### Task 6.3 — Extend `ConnectionStatus` to render failure context + Reconnect (M)

Edit `frontend/src/components/features/calendar/connection-status.tsx`:

1. Import `useReauthorizeCalendar` and `errorCodeToMessage`.
2. Compute:
   ```ts
   const isHealthy = attrs.status === "connected" || attrs.status === "syncing";
   const isDisconnected = attrs.status === "disconnected";
   const isError = attrs.status === "error";
   const failureMessage =
     errorCodeToMessage(attrs.errorCode) ??
     (isDisconnected ? "This calendar is disconnected." : null);
   ```
3. Status badge:
   ```tsx
   {isError && (
     <Badge variant="outline" className="bg-amber-500/10 text-amber-700 border-amber-200">
       Sync issues
     </Badge>
   )}
   {isDisconnected && (
     <Badge variant="outline" className="bg-red-500/10 text-red-700 border-red-200">
       Disconnected
     </Badge>
   )}
   ```
4. When `failureMessage` is non-null, render a second-line block beneath the existing row containing:
   - The failure message.
   - A "Tried X ago · Last success Y ago" subline when both timestamps are present and `lastSyncAttemptAt !== lastSyncAt`. Use a small relative-time formatter (extract from existing code if one exists in `frontend/src/lib/`; if not, write a one-off helper inline with `Intl.RelativeTimeFormat`).
   - The Reconnect button:
     ```tsx
     <Button
       variant={isDisconnected ? "default" : "outline"}
       size="sm"
       onClick={() => reauthorize.mutate(window.location.origin + "/app/calendar")}
       disabled={reauthorize.isPending}
     >
       {isDisconnected ? "Reconnect" : "Reconnect anyway"}
     </Button>
     ```
5. Hide the manual sync button when `isDisconnected` (but leave it on `isError`).
6. The existing layout uses a single horizontal `flex` row. The failure block needs to wrap below the row — restructure the outer container into a `flex flex-col gap-1` with the existing horizontal row as the first child and the failure block as the second.

Match mobile UI preferences from memory: tap-only, no swipe gestures. The Reconnect button is a tap target — keep it sized like the existing buttons.

- **Acceptance:** visual check in dev — a `disconnected` row shows red badge, message, timestamps, prominent Reconnect; an `error` row shows amber badge, message, timestamps, tertiary "Reconnect anyway"; healthy rows are visually unchanged.

#### Task 6.4 — Confirm `useReauthorizeCalendar` is sufficient (S)

Re-read `frontend/src/lib/hooks/api/use-calendar.ts:163-174`. The hook already POSTs to `/calendar/connections/google/authorize` with `reauthorize: true` semantics (verify by reading `calendarService.reauthorizeGoogle` in `frontend/src/services/api/calendar.ts`). If it does, no change is needed.

- **Acceptance:** PR description records that the hook is reused as-is; no code change.

### Phase 7 — Verification & ship

#### Task 7.1 — Build and test all affected services (S)

Per CLAUDE.md, run for calendar-service and the frontend:
- `cd services/calendar-service && go build ./... && go test ./...`
- `cd frontend && npm run build` (or the existing CI command)
- If any shared library was touched (none anticipated), verify Docker builds.

- **Acceptance:** all green.

#### Task 7.2 — Local end-to-end smoke (S)

Per memory, use `scripts/local-up.sh`. Steps:
1. Bring up the stack.
2. Connect a Google calendar through the existing flow.
3. Manually corrupt the row's `refresh_token` in psql to trigger a decrypt failure on the next sync.
4. Confirm: row flips to `disconnected`, `error_code = token_decrypt_failed`, `consecutive_failures >= 3`, frontend renders the red badge + message + Reconnect button.
5. Click Reconnect, complete the OAuth dance, confirm the row resets to `connected` with all error fields cleared.
6. Repeat with a forced 500 from a stub (or skip if too costly — at minimum verify the transient counter via a unit test).

- **Acceptance:** documented in `notes.md` if created, otherwise in the PR description.

#### Task 7.3 — PR (S)

Open PR with:
- Title referencing task-025.
- Body summarizing the schema additions, classification taxonomy, processor methods, frontend changes.
- Test plan checklist mirroring PRD §10.

- **Acceptance:** PR opened, links to PRD.

## Risk Assessment & Mitigation

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------|--------|------------|
| R1 | Counter increment / status transition race | Eliminated | — | `updateSyncFailure` uses a single atomic `UPDATE` with `CASE`/`GREATEST` expressions; no SELECT-then-UPDATE window. See Task 3.4. |
| R2 | `gorm.Expr("NULL")` behaves differently across GORM v1/v2 versions and silently fails to nullify the columns | Low | Error fields stay populated after recovery | Test coverage in Task 5.1 (`TestRecordSyncSuccess_ClearsErrorAndResetsCounter`) reads back the row and asserts `nil`. Alternative: use `*string`/`*time.Time` zero pointers in the map — but `gorm.Expr` is the explicit, version-stable form. |
| R3 | `error_message` accidentally contains a refresh or access token from a malformed Google response | Low | Sensitive data on disk | `RecordSyncFailure` truncates `message` to 500 chars before storing (Task 3.4). Google's documented `invalid_grant` body is well within that bound; 5xx HTML bodies are clipped. |
| R4 | Adding `last_sync_attempt_at` write at the top of every `syncOne` doubles the per-connection write count | Low | Marginal DB load | The sync loop already writes per connection on success/failure; one extra write per cycle is negligible. PRD §8 explicitly accepts this. |
| R5 | Frontend reuses `useReauthorizeCalendar` and the OAuth state's `Reauthorize` flag forces consent — could surprise users on `error`-status rows where the system was still retrying | Low | UX nit | This is by design (PRD §4.5). The "Reconnect anyway" copy makes the user-initiated nature clear. |
| R6 | Existing frontend renders `error` status as a generic red badge today, but this code path hasn't been exercised because no backend writes `error`. After this task it starts firing — there could be lurking type/visual bugs | Low | Visual glitch on first appearance | Phase 6 explicitly redesigns the badge for `error` (amber, "Sync issues"). Manual smoke in Task 7.2 covers it. |
| R7 | The `failureEscalationThreshold = 3` constant is too low and produces `error` status for completely transient single-cycle blips | Low–Medium | UI noise | The threshold is "3 consecutive failures" — at the default sync cadence this is 15+ minutes of sustained outage, not a single blip. Tunable via the constant if needed; PRD §4.2 explicitly hardcodes it for v1. |
| R8 | `isError` rows still show the manual sync button and a user clicking it repeatedly racks up rate-limited toasts | Low | Annoyance | Existing 5-minute rate limit handles this; the button is `disabled` when not `connected` per existing code (`connection-status.tsx:44`) — keep that disable in place for `error` so the user must Reconnect to manually sync. **Actually:** revisit Task 6.3 step 5 — the spec says "leaves it enabled on `error`" but the safer default is to leave the existing `disabled={... attrs.status !== "connected"}` in place. **Decision: defer to existing behavior — sync button disabled on `error`.** Update Task 6.3 accordingly during implementation. |

## Success Metrics

- **Functional:** A connection that hits `invalid_grant` lands in `disconnected` with `error_code = token_revoked` within one sync cycle, and the frontend immediately shows the revocation message with a Reconnect button.
- **Resilience:** Three consecutive transport errors flip the row to `error` (not `disconnected`); a fourth successful cycle restores it to `connected` with counter at 0.
- **Recoverability:** A user with a `disconnected` row can recover entirely in-app via the Reconnect button without deleting and re-adding the connection.
- **Observability:** Server logs include `error_code` as a structured field on every failure path, enabling dashboard grouping per PRD §8.
- **No regressions:** existing healthy connections continue to sync without behavior changes; `LastSyncAt` continues to mean "last successful sync"; `lastSyncEventCount` continues to mean what it means today.
- **Test coverage:** the eight processor tests from Task 5.1, the five classification tests from Task 5.2, and the three googlecal tests from Task 5.3 all pass.

## Required Resources & Dependencies

- Local dev stack (`scripts/local-up.sh`) with calendar-service, postgres, frontend.
- A real Google OAuth client (already configured for local) for the manual smoke in Task 7.2.
- No new third-party libraries.
- No schema changes outside `calendar_connections`.
- No changes to other services or shared libraries.

## Timeline / Effort

Per CLAUDE.md, no calendar estimates. Effort sizing only:

| Phase | Effort | Notes |
|-------|--------|-------|
| Phase 1 — Schema & model | S | Five mechanical struct edits. |
| Phase 2 — Classification | M | New typed error, sync engine restructure, decrypt sentinel. |
| Phase 3 — Processor methods | M | Four new methods, one extension to existing helper. |
| Phase 4 — JSON:API & callback | S | Additive REST fields; callback already wired via Task 3.6. |
| Phase 5 — Backend tests | M | Eight processor tests + five classify + three googlecal. |
| Phase 6 — Frontend | M | Type, helper, component restructure. |
| Phase 7 — Verify & ship | S | Build, local smoke, PR. |
| **Total** | **L** | Single PR; backend and frontend land together. |

## Out of Scope (per PRD §2 non-goals)

- Push, email, or any out-of-band failure notifications.
- OAuth flow rework, encryption scheme changes, sync scheduler changes.
- Multi-provider support.
- Historical failure log.
- Automated reconnection without user consent.
- Backfilling failure context onto existing rows.
- Changes to `ReauthorizeBanner` (write-access upgrade flow).
