# Household Management — Implementation Plan

Last Updated: 2026-03-26

---

## Executive Summary

This plan implements a household invitation system and member management UI for Home Hub. The work spans three services (account-service, auth-service, frontend) and introduces one new domain (`invitation`) in the account-service, modifies the existing `membership` and `appcontext` domains, adds a batch user lookup endpoint to auth-service, and builds new frontend pages/flows for invitation management and household members.

The plan is divided into 5 phases executed sequentially, with tasks within each phase parallelizable where noted.

---

## Current State Analysis

### What Exists
- **account-service**: Tenant, household, membership, preference, and appcontext domains fully operational. Memberships have roles (owner, admin, editor, viewer). Appcontext resolves active household and role.
- **auth-service**: User domain with `/users/me` endpoint. No batch lookup. No filtering by IDs.
- **frontend**: Onboarding flow (create tenant → create household). Households page with card layout. Navigation with collapsible groups. Auth/tenant providers. React Query hooks pattern established.

### What's Missing
- No invitation model or API endpoints
- No authorization enforcement beyond tenant isolation (no role-based action checks)
- No household members UI
- No batch user lookup
- No pending invitation count in app context
- No first-login join flow for invited users
- No self-removal (leave household) capability

---

## Proposed Future State

After implementation:
1. Privileged users (owner/admin) can invite users by email, manage roles, and remove members
2. New users with invitations see a join flow instead of onboarding
3. All members can view household membership and leave voluntarily
4. App context includes pending invitation count for nav badge
5. Auth-service supports batch user lookup restricted to co-members

---

## Phase 1: Invitation Domain (account-service)

Build the complete `invitation` domain package following existing patterns.

### 1.1 — Entity & Model (S)
- Create `invitation/entity.go` with GORM entity matching data model spec
- Create `invitation/model.go` with immutable domain model and accessors
- Include `Migration()` function with proper indexes (email+status, household+email partial unique on pending)
- **Acceptance**: Entity auto-migrates. Model has getters for all fields.
- **Dependencies**: None

### 1.2 — Builder (S)
- Create `invitation/builder.go` with fluent builder
- Validate: email required, role in {admin, editor, viewer}, householdId required, invitedBy required
- Default role to `viewer` if omitted
- Set `expiresAt` to `createdAt + 7 days`
- Set `status` to `pending` on creation
- **Acceptance**: Builder rejects owner role, missing email, missing household. Defaults work.
- **Dependencies**: 1.1

### 1.3 — Administrator & Provider (S)
- Create `invitation/administrator.go` with: `create()`, `updateStatus()`
- Create `invitation/provider.go` with queries:
  - `getByID()` — single invitation by ID
  - `getByHouseholdPending()` — pending, non-expired invitations for a household
  - `getByEmailPending()` — pending, non-expired invitations for an email (bypasses tenant filter)
  - `getByHouseholdAndEmailPending()` — uniqueness check
  - `countByEmailPending()` — for appcontext pending count
- All "pending" queries include `WHERE status = 'pending' AND expires_at > NOW()`
- **Acceptance**: All queries filter expired correctly. Email lookup bypasses tenant.
- **Dependencies**: 1.1

### 1.4 — Processor (M)
- Create `invitation/processor.go` with business logic:
  - `Create(tenantID, householdID, email, role, inviterID)` — checks uniqueness, checks no existing membership for email, validates inviter has owner/admin role
  - `Revoke(id, revokerID)` — validates pending status, validates revoker has owner/admin role
  - `Accept(id, userID, userEmail, userTenantID)` — validates pending + not expired, validates email match, creates membership, handles tenant assignment, creates preference, sets active household
  - `Decline(id, userID, userEmail)` — validates pending, validates email match
  - `ByHouseholdProvider(householdID)` — list pending for household
  - `ByEmailProvider(email)` — list pending for email (join flow)
  - `CountByEmailProvider(email)` — count for appcontext
- **Acceptance**: All status transitions correct. Authorization enforced. Cross-tenant blocked. Auto-switch on accept.
- **Dependencies**: 1.2, 1.3, membership domain (for creating memberships and checking roles), preference domain

### 1.5 — REST & Resource (M)
- Create `invitation/rest.go` with RestModel, Transform(), CreateRequest
- Create `invitation/resource.go` with route registration:
  - `GET /invitations` with `filter[householdId]` — list by household
  - `GET /invitations/mine` with `filter[status]` — list by current user email
  - `POST /invitations` — create
  - `DELETE /invitations/{id}` — revoke
  - `POST /invitations/{id}/accept` — accept
  - `POST /invitations/{id}/decline` — decline
- Register routes in `cmd/main.go`
- The `/invitations/mine` endpoint includes household resources (JSON:API `included`)
- **Acceptance**: All endpoints respond per API contract. Proper HTTP status codes.
- **Dependencies**: 1.4

### 1.6 — Wire Up & Migration (S)
- Register `invitation.Migration` in `cmd/main.go`
- Register `invitation.InitializeRoutes` in `cmd/main.go`
- Add `/api/v1/invitations` route to nginx/ingress config
- **Acceptance**: Service starts, migration runs, routes respond.
- **Dependencies**: 1.5

---

## Phase 2: Membership & AppContext Modifications (account-service)

### 2.1 — Membership Authorization (M)
- Modify `membership/processor.go` and `membership/resource.go`:
  - `Delete`: Allow self-deletion (user deleting own membership). Add last-owner guard. On self-delete, null out `activeHouseholdId` in preferences if it matches.
  - `UpdateRole`: Enforce owner/admin required. Admin cannot modify owner's role. User cannot modify own role.
  - `List by household`: Add `filter[householdId]` query parameter support. Compute `isLastOwner` boolean.
- Modify `membership/rest.go`: Add `IsLastOwner` attribute to RestModel
- Add `membership/provider.go`: Add `getByHousehold()` query, `countOwnersByHousehold()` query
- **Acceptance**: Self-removal works. Last-owner blocked. Role change authorization enforced. isLastOwner computed.
- **Dependencies**: Phase 1 (for testing integration)

### 2.2 — AppContext Pending Count (S)
- Modify `appcontext/context.go`: Query invitation count by email (bypassing tenant filter)
- Modify `appcontext/rest.go`: Add `PendingInvitationCount` attribute
- **Acceptance**: Context response includes `pendingInvitationCount` integer.
- **Dependencies**: 1.3 (invitation provider for count query)

---

## Phase 3: Auth-Service Batch User Lookup

### 3.1 — Batch User Endpoint (M)
- Modify `user/resource.go`: Add list handler for `GET /users?filter[ids]=uuid1,uuid2`
- Modify `user/provider.go`: Add `getByIDs(ids []uuid.UUID)` query
- Add authorization check: The endpoint needs the requester's household memberships to filter results. Since auth-service doesn't have membership data, implement one of:
  - **Option A**: Accept a list of IDs and return all matching users (simpler, relies on frontend only requesting co-member IDs). The PRD says "restricted to users sharing a household" — this can be enforced by the account-service providing the IDs, since the frontend only gets member user IDs from the membership list.
  - **Option B**: Cross-service call to account-service for membership data (violates service boundaries).
  - **Recommended: Option A** — The auth-service returns users matching the provided IDs. The security boundary is that only authenticated users can call this, and the frontend only provides IDs it obtained from the membership list. Cap at 50 IDs per request.
- Modify `user/rest.go`: Ensure `TransformSlice` works for batch response
- Add nginx/ingress route if needed (already exists: `/api/v1/users` → auth-service)
- **Acceptance**: Batch lookup returns user details for valid IDs. Max 50 IDs enforced. Unknown IDs silently omitted.
- **Dependencies**: None (can be done in parallel with Phase 1-2)

---

## Phase 4: Frontend — API Layer & Hooks

### 4.1 — Invitation API Service (S)
- Add invitation methods to `services/api/account.ts`:
  - `listInvitationsByHousehold(tenant, householdId)`
  - `listMyInvitations()`
  - `createInvitation(tenant, attrs)`
  - `revokeInvitation(tenant, id)`
  - `acceptInvitation(id)`
  - `declineInvitation(id)`
- **Acceptance**: All methods make correct HTTP calls with proper JSON:API payloads.
- **Dependencies**: Phase 1 backend complete

### 4.2 — Batch User Lookup Service (S)
- Add user lookup method to `services/api/auth.ts`:
  - `getUsersByIds(ids: string[])`
- **Acceptance**: Returns user data array for given IDs.
- **Dependencies**: Phase 3 backend complete

### 4.3 — Invitation Types (S)
- Add `types/models/invitation.ts` with TypeScript types matching JSON:API response
- **Acceptance**: Types match API contract.
- **Dependencies**: None

### 4.4 — Invitation Hooks (M)
- Create `lib/hooks/api/use-invitations.ts`:
  - Query key factory: `invitationKeys`
  - `useHouseholdInvitations(householdId)` — list invitations for a household
  - `useMyInvitations()` — list current user's pending invitations
  - `useCreateInvitation()` — mutation with cache invalidation
  - `useRevokeInvitation()` — mutation with cache invalidation
  - `useAcceptInvitation()` — mutation, invalidates context + households + invitations
  - `useDeclineInvitation()` — mutation with cache invalidation
- **Acceptance**: All hooks follow existing patterns. Cache invalidation correct.
- **Dependencies**: 4.1, 4.3

### 4.5 — Membership Hooks Updates (S)
- Update `lib/hooks/api/use-households.ts` or create `use-memberships.ts`:
  - `useHouseholdMembers(householdId)` — list memberships for household
  - `useUpdateMemberRole()` — mutation
  - `useRemoveMember()` — mutation
  - `useLeaveHousehold()` — mutation, invalidates context
- **Acceptance**: Hooks handle all membership management operations.
- **Dependencies**: Phase 2 backend complete

### 4.6 — User Lookup Hook (S)
- Create `lib/hooks/api/use-users.ts`:
  - `useUsersByIds(ids: string[])` — batch user lookup, enabled when IDs available
- **Acceptance**: Returns enriched user data. Handles empty/null ID arrays.
- **Dependencies**: 4.2

### 4.7 — App Context Update (S)
- Update `types/models/context.ts`: Add `pendingInvitationCount` to AppContext attributes
- **Acceptance**: Type includes new field.
- **Dependencies**: Phase 2 backend complete

---

## Phase 5: Frontend — UI Components & Pages

### 5.1 — Invitation Schemas (S)
- Create `lib/schemas/invitation.schema.ts`:
  - `createInvitationSchema` — email (required, email format), role (enum: viewer, editor, admin)
- **Acceptance**: Zod schema validates correctly.
- **Dependencies**: None

### 5.2 — Invite Member Dialog (M)
- Create `components/features/households/invite-member-dialog.tsx`:
  - Email input field
  - Role select (Viewer default, Editor, Admin)
  - Uses `useCreateInvitation()` hook
  - Error handling for 409 (already invited) and 422 (already member)
  - Toast on success
- **Acceptance**: Dialog creates invitations. Validation works. Error messages display correctly.
- **Dependencies**: 4.4, 5.1

### 5.3 — Household Members Page (L)
- Create `pages/HouseholdMembersPage.tsx` at route `/app/households/:id/members`:
  - Members list with user details (name, email, avatar, role, join date)
  - Pending invitations list (email, role, inviter, created, expires)
  - Privileged view: role dropdown, remove button, revoke button, invite button
  - Non-privileged view: read-only display
  - Sole owner warning indicator
  - Leave household button (all users)
  - Uses `useHouseholdMembers()`, `useHouseholdInvitations()`, `useUsersByIds()`
  - Role-based rendering driven by `resolvedRole` from app context
- Add route to `App.tsx`
- **Acceptance**: Both views render correctly. All actions work per PRD.
- **Dependencies**: 4.4, 4.5, 4.6, 5.2

### 5.4 — Confirmation Dialogs (S)
- Create confirmation dialog components (or use a shared pattern):
  - Remove member confirmation
  - Leave household confirmation
  - Last-owner block dialog
  - Revoke invitation confirmation
- **Acceptance**: All confirmations prevent accidental actions. Last-owner shows appropriate message.
- **Dependencies**: None

### 5.5 — Households Page Link (S)
- Modify `pages/HouseholdsPage.tsx`:
  - Add "Members" link/button on each household card linking to `/app/households/:id/members`
  - Add pending invitations section showing current user's own invitations with accept/decline
- **Acceptance**: Navigation to members page works. User's own invitations displayed.
- **Dependencies**: 4.4, 5.3

### 5.6 — Navigation Badge (S)
- Modify `components/features/navigation/nav-group.tsx` or `nav-config.ts`:
  - Add badge support to nav items
  - Show `pendingInvitationCount` badge on "Households" nav item
  - Badge driven by `appContext.pendingInvitationCount` from auth provider
- **Acceptance**: Badge appears when count > 0. Updates on accept/decline.
- **Dependencies**: 4.7

### 5.7 — Modified Onboarding Flow (M)
- Modify `pages/OnboardingPage.tsx`:
  - Before showing standard onboarding, check for pending invitations via `useMyInvitations()`
  - If invitations exist, show invitation selection screen (per UX spec)
  - Accept → creates membership, navigates to dashboard
  - Decline all → falls through to standard onboarding
  - "Create my own household" → standard onboarding
- **Acceptance**: New users with invitations see join flow. Accepting works end-to-end. Declining falls through.
- **Dependencies**: 4.4

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Tenant bypass on `/invitations/mine` introduces security hole | Medium | High | Email derived from JWT only. Explicit `WithoutTenantFilter` usage with careful scoping. Code review. |
| Race condition on partial unique index (concurrent invitations) | Low | Low | Database constraint enforces uniqueness. Application handles 409 gracefully. |
| Cross-tenant accept edge case | Medium | High | Explicit tenant comparison in processor. Return 422 with clear message. |
| Stale app context after accept/leave | Medium | Medium | Invalidate context query key on all mutations that change membership state. |
| Auth-service user lookup enumeration | Low | Medium | Cap at 50 IDs. Require authentication. Frontend only sends IDs from membership lists. |

---

## Success Metrics

- All 24 acceptance criteria from PRD pass
- All existing tests continue to pass
- Docker builds for all three affected services succeed
- No regressions in onboarding flow for users without invitations
- Invitation lifecycle (create → accept/decline/revoke/expire) works end-to-end
- Role-based UI rendering matches privileged vs non-privileged views

---

## Required Resources & Dependencies

- **Services modified**: account-service, auth-service, frontend
- **New domain**: `invitation` (account-service)
- **Shared libraries**: No changes needed (existing patterns sufficient)
- **Infrastructure**: nginx/ingress route for `/api/v1/invitations` (account-service)
- **External dependencies**: None (no email delivery, no new third-party libraries)

---

## Timeline Estimates

| Phase | Effort | Can Parallelize With |
|-------|--------|---------------------|
| Phase 1: Invitation Domain | L | Phase 3 |
| Phase 2: Membership & AppContext Mods | M | Phase 3 |
| Phase 3: Auth-Service Batch Lookup | M | Phase 1, Phase 2 |
| Phase 4: Frontend API Layer | M | — (needs backend) |
| Phase 5: Frontend UI | XL | — (needs Phase 4) |
