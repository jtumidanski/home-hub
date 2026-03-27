# Task & Reminder Ownership — Implementation Plan

Last Updated: 2026-03-26

---

## Executive Summary

Add an `owner_user_id` field to tasks and reminders in productivity-service, a new `GET /households/{id}/members` endpoint on account-service, and frontend changes for owner display, filtering, sorting, and dashboard widget navigation. This is a cross-cutting feature touching three services (productivity-service, account-service, frontend) with no breaking changes to existing behavior.

---

## Current State Analysis

### productivity-service
- **Task model** (`internal/task/model.go`): immutable struct with accessors. No owner field.
- **Task entity** (`internal/task/entity.go`): GORM entity with `CompletedByUserId` (nullable UUID) as precedent for user-reference fields. Migration via `db.AutoMigrate`.
- **Task builder** (`internal/task/builder.go`): fluent builder with setters for all fields.
- **Task REST** (`internal/task/rest.go`): `RestModel`, `CreateRequest`, `UpdateRequest` structs with JSON:API marshaling. No `ownerUserId` field.
- **Task resource** (`internal/task/resource.go`): routes for CRUD. `createHandler` reads tenant context for IDs. `Create` processor method takes individual args (not struct).
- **Task administrator** (`internal/task/administrator.go`): `create()` and `update()` functions that build and persist entities.
- **Reminder model/entity/rest**: Same pattern as tasks, no owner field.
- **Summary**: processor counts (pending, overdue, completed today) — no changes needed per PRD.

### account-service
- **Membership model** (`internal/membership/model.go`): has `userID`, `householdID`, `role`. No `displayName`.
- **Membership entity** (`internal/membership/entity.go`): table `memberships` with user/household foreign keys.
- **Membership REST** (`internal/membership/rest.go`): returns `role`, `isLastOwner`, and relationships to `users` and `households` — but no display name.
- **No `/households/{id}/members` endpoint** exists. Current membership listing is via `/memberships?filter[householdId]=uuid`.
- **User display names** live in `auth.users` table (auth-service schema in same database).

### frontend
- **Task types** (`types/models/task.ts`): `TaskAttributes` interface, no `ownerUserId`.
- **Reminder types** (`types/models/reminder.ts`): `ReminderAttributes` interface, no `ownerUserId`.
- **TasksPage** (`pages/TasksPage.tsx`): uses `@tanstack/react-table` columns, no filtering or sorting UI. Supports mobile card view.
- **RemindersPage** (`pages/RemindersPage.tsx`): same pattern, no filtering or sorting.
- **DashboardPage** (`pages/DashboardPage.tsx`): three Card widgets (Pending Tasks, Active Reminders, Overdue) — not clickable, no navigation.
- **No household-members hook** exists. No owner dropdown component.

---

## Proposed Future State

1. Tasks and reminders have a nullable `owner_user_id` UUID column. Existing records remain `NULL` (household-wide).
2. API create/update endpoints accept `ownerUserId`. Create defaults to authenticated user when omitted; explicit `null` means household-wide.
3. API list/get responses include `ownerUserId` in attributes.
4. `GET /api/v1/households/{id}/members` on account-service returns membership records enriched with `displayName` from auth.users (cross-schema join within shared database).
5. Frontend resolves owner names via the members endpoint, cached with React Query.
6. Task/Reminder list pages gain: title search, status filter, owner filter, column sorting (title, status, owner, date).
7. Filter state is reflected in URL query params for deep linking.
8. Dashboard widgets link to pre-filtered list pages.

---

## Implementation Phases

### Phase 1: Backend — productivity-service (owner field)

**1.1 Task domain — add OwnerUserID**
- `model.go`: add `ownerUserID *uuid.UUID` field + accessor `OwnerUserID() *uuid.UUID`
- `entity.go`: add `OwnerUserID *uuid.UUID \`gorm:"type:uuid"\`` to Entity. Update `ToEntity()` and `Make()`.
- `builder.go`: add `ownerUserID *uuid.UUID` field + `SetOwnerUserID(*uuid.UUID)` setter. Thread through `Build()`.
- `rest.go`: add `OwnerUserId *string \`json:"ownerUserId"\`` to `RestModel`, `CreateRequest`, `UpdateRequest`. Use `*string` for nullable JSON representation. Update `Transform()`.
- `administrator.go`: update `create()` to accept `ownerUserID *uuid.UUID`. Update `update()` similarly.
- `processor.go`: update `Create()` and `Update()` signatures to accept `ownerUserID`. In `Create()`, thread ownerUserID through. Resource layer sets default to authenticated user when not provided in request.
- `resource.go`: in `createHandler`, extract `ownerUserId` from input. If not present in request (Go zero-value check), default to `t.UserId()`. Pass to processor. In `updateHandler`, pass through.
- `provider.go`: no changes (no server-side filtering by owner).
- **Tests**: update `builder_test.go`, `processor_test.go`.

**1.2 Reminder domain — add OwnerUserID**
- Same changes as 1.1 applied to `internal/reminder/` files.

**1.3 Verify**
- Run `go test ./...` for productivity-service.
- Docker build passes.

### Phase 2: Backend — account-service (members endpoint)

**2.1 New household members endpoint**
- Add `GET /households/{id}/members` route in `household/resource.go` (co-located with household routes).
- Handler: lookup memberships by household ID via membership processor. Cross-schema query to `auth.users` to get display names for each user ID. Return JSON:API response with type `"members"`, attributes `{ displayName, role }`, relationships to `user` and `household`.
- Implementation approach: define a lightweight `MemberView` struct in `household/` package for the join result. Use raw GORM query joining `account.memberships` with `auth.users` on `user_id = id`. No new domain model needed — this is a read-only projection.
- 404 if household not found or user not a member (tenant scoping handles this implicitly).

**2.2 Verify**
- Run `go test ./...` for account-service.
- Docker build passes.

### Phase 3: Frontend — types, API client, member hooks

**3.1 Update TypeScript types**
- `task.ts`: add `ownerUserId?: string | null` to `TaskAttributes`, `TaskCreateAttributes`, `TaskUpdateAttributes`.
- `reminder.ts`: add `ownerUserId?: string | null` to `ReminderAttributes`, `ReminderCreateAttributes`, `ReminderUpdateAttributes`.

**3.2 New member types and hook**
- Create `types/models/member.ts`: `Member` interface with `id`, `type: "members"`, `attributes: { displayName, role }`, `relationships`.
- Create `lib/hooks/api/use-household-members.ts`: React Query hook to fetch `GET /api/v1/households/{id}/members`. Long staleTime (members change infrequently). Export `useHouseholdMembers()` and a helper `useMemberMap()` that returns `Map<userId, displayName>`.

**3.3 Update API client**
- `productivityService.createTask()`: include `ownerUserId` in payload.
- `productivityService.updateTask()`: include `ownerUserId` in payload.
- Same for reminder endpoints.
- Add account-service call for household members.

### Phase 4: Frontend — create/edit dialogs (owner dropdown)

**4.1 Owner select component**
- Create reusable `OwnerSelect` component: dropdown with household members + "Everyone" option. Uses `useHouseholdMembers()` hook.

**4.2 Task create dialog**
- `create-task-dialog.tsx`: add OwnerSelect field. Update schema (`task.schema.ts`) with optional `ownerUserId` field.
- Default selection: current user (from auth context).
- "Everyone" maps to `null`.

**4.3 Reminder create dialog**
- Same changes to `create-reminder-dialog.tsx` and `reminder.schema.ts`.

**4.4 Edit functionality**
- If edit dialogs exist or are created: include OwnerSelect pre-populated with current owner.

### Phase 5: Frontend — list pages (filtering, sorting, owner column)

**5.1 Filter bar component**
- Create `ListFilterBar` component with: title search input, status dropdown, owner dropdown.
- Filter state managed via URL search params (`useSearchParams`).

**5.2 Tasks page**
- Add owner column to table (display name from member map, or "Everyone" for null, or "Former member" for unresolvable).
- Add owner column to mobile card view.
- Integrate filter bar. Apply client-side filtering (title search, status filter, owner filter).
- Enable column sorting (title, status, owner, date).
- Handle `?status=overdue` virtual status (pending + past due date).
- Empty state with "Clear filters" action when filters produce no results.

**5.3 Reminders page**
- Same changes as 5.2 adapted for reminder statuses (active, dismissed, snoozed, upcoming, all).

### Phase 6: Frontend — dashboard widget navigation

**6.1 Make widgets clickable**
- `DashboardPage.tsx`: wrap each Card in a `Link` (react-router-dom).
- Pending Tasks → `/tasks?status=pending`
- Active Reminders → `/reminders?status=active`
- Overdue → `/tasks?status=overdue`

### Phase 7: Testing & verification

**7.1 Backend tests**
- productivity-service: builder tests, processor tests for owner field.
- account-service: members endpoint test.

**7.2 Frontend tests**
- Update existing TasksPage and RemindersPage tests.
- Test filter bar rendering and interaction.
- Test dashboard widget links.
- Test owner display (member name, "Everyone", "Former member").

**7.3 Docker builds**
- Verify productivity-service builds.
- Verify account-service builds.
- Verify frontend builds.

---

## Risk Assessment and Mitigation

| Risk | Impact | Mitigation |
|---|---|---|
| Cross-schema join (account→auth.users) couples services | Medium | Read-only projection, limited to display_name lookup. No writes. Documented as intentional design decision. |
| `ownerUserId` references a user removed from household | Low | Frontend shows "Former member". Data remains functional. Owner can be manually changed. |
| Client-side filtering performance with large datasets | Low | PRD explicitly accepts this. Current data volumes are small. Server-side filtering is a non-goal. |
| URL query param format conflicts with future features | Low | Use simple `?status=X&owner=Y&q=Z` format. Standard and extensible. |

---

## Success Metrics

- All 22 acceptance criteria from PRD section 10 pass.
- Docker builds succeed for productivity-service, account-service, frontend.
- All existing tests continue to pass.
- Existing tasks/reminders display as "Everyone" (null owner) after migration.

---

## Required Resources and Dependencies

- **productivity-service**: GORM AutoMigrate handles column addition.
- **account-service**: Cross-schema read access to `auth.users` (same database, already available).
- **frontend**: No new npm dependencies. Uses existing `@tanstack/react-table`, `react-hook-form`, `react-router-dom`.
- **No infrastructure changes** required.

---

## Timeline Estimates

| Phase | Effort |
|---|---|
| Phase 1: productivity-service task+reminder owner | M |
| Phase 2: account-service members endpoint | M |
| Phase 3: Frontend types, hooks, API client | S |
| Phase 4: Create/edit dialogs (owner dropdown) | M |
| Phase 5: List pages (filtering, sorting, owner) | L |
| Phase 6: Dashboard widget navigation | S |
| Phase 7: Testing & verification | M |
