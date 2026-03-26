# Task 009: Calendar Sync — Task Checklist

Last Updated: 2026-03-26

---

## Phase 1: Backend Foundation

### 1.1 Service Scaffold [L]
- [ ] Create `services/calendar-service/` directory structure
- [ ] Create `cmd/main.go` with logger, config, tracing, database, server init
- [ ] Create `internal/config/config.go` with all env vars (DB, JWKS_URL, PORT, Google OAuth, encryption key, sync interval)
- [ ] Create `go.mod` referencing shared modules
- [ ] Create `Dockerfile` (multi-stage build)
- [ ] Verify service starts, connects to DB, and serves health check

### 1.2 Token Encryption [S]
- [ ] Implement AES-256-GCM encrypt/decrypt functions (key from env var)
- [ ] Unit tests with known test vectors
- [ ] Handle key decoding (base64 or hex)

### 1.3 OAuth State Domain [M]
- [ ] Create `internal/oauthstate/` — entity, model, builder, processor, provider
- [ ] Entity: `calendar_oauth_states` table (id=state UUID, tenant_id, household_id, user_id, redirect_uri, expires_at)
- [ ] Processor: CreateState, ValidateAndConsume (check exists + not expired), CleanupExpired
- [ ] Migration registered in main.go

### 1.4 Connection Domain [XL]
- [ ] Create `internal/connection/` — entity, model, builder, processor, provider, resource, rest
- [ ] Entity: `calendar_connections` table with all columns per data-model.md
- [ ] Model: immutable domain model with accessors
- [ ] Builder: validation for provider, email, status, token encryption on build
- [ ] Processor: Create, GetByID, ListByUserAndHousehold, UpdateStatus, UpdateSyncInfo, Delete (cascade)
- [ ] Provider: query builders for tenant+household, user_id, id filters
- [ ] Resource: JSON:API `calendar-connections` type with Transform (never expose tokens)
- [ ] REST: POST /connections/google/authorize — generate state, build Google OAuth URL, return JSON
- [ ] REST: GET /connections/google/callback — validate state, exchange code, encrypt tokens, create connection, fetch calendar list, trigger first sync, redirect to frontend
- [ ] REST: GET /connections — list user's connections in household
- [ ] REST: DELETE /connections/{id} — revoke token, delete events/sources/connection
- [ ] REST: POST /connections/{id}/sync — manual sync with 5-min rate limit
- [ ] Register routes in main.go under `/api/v1/calendar` prefix

### 1.5 Google Calendar API Client [L]
- [ ] Create `internal/googlecal/` package
- [ ] OAuth token exchange (POST to Google token endpoint with auth code)
- [ ] Token refresh (POST to Google token endpoint with refresh token)
- [ ] Token revocation (POST to Google revoke endpoint)
- [ ] CalendarList.list — fetch user's Google Calendar list
- [ ] Events.list — fetch events with syncToken support, singleEvents=true, timeMin/timeMax
- [ ] Response type structs (CalendarListResponse, EventsResponse, Token, etc.)
- [ ] Exponential backoff on 429/5xx (1s, 2s, 4s, max 3 retries)
- [ ] Structured logging for all API calls

---

## Phase 2: Source, Event & Sync

### 2.1 Source Domain [L]
- [ ] Create `internal/source/` — entity, model, builder, processor, provider, resource, rest
- [ ] Entity: `calendar_sources` table per data-model.md
- [ ] Processor: CreateOrUpdate from Google calendar list, ToggleVisibility, ListByConnection, DeleteByConnection
- [ ] Resource: JSON:API `calendar-sources` type
- [ ] REST: GET /connections/{id}/calendars — list sources
- [ ] REST: PATCH /connections/{id}/calendars/{calId} — toggle visibility

### 2.2 Event Domain [L]
- [ ] Create `internal/event/` — entity, model, builder, processor, provider, resource, rest
- [ ] Entity: `calendar_events` table per data-model.md
- [ ] Processor: BulkUpsert (insert or update by source_id+external_id), DeleteByExternalIDs, DeleteByConnection, DeleteBySource, QueryByHouseholdAndTimeRange
- [ ] Resource: JSON:API `calendar-events` type with privacy masking in Transform
- [ ] Privacy: if visibility=private/confidential AND user_id != requester → title="Busy", description=null, location=null
- [ ] REST: GET /events?start=...&end=... — list household events, enforce 90-day max range, sort by start_time (all-day first)

### 2.3 User Color Assignment [S]
- [ ] Define 8-color palette constant
- [ ] Color assigned based on connection creation order within household (count existing connections)
- [ ] Store color on connection record
- [ ] Denormalize color onto event records during sync

### 2.4 Background Sync Engine [XL]
- [ ] Create `internal/sync/` package
- [ ] Sync orchestrator: runs on time.Ticker (configurable interval from config)
- [ ] Per tick: query all `connected` connections, stagger with 0-60s random jitter per connection
- [ ] Per connection sync:
  - [ ] Refresh access token if expired (using googlecal client)
  - [ ] On token refresh failure → mark connection `disconnected`, skip
  - [ ] Refresh calendar list (add new sources, remove deleted ones)
  - [ ] For each source with visible=true (sync all, filter at query time):
    - [ ] Fetch events with syncToken (or full sync if token invalid/missing)
    - [ ] Upsert events into calendar_events
    - [ ] Delete removed events (from Google's cancelled items)
    - [ ] Update syncToken on source record
  - [ ] Update connection last_sync_at and last_sync_event_count
- [ ] Immediate sync on new connection (async goroutine from callback)
- [ ] Manual sync trigger (called from REST, respects rate limit)
- [ ] Expired OAuth state cleanup (run during sync loop)
- [ ] Start sync loop goroutine from main.go
- [ ] Graceful shutdown (context cancellation)

---

## Phase 3: Infrastructure & Deployment

### 3.1 Docker Compose [S]
- [ ] Add `calendar-service` to `deploy/compose/docker-compose.yml`
- [ ] Add environment variables (DB, JWKS_URL, Google OAuth, encryption key)
- [ ] Add nginx route: `location /api/v1/calendar/`

### 3.2 Kubernetes [M]
- [ ] Create calendar-service Deployment manifest
- [ ] Create calendar-service Service manifest
- [ ] Add Ingress rule for `/api/v1/calendar/`
- [ ] Add Secret references for Google credentials and encryption key

### 3.3 CI/CD [S]
- [ ] Add GitHub Actions workflow for calendar-service (build, test, lint)
- [ ] Add Docker image build: `ghcr.io/<owner>/home-hub-calendar`

### 3.4 Service Documentation [M]
- [ ] Create `services/calendar-service/docs/domain.md`
- [ ] Create `services/calendar-service/docs/rest.md`
- [ ] Create `services/calendar-service/docs/storage.md`

### 3.5 Architecture Update [S]
- [ ] Update `docs/architecture.md` — add calendar-service to core services list, routing table, service responsibilities

---

## Phase 4: Frontend

### 4.1 API Service & Types [M]
- [ ] Create `services/api/calendar.ts` — CalendarService class
  - [ ] `getConnections()` — GET /calendar/connections
  - [ ] `authorizeGoogle(redirectUri)` — POST /calendar/connections/google/authorize
  - [ ] `deleteConnection(id)` — DELETE /calendar/connections/{id}
  - [ ] `getCalendarSources(connectionId)` — GET /calendar/connections/{id}/calendars
  - [ ] `toggleCalendarSource(connectionId, calId, visible)` — PATCH /calendar/connections/{id}/calendars/{calId}
  - [ ] `triggerSync(connectionId)` — POST /calendar/connections/{id}/sync
  - [ ] `getEvents(start, end)` — GET /calendar/events
- [ ] TypeScript types for CalendarConnection, CalendarSource, CalendarEvent

### 4.2 React Query Hooks [M]
- [ ] Create `lib/hooks/api/use-calendar.ts`
- [ ] Key factory: `calendarKeys` (connections, sources, events)
- [ ] `useCalendarConnections()` — staleTime: 60s, refetchOnWindowFocus
- [ ] `useCalendarSources(connectionId)` — enabled when connectionId provided
- [ ] `useCalendarEvents(start, end)` — staleTime: 60s, refetchOnWindowFocus
- [ ] `useConnectGoogleCalendar()` — mutation, navigates to Google on success
- [ ] `useDisconnectCalendar()` — mutation, invalidates connections + events
- [ ] `useToggleCalendarSource()` — mutation, invalidates sources + events
- [ ] `useTriggerSync()` — mutation, invalidates connection on settle

### 4.3 Calendar Grid Component [XL]
- [ ] Create `components/features/calendar/CalendarGrid.tsx`
- [ ] 7-column day layout with hour rows (6 AM – 11 PM default)
- [ ] Day column headers with day name + date
- [ ] Today column visual highlight
- [ ] Current time indicator line (horizontal red line)
- [ ] Scrollable hour rows
- [ ] Event block positioning: calculate top/height from start/end times relative to hour grid
- [ ] Overlapping event detection and side-by-side rendering (split column width)
- [ ] All-day event section above the grid (collapsible)
- [ ] Multi-day events spanning across day columns in all-day section
- [ ] All times converted to household timezone (from useTenant context)

### 4.4 Event Components [L]
- [ ] Create `components/features/calendar/EventBlock.tsx` — colored event block in the grid
- [ ] Create `components/features/calendar/EventPopover.tsx` — click popover with details or "Busy"
- [ ] Create `components/features/calendar/AllDayEventRow.tsx` — all-day/multi-day event display
- [ ] Color coding by userColor attribute
- [ ] Muted styling for "Busy" (private) events
- [ ] User legend component showing connected users and their colors

### 4.5 Connection Management UI [L]
- [ ] Create `components/features/calendar/ConnectCalendarButton.tsx` — initiates OAuth
- [ ] Create `components/features/calendar/ConnectionStatus.tsx` — status, last sync time, sync/disconnect actions
- [ ] Create `components/features/calendar/CalendarSelectionPanel.tsx` — toggle source calendars
- [ ] Create `components/features/calendar/DisconnectDialog.tsx` — confirmation dialog
- [ ] Handle 429 on manual sync (show "Try again in X minutes")

### 4.6 Calendar Page [L]
- [ ] Create `pages/CalendarPage.tsx`
- [ ] Week navigation: prev/next buttons, "Today" button
- [ ] Date range header display (e.g., "March 23 – March 29, 2026")
- [ ] State: current week start date (default: start of current week)
- [ ] Empty state when no connections in household
- [ ] Handle query params: `?connected=true` → success toast, `?error=...` → error toast
- [ ] Connect button in header area
- [ ] Connection status display
- [ ] User legend

### 4.7 Navigation & Routing [S]
- [ ] Add "Calendar" entry to sidebar navigation
- [ ] Add `/app/calendar` route in App.tsx
- [ ] Ensure OAuth callback redirect works (`/app/calendar?connected=true`)

---

## Phase 5: Testing & Polish

### 5.1 Backend Unit Tests [L]
- [ ] Token encryption: encrypt/decrypt round-trip, wrong key fails
- [ ] Privacy masking: private event → Busy for non-owner, full details for owner
- [ ] Color assignment: deterministic order-based, wraps palette
- [ ] Rate limiting: manual sync within 5 min rejected with 429
- [ ] Sync token: incremental sync when token present, full sync when missing/invalid
- [ ] OAuth state: create, validate, reject expired, cleanup

### 5.2 Frontend Verification [M]
- [ ] Calendar grid renders events at correct hour positions
- [ ] Overlapping events split column width correctly
- [ ] All-day events render in dedicated section
- [ ] Privacy masking shows "Busy" for non-owner private events
- [ ] Empty state renders when no connections
- [ ] Week navigation updates date range and refetches events
- [ ] Today highlight and current time indicator display correctly

### 5.3 End-to-End Flow [M]
- [ ] Full OAuth connect → sync → display → disconnect flow
- [ ] Multi-user household with merged events
- [ ] Calendar source toggle hides/shows events
- [ ] Connection status updates (syncing, error, disconnected)
- [ ] Post-connection toast and calendar selection panel

### 5.4 Build Verification [S]
- [ ] All backend services build (verify shared module changes don't break others)
- [ ] Frontend builds without errors
- [ ] Docker compose stack starts with calendar-service
- [ ] Calendar-service migrates schema on first startup
