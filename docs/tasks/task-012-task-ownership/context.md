# Task & Reminder Ownership — Context

Last Updated: 2026-03-26

---

## Key Files

### productivity-service — Task domain
| File | Purpose |
|---|---|
| `services/productivity-service/internal/task/model.go` | Immutable domain model. Add `ownerUserID *uuid.UUID` field + accessor. |
| `services/productivity-service/internal/task/entity.go` | GORM entity, migration, Make/ToEntity. Add `OwnerUserID` column. |
| `services/productivity-service/internal/task/builder.go` | Fluent builder. Add `SetOwnerUserID` setter. |
| `services/productivity-service/internal/task/rest.go` | REST models (`RestModel`, `CreateRequest`, `UpdateRequest`). Add `ownerUserId` JSON field. |
| `services/productivity-service/internal/task/resource.go` | HTTP handlers. Update create/update handlers to pass ownerUserId. Default to auth user on create. |
| `services/productivity-service/internal/task/processor.go` | Business logic. Update `Create()` and `Update()` signatures. |
| `services/productivity-service/internal/task/administrator.go` | DB operations (`create`, `update`). Accept and persist ownerUserID. |
| `services/productivity-service/internal/task/provider.go` | DB queries. No changes needed. |
| `services/productivity-service/internal/task/builder_test.go` | Builder tests. Update for new field. |
| `services/productivity-service/internal/task/processor_test.go` | Processor tests. Update for new field. |

### productivity-service — Reminder domain
| File | Purpose |
|---|---|
| `services/productivity-service/internal/reminder/model.go` | Same pattern as task. Add `ownerUserID`. |
| `services/productivity-service/internal/reminder/entity.go` | Add `OwnerUserID` column. |
| `services/productivity-service/internal/reminder/builder.go` | Add `SetOwnerUserID` setter. |
| `services/productivity-service/internal/reminder/rest.go` | Add `ownerUserId` to REST models. |
| `services/productivity-service/internal/reminder/resource.go` | Update create/update handlers. |
| `services/productivity-service/internal/reminder/processor.go` | Update signatures. |
| `services/productivity-service/internal/reminder/administrator.go` | Accept and persist ownerUserID. |

### account-service — Members endpoint
| File | Purpose |
|---|---|
| `services/account-service/internal/household/resource.go` | Add `GET /households/{id}/members` route. |
| `services/account-service/internal/membership/processor.go` | Has `ByHouseholdProvider()` — reuse for member lookup. |
| `services/account-service/internal/membership/entity.go` | Membership entity (join source). |
| `services/account-service/cmd/main.go` | Route wiring — household routes already initialized. |

### auth-service — User data (read-only reference)
| File | Purpose |
|---|---|
| `services/auth-service/internal/user/entity.go` | `auth.users` table with `display_name` column. Cross-schema read target. |

### frontend — Types
| File | Purpose |
|---|---|
| `frontend/src/types/models/task.ts` | Add `ownerUserId` to attributes interfaces. |
| `frontend/src/types/models/reminder.ts` | Add `ownerUserId` to attributes interfaces. |
| `frontend/src/types/models/summary.ts` | No changes (summaries remain household-wide). |

### frontend — Hooks & API
| File | Purpose |
|---|---|
| `frontend/src/lib/hooks/api/use-tasks.ts` | Update mutation payloads for ownerUserId. |
| `frontend/src/lib/hooks/api/use-reminders.ts` | Update mutation payloads for ownerUserId. |
| `frontend/src/lib/hooks/api/use-household-members.ts` | **New.** Hook to fetch household members. |

### frontend — Schemas
| File | Purpose |
|---|---|
| `frontend/src/lib/schemas/task.schema.ts` | Add optional `ownerUserId` to create/update schemas. |
| `frontend/src/lib/schemas/reminder.schema.ts` | Add optional `ownerUserId` to create/update schemas. |

### frontend — Components
| File | Purpose |
|---|---|
| `frontend/src/components/features/tasks/create-task-dialog.tsx` | Add owner dropdown. |
| `frontend/src/components/features/reminders/create-reminder-dialog.tsx` | Add owner dropdown. |
| `frontend/src/components/features/tasks/task-card.tsx` | Add owner display to mobile card. |
| `frontend/src/components/features/reminders/reminder-card.tsx` | Add owner display to mobile card. |
| `frontend/src/pages/TasksPage.tsx` | Add owner column, filter bar, sorting, empty state. |
| `frontend/src/pages/RemindersPage.tsx` | Add owner column, filter bar, sorting, empty state. |
| `frontend/src/pages/DashboardPage.tsx` | Make widget cards clickable links. |

---

## Key Decisions

1. **Cross-schema join for display names**: The account-service will query `auth.users` directly (same database, different schema) rather than making an HTTP call to auth-service. This is a read-only projection for display names only, keeping the endpoint lightweight. Documented as an intentional trade-off.

2. **`ownerUserId` as nullable `*string` in REST**: Using `*string` (not `*uuid.UUID`) in REST models allows JSON `null` representation and avoids UUID zero-value confusion. The field is omitted from JSON when nil, and explicitly `null` when set to household-wide.

3. **Default owner on create**: The resource layer (HTTP handler) sets `ownerUserId` to the authenticated user's ID when the field is not present in the request. The processor receives the resolved value. This keeps the defaulting logic at the transport boundary.

4. **Client-side filtering with URL params**: All filtering/sorting is client-side per PRD. Filter state is stored in URL search params (`?status=pending&owner=uuid&q=search`), enabling dashboard widget deep links and shareable URLs.

5. **"Overdue" as virtual status**: The frontend computes overdue as `status=pending AND dueOn < today`. The `?status=overdue` query param is interpreted by frontend filter logic, not the API. No backend changes needed.

6. **Members endpoint on household routes**: `GET /households/{id}/members` is added to the existing household `InitializeRoutes` in account-service, not as a separate domain. This keeps the endpoint co-located with the resource it's scoped to.

---

## Dependencies

| Dependency | Type | Notes |
|---|---|---|
| `auth.users` table | Database | Cross-schema read for display names. Same PostgreSQL database. |
| `@tanstack/react-table` | npm (existing) | Already used by TasksPage/RemindersPage for table rendering. Sorting is built-in. |
| `react-router-dom` | npm (existing) | Already used for routing. `Link` component for dashboard navigation. `useSearchParams` for filter state. |
| `react-hook-form` + `zod` | npm (existing) | Already used for form dialogs. Schema updates for ownerUserId field. |

---

## Patterns to Follow

- **Immutable domain models** with private fields and public accessors (see existing task/reminder models).
- **Fluent builder** with validation in `Build()` (see existing builders).
- **GORM AutoMigrate** for schema changes — add field to entity struct, migration runs on startup.
- **REST model separation**: domain model → REST model transformation via `Transform()` functions.
- **JSON:API envelope**: all responses wrapped in `{ data: { type, id, attributes } }` via `server.MarshalResponse`.
- **React Query hooks**: key factory pattern, `staleTime`/`gcTime` configuration, mutation with `invalidateQueries`.
- **Form schemas**: Zod schemas with `react-hook-form` resolver.
