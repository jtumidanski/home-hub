# task-001-init — Task Checklist

Last Updated: 2026-03-24

---

## Phase 1 — Monorepo Scaffold

- [x] 1.1 Create top-level go.work
- [x] 1.2 Create services/auth-service skeleton (go.mod, cmd/, internal/)
- [x] 1.3 Create services/account-service skeleton
- [x] 1.4 Create services/productivity-service skeleton
- [x] 1.5 Create shared/go/auth module
- [x] 1.6 Create shared/go/http module
- [x] 1.7 Create shared/go/logging module
- [x] 1.8 Create shared/go/testing module
- [x] 1.9 Initialize frontend (React + Vite + ShadCN)
- [x] 1.10 Create deploy/compose/ and deploy/k8s/ directories
- [x] 1.11 Create scripts/ directory with placeholder scripts
- [x] 1.12 Create bruno/ directory structure
- [x] 1.13 Verify `go work sync` succeeds

## Phase 2 — Common Service Baseline

- [x] 2.1 Implement env-based config loading
- [x] 2.2 Implement Logrus structured JSON logging
- [x] 2.3 Implement OpenTelemetry tracing initialization
- [x] 2.4 Implement request ID middleware
- [x] 2.5 Implement GORM database connection setup
- [x] 2.6 Implement UUID generation in application layer
- [x] 2.7 Implement startup migration runner
- [x] 2.8 Implement /healthz and /readyz endpoints
- [x] 2.9 Implement JWT validation helper (shared/go/auth)
- [x] 2.10 Implement tenant context middleware
- [x] 2.11 Implement service bootstrap pattern (main.go template)

## Phase 3 — CI Foundation

- [x] 3.1 Create build scripts
- [x] 3.2 Create test scripts
- [x] 3.3 Create lint scripts
- [x] 3.4 Create local-up.sh and local-down.sh
- [x] 3.5 Create PR GitHub Actions workflow
- [x] 3.6 Create main branch GitHub Actions workflow
- [x] 3.7 Create Dockerfiles for each service and frontend

## Phase 4 — Auth Service

- [x] 4.1 Create auth schema migrations
- [x] 4.2 Implement GORM entities for auth tables
- [x] 4.3 Implement domain models (User, ExternalIdentity, OIDCProvider, RefreshToken)
- [x] 4.4 Implement OIDC provider management (Google)
- [x] 4.5 Implement OIDC login flow (GET /auth/login/{provider})
- [x] 4.6 Implement OIDC callback (GET /auth/callback/{provider})
- [x] 4.7 Implement asymmetric JWT issuance
- [x] 4.8 Implement JWKS endpoint
- [x] 4.9 Implement refresh token sessions
- [x] 4.10 Implement HTTP-only cookie handling
- [x] 4.11 Implement token refresh (POST /auth/token/refresh)
- [x] 4.12 Implement logout (POST /auth/logout)
- [x] 4.13 Implement GET /users/me
- [x] 4.14 Implement GET /auth/providers
- [x] 4.15 Write unit tests for auth domain logic

## Phase 5 — Account Service

- [x] 5.1 Create account schema migrations
- [x] 5.2 Implement GORM entities
- [x] 5.3 Implement domain models (Tenant, Household, Membership, Preference)
- [x] 5.4 Implement tenant CRUD
- [x] 5.5 Implement household CRUD
- [x] 5.6 Implement membership management
- [x] 5.7 Implement preference management
- [x] 5.8 Implement GET /contexts/current
- [x] 5.9 Implement context includes support
- [x] 5.10 Implement fallback when active household is invalid
- [x] 5.11 Write unit tests for account domain logic

## Phase 6 — Frontend Auth + Onboarding

- [x] 6.1 Set up React Router with routes
- [x] 6.2 Implement JSON:API client
- [x] 6.3 Implement login page
- [x] 6.4 Implement auth bootstrap (users/me + contexts/current)
- [x] 6.5 Implement onboarding flow
- [x] 6.6 Implement ShadCN theme toggle
- [x] 6.7 Implement household switcher
- [x] 6.8 Implement protected route wrapper
- [x] 6.9 Implement app shell (sidebar/nav, layout)

## Phase 7 — Productivity Service

- [x] 7.1 Create productivity schema migrations
- [x] 7.2 Implement GORM entities
- [x] 7.3 Implement domain models (Task, Reminder, etc.)
- [x] 7.4 Implement task CRUD
- [x] 7.5 Implement task restoration
- [x] 7.6 Implement reminder CRUD
- [x] 7.7 Implement reminder snooze
- [x] 7.8 Implement reminder dismissal
- [x] 7.9 Implement task summary
- [x] 7.10 Implement reminder summary
- [x] 7.11 Implement dashboard summary
- [x] 7.12 Implement include support for summary endpoints
- [x] 7.13 Write unit tests for productivity domain logic

## Phase 8 — Frontend Productivity UI

- [x] 8.1 Implement dashboard page
- [x] 8.2 Implement tasks list page
- [x] 8.3 Implement task detail/edit view
- [x] 8.4 Implement task restore UI
- [x] 8.5 Implement reminders list page
- [x] 8.6 Implement reminder detail/edit view
- [x] 8.7 Implement snooze UI
- [x] 8.8 Implement dismiss UI
- [x] 8.9 Implement settings page
- [x] 8.10 Implement households management page

## Phase 9 — Local Environment

- [ ] 9.1 Create docker-compose.yml
- [ ] 9.2 Create nginx.conf with path-prefix routing
- [ ] 9.3 Create .env.example
- [ ] 9.4 Verify end-to-end login flow locally
- [ ] 9.5 Verify all API endpoints work through proxy

## Phase 10 — k3s Deployment

- [ ] 10.1 Create k8s manifests for auth-service
- [ ] 10.2 Create k8s manifests for account-service
- [ ] 10.3 Create k8s manifests for productivity-service
- [ ] 10.4 Create k8s manifests for frontend
- [ ] 10.5 Create Ingress YAML with path-prefix routing
- [ ] 10.6 Document secret management approach

## Phase 11 — Bruno Collections

- [ ] 11.1 Create auth collection
- [ ] 11.2 Create account collection
- [ ] 11.3 Create productivity collection
- [ ] 11.4 Create environment files (local, prod)

## Phase 12 — Renovate + Maturity

- [ ] 12.1 Create renovate.json
- [ ] 12.2 Verify all docs are accurate against implementation
- [ ] 12.3 Verify CI is strict (branch protection)
