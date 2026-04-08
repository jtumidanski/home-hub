# Tasks — Task 025: Calendar Reconnect UX

Last Updated: 2026-04-08

Tracking checklist. See `plan.md` for full descriptions and `context.md` for code pointers.

## Phase 1 — Schema & model

- [ ] **1.1** Add five columns to `connection.Entity` (`services/calendar-service/internal/connection/entity.go`): `ErrorCode *string`, `ErrorMessage *string`, `LastErrorAt *time.Time`, `LastSyncAttemptAt *time.Time`, `ConsecutiveFailures int` (default 0). _(S)_
- [ ] **1.2** Mirror new fields in `connection.Model` (getters), `connection.Builder` (setters, no validation), `entity.go::ToEntity` and `entity.go::Make`. _(S, depends on 1.1)_
- [ ] **1.3** Run `scripts/local-up.sh`; verify via psql that `calendar_connections` has the five new columns with documented defaults; existing rows load and serialize without errors. _(S, depends on 1.2)_

## Phase 2 — Failure classification

- [ ] **2.1** Add `googlecal.TokenRefreshError` (StatusCode/OAuthError/Body) and `IsInvalidGrant(err)` helper next to `SyncTokenInvalidError` in `googlecal/client.go`. Restructure `Client.RefreshToken` non-200 branch to parse the JSON body and return the typed error. _(S)_
- [ ] **2.2** Add `Engine.classifyTokenError(err) (code string, hard bool)` to `sync/sync.go`. Restructure `syncOne` to: (a) call `RecordSyncAttempt` at entry, (b) classify on `getValidAccessToken` error and call `RecordSyncFailure`, (c) call `RecordSyncSuccess` instead of `UpdateSyncInfo` on the success path. _(M, depends on 2.1, 3.2-3.4)_
- [ ] **2.3** Add `crypto.ErrDecryptFailed` sentinel and wrap existing decrypt failures with `fmt.Errorf("%w: %v", ErrDecryptFailed, err)`. _(S)_

## Phase 3 — Processor write methods

- [ ] **3.1** Add exported constant `connection.FailureEscalationThreshold = 3` to `processor.go`. _(S)_
- [ ] **3.2** Add `Processor.RecordSyncAttempt(id, at)` + `administrator.go::updateSyncAttempt`. _(S)_
- [ ] **3.3** Add `Processor.RecordSyncSuccess(id, eventCount, at)` + `administrator.go::updateSyncSuccess`. Sets `status='connected'`, both timestamps, event count; clears all error fields and counter. Use `gorm.Expr("NULL")` for nullable clears. _(M)_
- [ ] **3.4** Add `Processor.RecordSyncFailure(id, code, message, at)` + `administrator.go::updateSyncFailure` + `isHardErrorCode`. Truncate `message` to 500 chars before storing. Use a single atomic `UPDATE` with `CASE`/`GREATEST` expressions so the counter increment, threshold check, and status transition happen in one SQL statement (no read-modify-write race). Hard codes → `status='disconnected'`, `consecutive_failures = GREATEST(... + 1, 3)`; transient codes → `consecutive_failures + 1`, `status = CASE WHEN ... + 1 >= 3 THEN 'error' ELSE status END`. _(M, depends on 3.1)_
- [ ] **3.5** Add `Processor.ClearErrorState(id)` + `administrator.go::clearErrorState`. _(S)_
- [ ] **3.6** Extend `administrator.go::updateTokensAndWriteAccess` to ALSO clear error fields and reset counter in the same write (chosen over a separate `ClearErrorState` call from the callback handler). _(S)_

## Phase 4 — JSON:API surface & callback verification

- [ ] **4.1** Extend `connection.RestModel` with `LastSyncAttemptAt`, `ErrorCode`, `ErrorMessage`, `LastErrorAt`, `ConsecutiveFailures` (JSON tags per PRD §5). Update `Transform`. _(S, depends on 1.2)_
- [ ] **4.2** Re-read `connection/resource.go:152-173` (callback reauthorize branch); confirm no code change is needed because Task 3.6 covers the reset. Document verification in PR description. _(S, depends on 3.6)_

## Phase 5 — Backend tests

- [ ] **5.1** Extend `connection/processor_test.go` with: `TestRecordSyncAttempt_SetsTimestamp`, `TestRecordSyncSuccess_ClearsErrorAndResetsCounter`, `TestRecordSyncFailure_TransientUnderThresholdLeavesStatus`, `TestRecordSyncFailure_TransientReachesThresholdSetsError`, `TestRecordSyncFailure_HardFailureSetsDisconnectedAndForcesCounter`, `TestRecordSyncFailure_HardFailurePreservesHigherCounter`, `TestClearErrorState_ResetsAllFields`, `TestUpdateTokensAndWriteAccess_ClearsErrorState`. _(M, depends on Phase 3)_
- [ ] **5.2** Add `sync/classify_test.go` (or extend existing sync test) with: `TestClassifyTokenError_InvalidGrantHard`, `TestClassifyTokenError_Unauthorized401Hard`, `TestClassifyTokenError_5xxTransient`, `TestClassifyTokenError_TransportTransient`, `TestClassifyTokenError_DecryptFailedHard`. _(S, depends on 2.2)_
- [ ] **5.3** Add `googlecal/client_test.go` cases: `TestRefreshToken_InvalidGrantReturnsTypedError` (httptest server returning 400 + invalid_grant body), `TestRefreshToken_5xxReturnsTypedErrorWithoutOAuthCode`, `TestIsInvalidGrant_NilFalse`. _(S, depends on 2.1)_

## Phase 6 — Frontend

- [ ] **6.1** Extend `CalendarConnectionAttributes` in `frontend/src/types/models/calendar.ts` with `lastSyncAttemptAt`, `errorCode`, `errorMessage`, `lastErrorAt`, `consecutiveFailures`. _(S)_
- [ ] **6.2** Add `frontend/src/components/features/calendar/error-code-message.ts` exporting `errorCodeToMessage(code)` per PRD §4.6 mapping. _(S)_
- [ ] **6.3** Restructure `connection-status.tsx`: amber "Sync issues" badge for `error`, red "Disconnected" badge for `disconnected`. Add second-line failure block with message + "Tried X ago · Last success Y ago" subline (when timestamps differ) + Reconnect button (`default` variant on `disconnected`, `outline` on `error` with "Reconnect anyway" copy). Wire button to `useReauthorizeCalendar`. Keep sync button disabled on non-connected statuses (preserve existing behavior; see plan risk R8). _(M, depends on 6.1, 6.2)_
- [ ] **6.4** Re-read `useReauthorizeCalendar` and `calendarService.reauthorizeGoogle`; confirm it's reusable as-is. Document in PR. _(S)_

## Phase 7 — Verification & ship

- [ ] **7.1** Run `go build ./...` and `go test ./...` for `services/calendar-service`; run frontend build/typecheck. _(S, depends on Phase 5, Phase 6)_
- [ ] **7.2** Local end-to-end smoke via `scripts/local-up.sh`: connect a Google calendar; corrupt the row's `refresh_token` in psql; confirm flips to `disconnected` with `error_code = token_decrypt_failed` and frontend renders red badge + Reconnect; click Reconnect, complete OAuth dance, confirm full reset. _(S, depends on 7.1)_
- [ ] **7.3** Open PR referencing task-025; body summarizes schema/classification/processor/frontend changes and includes a test plan checklist mirroring PRD §10. _(S, depends on 7.2)_

## PRD Acceptance Criteria Checklist (PRD §10)

- [ ] `calendar_connections` has the five new columns with documented types and defaults; migration runs cleanly on empty and populated DBs. _(Tasks 1.1, 1.3)_
- [ ] Successful sync cycle updates `last_sync_at`, `last_sync_attempt_at`, `last_sync_event_count` and clears error fields + counter. _(Tasks 3.3, 5.1, 7.2)_
- [ ] Transient sync failure updates `last_sync_attempt_at`/`error_code`/`error_message`/`last_error_at`, increments counter, status stays `connected` until counter reaches 3 then transitions to `error`. _(Tasks 3.4, 5.1)_
- [ ] Hard sync failure (`invalid_grant`, 401, decrypt) immediately transitions status to `disconnected`, sets matching `error_code`, counter ≥ 3. _(Tasks 3.4, 5.1)_
- [ ] Google's `invalid_grant` is detected and produces `error_code = token_revoked`, distinct from generic refresh HTTP errors. _(Tasks 2.1, 2.2, 5.2, 5.3)_
- [ ] JSON:API `calendar_connections` resource serializes all five new attributes with documented names and types. _(Task 4.1)_
- [ ] Frontend `ConnectionStatus` renders user-facing message from PRD §4.6 next to status badge when `errorCode` is present. _(Tasks 6.2, 6.3)_
- [ ] Frontend renders Reconnect button for `disconnected` and `error`; click triggers existing reauthorize flow. _(Task 6.3)_
- [ ] On successful reauthorize, callback handler clears error fields, sets `status = connected`, triggers immediate sync. _(Tasks 3.6, 4.2)_
- [ ] When `lastSyncAt` and `lastSyncAttemptAt` differ, UI shows both. _(Task 6.3)_
- [ ] No automated notifications. _(Out of scope; nothing to do — verify in PR.)_
- [ ] All affected services build cleanly; calendar-service unit tests cover new processor methods and classification logic. _(Tasks 5.1, 5.2, 5.3, 7.1)_
