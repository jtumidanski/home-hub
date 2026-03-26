# Household Management — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-26

---

## 1. Overview

Home Hub already supports tenants, households, memberships with roles (owner, admin, editor, viewer), and user preferences. However, there is no mechanism to invite users to a household, no join flow for invited users, and no UI for managing household membership.

This feature introduces an **invitation system** and a **household members management UI**. Privileged users (owner, admin) can invite new users by email address, manage role assignments, and remove members. Non-privileged users (editor, viewer) can view membership and invitation information but cannot modify it. When a new user logs in for the first time with a pending invitation, they skip the existing onboarding flow and instead join the inviter's tenant and household directly.

Users can also leave a household voluntarily without requiring action from a privileged role.

## 2. Goals

Primary goals:
- Enable privileged users to invite others to their household by email address
- Provide a join flow for first-time users who have pending invitations
- Display household members with their roles in the web UI
- Allow privileged users to manage role assignments, membership, and invitations
- Allow non-privileged users to view membership information read-only
- Allow any member to leave a household voluntarily

Non-goals:
- Email or push notification delivery of invitations
- Cross-tenant invitations
- Ownership transfer
- Audit logging of membership changes
- Invitation link/code sharing (invitations are email-based lookup only)

## 3. User Stories

- As an **owner or admin**, I want to invite a user to my household by entering their email address and selecting a role, so that they can join when they sign up.
- As an **owner or admin**, I want to revoke a pending invitation, so that I can cancel an invite before it is accepted.
- As an **owner or admin**, I want to change the role of an existing household member, so that I can adjust their permissions.
- As an **owner or admin**, I want to remove a member from the household, so that they no longer have access.
- As a **new user** with a pending invitation, I want to see my invitations on first login and pick one to accept, so that I can join a household without creating my own.
- As a **new user** with pending invitations, I want to decline invitations I do not wish to accept.
- As a **new user** with no pending invitations, I want to go through the existing onboarding flow (create tenant + household), so that my experience is unchanged.
- As any **household member**, I want to view the list of members and their roles, so that I know who is in my household.
- As any **household member**, I want to view pending invitations, so that I know who has been invited.
- As any **household member**, I want to leave a household voluntarily.
- As an **owner or admin**, I want to see and manage invitations in the same UI where I manage members.

## 4. Functional Requirements

### 4.1 Invitation Model

- An invitation records: the inviter, the invitee email, the target household, the assigned role, status, and expiration.
- Valid statuses: `pending`, `accepted`, `declined`, `revoked`. There is no stored `expired` status — expiration is a virtual state derived at query time by filtering `WHERE status = 'pending' AND expires_at > NOW()`.
- Invitations expire after **7 days** from creation.
- Only one pending invitation may exist per email per household at a time.
- The inviter may specify a role; if omitted, it defaults to `viewer`.
- Valid roles for invitation: `admin`, `editor`, `viewer`. An invitation cannot assign `owner`.
- Invitations are scoped to a tenant (via the household's tenant).

### 4.2 Invitation Lifecycle

- **Create**: A privileged user (owner or admin) creates an invitation by providing an email and optionally a role.
- **Revoke**: A privileged user may revoke a pending invitation, setting its status to `revoked`.
- **Accept**: The invitee accepts a pending invitation. A membership is created with the specified role, and the invitation status becomes `accepted`.
- **Decline**: The invitee declines a pending invitation, setting its status to `declined`.
- **Expire**: Invitations past their expiration timestamp are treated as expired. Expiration is enforced at the database query level (`WHERE status = 'pending' AND expires_at > NOW()`), not in application code. No `expired` status is ever written.

### 4.3 First-Login Join Flow

- When a new user authenticates via OIDC for the first time, the frontend checks for pending invitations by the user's email.
- If **one or more** pending invitations exist:
  - The user is shown an invitation selection screen instead of the standard onboarding.
  - Each invitation displays: household name, inviter info (if available), assigned role, and expiration.
  - The user selects one invitation to accept. Remaining invitations stay pending (they can be managed later).
  - Accepting an invitation: the user joins the inviter's tenant and household with the specified role. A preference record is created. The user proceeds to the dashboard.
  - The user may also decline all invitations, in which case they fall through to the standard onboarding flow.
- If **no** pending invitations exist:
  - The existing onboarding flow (create tenant + create household) proceeds unchanged.

### 4.4 Household Members UI

- A sub-page of the existing Households page at `/app/households/:id/members` displays:
  - **Members list**: Each member's display name, email, role, and join date.
  - **Pending invitations list**: Each invitation's email, assigned role, inviter, creation date, and expiration.
- **Privileged view** (owner or admin):
  - Can change a member's role (dropdown/select).
  - Can remove a member (with confirmation).
  - Can revoke a pending invitation.
  - Can create a new invitation (email + role form).
  - Owner cannot be removed by an admin. Only another owner can demote/remove an owner.
  - A user cannot change their own role.
- **Non-privileged view** (editor or viewer):
  - Same information displayed, but all management actions are hidden/disabled.
- **Self-actions**:
  - Any member can leave the household (with confirmation dialog).
  - The last owner cannot leave (must transfer ownership first — but ownership transfer is out of scope, so this is simply blocked with an error message).

### 4.5 Role Hierarchy & Permissions

| Action | Owner | Admin | Editor | Viewer |
|--------|-------|-------|--------|--------|
| View members & invitations | Yes | Yes | Yes | Yes |
| Invite user | Yes | Yes | No | No |
| Revoke invitation | Yes | Yes | No | No |
| Change member role | Yes | Yes (not owners) | No | No |
| Remove member | Yes | Yes (not owners) | No | No |
| Leave household | Yes* | Yes | Yes | Yes |

*Owner can leave only if another owner exists.

Admin cannot modify owners (cannot change an owner's role or remove an owner).

### 4.6 Post-Onboarding Invitation Management

- After the initial join flow, users may receive additional invitations to other households.
- The `pendingInvitationCount` from app context drives a badge on the existing "Households" nav item.
- The households page surfaces a section showing the user's own pending invitations with accept/decline actions.

## 5. API Surface

### 5.1 Invitation Endpoints (account-service)

All endpoints require JWT authentication and are tenant-scoped.

**List invitations for a household:**
```
GET /api/v1/invitations?filter[householdId]={householdId}
```
- Returns pending invitations for the specified household.
- Available to all household members.
- Expired invitations are excluded by default.

**List invitations for the current user:**
```
GET /api/v1/invitations/mine?filter[status]=pending
```
- Returns pending, non-expired invitations addressed to the authenticated user's email (derived from JWT, not a query parameter).
- Used by the join flow and post-onboarding notification.
- This endpoint operates without tenant filtering (the user may not have a tenant yet).

**Create invitation:**
```
POST /api/v1/invitations
```
```json
{
  "data": {
    "type": "invitations",
    "attributes": {
      "email": "invitee@example.com",
      "role": "editor"
    },
    "relationships": {
      "household": {
        "data": { "type": "households", "id": "{householdId}" }
      }
    }
  }
}
```
- Requires owner or admin role in the specified household.
- Role defaults to `viewer` if omitted.
- Returns 409 if a pending invitation already exists for this email + household.
- Returns 422 if the email already has a membership in this household.

**Revoke invitation:**
```
DELETE /api/v1/invitations/{id}
```
- Sets status to `revoked`.
- Requires owner or admin role in the invitation's household.
- Returns 404 if invitation is not found or not pending.

**Accept invitation:**
```
POST /api/v1/invitations/{id}/accept
```
- Sets status to `accepted`.
- Creates a membership for the user in the invitation's household with the specified role.
- If the user has no tenant, assigns the invitation's tenant.
- If the user already has a tenant that differs from the invitation's tenant, returns 422 (cross-tenant blocked).
- Creates a preference record if none exists.
- Sets `activeHouseholdId` on the user's preference to the invitation's household (auto-switch).
- Returns 404 if invitation is not found or not pending.
- Returns 410 if invitation is expired.

**Decline invitation:**
```
POST /api/v1/invitations/{id}/decline
```
- Sets status to `declined`.
- Returns 404 if invitation is not found or not pending.

### 5.2 Membership Endpoint Changes

**Leave household (self-remove):**
```
DELETE /api/v1/memberships/{id}
```
- The existing delete endpoint is extended: a user may delete their own membership (leave).
- If the user is the last owner, return 422 with an error message indicating they cannot leave.
- If the deleted membership's household was the user's active household, set `activeHouseholdId` to null in preferences. The existing appcontext resolution logic will fall back to the user's first remaining membership's household on the next context fetch.

**Update membership role:**
```
PATCH /api/v1/memberships/{id}
```
- The existing endpoint. Authorization rules tightened:
  - Requires owner or admin role.
  - Admin cannot change an owner's role.
  - A user cannot change their own role.

### 5.3 App Context Changes

**GET /api/v1/contexts/current**
- Add a new attribute: `pendingInvitationCount` (integer) — number of pending invitations for the current user's email across all households.
- This allows the frontend to show a badge/indicator without a separate API call.

## 6. Data Model

### 6.1 New Entity: Invitation

Table: `account.invitations`

| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PRIMARY KEY |
| tenant_id | UUID | NOT NULL, INDEX |
| household_id | UUID | NOT NULL, INDEX |
| email | TEXT | NOT NULL |
| role | TEXT | NOT NULL, DEFAULT 'viewer' |
| status | TEXT | NOT NULL, DEFAULT 'pending' |
| invited_by | UUID | NOT NULL |
| expires_at | TIMESTAMP | NOT NULL |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

Indexes:
- `idx_invitations_household_email_status` on (household_id, email, status) — enforces uniqueness for pending invitations and supports lookup.
- `idx_invitations_email_status` on (email, status) — supports lookup by invitee email for join flow.

Unique partial index: `(household_id, email) WHERE status = 'pending'` — only one pending invitation per email per household.

### 6.2 Domain Files (account-service)

Following the existing service code pattern:

| File | Purpose |
|------|---------|
| `model.go` | Immutable invitation domain model |
| `entity.go` | GORM entity with Migration() |
| `builder.go` | Fluent builder for invitation creation |
| `processor.go` | Business logic (create, revoke, accept, decline, expiration checks) |
| `provider.go` | Database access layer |
| `resource.go` | Route registration and HTTP handlers |
| `rest.go` | JSON:API resource mappings |

### 6.3 Membership Changes

No schema changes to `account.memberships`. Authorization logic changes only.

## 7. Service Impact

### 7.1 account-service

- **New domain**: `invitation` — full domain package with all standard files.
- **Modified domain**: `membership` — authorization rules for role changes and self-removal.
- **Modified domain**: `appcontext` — include `pendingInvitationCount` in context response.
- **New routes**: `/api/v1/invitations`, `/api/v1/invitations/mine`, `/api/v1/invitations/{id}`, `/api/v1/invitations/{id}/accept`, `/api/v1/invitations/{id}/decline`.
- **Authorization**: New middleware or processor logic to enforce owner/admin checks for invitation and membership management.

### 7.2 frontend

- **Modified**: Onboarding flow — detect pending invitations, show join screen with invitation picker.
- **Modified**: Auth provider — surface `pendingInvitationCount` from app context.
- **New page**: Household members sub-page at `/app/households/:id/members` — members list, invitations list, role management, invite form.
- **Modified**: Navigation — badge on "Households" nav item driven by `pendingInvitationCount`.
- **New components**: Invite dialog, role selector, member removal confirmation, leave household confirmation.
- **New API service methods**: Invitation CRUD operations.
- **New hooks**: `useInvitations`, `useCreateInvitation`, `useRevokeInvitation`, `useAcceptInvitation`, `useDeclineInvitation`.

### 7.3 auth-service

- **New endpoint**: `GET /api/v1/users?filter[ids]=uuid1,uuid2` — batch user lookup by IDs. Returns display name, email, and avatar for the requested user IDs. Required by the frontend to display member details on the household members page. The account-service does not call this directly (service boundaries); the frontend joins the data client-side.

## 8. Non-Functional Requirements

### 8.1 Security
- The `/invitations/mine` endpoint operates without tenant context (user may not have a tenant yet). It derives the email from the JWT — the email is never a query parameter, preventing enumeration of other users' invitations.
- The batch user lookup endpoint (`GET /api/v1/users?filter[ids]=...`) only returns users who share at least one household membership with the requester, preventing arbitrary user profile enumeration.
- Role assignment via invitation is capped at `admin` — `owner` cannot be assigned via invitation.
- Authorization checks must be enforced server-side, not just in UI.

### 8.2 Performance
- Invitation queries should be efficient with proper indexing (email + status, household + email + status).
- `pendingInvitationCount` in app context should use a COUNT query, not load full invitation records.

### 8.3 Multi-Tenancy
- Invitations are tenant-scoped via their household's tenant.
- The join flow (accept invitation) must correctly assign the user to the invitation's tenant.
- The invitation lookup by email (for join flow) intentionally bypasses tenant filtering since the user has no tenant yet.

### 8.4 Observability
- Log invitation lifecycle events (create, accept, decline, revoke) with `request_id`, `user_id`, `tenant_id`, `household_id`.

## 9. Resolved Questions

1. **Expiration cleanup** — Filter at query time via `WHERE status = 'pending' AND expires_at > NOW()`. No background cleanup job. No `expired` status written.
2. **Re-invitation** — Yes. After an invitation is declined or revoked, the same email can be invited again. The partial unique index only constrains `pending` invitations.
3. **Cross-tenant invitation acceptance** — Blocked. If a user already belongs to tenant A and accepts an invitation to a household in tenant B, the accept returns 422 with an error explaining the user already belongs to a different tenant.
4. **Batch user lookup** — Auth-service exposes `GET /api/v1/users?filter[ids]=...` for the frontend to resolve member details. Account-service does not call auth-service directly (service boundaries preserved). Endpoint restricted to return only users sharing a household with the requester.
5. **Auto-switch on accept** — Accepting an invitation automatically sets the accepted household as the user's active household.
6. **Sole owner visibility** — The membership resource includes a computed `isLastOwner` boolean attribute when listing by household. Server computes this via `COUNT(*) WHERE household_id = ? AND role = 'owner'`. The UI uses this to show a warning indicator.
7. **Invitation lookup pattern** — Dedicated sub-resource `GET /api/v1/invitations/mine` instead of filter parameter. Email derived from JWT, not a query parameter.
8. **Active household on leave** — Set `activeHouseholdId` to null. Existing appcontext resolution falls back to first remaining membership's household.
9. **Navigation placement** — Household members is a sub-page at `/app/households/:id/members`, not a new top-level nav item. Pending invitation badge goes on the existing "Households" nav item.

## 10. Acceptance Criteria

- [ ] Privileged users (owner, admin) can create invitations by entering an email and selecting a role.
- [ ] Invitations default to `viewer` role when no role is specified.
- [ ] Only one pending invitation per email per household is allowed.
- [ ] Privileged users can revoke pending invitations.
- [ ] A new user with pending invitations sees an invitation selection screen instead of onboarding.
- [ ] Accepting an invitation creates a membership and assigns the user to the inviter's tenant.
- [ ] Declining all invitations falls through to standard onboarding.
- [ ] Invitations expire after 7 days and are excluded from active queries.
- [ ] All household members can view the members list and pending invitations.
- [ ] Privileged users can change member roles (admin cannot modify owners).
- [ ] Privileged users can remove members (admin cannot remove owners).
- [ ] A user cannot change their own role.
- [ ] Any member can leave a household voluntarily.
- [ ] The last owner of a household cannot leave.
- [ ] The UI displays a warning indicator when a member is the sole owner.
- [ ] Non-privileged users see membership information in read-only mode.
- [ ] Cross-tenant invitation acceptance is blocked with a clear error message.
- [ ] Accepting an invitation auto-switches the user's active household to the accepted household.
- [ ] Re-inviting a previously declined or revoked email is allowed.
- [ ] Auth-service batch user lookup endpoint returns user details for a list of IDs (restricted to users sharing a household with the requester).
- [ ] Membership resource includes computed `isLastOwner` boolean when listing by household.
- [ ] Active household set to null on leave; appcontext falls back to first remaining membership.
- [ ] App context includes `pendingInvitationCount` for notification purposes.
- [ ] All invitation and membership management endpoints enforce server-side authorization.
