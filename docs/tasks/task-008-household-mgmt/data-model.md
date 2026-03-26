# Household Management — Data Model

## New Entity: Invitation

### Domain Model (model.go)

Immutable domain model with accessor methods.

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary key |
| tenantId | UUID | Tenant scope (derived from household) |
| householdId | UUID | Target household |
| email | string | Invitee's email address |
| role | string | Assigned role on acceptance (`admin`, `editor`, `viewer`) |
| status | string | Lifecycle status (`pending`, `accepted`, `declined`, `revoked`, `expired`) |
| invitedBy | UUID | User ID of the inviter |
| expiresAt | time.Time | Expiration timestamp (created_at + 7 days) |
| createdAt | time.Time | Creation timestamp |
| updatedAt | time.Time | Last update timestamp |

### GORM Entity (entity.go)

```
Table: account.invitations

id              UUID            PRIMARY KEY
tenant_id       UUID            NOT NULL, INDEX
household_id    UUID            NOT NULL, INDEX
email           TEXT            NOT NULL
role            TEXT            NOT NULL DEFAULT 'viewer'
status          TEXT            NOT NULL DEFAULT 'pending'
invited_by      UUID            NOT NULL
expires_at      TIMESTAMP       NOT NULL
created_at      TIMESTAMP       NOT NULL
updated_at      TIMESTAMP       NOT NULL
```

### Indexes

| Name | Columns | Type | Purpose |
|------|---------|------|---------|
| idx_invitations_tenant_id | tenant_id | Standard | Tenant filtering |
| idx_invitations_household_id | household_id | Standard | Household lookup |
| idx_invitations_email_status | email, status | Standard | Join flow lookup (find pending by email) |
| idx_invitations_unique_pending | household_id, email | Unique partial (WHERE status = 'pending') | One pending invitation per email per household |

### Status Transitions

```
                 ┌──────────┐
       create ──►│ pending  │
                 └────┬─────┘
                      │
          ┌───────────┼───────────┐
          ▼           ▼           ▼
     ┌─────────┐ ┌─────────┐ ┌─────────┐
     │accepted │ │declined │ │ revoked │
     └─────────┘ └─────────┘ └─────────┘

     Expired: pending + expires_at < now()
     (not a stored status — derived at query time)
```

- `pending` → `accepted`: Invitee accepts. Creates membership. Auto-switches active household.
- `pending` → `declined`: Invitee declines.
- `pending` → `revoked`: Privileged user cancels.
- `pending` with `expires_at < now()`: Treated as expired. All queries for pending invitations include `WHERE status = 'pending' AND expires_at > NOW()` at the database level. No `expired` status is ever written.

Valid stored statuses: `pending`, `accepted`, `declined`, `revoked` (4 values, not 5).

### Invariants

1. Only one `pending` invitation per (household_id, email) combination.
2. Role must be one of: `admin`, `editor`, `viewer`. Cannot be `owner`.
3. `expires_at` is set to `created_at + 7 days` on creation.
4. `invited_by` must be a user with `owner` or `admin` role in the target household.
5. Cannot create an invitation if the email already has a membership in the household.

---

## Existing Entity Changes

### Membership

No schema changes. Logic changes only.

**Computed attribute**: `isLastOwner` (boolean) — added to the membership JSON:API resource when listing by household. Computed server-side via `COUNT(*) WHERE household_id = ? AND role = 'owner'`. True when the count is 1 and the membership's role is `owner`. Not stored in the database.

Changes:

- **Self-deletion**: A user can delete their own membership (leave household).
- **Last-owner guard**: If the membership being deleted has role `owner` and no other `owner` exists in the household, the deletion is rejected (uses `isLastOwner` check).
- **Role update authorization**: Admin cannot modify a membership with role `owner`. A user cannot modify their own membership's role.
- **Preference cleanup**: When a membership is deleted and its household matches the user's `activeHouseholdId`, set `activeHouseholdId` to null. The existing appcontext resolution logic falls back to the first remaining membership's household.

### App Context

No schema changes. Response changes only:

- Add `pendingInvitationCount` attribute: COUNT of invitations where `email = current_user.email AND status = 'pending' AND expires_at > now()`.
- This query bypasses tenant filtering (invitations may exist across tenants).

---

## Entity Relationships

```
Tenant ──1:N──► Household ──1:N──► Membership ◄── User
                    │
                    └──1:N──► Invitation
                                  │
                                  └── invited_by ──► User
```

- A household has many memberships and many invitations.
- An invitation belongs to one household and references one inviter (user).
- Accepting an invitation creates a membership.
