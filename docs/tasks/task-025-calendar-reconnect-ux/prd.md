# Calendar Reconnect UX — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-04-08
---

## 1. Overview

When the calendar background sync fails to refresh a user's Google OAuth tokens, the calendar-service silently flips the connection's status to `disconnected` and logs the error server-side. The frontend renders a bare red "disconnected" badge with no explanation, and the only existing reconnect path (`ReauthorizeBanner`) triggers exclusively for write-access upgrades on still-connected calendars. The user has no way to learn *why* their calendar broke and no in-app way to fix it — they must delete the connection and re-add it from scratch.

This task closes that gap end-to-end. The backend will persist a structured failure reason and a consecutive-failure counter on the connection, distinguishing transient sync errors from hard authorization failures. The status enum's existing `error` and `disconnected` values will be given clear, distinct semantics. The frontend will surface the failure context next to the affected connection and offer a one-click "Reconnect" action that reuses the existing OAuth authorize flow.

The result: users see what's wrong, know whether the system is retrying on its own, and can recover without administrative steps.

## 2. Goals

Primary goals:
- Persist structured failure context (`error_code`, `error_message`, `last_error_at`, `consecutive_failures`) on each calendar connection.
- Distinguish transient sync failures (status `error`, system keeps retrying) from hard auth failures (status `disconnected`, requires user action).
- Surface a human-readable failure reason in the calendar connection UI.
- Provide an in-app "Reconnect" action for `disconnected` and `error` connections that reuses the existing OAuth reauthorize flow.
- Detect Google's `invalid_grant` response explicitly so revocation can be reported as such instead of as a generic failure.
- Expose `lastSyncAttemptAt` so users can distinguish "tried recently and failed" from "hasn't been tried in hours."

Non-goals:
- Push, email, or any out-of-band failure notifications.
- Reworking the OAuth flow itself, the encryption scheme, or the sync engine's scheduling.
- Multi-provider support (Google remains the only provider).
- Storing a historical log of past sync failures — only the latest failure is retained.
- Automated reconnection without user consent.
- Backfilling failure context onto existing rows.

## 3. User Stories

- As a household member, when my calendar stops syncing I want to see *why* it stopped so that I can tell whether it's a temporary blip or something I need to act on.
- As a household member, I want a one-click way to reconnect a broken calendar so that I don't have to delete and re-add it.
- As a household member, I want to know whether the system is still trying to recover on its own so that I don't intervene unnecessarily for a transient issue.
- As a household member who revoked Google access from my Google account settings, I want the app to tell me clearly that I revoked access (not just "disconnected") so that I understand what happened.
- As a household member, I want to see when the last sync attempt happened (not just the last successful sync) so that I can trust the app's status indicator.

## 4. Functional Requirements

### 4.1 Failure Classification

Sync failures must be classified into one of these `error_code` values when the connection is updated with an error:

| Code | Meaning | Severity | Status transition |
|---|---|---|---|
| `token_revoked` | Google returned `invalid_grant` on token refresh | hard | → `disconnected` immediately |
| `refresh_unauthorized` | Google returned 401 on token refresh (non-`invalid_grant`) | hard | → `disconnected` immediately |
| `token_decrypt_failed` | Local encryption key cannot decrypt the stored token | hard | → `disconnected` immediately |
| `refresh_http_error` | Transport error or 5xx from Google's token endpoint | transient | increment counter; → `error` after 3 consecutive failures |
| `unknown` | Any uncategorized failure path during refresh | transient | increment counter; → `error` after 3 consecutive failures |

The `error_message` field stores the raw error string for debugging; it must never be the user-facing message.

### 4.2 Retry Counter

- A new `consecutive_failures` integer column tracks the number of consecutive transient sync failures.
- A successful sync (any successful refresh + sync cycle) resets `consecutive_failures` to 0 and clears `error_code`, `error_message`, and `last_error_at`.
- Transient failures increment the counter without changing the visible `status` until the counter reaches 3, at which point status transitions to `error`.
- Hard failures bypass the counter and set status to `disconnected` immediately. The counter is also set to its escalation threshold so the UI does not need to special-case the transition.
- The threshold (3) is a constant in code, not configurable in this iteration.

### 4.3 Status Semantics

- `connected` — healthy; tokens valid; last sync succeeded or has not yet been attempted.
- `syncing` — sync cycle in progress (existing behavior, unchanged).
- `error` — recent sync attempts have failed transiently; the system will continue retrying on its normal schedule. UI shows yellow/amber badge with reason and a "Reconnect anyway" affordance.
- `disconnected` — tokens are invalid or revoked; no automatic recovery is possible. UI shows red badge with reason and a prominent "Reconnect" button.

### 4.4 Last Sync Attempt Timestamp

- A new `last_sync_attempt_at` timestamp records the time of the most recent sync attempt regardless of outcome.
- `last_sync_at` retains its existing meaning: the time of the most recent *successful* sync.
- Both fields are exposed in the JSON:API resource so the UI can show a meaningful relative timestamp ("tried 4 minutes ago, last success 6 hours ago").

### 4.5 Reconnect Action

- The frontend exposes a "Reconnect" button on the connection row when status is `disconnected` or `error`.
- The button triggers the existing `useReauthorizeCalendar` hook, which calls `POST /calendar/connections/google/authorize` with `reauthorize: true`.
- On successful OAuth callback, the callback handler must:
  - Update tokens (existing behavior).
  - Reset `status` to `connected`.
  - Clear `error_code`, `error_message`, `last_error_at`, and `consecutive_failures`.
  - Trigger an immediate sync (existing behavior).
- The reconnect flow does **not** clear source-level `syncToken`s. The existing fallback-to-full-sync path will handle any token invalidation that surfaces during the next sync.
- The Reauthorize banner for write-access upgrades remains as-is; it is a separate concern.

### 4.6 User-Facing Messaging

The frontend maps `error_code` to a localized human-readable message. Initial mapping:

| Code | User-facing message |
|---|---|
| `token_revoked` | "Access was revoked from your Google account. Reconnect to resume syncing." |
| `refresh_unauthorized` | "Google rejected your saved credentials. Reconnect to resume syncing." |
| `token_decrypt_failed` | "We can't read your stored credentials. Reconnect to resume syncing." |
| `refresh_http_error` | "Couldn't reach Google. Retrying automatically." |
| `unknown` | "Sync is failing. Retrying automatically." |
| (no code) | (no message) |

The raw `error_message` is not displayed by default but may appear in a "Details" expander for debugging.

### 4.7 OAuth Callback Failure Path

The initial-connect callback already redirects to `/app/calendar?error=...` on failure. This task leaves that flow unchanged — initial-connect failures are not stored on a connection because no connection exists yet at that point. Only post-connect sync failures populate the new fields.

## 5. API Surface

No new endpoints. The existing JSON:API resource for `calendar_connections` gains new attributes:

| Attribute | Type | Nullable | Notes |
|---|---|---|---|
| `errorCode` | string | yes | One of the codes in §4.1, or null when healthy |
| `errorMessage` | string | yes | Raw debug string; UI uses for "Details" only |
| `lastErrorAt` | string (ISO-8601) | yes | Time of the latest failure |
| `lastSyncAttemptAt` | string (ISO-8601) | yes | Time of the most recent sync attempt regardless of outcome |
| `consecutiveFailures` | integer | no | Defaults to 0 |

`status` continues to be a string with the values listed in §4.3.

The existing `POST /calendar/connections/google/authorize` endpoint and `GET /calendar/connections/google/callback` endpoint are reused as-is from the request side. The callback handler's *internal* behavior is extended per §4.5 but its URL contract is unchanged.

## 6. Data Model

New columns on `calendar_connections`:

| Column | Type | Nullable | Default |
|---|---|---|---|
| `error_code` | varchar(40) | yes | NULL |
| `error_message` | text | yes | NULL |
| `last_error_at` | timestamptz | yes | NULL |
| `last_sync_attempt_at` | timestamptz | yes | NULL |
| `consecutive_failures` | integer | no | 0 |

Migration notes:
- Add columns via GORM `AutoMigrate` (consistent with existing connection migrations).
- No backfill — existing rows get the column defaults. Healthy connections will populate `last_sync_attempt_at` on the next sync cycle. Already-broken connections will populate error fields on the next sync attempt and become recoverable through the new UI at that point.
- All existing immutability invariants on `connection.Model` continue to apply: a new builder/setter is added for each field; the model remains immutable post-construction.

## 7. Service Impact

**calendar-service (Go)**
- `connection/entity.go` — add five new columns; extend `ToEntity` / `Make`.
- `connection/model.go` — add fields, getters, and builder methods; remain immutable.
- `connection/processor.go` — add new methods (or extend existing ones) for:
  - `RecordSyncAttempt(id, at)` — sets `last_sync_attempt_at`.
  - `RecordSyncSuccess(id, eventCount, at)` — extends `UpdateSyncInfo` to also clear error fields, reset `consecutive_failures`, set `last_sync_at`, and set `last_sync_attempt_at`.
  - `RecordSyncFailure(id, code, message, at, escalated)` — sets error fields, increments counter, optionally transitions status.
- `connection/resource.go` — add new attributes to the JSON:API `RestModel` and `Transform` function; extend the callback handler to clear error state on successful reauthorize (per §4.5).
- `sync/sync.go` — wire `getValidAccessToken` to classify the error and call the appropriate processor method; record an attempt timestamp at the top of `syncOne`; classify Google API errors at the source (specifically detecting `invalid_grant` vs other failures).
- `googlecal/` — surface enough detail from token refresh failures for the sync engine to distinguish `invalid_grant` from generic transport errors. A typed error or a sentinel suffices; reuse existing patterns in the package.

**frontend**
- `src/types/models/calendar.ts` — add the new attribute fields to `CalendarConnection`.
- `src/components/features/calendar/connection-status.tsx` — render the failure reason inline when present; show both `lastSyncAt` and `lastSyncAttemptAt` when they differ; render a "Reconnect" button for `disconnected` and `error` statuses; route the click through `useReauthorizeCalendar`.
- `src/components/features/calendar/` — add a small `error-code-message.ts` helper (or inline map) implementing the §4.6 mapping. No new component file required unless the inline rendering becomes unwieldy.
- `src/lib/hooks/api/use-calendar.ts` — no changes expected; the existing reauthorize hook is reused.

No other services are affected. Multi-tenancy invariants are preserved because all changes are on rows already scoped by `tenant_id` and `household_id`.

## 8. Non-Functional Requirements

- **Performance**: New columns add a small write per sync attempt. The sync engine already writes per-connection on success; this adds at most one additional write per failure. No new queries are introduced.
- **Security**: No new secrets are stored. `error_message` may contain raw error strings from Google's token endpoint — these must not contain refresh tokens or access tokens (verify by inspecting `googlecal` error formatting before storing).
- **Observability**: Existing `logrus` warnings on failure paths are retained. Add a structured field `error_code` to those log entries so dashboards can group failures by category.
- **Multi-tenancy**: All reads and writes go through processors that already enforce tenant scoping. No new endpoints means no new authorization surface.
- **Backwards compatibility**: New JSON:API attributes are additive. Existing clients that ignore unknown attributes are unaffected. The frontend is the only known client and is updated in this task.

## 9. Open Questions

None remaining at PRD time. Decisions captured during scope discussion:

- Use a structured `error_code` enum + raw `error_message` (not a single string).
- Retain transient errors with a counter; escalate to `error` after 3 consecutive failures.
- Use the existing `error` and `disconnected` statuses with distinct semantics.
- Reconnect button is inline in `ConnectionStatus`, not a banner.
- Reuse the existing reauthorize endpoint; extend the callback handler to clear error state.
- Detect Google's `invalid_grant` explicitly to enable revocation-specific messaging.
- No backfill; new columns default to NULL/0.

## 10. Acceptance Criteria

- [ ] `calendar_connections` table has the five new columns with the documented types and defaults; migration runs cleanly on an empty DB and on a DB with existing rows.
- [ ] On a successful sync cycle, `last_sync_at`, `last_sync_attempt_at`, `last_sync_event_count` are updated and `error_code`, `error_message`, `last_error_at`, `consecutive_failures` are cleared.
- [ ] On a transient sync failure, `last_sync_attempt_at`, `error_code`, `error_message`, `last_error_at` are updated, `consecutive_failures` increments, and `status` remains `connected` until the counter reaches 3, at which point status transitions to `error`.
- [ ] On a hard sync failure (`invalid_grant`, 401, decrypt failure), `status` immediately transitions to `disconnected`, `error_code` is set to the matching value, and the counter is set to ≥3.
- [ ] Google's `invalid_grant` response is detected and produces `error_code = token_revoked`, distinct from generic refresh HTTP errors.
- [ ] The JSON:API `calendar_connections` resource serializes all five new attributes with the documented names and types.
- [ ] The frontend `ConnectionStatus` component renders the user-facing message from §4.6 next to the status badge when `errorCode` is present.
- [ ] The frontend renders a "Reconnect" button for `disconnected` and `error` connections; clicking it triggers the existing reauthorize OAuth flow.
- [ ] On successful reauthorize, the OAuth callback handler clears `error_code`, `error_message`, `last_error_at`, and `consecutive_failures`, sets `status` to `connected`, and triggers an immediate sync.
- [ ] When `lastSyncAt` and `lastSyncAttemptAt` differ, the UI shows both (e.g., "tried 4 minutes ago, last success 6 hours ago").
- [ ] No automated notifications are sent when status changes — failure surfacing is in-app only.
- [ ] All affected services build cleanly; calendar-service unit tests cover the new processor methods and the classification logic in `getValidAccessToken`.
