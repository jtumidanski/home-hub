# Task & Reminder Ownership — Product Requirements Document

Version: v1
Status: Draft
Created: 2026-03-26
---

## 1. Overview

Tasks and reminders in Home Hub are currently household-wide with no concept of individual ownership. Every item is visible to all members and there is no way to indicate who is responsible for completing it.

This feature adds an **owner** to tasks and reminders. The owner defaults to the creating user but can be set to any household member or left unset to indicate "household-wide" (anyone can do it). The Tasks and Reminders list pages gain sorting and filtering capabilities, and the three dashboard widgets become navigation entry points into pre-filtered list views.

## 2. Goals

Primary goals:
- Add an owner field to tasks and reminders with a default of the creating user
- Support a "household-wide" / everyone option (no specific owner)
- Display owner in task and reminder list pages
- Enable client-side sorting and filtering by title, status, owner, and date
- Make dashboard widgets clickable, navigating to pre-filtered list pages

Non-goals:
- Notifications when a task/reminder is assigned to someone
- Reassignment history or audit trail
- Per-user filtered dashboard summaries (summaries remain household-wide)
- Server-side filtering or pagination
- Bulk assignment operations

## 3. User Stories

- As a household member, I want to assign a task to a specific person so it's clear who is responsible.
- As a household member, I want to create a task without specifying an owner so anyone in the household can pick it up.
- As a household member, I want to see who owns each task and reminder in the list view so I know what's mine.
- As a household member, I want to filter the task list by owner so I can see only my tasks or a specific person's tasks.
- As a household member, I want to sort tasks and reminders by title, status, owner, or date to find what I need quickly.
- As a household member, I want to click a dashboard widget to navigate to the relevant list page pre-filtered to the data shown in the widget.

## 4. Functional Requirements

### 4.1 Owner Field

- Tasks and reminders gain an `owner_user_id` field (UUID, nullable).
- `NULL` means "household-wide" — no specific owner, anyone can do it.
- A non-null value references a user in the household.
- On create, `owner_user_id` defaults to the authenticated user's ID.
- On create or update, the user may explicitly set `owner_user_id` to any household member's user ID or to `null` (everyone).
- No server-side validation that `owner_user_id` is a valid household member (the frontend controls the dropdown options).

### 4.2 API Changes

#### Tasks

- `POST /api/v1/tasks` — accepts optional `ownerUserId` in attributes. Defaults to the authenticated user if omitted.
- `PATCH /api/v1/tasks/{id}` — accepts `ownerUserId` (set to a user ID or `null` to clear).
- `GET /api/v1/tasks` and `GET /api/v1/tasks/{id}` — response includes `ownerUserId` in attributes.

#### Reminders

- `POST /api/v1/reminders` — accepts optional `ownerUserId` in attributes. Defaults to the authenticated user if omitted.
- `PATCH /api/v1/reminders/{id}` — accepts `ownerUserId` (set to a user ID or `null` to clear).
- `GET /api/v1/reminders` and `GET /api/v1/reminders/{id}` — response includes `ownerUserId` in attributes.

#### Summary Endpoints

- No changes. Summary counts remain household-wide.

### 4.3 List Pages — Sorting & Filtering

Both the Tasks page and Reminders page gain:

- **Title search** — text input that filters items whose title contains the search string (case-insensitive).
- **Status filter** — dropdown to filter by status. Tasks: pending, completed, all. Reminders: active, dismissed, snoozed, upcoming, all.
- **Owner filter** — dropdown listing household members (by display name) plus "Everyone" and "All". "Everyone" shows only household-wide items (null owner). "All" shows all items regardless of owner. Selecting a member shows only their items.
- **Date sort** — sortable by due date (tasks) or scheduled time (reminders).
- **Title sort** — sortable alphabetically by title.
- **Status sort** — sortable by status.
- **Owner sort** — sortable by owner display name, with household-wide items sorting last.

All filtering and sorting is client-side. Filter/sort state is preserved in URL query parameters so dashboard links can pre-set them.

### 4.4 Owner Display

- Owner appears as a column in the list table (desktop) and as a field in the card view (mobile).
- Household-wide items display "Everyone" as the owner.
- Owner names are resolved by fetching household members via the account-service's `GET /households/{id}/members` endpoint (see section 5), which returns user IDs with display names.
- If an `owner_user_id` references a user no longer in the household (removed member), the UI displays "Former member". The data is not reassigned — items remain functional and the owner can be manually changed.

### 4.5 Filtered Empty States

- When filters produce zero results, show a generic "No items found" message with a "Clear filters" action to reset all filters.

### 4.6 Overdue as a Virtual Status

- "Overdue" is not a stored status value. The frontend computes it client-side as `status = pending AND dueOn < today`. The `?status=overdue` query parameter is interpreted by the frontend filter logic, not the API.

### 4.7 Dashboard Widget Navigation

The three dashboard widgets become clickable links:

| Widget | Navigates to | Pre-applied filter |
|---|---|---|
| Pending Tasks | `/tasks` | `?status=pending` |
| Active Reminders | `/reminders` | `?status=active` |
| Overdue | `/tasks` | `?status=overdue` |

## 5. API Surface

### Modified Endpoints

#### `POST /api/v1/tasks`

Request body additions:
```json
{
  "data": {
    "type": "tasks",
    "attributes": {
      "title": "Buy groceries",
      "ownerUserId": "uuid-or-null"
    }
  }
}
```

- `ownerUserId` is optional. If omitted, defaults to the authenticated user's ID.
- If explicitly set to `null`, the task is household-wide.

#### `PATCH /api/v1/tasks/{id}`

```json
{
  "data": {
    "type": "tasks",
    "id": "task-uuid",
    "attributes": {
      "ownerUserId": "uuid-or-null"
    }
  }
}
```

#### `GET /api/v1/tasks` — Response

Each task resource now includes `ownerUserId`:
```json
{
  "data": [{
    "type": "tasks",
    "id": "task-uuid",
    "attributes": {
      "title": "Buy groceries",
      "status": "pending",
      "dueOn": "2026-03-28",
      "ownerUserId": "user-uuid",
      "notes": "",
      "rolloverEnabled": false,
      "completedAt": null,
      "completedByUserId": null,
      "createdAt": "2026-03-26T10:00:00Z",
      "updatedAt": "2026-03-26T10:00:00Z"
    }
  }]
}
```

`ownerUserId` is `null` for household-wide tasks.

#### Reminders — Same pattern

`POST`, `PATCH`, `GET` for reminders follow the identical pattern with `ownerUserId` added to attributes.

### New Endpoint

#### `GET /api/v1/households/{id}/members`

Returns household members with display names for use in owner dropdowns and name resolution. This is a new lightweight endpoint on the account-service.

Response:
```json
{
  "data": [{
    "type": "members",
    "id": "membership-uuid",
    "attributes": {
      "displayName": "Alice",
      "role": "owner"
    },
    "relationships": {
      "user": { "data": { "type": "users", "id": "user-uuid" } },
      "household": { "data": { "type": "households", "id": "household-uuid" } }
    }
  }]
}
```

The frontend uses this to map `ownerUserId` values to display names and to populate the owner dropdown in create/edit forms.

### Error Cases

- No new error cases on productivity-service endpoints. Invalid UUIDs are handled by existing validation. No server-side membership validation.
- `GET /households/{id}/members` returns `404` if the household does not exist or the user is not a member.

## 6. Data Model

### Tasks Table — Migration

Add column:
```sql
ALTER TABLE tasks ADD COLUMN owner_user_id UUID;
```

- Nullable. `NULL` = household-wide.
- No foreign key constraint (user data lives in a separate service/database).
- No index needed (no server-side filtering by owner).

### Reminders Table — Migration

Add column:
```sql
ALTER TABLE reminders ADD COLUMN owner_user_id UUID;
```

- Same semantics as tasks.

### Existing Data

- Existing tasks and reminders get `owner_user_id = NULL` (household-wide) by default from the migration. This is correct — they were created before ownership existed and belong to the whole household.

## 7. Service Impact

### productivity-service

- **Models**: Add `OwnerUserID` field to task and reminder domain models.
- **Entities**: Add `OwnerUserID` column to task and reminder GORM entities. Run migration.
- **Builders**: Accept `OwnerUserID` option. Default to authenticated user ID on create when not explicitly provided.
- **Processors**: Pass `OwnerUserID` through on create and update.
- **REST layer**: Map `ownerUserId` in JSON:API request/response transformations.
- **Summary**: No changes.

### frontend

- **Types**: Add `ownerUserId` to task and reminder TypeScript interfaces.
- **API client**: Include `ownerUserId` in create/update payloads.
- **Member hooks**: Add `use-household-members` hook to fetch members via the new `/households/{id}/members` endpoint for owner dropdowns and name resolution. Cache with React Query (members change infrequently).
- **Create/Edit dialogs**: Add owner dropdown (members + "Everyone").
- **List pages**: Add filter bar (title search, status dropdown, owner dropdown), enable column sorting via tanstack/react-table.
- **Dashboard**: Make widget cards clickable with navigation to filtered list pages.

### account-service

- **New endpoint**: `GET /api/v1/households/{id}/members` — returns membership records with user display names. This provides the frontend with the data needed for owner dropdowns and name resolution without requiring separate user profile lookups.
- **No other changes**. Existing membership CRUD and household endpoints remain as-is.

## 8. Non-Functional Requirements

- **Performance**: Client-side filtering is acceptable for current data volumes. One additional API call for household members (cacheable, members change infrequently) beyond the existing list endpoints.
- **Multi-tenancy**: `owner_user_id` is scoped within the existing `tenant_id` + `household_id` context. No cross-household assignment possible.
- **Security**: No server-side validation that `owner_user_id` belongs to the household. The frontend constrains the dropdown to valid members. Malformed requests with invalid UUIDs are harmless (the UUID simply won't match any user).
- **Observability**: No new metrics or logging beyond existing patterns.

## 9. Open Questions

None — all questions resolved during scoping.

## 10. Acceptance Criteria

- [ ] Tasks have an `owner_user_id` column (nullable UUID) with migration applied.
- [ ] Reminders have an `owner_user_id` column (nullable UUID) with migration applied.
- [ ] Creating a task without specifying an owner defaults to the authenticated user.
- [ ] Creating a task with `ownerUserId: null` creates a household-wide task.
- [ ] Creating a task with a specific `ownerUserId` assigns it to that user.
- [ ] Same three behaviors work for reminders.
- [ ] Updating a task/reminder can change the owner or set to null.
- [ ] `GET /tasks` and `GET /reminders` responses include `ownerUserId`.
- [ ] Tasks page displays owner column showing member name or "Everyone".
- [ ] Reminders page displays owner column showing member name or "Everyone".
- [ ] Both pages support filtering by title (text search).
- [ ] Both pages support filtering by status.
- [ ] Both pages support filtering by owner (specific member, "Everyone", or "All").
- [ ] Both pages support sorting by title, status, owner, and date.
- [ ] Filter state is reflected in URL query parameters.
- [ ] Pending Tasks dashboard widget links to `/tasks?status=pending`.
- [ ] Active Reminders dashboard widget links to `/reminders?status=active`.
- [ ] Overdue dashboard widget links to `/tasks?status=overdue`.
- [ ] `GET /households/{id}/members` returns members with display names.
- [ ] Owner column shows "Former member" for removed household members.
- [ ] Filtered empty state shows "No items found" with a "Clear filters" action.
- [ ] "Overdue" filter works as a client-side computed status (pending + past due).
- [ ] Existing tasks/reminders show as "Everyone" (null owner) after migration.
- [ ] Docker builds pass for productivity-service and account-service.
