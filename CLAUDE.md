# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## Project Overview

**Home Hub** is a modular, microservice-based home information platform designed to power kiosk-style displays and administrative interfaces for shared households. The project is currently in the planning/architecture phase with comprehensive documentation but no implementation yet.

**Key Architecture Documents:**
- `/docs/PROJECT_KNOWLEDGE.md` - Complete architecture specification
- `/dev/README.md` - Development methodology and dev docs pattern

---

## Core Architectural Principles

### 1. Microservices by Domain
Each domain (users, calendar, weather, tasks, meals, reminders) owns:
- Its own PostgreSQL database schema
- Independent REST API
- Dedicated migrator service
- Domain-specific workers (background jobs)

### 2. Security Model
- **Gateway pattern**: Single public entry point (`gateway` service)
- **Header-based context**: Gateway injects `X-Tenant-ID`, `X-Household-ID`, `X-User-ID`, and service JWT

### 3. Communication Flow
```
Frontend → nginx proxy → Gateway → Domain Services → PostgreSQL
                                     ↓
                                  Workers
```

**Critical**: Services NEVER receive tenant information in request/response bodies - only via headers. The database session sets `app.household_id`.

### 4. Stateless Design
- All state persisted in PostgreSQL (one schema per service)
- Workers handle time-based logic (syncs, rollovers, sweeps)
- Frontend uses polling (30s interval), not WebSockets
- Suitable for kiosk deployments on Raspberry Pi 4

---

## Repository Structure (Planned)

```
home-hub/                          # Mono-repo for all services
  apps/
    gateway/                       # Public API, auth, routing
    dashboard-composer/            # Aggregates data from services
    svc-users/                     # Tenants, households, members
    svc-calendar/                  # Google Calendar integration
    svc-weather/                   # Weather caching
    svc-tasks/                     # Tasks CRUD + rollover
    svc-meals/                     # Recipes + weekly planner
    svc-reminders/                 # Reminders + snooze/dismiss
    workers/                       # Background jobs per domain
    migrators/                     # One per service (AutoMigrate)
    kiosk/                         # React/Tailwind (Next.js)
    admin/                         # React/Tailwind (Next.js)
```

---

## Service Responsibilities

| Service | Purpose | Database Schema | Notes |
|---------|---------|-----------------|-------|
| **gateway** | Entry point, auth, routing | None | Stateless |
| **dashboard-composer** | Aggregates service data | None | Read-only |
| **svc-users** | Tenants, households, invites, devices | `users`, `tenants`, `households`, `devices` | Defines household timezone |
| **svc-calendar** | Calendar sync and storage | `calendars`, `events`, `tokens` | 5-min sync worker |
| **svc-weather** | Weather caching | `locations`, `current`, `forecast` | 30-min refresh |
| **svc-tasks** | Tasks CRUD and rollover | `tasks`, `audit_log` | Daily rollover worker |
| **svc-meals** | Recipes and meal planning | `recipes`, `plan_weeks`, `plan_items` | Sunday planner worker |
| **svc-reminders** | Reminders CRUD and snooze | `reminders`, `audit_log` | 1-min sweep worker |

---

## Workers and Scheduling

| Worker | Schedule | Description |
|--------|----------|-------------|
| `calendar-sync` | Every 5m | Sync Google Calendar events |
| `weather-sync` | Every 30m | Refresh weather forecast |
| `tasks-rollover` | 00:01 local | Rollover incomplete tasks |
| `meals-planner` | Sunday 08:00 local | Generate weekly plan |
| `reminders-sweep` | Every 1m | Manage reminders, respect quiet hours (00:00-05:00) |

Workers are headless and communicate with their owning service databases.

---

## Development Methodology: Dev Docs Pattern

This project uses a **three-file dev docs pattern** to maintain context across Claude Code sessions and context resets.

### Pattern Structure
For each complex task, create:
```
dev/active/[task-name]/
├── [task-name]-plan.md       # Strategic implementation plan
├── [task-name]-context.md    # Current state, key files, decisions
└── [task-name]-tasks.md      # Checklist format task tracking
```

### When to Use
- ✅ Complex multi-day tasks
- ✅ Features with many moving parts
- ✅ Work spanning multiple sessions
- ✅ Refactoring large systems
- ❌ Simple bug fixes
- ❌ Single-file changes

**Rule of thumb**: If it takes >2 hours or spans sessions, use dev docs.

### Available Slash Commands
- `/dev-docs [description]` - Create new task documentation
- `/dev-docs-update` - Update docs before context reset

### Key Practices
1. **Update frequently**: Especially `[task-name]-context.md` SESSION PROGRESS section
2. **After context reset**: Read all three files to resume instantly
3. **Mark tasks complete**: Update tasks.md immediately after completion
4. **Document decisions**: Capture "why" not just "what"

See `/dev/README.md` for comprehensive guide.

---

## Implementation Guidelines

### Service Development Pattern
When implementing a new service:

1. **Database Schema** (`migrators/`)
   - Define GORM models with tenant/household foreign keys
   - Create `Migration()` function for AutoMigrate

2. **Service API** (`apps/svc-*/`)
   - REST endpoints following OpenAPI spec in `packages/openapi/`
   - Extract context from headers: `X-Tenant-ID`, `X-Household-ID`, `X-User-ID`
   - Set database session variables before queries
   - Never include tenant/household IDs in response bodies

3. **Worker Logic** (`apps/workers/`)
   - Implement scheduled tasks for domain
   - Use cron or similar for scheduling
   - Respect household timezones (from `svc-users`)

4. **OpenAPI Spec** (`packages/openapi/`)
   - Define request/response schemas
   - Document all endpoints
   - Generate clients for `dto-go` and `dto-js`

### Header Flow Pattern
```go
// In service handlers
tenantID := r.Header.Get("X-Tenant-ID")
householdID := r.Header.Get("X-Household-ID")
userID := r.Header.Get("X-User-ID")

// Set session variables before query
db.Exec("SET app.tenant_id = ?", tenantID)
db.Exec("SET app.household_id = ?", householdID)

var tasks []Task
db.Find(&tasks) // Only returns tenant's household's tasks
```

---

## Local Development (When Implemented)

### Expected Commands
```bash
# Start all services with Docker Compose
task dev

# Access points
# http://localhost:3000/kiosk/
# http://localhost:3000/admin/
# http://localhost:3000/api/
```

### Port Assignments (Planned)
- nginx proxy: `3000`
- gateway: `8080`
- kiosk: `5173`
- admin: `5174`
- Services: `8081-8087`

---

## Key Non-Functional Requirements

| Area | Goal |
|------|------|
| **Dashboard load (Pi 4)** | < 2 seconds |
| **API latency (P99)** | < 300 ms |
| **Offline cache window** | 1 day |
| **Audit retention** | 90 days |
| **Frontend polling interval** | 30 seconds |
| **Quiet hours (reminders)** | 00:00–05:00 |

---

## Critical Patterns to Follow

### ✅ DO
- Extract tenant context from headers, never from request body
- Use mono-repo structure with service isolation
- Follow gateway pattern for all external requests
- Implement workers as headless services
- Test cross-tenant isolation rigorously
- Use polling for kiosk UI updates (not WebSockets)
- Create OpenAPI specs before implementation
- Use shared libraries in `packages/shared-go/`

### ❌ NEVER
- Include `tenant_id` or `household_id` in API response bodies
- Allow direct database access from frontend
- Use WebSockets for kiosk updates (breaks caching)
- Allow cross-service direct database queries
- Hard-code tenant or household logic in business code
- Implement authentication outside of gateway
- Create services without corresponding migrators

---

## AI-Assisted Development Guidelines

### When Starting a New Service
1. Use `/dev-docs implement [service-name]` to create comprehensive plan
2. Reference existing architecture in `/docs/PROJECT_KNOWLEDGE.md`
3. Follow the service development pattern above
4. Implement in phases:
   - Phase 1: Database schema + migrator
   - Phase 2: REST API + OpenAPI spec
   - Phase 3: Worker logic (if applicable)
   - Phase 4: Integration tests
   - Phase 5: Client generation + frontend integration

### When Resuming Work
1. Read relevant dev docs in `/dev/active/[task-name]/`
2. Check `[task-name]-context.md` SESSION PROGRESS section
3. Review `[task-name]-tasks.md` for remaining work
4. Continue from the last in-progress task

### Before Context Reset
1. Run `/dev-docs-update` to capture current state
2. Mark all completed tasks in `[task-name]-tasks.md`
3. Update SESSION PROGRESS in `[task-name]-context.md`
4. Document any blockers or decisions made

---

## Technology Stack (Planned)

- **Backend**: Go 1.24+, PostgreSQL, Kafka (for future event streaming)
- **Frontend**: React, Tailwind CSS, Next.js
- **DevOps**: Docker, Kubernetes, Helm
- **Development**: Go modules, npm, Taskfile, Nix
- **Authentication**: Google OIDC
- **API Standard**: OpenAPI 3.0, REST with JSON

---

## Project Status

**Current Phase**: Architecture and documentation complete, implementation not started

**Next Steps**:
1. Set up mono-repo structure (`apps/`, `packages/`)
2. Implement `gateway` service with Google OIDC
4. Create shared libraries in `packages/shared-go/`
5. Implement remaining domain services following the pattern

**Use `/dev-docs [task]`** to create structured implementation plans for each service.
