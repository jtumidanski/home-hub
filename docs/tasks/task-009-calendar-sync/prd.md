# Household Calendar with Google Calendar Sync — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-26

---

## 1. Overview

Home Hub needs a unified household calendar that gives household members a shared view of upcoming events. Each user can optionally connect their Google Calendar, which syncs events into the household calendar on a read-only basis. When multiple household members connect their calendars, all events appear in a merged view, distinguished by user.

The calendar-service is a new microservice that manages Google Calendar OAuth connections, background event synchronization, event storage, and a REST API for the frontend. The frontend displays a 7-day hourly view of the household calendar.

This MVP is read-only sync. The architecture is designed to support future write-back (creating events from the UI that push to a user's Google Calendar) and additional providers (Outlook, Apple, CalDAV).

## 2. Goals

Primary goals:

- Allow users to optionally connect their Google Calendar to their household
- Sync Google Calendar events into a household-scoped calendar via background polling
- Display a merged 7-day hourly calendar view in the frontend
- Respect event privacy — private/confidential events show as "Busy" to other household members
- Support multiple users per household each with their own Google Calendar connection

Non-goals:

- Creating events in the UI or writing back to Google Calendar (future)
- Other calendar providers (Outlook, Apple, CalDAV)
- Recurring event editing or modification
- Calendar sharing across households
- Event notifications or reminders (productivity-service handles reminders)
- Real-time sync via Google push notifications / webhooks
- Mobile-specific calendar views

## 3. User Stories

- As a household member, I want to connect my Google Calendar so that my events appear on the household calendar
- As a household member, I want to choose which of my Google Calendars (personal, work, holidays) sync to the household calendar
- As a household member, I want to see a 7-day hourly view of all household events so I can plan my week
- As a household member, I want to see which family member each event belongs to so I can distinguish events
- As a household member, I want my private Google Calendar events to show as "Busy" to other members so my privacy is respected
- As a household member, I want to see my own private events with full details on the household calendar
- As a household member, I want to disconnect my Google Calendar if I no longer want to share events
- As a household member, I want the calendar to stay up-to-date automatically without manual refresh
- As a household member, I want events displayed in my household's timezone so times make sense in context

## 4. Functional Requirements

### 4.1 Google Calendar Connection

- Users connect their Google Calendar via a dedicated OAuth consent flow, separate from login
- The OAuth flow requests `https://www.googleapis.com/auth/calendar.readonly` scope
- OAuth tokens (access + refresh) are stored per user, per tenant in the calendar-service database
- Users can disconnect their Google Calendar, which revokes the token and deletes stored events for that user
- A user can have at most one Google Calendar connection per household (separate connections per household if the user is in multiple households)
- The connection status is visible in the UI (connected/disconnected, last sync time)
- When a user's membership in a household is removed, their calendar events are removed from that household's calendar and their connection is disassociated from the household

### 4.1.1 Calendar Selection

- After connecting, the service fetches the user's Google Calendar list (primary, secondary, subscribed)
- All calendars are synced by default
- Users can toggle which calendars are visible on the household calendar via the UI
- Calendar visibility toggles are stored per connection and applied as a filter on the events endpoint
- The calendar list is refreshed on each sync cycle to pick up newly added/removed Google Calendars

### 4.2 Event Synchronization

- Background polling syncs events from connected Google Calendars at a configurable interval (default: 15 minutes)
- Sync window: configurable, default past 7 days through next 30 days
- Recurring events are fetched in their expanded (single-instance) form from the Google API
- Events are stored in the calendar-service database, scoped to tenant + household + user
- Sync is incremental where possible using Google Calendar's `syncToken` per source calendar for efficient polling
- If a sync token is invalidated, the service performs a full sync for that user
- Deleted events in Google Calendar are removed from the local store during sync
- If a user's OAuth token refresh fails (revoked access), mark the connection as `disconnected` and stop syncing
- First sync is triggered immediately after connection creation (async goroutine), not waiting for the next background tick
- Sync operations are staggered across connections with random jitter (0–60s) to avoid burst traffic to Google API
- On 429 or 5xx responses from Google, use exponential backoff (1s, 2s, 4s, max 3 retries per sync cycle)
- Manual sync is rate-limited to once per 5 minutes per connection
- Duplicate events across source calendars are allowed (same event on "Personal" and "Work" calendars appears twice); users resolve this by toggling off redundant source calendars
- User display name on events is sourced from JWT claims at connection time and stored on the connection record

### 4.3 Event Display

- The household calendar merges events from all connected users in the household
- Each event displays: title, start time, end time, user attribution (name + color)
- Events marked `private` or `confidential` in Google Calendar show as "Busy" with user attribution to non-owner household members; full details are shown to the event owner
- All-day events are displayed in a dedicated section above the hourly grid
- Multi-day events span across days in the all-day section
- Events are color-coded per user for visual distinction
- Overlapping events within the same time slot are rendered side-by-side, splitting the column width
- All event times are displayed in the household's timezone (from the household model in account-service)
- User display names are sourced from the auth-service user profile (givenName/familyName)
- If a user disconnects and reconnects, they receive a new color assignment (creation-order based)

### 4.4 Calendar View

- 7-day view starting from the current day
- Vertical axis: hours of the day (e.g., 6 AM to 11 PM, or configurable)
- Horizontal axis: days of the week
- Navigation: previous/next week buttons
- Today indicator on the current day column

## 5. API Surface

See [api-contracts.md](api-contracts.md) for detailed endpoint specifications.

Summary of endpoints:

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/calendar/connections` | List user's calendar connections |
| POST | `/api/v1/calendar/connections/google/authorize` | Initiate Google OAuth flow |
| GET | `/api/v1/calendar/connections/google/callback` | Google OAuth callback |
| DELETE | `/api/v1/calendar/connections/{id}` | Disconnect a calendar |
| GET | `/api/v1/calendar/connections/{id}/calendars` | List Google Calendars for a connection |
| PATCH | `/api/v1/calendar/connections/{id}/calendars/{calId}` | Toggle calendar visibility |
| POST | `/api/v1/calendar/connections/{id}/sync` | Trigger manual sync |
| GET | `/api/v1/calendar/events` | List events for household calendar |

## 6. Data Model

See [data-model.md](data-model.md) for detailed schema.

Summary of entities:

| Entity | Purpose |
|--------|---------|
| `calendar_connections` | Stores OAuth connection state and tokens per user |
| `calendar_sources` | Stores individual Google Calendars per connection with visibility toggle |
| `calendar_events` | Stores synced calendar events |

All entities are scoped by `tenant_id` and `household_id`.

## 7. Service Impact

### 7.1 calendar-service (NEW)

New microservice following established patterns:

- `cmd/main.go` — service entry point, route registration, background sync loop
- `internal/config/` — environment variable configuration
- `internal/connection/` — OAuth connection domain (model, entity, builder, processor, provider, resource, rest)
- `internal/event/` — Calendar event domain (model, entity, builder, processor, provider, resource, rest)
- `internal/googlecal/` — Google Calendar API client (HTTP client, token refresh, types)
- `internal/sync/` — Background sync orchestrator
- `docs/` — Service documentation (domain.md, rest.md, storage.md)

### 7.2 auth-service

No changes required. The calendar-service manages its own Google OAuth flow with calendar-specific scopes, independent of the login OIDC flow. The same Google Cloud project client credentials may be reused (configured via separate env vars on the calendar-service).

### 7.3 frontend

- New `/app/calendar` route and page
- Calendar 7-day hourly grid component
- Google Calendar connect/disconnect UI (accessible from calendar page or settings)
- OAuth redirect handling for calendar connection flow
- New API service class for calendar endpoints
- New React Query hooks for calendar data
- Navigation sidebar entry for Calendar

### 7.4 Infrastructure

- nginx/ingress: add route `/api/v1/calendar/*` → calendar-service
- docker-compose: add calendar-service container
- k8s manifests: add calendar-service deployment, service, ingress rule
- secrets: add `GOOGLE_CALENDAR_CLIENT_ID`, `GOOGLE_CALENDAR_CLIENT_SECRET` (may reuse existing Google OAuth credentials)
- CI: add calendar-service build/test/lint workflows

## 8. Non-Functional Requirements

### 8.1 Performance

- Event list endpoint must respond within 500ms for a household with 5 connected calendars
- Background sync must not overload Google Calendar API — respect rate limits (use exponential backoff)
- Sync operations are staggered across connections with random jitter (0–60s) to avoid burst traffic
- Google Calendar API quota: ~1,000,000 queries/day per project, ~100 requests/100s per user — 15-minute polling is well within limits
- Max query range on events endpoint capped at 90 days server-side to prevent excessive data retrieval

### 8.2 Security

- Google OAuth tokens (access + refresh) must be stored encrypted at rest using AES-256-GCM
- Encryption key provided via `CALENDAR_TOKEN_ENCRYPTION_KEY` environment variable
- No key rotation mechanism for MVP — if the key must change, all connections must re-authorize (acceptable for small user base). Future: support `CALENDAR_TOKEN_ENCRYPTION_KEY_PREVIOUS` for graceful rotation (decrypt with old, re-encrypt with new on access)
- Tokens are never exposed via API responses
- Calendar event data is tenant-scoped and household-scoped — no cross-tenant or cross-household leakage
- OAuth state parameter must be validated to prevent CSRF
- The Google OAuth redirect URI must be validated server-side

### 8.3 Observability

- Structured JSON logging with request_id, user_id, tenant_id, household_id
- Log sync operations: start, completion, event count, errors
- Log OAuth flow: initiation, callback, token refresh, revocation
- OpenTelemetry tracing on all endpoints and sync operations

### 8.4 Multi-Tenancy

- All data scoped by tenant_id and household_id
- JWT validation via JWKS (same as other services)
- Tenant context extracted from request headers (same middleware as other services)

### 8.5 Resilience

- If Google API is unreachable, sync retries on next interval (no crash)
- If token refresh fails, mark connection as disconnected and notify user on next UI load
- Partial sync failure for one user does not block sync for other users in the household

## 9. Open Questions

All previously open questions have been resolved:

- **Google OAuth credentials:** Reuse the same Google Cloud project but configure via separate env vars (`GOOGLE_CALENDAR_CLIENT_ID`, `GOOGLE_CALENDAR_CLIENT_SECRET`) on the calendar-service. **Resolved.**
- **Sync interval:** Global for MVP, per-connection later. **Resolved.**
- **Calendar selection:** Sync all user calendars, let users toggle visibility per calendar in the UI. **Resolved.**
- **Timezone:** Use household timezone (already exists on the household model in account-service). **Resolved.**
- **Overlapping events:** Render side-by-side within the column, splitting width. **Resolved.**
- **Membership removal:** Remove events when a user leaves a household. **Resolved.**
- **Manual sync rate limit:** Once per 5 minutes per connection. **Resolved.**
- **User display name source:** Auth-service user profile (givenName/familyName). **Resolved.**
- **Color on reconnect:** New color assigned (creation-order based). **Resolved.**
- **Frontend calendar:** Custom-built 7-day hourly grid (no third-party calendar library). **Resolved.**
- **Token encryption key rotation:** No rotation for MVP; re-authorize on key change. Future: dual-key support. **Resolved.**
- **Pagination:** No pagination for MVP; max query range capped at 90 days server-side. **Resolved.**
- **Google API rate limits:** Exponential backoff on 429/5xx (1s, 2s, 4s, max 3 retries). Stagger sync with 0–60s jitter. **Resolved.**
- **Sync token placement:** Per source calendar (on `calendar_sources`), not per connection. Each Google Calendar has its own sync token. **Resolved.**
- **User display name source:** Pulled from JWT claims during OAuth callback, stored on the connection record. **Resolved.**
- **Connection-household scoping:** One connection per user per household per provider. `UNIQUE(tenant_id, household_id, user_id, provider)`. Users in multiple households connect separately. **Resolved.**
- **Cross-calendar deduplication:** Duplicates allowed across source calendars. Users toggle off redundant sources. Dedup within a single source via `UNIQUE(source_id, external_id)`. **Resolved.**
- **Initial sync timing:** First sync triggered immediately after connection (async goroutine), not waiting for background tick. **Resolved.**
- **Frontend polling:** React Query with 60-second stale time, refetch-on-window-focus, no active polling interval. Manual sync triggers refetch on completion. **Resolved.**

## 10. Acceptance Criteria

- [ ] A user can initiate a Google Calendar OAuth connection from the calendar page
- [ ] After connecting, the user can see and toggle which Google Calendars sync to the household
- [ ] After connecting, the user's Google Calendar events appear on the household calendar within one sync interval
- [ ] The household calendar displays a 7-day hourly grid with events from all connected household members
- [ ] Events are color-coded and attributed to the user who owns them
- [ ] Private/confidential events show as "Busy" to non-owner household members
- [ ] Private/confidential events show full details to the event owner
- [ ] All-day events appear in a dedicated section above the hourly grid
- [ ] Overlapping events render side-by-side within the same time slot column
- [ ] All event times are displayed in the household's timezone
- [ ] A user can disconnect their Google Calendar, removing their events from the household calendar
- [ ] When a user's household membership is removed, their events are removed from that household's calendar
- [ ] Connection status and last sync time are visible in the UI
- [ ] Background sync runs at a configurable interval (default 15 minutes)
- [ ] First sync triggers immediately after connection (async), not waiting for background tick
- [ ] Sync uses incremental sync tokens per source calendar for efficiency
- [ ] Sync operations are staggered with random jitter to avoid burst traffic
- [ ] Google API 429/5xx responses are handled with exponential backoff (max 3 retries)
- [ ] Manual sync can be triggered from the UI (rate-limited to once per 5 minutes)
- [ ] If a user revokes Google access externally, the connection is marked as disconnected
- [ ] OAuth tokens are encrypted at rest using AES-256-GCM
- [ ] The calendar-service starts, migrates its schema, and serves requests following existing service patterns
- [ ] All data is scoped by tenant_id and household_id
- [ ] Navigation includes a Calendar entry in the sidebar
- [ ] Week navigation (previous/next) works correctly
- [ ] Today is visually indicated on the calendar grid
- [ ] Events endpoint enforces a max query range of 90 days
- [ ] Calendar UI is custom-built (no third-party calendar library)
- [ ] Frontend uses React Query with 60-second stale time and refetch-on-window-focus
- [ ] Users in multiple households can connect separately per household
- [ ] User display name on events is sourced from JWT claims at connection time
