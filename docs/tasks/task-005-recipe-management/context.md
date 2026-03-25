# Recipe Management — Context

Last Updated: 2026-03-25

## Key Files to Reference

### Backend Patterns (copy from these)

| File | Purpose |
|------|---------|
| `services/productivity-service/cmd/main.go` | Service bootstrap: logging, tracing, DB connect, auth validator, server lifecycle, route registration |
| `services/productivity-service/internal/config/config.go` | Config loading from env vars |
| `services/productivity-service/internal/task/model.go` | Immutable domain model with private fields and accessors |
| `services/productivity-service/internal/task/entity.go` | GORM entity struct, table name, Migration function |
| `services/productivity-service/internal/task/builder.go` | Fluent builder with validation |
| `services/productivity-service/internal/task/processor.go` | Business logic: pure functions taking db + params |
| `services/productivity-service/internal/task/provider.go` | Lazy database access functions |
| `services/productivity-service/internal/task/resource.go` | HTTP handlers + route registration on mux.Router |
| `services/productivity-service/internal/task/rest.go` | JSON:API resource serialization/deserialization |
| `services/productivity-service/internal/task/restoration/` | Soft delete + restore pattern |
| `services/productivity-service/go.mod` | Module dependencies reference |
| `services/productivity-service/Dockerfile` | Multi-stage Docker build |

### Frontend Patterns (copy from these)

| File | Purpose |
|------|---------|
| `frontend/src/App.tsx` | Route registration under `<Route path="/app">` |
| `frontend/src/pages/TasksPage.tsx` | Page layout: responsive card/table, pull-to-refresh, create dialog |
| `frontend/src/services/api/base.ts` | BaseService with tenant headers, CRUD helpers |
| `frontend/src/services/api/productivity.ts` | Service class example (listTasks, createTask, etc.) |
| `frontend/src/lib/hooks/api/use-tasks.ts` | React Query hooks: key factory, queries, mutations, cache invalidation |
| `frontend/src/lib/schemas/task.schema.ts` | Zod validation schema + defaults |
| `frontend/src/types/models/task.ts` | TypeScript model interfaces (attributes, create/update types) |
| `frontend/src/types/api/responses.ts` | ApiResponse, ApiListResponse envelopes |
| `frontend/src/components/features/tasks/task-card.tsx` | Mobile card component with CardActionMenu |
| `frontend/src/components/features/tasks/create-task-dialog.tsx` | Dialog form with react-hook-form + Zod |
| `frontend/src/components/features/navigation/app-shell.tsx` | Nav items + sidebar/drawer |
| `frontend/src/components/common/data-table.tsx` | Desktop table component |
| `frontend/src/context/tenant-context.tsx` | Tenant/household context for multi-tenancy |

### Infrastructure (modify these)

| File | Action |
|------|--------|
| `go.work` | Add `./services/recipe-service` |
| `deploy/compose/docker-compose.yml` | Add recipe-service container |
| `deploy/compose/nginx.conf` | Add upstream + location for `/api/v1/recipes` |
| `deploy/k8s/ingress.yaml` | Add recipe routing rule |
| `.github/workflows/pr.yml` | Add recipe change detection + build job |
| `.github/workflows/main.yml` | Add recipe-service to docker build matrix |
| `scripts/build-all.sh` | Add call to build-recipe.sh |
| `scripts/test-all.sh` | Add recipe-service to test loop |
| `scripts/lint-all.sh` | Add recipe-service to lint loop |
| `scripts/ci-build.sh` | Add recipe-service build |
| `scripts/ci-test.sh` | Add recipe-service test |

## Key Decisions

1. **New service** — recipe-service is a standalone microservice, not part of productivity-service
2. **Cooklang format** — Recipes stored as raw Cooklang text; parsed server-side on read; client-side parser for live preview only
3. **Custom parser** — No external Cooklang library; parser built in-house in both Go and TypeScript
4. **Core spec subset** — v1 supports `@ingredients`, `#cookware`, `~timers`, comments (`--`, `[- -]`); defer metadata blocks and edge cases
5. **Single parser (Go only)** — No client-side TypeScript parser; live preview calls a server-side parse endpoint (debounced ~300ms); single source of truth
6. **Restore window** — Same duration as productivity-service
7. **Tag management** — Per-recipe editing only in v1; no rename/merge across recipes
6. **Tags as free-form strings** — Normalized lowercase, stored in separate table; no tag taxonomy or management UI in v1
7. **Soft delete with restore** — Matches productivity-service pattern
8. **Recipe detail as separate page** — Not a dialog/modal; uses route `/app/recipes/:id`
9. **Recipe form as separate page** — Not a dialog; uses routes `/app/recipes/new` and `/app/recipes/:id/edit` (same component)

## Dependencies Between Phases

```
Phase 1 (Scaffolding) ─────→ Phase 3 (CRUD API)
         │                         │
         │                         ↓
         │                    Phase 5 (Infra) ← can start early with 5.1, 5.2
         │
         └──→ Phase 2 (Parser) ──→ Phase 3 (needs parser for processor)
                    │
                    └──→ Phase 4 (Frontend) ← needs parser types for TS parser
```

Phase 4 tasks 4.1–4.4 can start as soon as the API contract is defined (it already is in the PRD). Task 4.5 (TS parser) benefits from Phase 2 as a reference. Tasks 4.6–4.8 need the API running for integration testing.

## New Files to Create

### Backend (services/recipe-service/)
```
cmd/main.go
internal/config/config.go
internal/recipe/model.go
internal/recipe/entity.go
internal/recipe/builder.go
internal/recipe/processor.go
internal/recipe/provider.go
internal/recipe/resource.go
internal/recipe/rest.go
internal/recipe/cooklang/types.go
internal/recipe/cooklang/parser.go
internal/recipe/cooklang/parser_test.go
docs/domain.md
docs/rest.md
docs/storage.md
go.mod
Dockerfile
```

### Frontend (frontend/src/)
```
types/models/recipe.ts
services/api/recipe.ts
lib/hooks/api/use-recipes.ts
lib/schemas/recipe.schema.ts
lib/hooks/use-cooklang-preview.ts
pages/RecipesPage.tsx
pages/RecipeDetailPage.tsx
pages/RecipeFormPage.tsx
components/features/recipes/recipe-card.tsx
components/features/recipes/recipe-ingredients.tsx
components/features/recipes/recipe-steps.tsx
components/features/recipes/cooklang-editor.tsx
components/features/recipes/cooklang-preview.tsx
components/features/recipes/cooklang-help.tsx
components/features/recipes/tag-input.tsx
```

### Infrastructure
```
scripts/build-recipe.sh
deploy/k8s/recipe-service.yaml
bruno/recipe/*.bru
```
