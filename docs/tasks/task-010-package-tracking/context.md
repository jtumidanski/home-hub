# Package Tracking — Context

Last Updated: 2026-03-26

---

## Key Files

### Reference Services (follow these patterns)

| File | Why it matters |
|------|---------------|
| `services/calendar-service/cmd/main.go` | Service entry point pattern (config, DB, server, background workers) |
| `services/calendar-service/internal/event/entity.go` | GORM entity pattern with `Migration()`, indexes, table name |
| `services/calendar-service/internal/event/model.go` | Immutable domain model with accessors |
| `services/calendar-service/internal/event/builder.go` | Fluent builder enforcing invariants |
| `services/calendar-service/internal/event/processor.go` | Business logic layer |
| `services/calendar-service/internal/event/provider.go` | Database access (lazy provider) |
| `services/calendar-service/internal/event/resource.go` | HTTP request/response mapping |
| `services/calendar-service/internal/event/rest.go` | Route registration and HTTP handlers |
| `services/calendar-service/internal/sync/sync.go` | Background worker pattern (ticker, polling loop) |
| `services/calendar-service/internal/crypto/crypto.go` | Token encryption pattern (AES-256-GCM) |
| `services/calendar-service/internal/config/config.go` | Config struct from env vars |

### Frontend Reference Files

| File | Why it matters |
|------|---------------|
| `frontend/src/App.tsx` | Route registration — add `/app/packages` here |
| `frontend/src/components/features/navigation/nav-config.ts` | Sidebar nav — add Packages entry here |
| `frontend/src/components/features/navigation/app-shell.tsx` | Layout shell with sidebar and content area |
| `frontend/src/lib/hooks/api/use-calendar.ts` | React Query hooks pattern for a service |
| `frontend/src/services/api/calendar.ts` | API service class pattern |
| `frontend/src/services/api/base.ts` | Base service with common HTTP methods |
| `frontend/src/types/models/calendar.ts` | TypeScript model types pattern |
| `frontend/src/lib/schemas/recipe.schema.ts` | Zod validation schema pattern |
| `frontend/src/pages/CalendarPage.tsx` | Page component pattern, calendar overlay target |
| `frontend/src/pages/DashboardPage.tsx` | Dashboard — add summary widget here |
| `frontend/src/components/features/calendar/all-day-event-row.tsx` | Calendar all-day event rendering — overlay integration point |

### Infrastructure Files

| File | Why it matters |
|------|---------------|
| `deploy/compose/docker-compose.yml` | Add package-service container |
| `deploy/compose/nginx.conf` | Add `/api/v1/packages` route |
| `go.work` | Add `./services/package-service` entry |
| `.github/workflows/` | CI pipeline — may need package-service build job |

### Documentation

| File | Why it matters |
|------|---------------|
| `docs/architecture.md` | Update with package-service routing and description |
| `docs/tasks/task-010-package-tracking/prd.md` | Full PRD with requirements |
| `docs/tasks/task-010-package-tracking/api-contracts.md` | API endpoint specifications |
| `docs/tasks/task-010-package-tracking/data-model.md` | Database schema details |

---

## Key Decisions

1. **CarrierClient interface** — Provider pattern from the start. Each carrier implements `Track(trackingNumber) → TrackingResult`. Zero extra cost since normalization is needed anyway.

2. **Frontend-only calendar integration** — No cross-service API calls. The frontend queries both calendar-service and package-service independently, then renders package ETAs as styled all-day events alongside calendar events.

3. **Pre-transit on "not found"** — When a carrier returns "not found" on initial poll, the package is still created as `pre_transit` with a user-facing warning. This avoids rejecting valid pre-ship labels.

4. **25-package active limit** — Keeps all three carriers within daily rate budgets even with aggressive polling. Configurable via env var.

5. **Privacy model mirrors calendar-service** — Private packages show "Package" placeholder to other household members (no tracking number, label, notes, or status details). 403 on detail endpoint for non-owners.

6. **Adaptive polling** — 30m default, 15m for `out_for_delivery`, stop for `delivered`/`exception`/`stale`. Stale after 14 days with no status change.

7. **Schema name: `package`** — Following the pattern of `calendar`, `productivity`, `recipe`, `weather`.

---

## Dependencies Between Tasks

```
Phase 1 (Backend Foundation)
  ├── 1. Service scaffold ← no deps
  ├── 2. Entity & model ← 1
  ├── 3. Builder ← 2
  ├── 4. Processor & provider ← 3
  ├── 5. Resource & REST ← 4
  ├── 6. Tracking event domain ← 1
  └── 7. Carrier detection ← 1

Phase 2 (Carrier Integration) ← Phase 1
  ├── 8. CarrierClient interface ← no deps
  ├── 9. OAuth token management ← 8
  ├── 10. USPS client ← 8, 9
  ├── 11. UPS client ← 8, 9
  ├── 12. FedEx client ← 8, 9
  ├── 13. Initial poll on create ← 5, 6, 10|11|12
  └── 14. Manual refresh ← 13

Phase 3 (Background Workers) ← Phase 2
  ├── 15. Polling scheduler ← 13
  ├── 16. Poll execution ← 15
  ├── 17. Stale detection ← 16
  └── 18. Archive/cleanup job ← 4

Phase 4 (Infrastructure) ← Phase 1
  ├── 19. Docker Compose ← 1
  ├── 20. Nginx routing ← 19
  ├── 21. Go workspace ← 1
  └── 22. Env template ← 19

Phase 5 (Frontend — Package List) ← Phase 1, Phase 4
  ├── 23. TypeScript types ← no deps
  ├── 24. API service ← 23
  ├── 25. React Query hooks ← 24
  ├── 26. Zod schemas ← 23
  ├── 27. Package list page ← 25, 28
  ├── 28. Package card ← 23
  ├── 29. Package detail/expand ← 28
  ├── 30. Create package dialog ← 25, 26
  ├── 31. Quick actions ← 25, 28
  ├── 32. Sidebar nav ← no deps
  └── 33. Route registration ← 27

Phase 6 (Calendar & Dashboard) ← Phase 5
  ├── 34. Calendar overlay ← 25
  └── 35. Dashboard widget ← 25

Phase 7 (Testing & Docs) ← all phases
  ├── 36. Backend tests ← Phase 1-3
  ├── 37. Frontend tests ← Phase 5-6
  ├── 38. Service documentation ← Phase 1-3
  └── 39. Bruno collection ← Phase 1-2
```

---

## External API References

| Carrier | API | Auth | Rate Limit | Docs |
|---------|-----|------|------------|------|
| USPS | Tracking API v3 | OAuth 2.0 client credentials | ~1000/day (soft) | developer.usps.com |
| UPS | Tracking API v1 | OAuth 2.0 client credentials | 250/day | developer.ups.com |
| FedEx | Track API v1 | OAuth 2.0 client credentials | 500/day | developer.fedex.com |

All three use OAuth 2.0 client credentials flow. Tokens are short-lived and must be refreshed before expiry.

---

## Naming Conventions

- Service directory: `services/package-service/`
- Database schema prefix: `package_` (tables: `package_packages`, `package_tracking_events`, `package_carrier_tokens`)
- Go package names: `package` → use `pkg` to avoid Go keyword collision (or name the domain `tracking`)
- API base path: `/api/v1/packages`
- Frontend route: `/app/packages`

**Important**: `package` is a Go reserved keyword. The internal domain package will need an alternative name. Options:
- `internal/tracking/` — for the package domain (recommended)
- `internal/pkg/` — shorter but less descriptive
- `internal/parcel/` — avoids collision, semantically similar
