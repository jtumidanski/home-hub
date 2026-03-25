# Home Hub — Product Requirements Document (PRD)

Version: v1  
Status: Draft  
Audience: Developers / Architects  
Purpose: Define product scope and constraints for deep technical design

---

# 1. Overview

Home Hub is a multi-tenant household productivity application designed to manage:

- households
- roles
- tasks
- reminders
- daily summary dashboard

The system is intended to run:

- locally via Docker Compose
- in Kubernetes (k3s)
- behind a reverse proxy using path-prefix routing

The product must be implemented as a microservice-based system with strict service boundaries.

This document defines:

- product scope
- functional requirements
- non-functional requirements
- architectural constraints
- API expectations
- persistence expectations

This document does NOT define:

- internal code structure
- exact package layout
- implementation details
- migration tooling choice
- HTTP framework choice beyond constraints

Those are left to the developer design phase.

---

# 2. Goals

Primary goals:

- multi-user household management
- simple task tracking
- simple reminder tracking
- derived dashboard summary
- clean multi-tenant model
- reproducible local environment
- deployable to Kubernetes
- versioned API from start
- CI from start

Secondary goals:

- extensible architecture
- clean separation of concerns
- safe schema evolution
- ability to add services later

Non-goals for v1:

- kiosk UI
- push notifications
- recurring reminders
- invitations flow
- mobile apps
- offline support

---

# 3. High-Level Architecture

System must be implemented as separate services.

Required services:

- auth-service
- account-service
- productivity-service
- frontend

Each service owns its own database schema.

Services communicate via HTTP.

All external requests go through a reverse proxy.

Routing must use path prefixes.

Example:

    /api/v1/auth/*
    /api/v1/tenants/*
    /api/v1/tasks/*

Frontend is a standalone container.

Authentication uses JWT.

JWT must be validated by downstream services.

No shared database tables across services.

---

# 4. Multi-Tenant Model

Hierarchy:

User
 → Tenant
   → Household
     → Membership

Rules:

- one tenant per user initially
- multiple households allowed
- roles per household
- all data scoped to tenant
- productivity data scoped to household

Every row must include:

    tenant_id

Household-scoped rows must include:

    household_id

Identity is global.

Context is tenant-scoped.

---

# 5. Authentication Requirements

System must support OIDC login.

Auth service must:

- support multiple providers
- issue JWT
- expose JWKS
- store refresh sessions
- expose current user endpoint

Downstream services must:

- validate JWT
- trust JWKS
- not call auth service for each request

Cookies must be HTTP-only.

Access tokens short-lived.

Refresh tokens server-side.

---

# 6. Account / Context Requirements

System must support:

- tenants
- households
- memberships
- preferences
- active context

Preferences must store:

- theme (light/dark)
- active household

There must be an endpoint:

    /contexts/current

This endpoint must return:

- tenant
- active household
- role
- theme
- memberships

Context must be derived server-side.

Frontend must not compute context.

---

# 7. Productivity Requirements

System must support:

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
- one-time only
- snooze
- dismiss

Summary:

- tasks summary
- reminder summary
- dashboard summary

Summary must be derived.

Summary must not be stored.

All productivity data must be tenant + household scoped.

---

# 8. API Requirements

API must be versioned from start.

Base path:

    /api/v1

API must be resource oriented.

Examples:

    /tasks
    /tasks/{id}
    /tasks/restorations

    /reminders
    /reminders/snoozes

    /summary/tasks

Context:

    /contexts/current

Media type should follow JSON:API conventions where possible.

API must support include where reasonable.

Breaking changes must require new version.

---

# 9. Persistence Requirements

Database:

PostgreSQL

ORM:

Gorm

Each service must own its schema.

Schemas:

    auth
    account
    productivity

IDs must be UUID.

Migrations must run on startup.

Migrations must be forward-only.

No cross-schema foreign keys required.

Indexes required on:

- tenant_id
- household_id
- user_id

Soft delete must use deleted_at.

---

# 10. Runtime Requirements

System must run locally with Docker Compose.

System must run in Kubernetes.

No Helm required.

Plain YAML acceptable.

Images must be built in CI.

Images must be versioned.

Secrets must be external.

Environment variables must configure services.

---

# 11. Frontend Requirements

Frontend must be standalone container.

Frontend must use:

- React
- ShadCN
- light/dark theme

Frontend must:

- call APIs
- store no business state
- bootstrap using /users/me and /contexts/current

No server-side rendering required.

---

# 12. CI/CD Requirements

Repository must include CI from start.

CI must:

- build changed services
- test changed services
- build frontend
- build docker images
- publish snapshot images

Renovate must be enabled.

Bruno collections must exist.

Scripts must exist for local dev.

---

# 13. Logging / Observability

Services must:

- log structured JSON
- include request id
- include user id
- include tenant id

Tracing must support OpenTelemetry.

Metrics optional.

---

# 14. Out of Scope for v1

Do not implement:

- notifications
- recurrence
- invitations
- kiosk UI
- mobile
- email
- audit UI
- background workers beyond scheduler

These may be added later.

---

# 15. Deliverable for Design Phase

Developer must produce:

- detailed service design
- DB migration plan
- API contract confirmation
- routing design
- compose design
- k8s design
- shared package design

before implementation begins.
