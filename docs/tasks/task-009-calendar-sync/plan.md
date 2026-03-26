# Task 009: Household Calendar with Google Calendar Sync — Implementation Plan

Last Updated: 2026-03-26

---

## Executive Summary

Build a new `calendar-service` microservice and corresponding frontend to provide a unified household calendar with Google Calendar sync. The service handles OAuth connections, background event polling, encrypted token storage, and a REST API. The frontend delivers a custom-built 7-day hourly grid with multi-user event display, privacy masking, and connection management. This is an MVP with read-only sync — no write-back to Google Calendar.

---

## Current State Analysis

- **Existing services:** auth-service, account-service, productivity-service, recipe-service, weather-service — all follow the standard DDD layered pattern (model/entity/builder/processor/provider/resource/rest)
- **Auth pattern established:** auth-service has Google OIDC login with state validation, code exchange, token storage, and cookie-based sessions. Calendar OAuth is a *separate* flow with different scopes (`calendar.readonly` vs `openid email profile`)
- **Shared libraries available:** `shared/go/` provides auth middleware (JWT/JWKS validation), database (GORM connection, tenant callbacks, migrations), server (handler registration, JSON:API helpers), tenant context, logging, and tracing
- **Frontend patterns established:** React + Vite + shadcn/ui, service classes per domain, React Query hooks with key factories, context providers (Auth, Tenant), JSON:API response types, AppShell with sidebar navigation
- **Infrastructure ready:** docker-compose with nginx reverse proxy, K8s manifests, GitHub Actions CI, path-prefix routing

**No calendar functionality exists today.** This is a greenfield service.

---

## Proposed Future State

```
Browser → nginx → /api/v1/calendar/* → calendar-service → Google Calendar API
                                                        → PostgreSQL (calendar.*)
```

New artifacts:
- `services/calendar-service/` — Go microservice (4 domains: connection, source, event, sync)
- `frontend/src/pages/CalendarPage.tsx` — 7-day hourly grid view
- `frontend/src/services/api/calendar.ts` — API service class
- `frontend/src/lib/hooks/api/use-calendar.ts` — React Query hooks
- `frontend/src/components/features/calendar/` — Calendar grid, event blocks, connection UI
- Infrastructure: docker-compose service, nginx route, K8s deployment, CI workflow

---

## Implementation Phases

### Phase 1: Backend Foundation (calendar-service scaffold + connection domain)

Stand up the new service and implement the Google Calendar OAuth connection flow. This is the critical path — everything else depends on having a running service with working OAuth.

**1.1 Service Scaffold**
- Create `services/calendar-service/` directory structure following existing patterns
- `cmd/main.go` — logger, config, tracing, database, server
- `internal/config/config.go` — env vars: DB, JWKS_URL, PORT, GOOGLE_CALENDAR_CLIENT_ID, GOOGLE_CALENDAR_CLIENT_SECRET, CALENDAR_TOKEN_ENCRYPTION_KEY, SYNC_INTERVAL_MINUTES
- `go.mod` referencing shared modules
- `Dockerfile` following the standard multi-stage build pattern

**1.2 Token Encryption**
- AES-256-GCM encrypt/decrypt utility in the connection domain or a shared crypto package within the service
- Key loaded from `CALENDAR_TOKEN_ENCRYPTION_KEY` env var
- Encrypt before entity persistence, decrypt after entity load
- Unit tests with known test vectors

**1.3 OAuth State Domain**
- `internal/oauthstate/` — model, entity, builder, processor, provider
- Entity: `calendar_oauth_states` table with UUID PK (= state param), tenant_id, household_id, user_id, redirect_uri, expires_at
- Processor: create state, validate state (check exists + not expired + matches user), delete expired states
- Cleanup: expired states deleted during sync loop

**1.4 Connection Domain**
- `internal/connection/` — model, entity, builder, processor, provider, resource, rest
- Entity: `calendar_connections` table per data-model.md
- Model: immutable with private fields, accessors for all columns
- Builder: validates provider, email, status, handles encrypted token fields
- Processor: create connection, get by ID, list by user+household, update status/sync times, delete (cascade events+sources)
- Provider: query builders for tenant+household, user, ID filters
- Resource: JSON:API `calendar-connections` type with Transform
- REST endpoints:
  - `POST /connections/google/authorize` — generate state, build Google auth URL, return JSON response
  - `GET /connections/google/callback` — validate state, exchange code, fetch user info, encrypt tokens, create connection, fetch calendar list, trigger first sync, redirect
  - `GET /connections` — list user's connections
  - `DELETE /connections/{id}` — disconnect (revoke token, delete events/sources/connection)
  - `POST /connections/{id}/sync` — manual sync trigger (rate-limited 5min)

**1.5 Google Calendar API Client**
- `internal/googlecal/` — HTTP client wrapper for Google Calendar API
- Token refresh: use refresh token to get new access token when expired
- Calendar list: `GET https://www.googleapis.com/calendar/v3/users/me/calendarList`
- Events list: `GET https://www.googleapis.com/calendar/v3/calendars/{calendarId}/events` with `syncToken` support
- OAuth token exchange: POST to Google's token endpoint
- Exponential backoff on 429/5xx (1s, 2s, 4s, max 3 retries)
- All responses typed as Go structs

### Phase 2: Source & Event Domains + Sync Engine

Build the remaining data domains and the background sync orchestrator.

**2.1 Source Domain (Calendar Sources)**
- `internal/source/` — model, entity, builder, processor, provider, resource, rest
- Entity: `calendar_sources` table per data-model.md
- Processor: create/upsert sources from Google calendar list, toggle visibility, get by connection
- REST endpoints:
  - `GET /connections/{id}/calendars` — list sources for a connection
  - `PATCH /connections/{id}/calendars/{calId}` — toggle visibility

**2.2 Event Domain**
- `internal/event/` — model, entity, builder, processor, provider, resource, rest
- Entity: `calendar_events` table per data-model.md
- Processor: bulk upsert events, delete by connection, delete by source+external_id, query by household+time range
- Privacy logic in resource Transform: if visibility is private/confidential AND requester != owner, mask title/description/location
- REST endpoint:
  - `GET /events?start=...&end=...` — list household events with privacy masking, 90-day max range, sorted by start_time (all-day first)

**2.3 User Color Assignment**
- Deterministic color from predefined palette based on connection creation order within household
- Color stored on connection and denormalized onto events during sync
- Palette: 8 colors cycling

**2.4 Background Sync Engine**
- `internal/sync/` — orchestrator
- Runs on configurable interval (default 15 min) via time.Ticker in a goroutine started from main.go
- Per tick: fetch all `connected` connections across all tenants/households, stagger with 0-60s random jitter
- Per connection: refresh access token if needed, for each visible source calendar, call Google Events API with syncToken (or full sync if token invalid), upsert/delete events, update syncToken, update connection last_sync_at and event count
- On token refresh failure: mark connection `disconnected`, skip
- On Google 429/5xx: exponential backoff per retry, max 3 retries, then skip to next connection
- On first connection: immediate async sync (goroutine) triggered from callback handler
- Refresh calendar list on each sync cycle (add/remove sources as needed)

### Phase 3: Infrastructure & Deployment

**3.1 Docker Compose**
- Add `calendar-service` to `deploy/compose/docker-compose.yml` with env vars
- Add nginx route: `location /api/v1/calendar/ { proxy_pass http://calendar-service:8080; }`

**3.2 Kubernetes**
- Deployment manifest for calendar-service
- Service manifest
- Ingress rule for `/api/v1/calendar/`
- Secret references for GOOGLE_CALENDAR_CLIENT_ID, GOOGLE_CALENDAR_CLIENT_SECRET, CALENDAR_TOKEN_ENCRYPTION_KEY

**3.3 CI/CD**
- GitHub Actions workflow: build, test, lint for calendar-service
- Docker image: `ghcr.io/<owner>/home-hub-calendar`

**3.4 Service Documentation**
- `services/calendar-service/docs/domain.md`
- `services/calendar-service/docs/rest.md`
- `services/calendar-service/docs/storage.md`

### Phase 4: Frontend — Calendar Page & Connection UI

**4.1 API Service & Types**
- `services/api/calendar.ts` — CalendarService class with methods for all endpoints
- TypeScript types for calendar-connections, calendar-sources, calendar-events API responses

**4.2 React Query Hooks**
- `lib/hooks/api/use-calendar.ts`
- Key factory: `calendarKeys.connections()`, `calendarKeys.sources(connectionId)`, `calendarKeys.events(start, end)`
- `useCalendarConnections()` — list connections
- `useCalendarSources(connectionId)` — list sources for a connection
- `useCalendarEvents(start, end)` — list events with 60s staleTime, refetchOnWindowFocus
- `useConnectGoogleCalendar()` — mutation to initiate OAuth
- `useDisconnectCalendar()` — mutation to delete connection
- `useToggleCalendarSource()` — mutation to toggle visibility
- `useTriggerSync()` — mutation to trigger manual sync

**4.3 Calendar Grid Component**
- `components/features/calendar/CalendarGrid.tsx` — 7-day hourly grid
- Vertical axis: hours (default 6 AM – 11 PM), scrollable
- Horizontal axis: 7 day columns starting from the provided start date
- Today column highlight with current time indicator line
- All-day event section above the grid (collapsible)
- Event blocks positioned by start/end time, colored by userColor
- Overlapping events rendered side-by-side splitting column width
- All times converted to household timezone (from TenantContext)

**4.4 Event Components**
- `components/features/calendar/EventBlock.tsx` — single event display in the grid
- `components/features/calendar/EventPopover.tsx` — click popover with full details or "Busy"
- `components/features/calendar/AllDayEventRow.tsx` — all-day / multi-day event display

**4.5 Connection Management UI**
- `components/features/calendar/ConnectCalendarButton.tsx` — initiates OAuth flow
- `components/features/calendar/ConnectionStatus.tsx` — shows status, last sync time, sync/disconnect actions
- `components/features/calendar/CalendarSelectionPanel.tsx` — toggle source calendars visible/hidden
- `components/features/calendar/DisconnectDialog.tsx` — confirmation dialog

**4.6 Calendar Page**
- `pages/CalendarPage.tsx` — main page component
- Week navigation: prev/next buttons, today button
- Header with date range display
- Empty state when no connections exist
- Handles query params: `?connected=true`, `?error=...` for post-OAuth toasts
- Legend showing connected users with their colors

**4.7 Navigation**
- Add "Calendar" entry to sidebar navigation
- Route: `/app/calendar`
- Add OAuth callback route handling (redirect from Google goes to backend callback, which redirects to `/app/calendar`)

### Phase 5: Integration Testing & Polish

**5.1 Backend Tests**
- Unit tests for token encryption/decryption
- Unit tests for privacy masking logic
- Unit tests for color assignment
- Unit tests for sync token handling
- Unit tests for rate limiting (manual sync)
- Processor tests for connection lifecycle

**5.2 Frontend Tests**
- Verify calendar grid renders events in correct positions
- Verify overlapping event layout
- Verify privacy masking (Busy display)
- Verify empty state and connection flow

**5.3 End-to-End Verification**
- Full OAuth flow: connect → sync → display → disconnect
- Multi-user household: two users connected, events merged
- Privacy: private events masked for non-owners
- Reconnection: new color assigned

---

## Risk Assessment and Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Google OAuth credentials misconfigured | Medium | High — blocks all functionality | Document setup steps clearly; validate credentials on service startup |
| Google Calendar API rate limiting | Low | Medium — delays sync | Exponential backoff + jitter already designed; 15-min interval well within quotas |
| Token encryption key loss | Low | High — all connections must re-authorize | Document key backup; future: dual-key rotation support |
| Sync token invalidation storms | Low | Medium — full re-sync burst | Already handled: fallback to full sync on invalid token |
| Complex calendar grid rendering | Medium | Medium — visual bugs | Build incrementally; start with basic positioning, add overlap handling |
| Timezone handling bugs | Medium | Medium — events shown at wrong times | Store all times UTC; convert in frontend using household timezone |
| Large event volumes impacting performance | Low | Medium — slow queries | Composite index on (tenant, household, start_time, end_time); 90-day query cap |

---

## Success Metrics

- User can connect Google Calendar and see events on household calendar within 60 seconds of connecting
- Background sync keeps events current within the 15-minute interval
- Calendar page loads and renders events within 500ms for a household with 5 connected users
- Privacy masking correctly hides details of private/confidential events from non-owners
- Zero cross-tenant or cross-household data leakage

---

## Required Resources and Dependencies

**External:**
- Google Cloud project with Calendar API enabled
- OAuth 2.0 client credentials (client ID + client secret) with `calendar.readonly` scope
- Authorized redirect URI configured in Google Cloud Console

**Internal:**
- PostgreSQL database (existing, new `calendar` schema)
- auth-service JWKS endpoint (existing)
- account-service household timezone (existing)
- Shared Go modules (existing)

**Environment Variables (new):**
- `GOOGLE_CALENDAR_CLIENT_ID`
- `GOOGLE_CALENDAR_CLIENT_SECRET`
- `CALENDAR_TOKEN_ENCRYPTION_KEY` (32-byte key, base64 or hex encoded)
- `SYNC_INTERVAL_MINUTES` (default: 15)

---

## Timeline Estimates

| Phase | Effort | Dependencies |
|-------|--------|-------------|
| Phase 1: Backend Foundation | XL | None |
| Phase 2: Source/Event/Sync | XL | Phase 1 |
| Phase 3: Infrastructure | M | Phase 1 |
| Phase 4: Frontend | XL | Phase 2 (API), Phase 3 (routing) |
| Phase 5: Testing & Polish | L | Phase 4 |

Phases 1-2 are sequential (sync depends on connections). Phase 3 can run in parallel with Phase 2. Phase 4 can start once Phase 2 API is stable. Phase 5 follows Phase 4.
