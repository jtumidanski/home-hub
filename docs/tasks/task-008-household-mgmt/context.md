# Household Management — Context

Last Updated: 2026-03-26

---

## Key Files to Modify

### account-service

| File | Change Type | Purpose |
|------|------------|---------|
| `services/account-service/cmd/main.go` | Modify | Register invitation migration + routes |
| `services/account-service/internal/invitation/` | **New package** | Full domain: entity, model, builder, administrator, processor, provider, resource, rest |
| `services/account-service/internal/membership/processor.go` | Modify | Self-deletion, last-owner guard, role authorization |
| `services/account-service/internal/membership/resource.go` | Modify | `filter[householdId]` support, authorization in handlers |
| `services/account-service/internal/membership/provider.go` | Modify | Add `getByHousehold()`, `countOwnersByHousehold()` |
| `services/account-service/internal/membership/rest.go` | Modify | Add `IsLastOwner` computed attribute |
| `services/account-service/internal/membership/administrator.go` | Modify | May need `deleteByID` updates for preference cleanup |
| `services/account-service/internal/appcontext/context.go` | Modify | Add pending invitation count to Resolved struct |
| `services/account-service/internal/appcontext/rest.go` | Modify | Add `PendingInvitationCount` to RestModel |

### auth-service

| File | Change Type | Purpose |
|------|------------|---------|
| `services/auth-service/internal/user/resource.go` | Modify | Add batch user list handler (`GET /users?filter[ids]=...`) |
| `services/auth-service/internal/user/provider.go` | Modify | Add `getByIDs()` query |

### frontend

| File | Change Type | Purpose |
|------|------------|---------|
| `frontend/src/services/api/account.ts` | Modify | Add invitation API methods |
| `frontend/src/services/api/auth.ts` | Modify | Add batch user lookup method |
| `frontend/src/types/models/invitation.ts` | **New** | Invitation TypeScript types |
| `frontend/src/types/models/context.ts` | Modify | Add `pendingInvitationCount` |
| `frontend/src/lib/hooks/api/use-invitations.ts` | **New** | Invitation React Query hooks |
| `frontend/src/lib/hooks/api/use-memberships.ts` | **New** | Membership management hooks |
| `frontend/src/lib/hooks/api/use-users.ts` | **New** | Batch user lookup hook |
| `frontend/src/lib/schemas/invitation.schema.ts` | **New** | Zod schema for invite form |
| `frontend/src/pages/HouseholdMembersPage.tsx` | **New** | Household members page |
| `frontend/src/pages/HouseholdsPage.tsx` | Modify | Add members link, user's pending invitations section |
| `frontend/src/pages/OnboardingPage.tsx` | Modify | Add invitation detection + join flow |
| `frontend/src/components/features/households/invite-member-dialog.tsx` | **New** | Invite form dialog |
| `frontend/src/components/features/navigation/nav-group.tsx` | Modify | Badge support |
| `frontend/src/components/providers/auth-provider.tsx` | Modify | Expose `pendingInvitationCount` |
| `frontend/src/App.tsx` | Modify | Add `/app/households/:id/members` route |

### Infrastructure

| File | Change Type | Purpose |
|------|------------|---------|
| `nginx.conf` or ingress config | Modify | Route `/api/v1/invitations` to account-service |

---

## Key Decisions

### 1. Invitation email lookup bypasses tenant filtering
The `/invitations/mine` endpoint and the appcontext pending count query must bypass GORM's automatic tenant filtering because the user may not have a tenant yet (first-login scenario). Use `database.WithoutTenantFilter(ctx)` for these queries.

### 2. No `expired` status in database
Expiration is a virtual state. All queries for "active" invitations include `WHERE status = 'pending' AND expires_at > NOW()`. No background job or status mutation needed.

### 3. Auth-service batch lookup: trust the caller
The auth-service `GET /users?filter[ids]=...` endpoint returns users matching provided IDs without verifying co-membership. Security relies on: (a) JWT authentication required, (b) frontend only provides IDs from the membership list, (c) cap at 50 IDs. This avoids cross-service calls that would violate service boundaries.

### 4. Membership authorization is per-handler, not middleware
Role-based authorization (owner/admin checks for invite, role change, removal) is enforced in the processor layer, not via shared middleware. This follows the existing pattern where tenant isolation is the only middleware-level concern.

### 5. Accept invitation creates membership via membership processor
The invitation processor's `Accept` method calls into the membership processor's `Create` method to create the membership. This avoids duplicating membership creation logic.

### 6. Frontend joins user data client-side
The members page fetches memberships from account-service and user details from auth-service separately, joining them client-side by user ID. This preserves service boundaries.

### 7. Active household auto-switch on accept
When a user accepts an invitation, their `activeHouseholdId` preference is immediately updated to the new household. This ensures the dashboard shows the right household context after accepting.

### 8. isLastOwner is computed per-request
Not stored in the database. When listing memberships by household, the server counts owners and annotates accordingly. This avoids stale data.

---

## Dependencies Between Tasks

```
Phase 1 (Invitation Domain)
  1.1 Entity/Model
    └─► 1.2 Builder
    └─► 1.3 Admin/Provider
         └─► 1.4 Processor ─► 1.5 REST/Resource ─► 1.6 Wire Up

Phase 2 (Membership/AppContext Mods)
  2.1 Membership Auth ◄── depends on Phase 1 for integration testing
  2.2 AppContext Count ◄── depends on 1.3 (invitation provider)

Phase 3 (Auth-Service) — independent, parallelizable with Phase 1+2
  3.1 Batch User Endpoint

Phase 4 (Frontend API) ◄── depends on backend phases
  4.1 Invitation API Service
  4.2 Batch User Service
  4.3 Invitation Types (independent)
  4.4 Invitation Hooks ◄── 4.1, 4.3
  4.5 Membership Hooks ◄── Phase 2
  4.6 User Lookup Hook ◄── 4.2
  4.7 Context Type Update

Phase 5 (Frontend UI) ◄── depends on Phase 4
  5.1 Schemas (independent)
  5.2 Invite Dialog ◄── 4.4, 5.1
  5.3 Members Page ◄── 4.4, 4.5, 4.6, 5.2
  5.4 Confirmation Dialogs (independent)
  5.5 Households Page Mods ◄── 4.4, 5.3
  5.6 Nav Badge ◄── 4.7
  5.7 Onboarding Flow Mod ◄── 4.4
```

---

## Reference Patterns

### Existing domain to follow: `membership`
- Entity: `services/account-service/internal/membership/entity.go`
- Model: `services/account-service/internal/membership/model.go`
- Builder: `services/account-service/internal/membership/builder.go`
- All 7 files follow the standard pattern documented in `docs/architecture.md` §17

### Existing hook to follow: `use-households.ts`
- Query key factory pattern
- `useQuery` / `useMutation` with proper `enabled`, `staleTime`, `gcTime`
- Cache invalidation on `onSettled`
- Located at `frontend/src/lib/hooks/api/use-households.ts`

### Existing dialog to follow: `create-task-dialog.tsx`
- React Hook Form + Zod + Shadcn Dialog
- `useForm` with `zodResolver`
- Toast on success, error handling
- Located at `frontend/src/components/features/tasks/create-task-dialog.tsx`

### Tenant filter bypass pattern
- Used in `appcontext/context.go` via `database.WithoutTenantFilter(ctx)`
- Applied to the GORM `db.WithContext(ctx)` call

---

## Spec Documents

All spec documents are in `docs/tasks/task-008-household-mgmt/`:

| Document | Purpose |
|----------|---------|
| `prd.md` | Full product requirements, acceptance criteria, resolved questions |
| `api-contracts.md` | JSON:API request/response contracts for all endpoints |
| `data-model.md` | Entity definitions, indexes, status transitions, invariants |
| `ux-flow.md` | Wireframes for join flow, members page, dialogs, nav badge |
