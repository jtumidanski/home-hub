# Task 009: Calendar Sync — Key Context

Last Updated: 2026-03-26

---

## Key Files — Existing Patterns to Follow

### Service Scaffold
- `services/account-service/cmd/main.go` — reference for service entry point pattern (logger, config, tracing, db, server)
- `services/weather-service/cmd/main.go` — reference for background ticker pattern (weather cache refresh is analogous to calendar sync)
- `services/auth-service/internal/config/config.go` — reference for OAuth-related config (OIDC client ID/secret)

### Domain Layer Pattern
- `services/account-service/internal/tenant/` — complete reference for the 7-file domain pattern (model, entity, builder, processor, provider, resource, rest)
- `services/account-service/internal/household/` — reference for household-scoped domain

### Google OAuth Flow
- `services/auth-service/internal/authflow/` — reference for OAuth state handling, code exchange, user creation
- `services/auth-service/internal/oidc/` — reference for OIDC discovery, token exchange, user info fetch (calendar-service will use Google's OAuth endpoints directly, not OIDC discovery, but token exchange pattern is similar)

### Shared Libraries
- `shared/go/auth/` — JWT validation middleware, JWKS fetching (reuse as-is)
- `shared/go/database/` — GORM connection, tenant callbacks, migration orchestration (reuse as-is)
- `shared/go/server/` — handler registration, JSON:API helpers (reuse as-is)
- `shared/go/tenant/` — tenant context extraction (reuse as-is)
- `shared/go/logging/` — structured logging + tracing init (reuse as-is)

### Frontend Patterns
- `frontend/src/services/api/` — API service class pattern
- `frontend/src/lib/hooks/api/` — React Query hook pattern with key factories
- `frontend/src/components/features/` — feature component organization
- `frontend/src/pages/` — page component pattern
- `frontend/src/components/features/navigation/` — sidebar nav structure (add Calendar entry)
- `frontend/src/App.tsx` — route registration

### Infrastructure
- `deploy/compose/docker-compose.yml` — add calendar-service container
- `deploy/compose/nginx.conf` — add `/api/v1/calendar/` route
- `deploy/k8s/` — K8s manifests (add calendar-service deployment)
- `.github/workflows/` — CI workflows (add calendar-service build/test)

---

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Separate OAuth flow | Calendar-service manages its own Google OAuth (not auth-service) | Different scopes, different token lifecycle, clean service boundary |
| State validation | Server-side DB table (not cookie) | More secure — cookie approach used by auth-service works but DB state is more robust for a JSON API flow |
| Token encryption | AES-256-GCM in application layer | Not in GORM hooks — keeps entity layer simple, encryption/decryption explicit in processor |
| Sync token per source | Stored on `calendar_sources`, not `calendar_connections` | Each Google Calendar has its own sync token for incremental sync |
| Color assignment | Creation-order based, stored on events | Deterministic, no join needed at query time |
| Privacy masking | Server-side in event resource Transform | Single point of enforcement, frontend never sees private data for non-owners |
| No pagination (MVP) | 90-day max query range instead | Sufficient for the 7-day view + nav; avoids complexity |
| Custom calendar grid | No third-party library | Full control over styling and behavior; shadcn/ui doesn't include a calendar grid |
| Frontend polling | React Query 60s staleTime + refetchOnWindowFocus | Simple, effective, no WebSocket complexity for MVP |

---

## Dependencies Between Components

```
Phase 1: Service Scaffold
    ↓
Phase 1: Token Encryption
    ↓
Phase 1: OAuth State Domain
    ↓
Phase 1: Connection Domain ←── Phase 1: Google Calendar API Client
    ↓
Phase 2: Source Domain
    ↓
Phase 2: Event Domain
    ↓
Phase 2: Sync Engine (depends on: Connection, Source, Event, GoogleCal client)
    ↓
Phase 3: Infrastructure (can start after Phase 1, must complete before Phase 4 E2E testing)
    ↓
Phase 4: Frontend (depends on: Phase 2 API surface, Phase 3 routing)
    ↓
Phase 5: Integration Testing
```

---

## External API References

- **Google Calendar API v3:** https://developers.google.com/calendar/api/v3/reference
  - CalendarList.list: fetches user's calendar list
  - Events.list: fetches events with sync token support, singleEvents=true for recurring expansion
- **Google OAuth 2.0:** https://developers.google.com/identity/protocols/oauth2/web-server
  - Authorization endpoint: `https://accounts.google.com/o/oauth2/v2/auth`
  - Token endpoint: `https://oauth2.googleapis.com/token`
  - Revoke endpoint: `https://oauth2.googleapis.com/revoke`
  - Scope: `https://www.googleapis.com/auth/calendar.readonly`

---

## Environment Variables (New)

| Variable | Service | Required | Description |
|----------|---------|----------|-------------|
| `GOOGLE_CALENDAR_CLIENT_ID` | calendar-service | Yes | Google OAuth client ID |
| `GOOGLE_CALENDAR_CLIENT_SECRET` | calendar-service | Yes | Google OAuth client secret |
| `CALENDAR_TOKEN_ENCRYPTION_KEY` | calendar-service | Yes | 32-byte AES-256 key (base64) |
| `SYNC_INTERVAL_MINUTES` | calendar-service | No (default 15) | Background sync interval |

---

## Spec Documents

- [prd.md](prd.md) — Full product requirements
- [api-contracts.md](api-contracts.md) — Detailed endpoint specifications
- [data-model.md](data-model.md) — Database schema
- [ux-flow.md](ux-flow.md) — UX wireframes and interaction flows
