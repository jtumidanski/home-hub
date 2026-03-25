# Home Hub — Appendix A: Design Constraints & Decisions

This document records architectural and technical decisions that have already been made.

These constraints must be respected during detailed design.

This appendix exists to avoid re-deciding core system direction during implementation.

This document does NOT define final implementation details.  
It defines boundaries within which the design must occur.

---

# 1. Service Architecture

The system must use a microservice architecture.

Required services:

- auth-service
- account-service
- productivity-service
- frontend

Services must be independently deployable.

Services must communicate via HTTP.

Services must not share tables.

Each service owns its schema.

---

# 2. Routing Model

All external requests must pass through a reverse proxy / ingress.

Routing must use path prefix routing.

Example:

    /api/v1/auth/*
    /api/v1/tenants/*
    /api/v1/tasks/*

The frontend must be a separate container.

The frontend must not be embedded in backend services.

---

# 3. API Versioning

All API endpoints must be versioned from the beginning.

Base path:

    /api/v1

Breaking changes must require a new version.

---

# 4. API Style

The API must be resource oriented.

Examples:

    /tasks
    /tasks/{id}
    /tasks/restorations

    /reminders
    /reminders/snoozes
    /reminders/dismissals

    /summary/tasks

Context endpoint:

    /contexts/current

JSON:API conventions should be followed where practical.

Exact compliance is not required, but resource structure should be similar.

---

# 5. Authentication Model

Authentication must use OIDC.

Auth service must:

- manage providers
- issue JWT
- expose JWKS
- store refresh tokens
- expose /users/me

Downstream services must:

- validate JWT locally
- not call auth service per request
- trust JWKS

Cookies must be HTTP-only.

Access tokens must be short-lived.

Refresh tokens must be server-side.

---

# 6. Identity vs Context

User identity belongs to auth-service.

Tenant / household context belongs to account-service.

Other services must not store identity information beyond user_id.

Context must be resolved via:

    /contexts/current

Frontend must not derive context.

---

# 7. Multi-Tenant Model

Hierarchy:

User
 → Tenant
   → Household
     → Membership

Rules:

- one tenant per user initially
- multiple households allowed
- roles per household
- preferences per tenant per user

All data must include:

    tenant_id

Household data must include:

    household_id

---

# 8. Database Model

Database: PostgreSQL

ORM: Gorm

Each service must use its own schema.

Schemas:

    auth
    account
    productivity

IDs must be UUID.

Timestamps must be UTC.

Soft delete must use:

    deleted_at

Migrations must run on startup.

Migrations must be forward-only.

No cross-schema foreign keys required.

Indexes required on:

- tenant_id
- household_id
- user_id

---

# 9. Context Model

Account service must support:

- tenants
- households
- memberships
- preferences

Preferences must include:

- theme
- active household

Context endpoint must return:

- tenant
- household
- role
- theme
- memberships

Endpoint:

    /contexts/current

Context must be derived server-side.

---

# 10. Productivity Model

Productivity service must support:

Tasks:

- title
- notes
- due date
- status
- soft delete
- restore window

Reminders:

- title
- scheduled time
- snooze
- dismiss
- one-time only (v1)

Summary:

- task summary
- reminder summary
- dashboard summary

Summary must be derived.

Summary must not be stored.

All rows must include:

    tenant_id
    household_id

---

# 11. Frontend Constraints

Frontend must be:

- standalone container
- React
- ShadCN
- light/dark theme

Frontend must:

- call API
- not hold business state
- bootstrap using:

    /users/me
    /contexts/current

No server-side rendering required.

---

# 12. Runtime Constraints

System must run locally using Docker Compose.

System must run in Kubernetes.

Helm is not required.

Plain YAML is acceptable.

Secrets must be external.

Config must use environment variables.

Scripts must exist for local development.

---

# 13. CI/CD Constraints

Repository must include CI from start.

CI must:

- build changed services
- test changed services
- build frontend
- build docker images
- publish snapshot images

Renovate must be configured.

Bruno must be used for API testing.

Scripts folder must exist.

---

# 14. Logging and Tracing

Services must:

- use structured logs
- include request id
- include user id
- include tenant id

Tracing must support OpenTelemetry.

Metrics optional.

---

# 15. Out of Scope for v1

Do not implement:

- kiosk UI
- push notifications
- recurring reminders
- invitations
- mobile apps
- offline support
- audit UI

These may be added later.

---

# 16. Design Phase Expectations

The developer performing detailed design must produce:

- service-level design
- API confirmation
- DB migration plan
- compose design
- k8s design
- shared package design

before implementation begins.
