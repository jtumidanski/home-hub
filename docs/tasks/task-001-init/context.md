# task-001-init — Context

Last Updated: 2026-03-24

---

## Key Files

| File | Purpose |
|------|---------|
| docs/tasks/task-001-init/prd.md | Product requirements — scope, constraints, non-goals |
| docs/tasks/task-001-init/appendix.md | Design constraints and decisions already made |
| docs/architecture.md | Service responsibilities, routing, auth model, persistence |
| docs/api.md | Full JSON:API specification for all endpoints |
| docs/schema.md | Database schema for all 3 services |
| docs/bootstrap-plan.md | 12-milestone execution plan (high-level) |
| docs/development.md | Development guide — tooling, scripts, conventions |
| docs/decisions.md | Decision log (template only, no decisions recorded yet) |
| DOCS.md | Documentation contract — how service docs must be structured |
| CLAUDE.md | Project-level instructions for Claude |

---

## Key Decisions Already Made

1. **Microservice architecture** — auth-service, account-service, productivity-service, frontend
2. **Go backend** — net/http, GORM, Logrus, OpenTelemetry
3. **React frontend** — Vite, ShadCN, no SSR
4. **PostgreSQL** — per-service schemas (auth, account, productivity), no cross-schema FKs
5. **JSON:API** — resource-oriented endpoints, versioned at /api/v1
6. **JWT auth** — asymmetric (RS256), JWKS endpoint, HTTP-only cookies, short-lived access tokens
7. **Multi-tenancy** — User → Tenant → Household → Membership hierarchy, all rows include tenant_id
8. **Context endpoint** — /contexts/current derives tenant, household, role, theme server-side
9. **Path-prefix routing** — nginx (local) and Ingress (k8s) route by /api/v1/<resource>
10. **CI from start** — GitHub Actions, PR validation, main branch publishes to GHCR
11. **UUID IDs** — generated in application layer
12. **Forward-only migrations** — run on service startup via GORM

---

## Key Dependencies Between Phases

```
Phase 1 (Scaffold)
  └─► Phase 2 (Baseline)
        ├─► Phase 3 (CI) — needs compilable services
        ├─► Phase 4 (Auth)
        │     └─► Phase 6 (Frontend Auth) — needs auth APIs
        ├─► Phase 5 (Account)
        │     └─► Phase 6 (Frontend Auth) — needs context API
        └─► Phase 7 (Productivity)
              └─► Phase 8 (Frontend UI) — needs productivity APIs
                    └─► Phase 9 (Local Env) — needs all services working
                          └─► Phase 10 (k3s)

Phase 11 (Bruno) — can build incrementally alongside Phases 4-7
Phase 12 (Renovate) — after everything else
```

---

## Parallelization Opportunities

- **Phases 4 + 5 + 7** can be built in parallel after Phase 2 (independent services)
- **Phase 3** can start as soon as services compile (after Phase 1, refined during Phase 2)
- **Phase 11** (Bruno) can be built incrementally as each service's API is completed
- **Frontend phases (6, 8)** must wait for their corresponding backend phases

---

## External Integration Points

| Integration | Service | Notes |
|------------|---------|-------|
| Google OIDC | auth-service | Requires client_id + secret, redirect URI configuration |
| PostgreSQL | all services | Each service uses its own schema, same database |
| GHCR | CI | Requires GitHub token with packages:write |
| k3s | deployment | External cluster, secrets managed outside manifests |

---

## Out of Scope (from PRD)

Do not implement in this task:
- Notifications / push
- Recurring reminders
- Invitations flow
- Kiosk UI
- Mobile apps
- Offline support
- Email
- Audit UI
- Background workers beyond scheduler
