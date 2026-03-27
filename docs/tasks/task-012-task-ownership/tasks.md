# Task & Reminder Ownership — Task Checklist

Last Updated: 2026-03-26

---

## Phase 1: productivity-service — Owner Field

### 1.1 Task domain
- [ ] Add `ownerUserID *uuid.UUID` to `task/model.go` with `OwnerUserID()` accessor
- [ ] Add `OwnerUserID *uuid.UUID` to `task/entity.go`, update `ToEntity()` and `Make()`
- [ ] Add `SetOwnerUserID(*uuid.UUID)` to `task/builder.go`, thread through `Build()`
- [ ] Add `OwnerUserId *string` to `RestModel`, `CreateRequest`, `UpdateRequest` in `task/rest.go`
- [ ] Update `Transform()` in `task/rest.go` to map ownerUserID UUID to string pointer
- [ ] Update `create()` in `task/administrator.go` to accept and persist `ownerUserID`
- [ ] Update `update()` in `task/administrator.go` to accept and persist `ownerUserID`
- [ ] Update `Create()` in `task/processor.go` to accept `ownerUserID` parameter
- [ ] Update `Update()` in `task/processor.go` to accept `ownerUserID` parameter
- [ ] Update `createHandler` in `task/resource.go`: extract ownerUserId from input, default to authenticated user
- [ ] Update `updateHandler` in `task/resource.go`: pass ownerUserId through
- [ ] Update `task/builder_test.go` for new field
- [ ] Update `task/processor_test.go` for new field

### 1.2 Reminder domain
- [ ] Add `ownerUserID *uuid.UUID` to `reminder/model.go` with `OwnerUserID()` accessor
- [ ] Add `OwnerUserID *uuid.UUID` to `reminder/entity.go`, update `ToEntity()` and `Make()`
- [ ] Add `SetOwnerUserID(*uuid.UUID)` to `reminder/builder.go`, thread through `Build()`
- [ ] Add `OwnerUserId *string` to `RestModel`, `CreateRequest`, `UpdateRequest` in `reminder/rest.go`
- [ ] Update `Transform()` in `reminder/rest.go` to map ownerUserID
- [ ] Update `create()` in `reminder/administrator.go` to accept and persist `ownerUserID`
- [ ] Update `update()` in `reminder/administrator.go` to accept and persist `ownerUserID`
- [ ] Update `Create()` in `reminder/processor.go` to accept `ownerUserID` parameter
- [ ] Update `Update()` in `reminder/processor.go` to accept `ownerUserID` parameter
- [ ] Update `createHandler` in `reminder/resource.go`: extract ownerUserId, default to authenticated user
- [ ] Update `updateHandler` in `reminder/resource.go`: pass ownerUserId through
- [ ] Update `reminder/builder_test.go` for new field
- [ ] Update `reminder/processor_test.go` for new field

### 1.3 Verify productivity-service
- [ ] `go test ./...` passes
- [ ] Docker build succeeds

---

## Phase 2: account-service — Members Endpoint

### 2.1 Household members endpoint
- [ ] Define `MemberView` struct in `household/` for join result (membership + display_name)
- [ ] Add `membersHandler` in `household/resource.go` with cross-schema query to `auth.users`
- [ ] Register `GET /households/{id}/members` route in `household/InitializeRoutes`
- [ ] Define `MemberRestModel` with JSON:API interface methods (type: "members")
- [ ] Return 404 if household not found / not accessible

### 2.2 Verify account-service
- [ ] `go test ./...` passes
- [ ] Docker build succeeds

---

## Phase 3: Frontend — Types, Hooks, API Client

### 3.1 TypeScript types
- [ ] Add `ownerUserId?: string | null` to `TaskAttributes` in `task.ts`
- [ ] Add `ownerUserId?: string | null` to `TaskCreateAttributes` and `TaskUpdateAttributes`
- [ ] Add `ownerUserId?: string | null` to `ReminderAttributes` in `reminder.ts`
- [ ] Add `ownerUserId?: string | null` to `ReminderCreateAttributes` and `ReminderUpdateAttributes`
- [ ] Create `types/models/member.ts` with `Member` and `MemberAttributes` interfaces

### 3.2 Household members hook
- [ ] Create `lib/hooks/api/use-household-members.ts`
- [ ] Implement `useHouseholdMembers()` query hook with long staleTime
- [ ] Implement `useMemberMap()` helper returning `Map<userId, displayName>`

### 3.3 API client updates
- [ ] Add household members fetch to account-service API client
- [ ] Update task create/update payloads to include `ownerUserId`
- [ ] Update reminder create/update payloads to include `ownerUserId`

---

## Phase 4: Frontend — Create/Edit Dialogs

### 4.1 Owner select component
- [ ] Create `components/common/owner-select.tsx` with member dropdown + "Everyone" option
- [ ] Use `useHouseholdMembers()` hook for data

### 4.2 Task create dialog
- [ ] Add `ownerUserId` to `task.schema.ts` (optional field)
- [ ] Add OwnerSelect to `create-task-dialog.tsx`
- [ ] Default to current user in form defaults

### 4.3 Reminder create dialog
- [ ] Add `ownerUserId` to `reminder.schema.ts` (optional field)
- [ ] Add OwnerSelect to `create-reminder-dialog.tsx`
- [ ] Default to current user in form defaults

---

## Phase 5: Frontend — List Pages (Filtering, Sorting, Owner Column)

### 5.1 Filter bar
- [ ] Create `components/common/list-filter-bar.tsx` with title search, status dropdown, owner dropdown
- [ ] Wire filter state to URL search params via `useSearchParams`
- [ ] Implement "Clear filters" action

### 5.2 Tasks page
- [ ] Add owner column to table (resolve via member map, "Everyone" for null, "Former member" for unknown)
- [ ] Add owner display to `task-card.tsx` mobile view
- [ ] Integrate filter bar component
- [ ] Implement client-side title search filter
- [ ] Implement client-side status filter (including virtual "overdue" status)
- [ ] Implement client-side owner filter ("All", "Everyone", specific member)
- [ ] Enable column sorting (title, status, owner, date)
- [ ] Show "No items found" with "Clear filters" when filtered results are empty

### 5.3 Reminders page
- [ ] Add owner column to table (resolve via member map)
- [ ] Add owner display to `reminder-card.tsx` mobile view
- [ ] Integrate filter bar component
- [ ] Implement client-side title search filter
- [ ] Implement client-side status filter (active, dismissed, snoozed, upcoming, all)
- [ ] Implement client-side owner filter
- [ ] Enable column sorting (title, status, owner, scheduled time)
- [ ] Show "No items found" with "Clear filters" when filtered results are empty

---

## Phase 6: Frontend — Dashboard Widget Navigation

- [ ] Wrap Pending Tasks card with `Link` to `/tasks?status=pending`
- [ ] Wrap Active Reminders card with `Link` to `/reminders?status=active`
- [ ] Wrap Overdue card with `Link` to `/tasks?status=overdue`
- [ ] Style cards with hover/focus states indicating clickability

---

## Phase 7: Testing & Verification

### 7.1 Backend tests
- [ ] Task builder test covers ownerUserID field
- [ ] Task processor test covers create with/without owner, update owner
- [ ] Reminder builder test covers ownerUserID field
- [ ] Reminder processor test covers create with/without owner, update owner
- [ ] Account-service members endpoint test

### 7.2 Frontend tests
- [ ] Update `TasksPage.test.tsx` for owner column and filter bar
- [ ] Update `RemindersPage.test.tsx` for owner column and filter bar
- [ ] Update `DashboardPage.test.tsx` for clickable widget links
- [ ] Test owner display: member name, "Everyone", "Former member"
- [ ] Test create dialog with owner select

### 7.3 Build verification
- [ ] productivity-service Docker build passes
- [ ] account-service Docker build passes
- [ ] frontend Docker build passes

---

## Acceptance Criteria (from PRD)

- [ ] Tasks have `owner_user_id` column (nullable UUID) with migration applied
- [ ] Reminders have `owner_user_id` column (nullable UUID) with migration applied
- [ ] Creating a task without specifying an owner defaults to the authenticated user
- [ ] Creating a task with `ownerUserId: null` creates a household-wide task
- [ ] Creating a task with a specific `ownerUserId` assigns it to that user
- [ ] Same three behaviors work for reminders
- [ ] Updating a task/reminder can change the owner or set to null
- [ ] `GET /tasks` and `GET /reminders` responses include `ownerUserId`
- [ ] Tasks page displays owner column showing member name or "Everyone"
- [ ] Reminders page displays owner column showing member name or "Everyone"
- [ ] Both pages support filtering by title (text search)
- [ ] Both pages support filtering by status
- [ ] Both pages support filtering by owner (specific member, "Everyone", or "All")
- [ ] Both pages support sorting by title, status, owner, and date
- [ ] Filter state is reflected in URL query parameters
- [ ] Pending Tasks dashboard widget links to `/tasks?status=pending`
- [ ] Active Reminders dashboard widget links to `/reminders?status=active`
- [ ] Overdue dashboard widget links to `/tasks?status=overdue`
- [ ] `GET /households/{id}/members` returns members with display names
- [ ] Owner column shows "Former member" for removed household members
- [ ] Filtered empty state shows "No items found" with a "Clear filters" action
- [ ] "Overdue" filter works as a client-side computed status (pending + past due)
- [ ] Existing tasks/reminders show as "Everyone" (null owner) after migration
- [ ] Docker builds pass for productivity-service and account-service
