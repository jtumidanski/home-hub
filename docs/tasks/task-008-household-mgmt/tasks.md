# Household Management — Task Checklist

Last Updated: 2026-03-26

---

## Phase 1: Invitation Domain (account-service)

- [ ] **1.1** Create `invitation/entity.go` — GORM entity with Migration(), indexes, partial unique constraint (S)
- [ ] **1.2** Create `invitation/model.go` — Immutable domain model with accessors (S)
- [ ] **1.3** Create `invitation/builder.go` — Fluent builder with validation (email, role, householdId, invitedBy), defaults (viewer, 7-day expiry, pending status) (S)
- [ ] **1.4** Create `invitation/administrator.go` — DB operations: create(), updateStatus() (S)
- [ ] **1.5** Create `invitation/provider.go` — Queries: getByID, getByHouseholdPending, getByEmailPending, getByHouseholdAndEmailPending, countByEmailPending (S)
- [ ] **1.6** Create `invitation/processor.go` — Business logic: Create, Revoke, Accept, Decline with authorization + validation (M)
- [ ] **1.7** Create `invitation/rest.go` — RestModel, Transform(), CreateRequest (S)
- [ ] **1.8** Create `invitation/resource.go` — Routes: GET list, GET mine, POST create, DELETE revoke, POST accept, POST decline (M)
- [ ] **1.9** Register invitation in `cmd/main.go` — Migration + InitializeRoutes (S)
- [ ] **1.10** Add tests for invitation processor — Create, revoke, accept, decline, expiration, authorization (M)
- [ ] **1.11** Update nginx/ingress routing — `/api/v1/invitations` → account-service (S)

## Phase 2: Membership & AppContext Modifications (account-service)

- [ ] **2.1** Add `getByHousehold()` query to `membership/provider.go` (S)
- [ ] **2.2** Add `countOwnersByHousehold()` query to `membership/provider.go` (S)
- [ ] **2.3** Add `filter[householdId]` support to membership list handler in `membership/resource.go` (S)
- [ ] **2.4** Add `IsLastOwner` computed attribute to `membership/rest.go` (S)
- [ ] **2.5** Implement self-deletion (leave) in membership delete handler — own membership check, preference cleanup (M)
- [ ] **2.6** Implement last-owner guard in membership delete — block if sole owner (S)
- [ ] **2.7** Tighten role update authorization — owner/admin required, admin can't modify owner, can't modify self (M)
- [ ] **2.8** Add admin can't remove owner guard to membership delete (S)
- [ ] **2.9** Add `PendingInvitationCount` to `appcontext/context.go` Resolved struct — count query bypassing tenant filter (S)
- [ ] **2.10** Add `PendingInvitationCount` to `appcontext/rest.go` RestModel (S)
- [ ] **2.11** Add tests for membership authorization changes (M)

## Phase 3: Auth-Service Batch User Lookup

- [ ] **3.1** Add `getByIDs()` query to `user/provider.go` (S)
- [ ] **3.2** Add list handler for `GET /users?filter[ids]=...` to `user/resource.go` — parse IDs, cap at 50, return matching users (M)
- [ ] **3.3** Add tests for batch user lookup (S)

## Phase 4: Frontend — API Layer & Hooks

- [ ] **4.1** Create `types/models/invitation.ts` — Invitation type definitions (S)
- [ ] **4.2** Update `types/models/context.ts` — Add `pendingInvitationCount` attribute (S)
- [ ] **4.3** Add invitation methods to `services/api/account.ts` — list, listMine, create, revoke, accept, decline (S)
- [ ] **4.4** Add batch user lookup to `services/api/auth.ts` — getUsersByIds (S)
- [ ] **4.5** Create `lib/hooks/api/use-invitations.ts` — Query + mutation hooks with cache invalidation (M)
- [ ] **4.6** Create `lib/hooks/api/use-memberships.ts` — Household members, update role, remove, leave hooks (M)
- [ ] **4.7** Create `lib/hooks/api/use-users.ts` — Batch user lookup hook (S)
- [ ] **4.8** Create `lib/schemas/invitation.schema.ts` — Zod schema for invite form (S)

## Phase 5: Frontend — UI Components & Pages

- [ ] **5.1** Create invite member dialog — Email + role form, error handling for 409/422 (M)
- [ ] **5.2** Create household members page — Members list, invitations list, role management, computed from user role (L)
- [ ] **5.3** Create confirmation dialogs — Remove member, leave household, last-owner block, revoke invitation (S)
- [ ] **5.4** Add route `/app/households/:id/members` to `App.tsx` (S)
- [ ] **5.5** Modify households page — Add "Members" link per card, user's pending invitations section with accept/decline (M)
- [ ] **5.6** Add nav badge support — Badge on "Households" driven by `pendingInvitationCount` (S)
- [ ] **5.7** Modify onboarding flow — Detect invitations before standard flow, show invitation selection screen, accept/decline/fallthrough (M)

## Verification

- [ ] **V.1** Docker build: account-service (S)
- [ ] **V.2** Docker build: auth-service (S)
- [ ] **V.3** Docker build: frontend (S)
- [ ] **V.4** End-to-end: Create invitation → new user accepts → joins household (M)
- [ ] **V.5** End-to-end: Role management — change role, remove member, leave household (M)
- [ ] **V.6** End-to-end: Onboarding flow unchanged for users without invitations (S)
- [ ] **V.7** Review all 24 PRD acceptance criteria (S)
