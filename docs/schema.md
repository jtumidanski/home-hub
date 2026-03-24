# Home Hub — Database Schema

This document defines the database schema for all services.

Rules:

- each service owns its schema
- no cross-service tables
- all tables use UUID primary keys
- all timestamps are UTC
- soft delete uses deleted_at
- migrations run on service startup
- Gorm is used for persistence

Schemas:

- auth
- account
- productivity

Each service connects to the same database, but only accesses its own schema.

---

## 1. Conventions

### 1.1 Primary Keys

All primary keys are UUID.

Column name:

    id UUID PRIMARY KEY

Generated in application.

---

### 1.2 Timestamps

All tables include:

    created_at TIMESTAMP WITH TIME ZONE NOT NULL
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL

Optional:

    deleted_at TIMESTAMP WITH TIME ZONE NULL

Soft delete uses deleted_at.

---

### 1.3 Tenant Scoping

All tenant data must include:

    tenant_id UUID NOT NULL

Household-scoped data must include:

    household_id UUID NOT NULL

---

### 1.4 User Identity

User identity lives only in auth schema.

Other schemas reference user_id only.

---

## 2. Auth Schema

Schema:

    auth

Tables:

- auth.users
- auth.external_identities
- auth.oidc_providers
- auth.refresh_tokens

---

### 2.1 auth.users

Stores identity-only user records.

Columns:

    id UUID PK
    email TEXT NOT NULL
    display_name TEXT
    given_name TEXT
    family_name TEXT
    avatar_url TEXT

    created_at TIMESTAMPTZ NOT NULL
    updated_at TIMESTAMPTZ NOT NULL

Rules:

- email must be unique
- no tenant data here

---

### 2.2 auth.external_identities

Maps external provider identity to user.

Columns:

    id UUID PK
    user_id UUID NOT NULL
    provider TEXT NOT NULL
    provider_subject TEXT NOT NULL

    created_at TIMESTAMPTZ NOT NULL
    updated_at TIMESTAMPTZ NOT NULL

Constraints:

    UNIQUE(provider, provider_subject)

---

### 2.3 auth.oidc_providers

Configured login providers.

Columns:

    id UUID PK
    name TEXT NOT NULL
    issuer_url TEXT NOT NULL
    client_id TEXT NOT NULL
    enabled BOOLEAN NOT NULL

    created_at TIMESTAMPTZ NOT NULL
    updated_at TIMESTAMPTZ NOT NULL

Secrets stored outside DB.

---

### 2.4 auth.refresh_tokens

Server-side refresh sessions.

Columns:

    id UUID PK
    user_id UUID NOT NULL
    token_hash TEXT NOT NULL
    expires_at TIMESTAMPTZ NOT NULL
    revoked BOOLEAN NOT NULL

    created_at TIMESTAMPTZ NOT NULL
    updated_at TIMESTAMPTZ NOT NULL

Rules:

- token stored hashed
- revoke on logout

---

## 3. Account Schema

Schema:

    account

Tables:

- account.tenants
- account.households
- account.memberships
- account.preferences

---

### 3.1 account.tenants

Top-level container.

Columns:

    id UUID PK
    name TEXT NOT NULL

    created_at TIMESTAMPTZ NOT NULL
    updated_at TIMESTAMPTZ NOT NULL

Rules:

- one tenant per user initially
- may allow more later

---

### 3.2 account.households

Households belong to tenant.

Columns:

    id UUID PK
    tenant_id UUID NOT NULL

    name TEXT NOT NULL
    timezone TEXT NOT NULL
    units TEXT NOT NULL

    created_at TIMESTAMPTZ NOT NULL
    updated_at TIMESTAMPTZ NOT NULL

Units:

- imperial
- metric

Rules:

- household belongs to tenant
- multiple households allowed

---

### 3.3 account.memberships

User role per household.

Columns:

    id UUID PK

    tenant_id UUID NOT NULL
    household_id UUID NOT NULL
    user_id UUID NOT NULL

    role TEXT NOT NULL

    created_at TIMESTAMPTZ NOT NULL
    updated_at TIMESTAMPTZ NOT NULL

Roles:

- owner
- admin
- editor
- viewer

Constraints:

    UNIQUE(household_id, user_id)

Rules:

- roles differ per household
- owner required for household

---

### 3.4 account.preferences

One per user per tenant.

Columns:

    id UUID PK

    tenant_id UUID NOT NULL
    user_id UUID NOT NULL

    theme TEXT NOT NULL

    active_household_id UUID NULL

    created_at TIMESTAMPTZ NOT NULL
    updated_at TIMESTAMPTZ NOT NULL

Theme:

- light
- dark

Constraints:

    UNIQUE(tenant_id, user_id)

Rules:

- preference controls theme
- preference controls active household
- active household must belong to tenant

---

## 4. Productivity Schema

Schema:

    productivity

Tables:

- productivity.tasks
- productivity.task_restorations
- productivity.reminders
- productivity.reminder_snoozes
- productivity.reminder_dismissals

---

### 4.1 productivity.tasks

Columns:

    id UUID PK

    tenant_id UUID NOT NULL
    household_id UUID NOT NULL

    title TEXT NOT NULL
    notes TEXT

    status TEXT NOT NULL

    due_on DATE

    rollover_enabled BOOLEAN NOT NULL

    completed_at TIMESTAMPTZ NULL
    completed_by_user_id UUID NULL

    deleted_at TIMESTAMPTZ NULL

    created_at TIMESTAMPTZ NOT NULL
    updated_at TIMESTAMPTZ NOT NULL

Status:

- pending
- completed

Rules:

- soft delete only
- restore allowed within window
- completed_at managed by service

Indexes:

    tenant_id
    household_id
    due_on
    deleted_at

---

### 4.2 productivity.task_restorations

Columns:

    id UUID PK

    tenant_id UUID NOT NULL
    household_id UUID NOT NULL

    task_id UUID NOT NULL

    created_by_user_id UUID NOT NULL

    created_at TIMESTAMPTZ NOT NULL

Rules:

- used for audit
- restoration clears deleted_at

---

### 4.3 productivity.reminders

Columns:

    id UUID PK

    tenant_id UUID NOT NULL
    household_id UUID NOT NULL

    title TEXT NOT NULL
    notes TEXT

    scheduled_for TIMESTAMPTZ NOT NULL

    last_dismissed_at TIMESTAMPTZ NULL
    last_snoozed_until TIMESTAMPTZ NULL

    created_at TIMESTAMPTZ NOT NULL
    updated_at TIMESTAMPTZ NOT NULL

Rules:

- one-time reminders only in v1
- active state derived

Indexes:

    tenant_id
    household_id
    scheduled_for

---

### 4.4 productivity.reminder_snoozes

Columns:

    id UUID PK

    tenant_id UUID NOT NULL
    household_id UUID NOT NULL

    reminder_id UUID NOT NULL

    duration_minutes INT NOT NULL

    snoozed_until TIMESTAMPTZ NOT NULL

    created_by_user_id UUID NOT NULL

    created_at TIMESTAMPTZ NOT NULL

Allowed durations:

- 10
- 30
- 60

Rules:

- service computes snoozed_until

---

### 4.5 productivity.reminder_dismissals

Columns:

    id UUID PK

    tenant_id UUID NOT NULL
    household_id UUID NOT NULL

    reminder_id UUID NOT NULL

    created_by_user_id UUID NOT NULL

    created_at TIMESTAMPTZ NOT NULL

Rules:

- dismissal tracked separately
- reminder remains for audit

---

## 5. Foreign Key Rules

Foreign keys may be enforced within schema.

Do not enforce across schemas.

Examples:

Allowed:

    account.memberships -> account.households

Not allowed:

    productivity.tasks -> account.households

Cross-service validation done in application.

---

## 6. Migration Rules

Each service owns migrations.

Location:

    services/<service>/migrations/

Rules:

- forward only
- no destructive changes in v1
- no editing old migrations
- migrations run on startup

---

## 7. Index Guidelines

All tables should index:

- tenant_id
- household_id (if present)
- user_id (if present)

High-use filters must be indexed.

---

## 8. Future Extensions

Planned additions:

- recurring reminders
- task templates
- invitations
- tenant sharing
- audit logs
- notification service

Schema should allow additive changes.
