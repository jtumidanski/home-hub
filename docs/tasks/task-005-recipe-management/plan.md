# Recipe Management — Implementation Plan

Last Updated: 2026-03-25

## Executive Summary

This plan covers the implementation of recipe management for Home Hub — a new `recipe-service` microservice with Cooklang-based recipe storage and parsing, plus frontend pages for browsing, viewing, creating, and editing recipes with a live preview editor. The feature spans a new Go service, TypeScript frontend pages, infrastructure integration, and CI/CD updates.

The implementation is structured into 5 phases: service scaffolding, Cooklang parser, CRUD + search API, frontend pages, and infrastructure integration. Phases 1–3 are backend-focused and sequential. Phase 4 (frontend) can begin in parallel once Phase 2 is complete (parser provides the TypeScript reference). Phase 5 (infra) can be done incrementally alongside other phases.

## Current State Analysis

- **Backend**: 4 existing Go microservices (auth, account, productivity, weather) following a consistent DDD pattern with model/entity/builder/processor/provider/resource/rest files
- **Frontend**: React + Vite + shadcn/ui with TanStack Query, react-hook-form + Zod, mobile-first responsive design using card views on mobile and DataTable on desktop
- **Infrastructure**: Docker Compose with nginx reverse proxy, K8s manifests, GitHub Actions CI with per-service change detection, Bruno API testing collections
- **No existing recipe or food-related code** — this is entirely greenfield

## Proposed Future State

- New `recipe-service` microservice with `recipe.*` schema
- Cooklang parser (Go) for server-side validation and structured output
- Cooklang parser (TypeScript) for client-side live preview
- 7 REST endpoints for recipe CRUD, tags, and restorations
- 3 new frontend pages: recipe list (with search/filter), recipe detail, recipe create/edit (with live Cooklang preview)
- Full infrastructure integration: compose, nginx, k8s, CI, build scripts, Bruno collection

---

## Phase 1: Service Scaffolding

Bootstrap the recipe-service following existing patterns. No business logic yet — just the skeleton that builds and runs.

### 1.1 Initialize Go module and directory structure
**Effort**: S
**Dependencies**: None
**Acceptance**: `go.mod` exists at `services/recipe-service/`, directory structure matches the service pattern (`cmd/`, `internal/config/`, `internal/recipe/`)

Create:
```
services/recipe-service/
  go.mod
  cmd/
    main.go
  internal/
    config/
      config.go
```

- `go.mod`: module `github.com/jtumidanski/home-hub/services/recipe-service`, Go 1.26, require shared modules
- `config.go`: Load config from env vars (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, PORT, JWKS_URL) — mirror productivity-service config
- `main.go`: Bootstrap with logging, tracing, database connection, auth validator, server lifecycle — mirror productivity-service main.go

### 1.2 Add recipe-service to Go workspace
**Effort**: S
**Dependencies**: 1.1
**Acceptance**: `go work sync` succeeds, `go build ./services/recipe-service/...` compiles

Add `./services/recipe-service` to `go.work`.

### 1.3 Create GORM entities and migration
**Effort**: M
**Dependencies**: 1.1
**Acceptance**: Service starts and creates `recipe.recipes`, `recipe.recipe_tags`, `recipe.recipe_restorations` tables via AutoMigrate

Files to create in `internal/recipe/`:
- `entity.go` — Entity, TagEntity, RestorationEntity structs with GORM tags per data-model.md; Migration function
- Register migration in main.go

### 1.4 Create domain model and builder
**Effort**: M
**Dependencies**: 1.3
**Acceptance**: Model is immutable with accessor methods; Builder enforces required fields (title, source, tenant_id, household_id); entity-to-model and model-to-entity conversions work

Files:
- `model.go` — Immutable Recipe model with private fields, accessor methods, Tag as a value type
- `builder.go` — Fluent builder with SetTitle, SetSource, SetDescription, SetServings, SetPrepTimeMinutes, SetCookTimeMinutes, SetSourceURL, SetTags; Build() validates required fields

### 1.5 Create provider (database access)
**Effort**: M
**Dependencies**: 1.3, 1.4
**Acceptance**: Provider can create, read (by ID), list (with pagination, search, tag filter), update, soft-delete, and restore recipes; tenant/household scoping enforced

File: `provider.go` — Lazy database access following productivity-service pattern. Key operations:
- `Create(db, entity) → entity`
- `GetByID(db, tenantID, householdID, id) → entity`
- `List(db, tenantID, householdID, search, tags, page, pageSize) → []entity, total`
- `Update(db, entity) → entity`
- `Delete(db, tenantID, householdID, id)`
- `Restore(db, tenantID, householdID, id) → entity`
- `ListTags(db, tenantID, householdID) → []TagCount`

### 1.6 Verify service builds and starts
**Effort**: S
**Dependencies**: 1.1–1.5
**Acceptance**: `go build ./services/recipe-service/...` succeeds; service starts with database connection and creates schema

---

## Phase 2: Cooklang Parser

Implement the Cooklang parser as an internal package. This is the core differentiating logic.

### 2.1 Define parser types
**Effort**: S
**Dependencies**: None
**Acceptance**: Types compile; represent ingredients, cookware, timers, text segments, steps, and parse results

File: `internal/recipe/cooklang/types.go`

Types needed:
- `Ingredient{Name, Quantity, Unit string}`
- `Cookware{Name string}`
- `Timer{Quantity, Unit string}`
- `Segment` — union type (text, ingredient, cookware, timer) via Type field + relevant fields
- `Step{Number int, Segments []Segment}`
- `ParseResult{Ingredients []Ingredient, Steps []Step}`
- `ParseError{Line, Column int, Message string}`

### 2.2 Implement core parser
**Effort**: L
**Dependencies**: 2.1
**Acceptance**: Parser handles core Cooklang syntax: `@ingredient{qty%unit}`, `#cookware{}`, `~timer{qty%unit}`, `--` line comments, `[- -]` block comments, blank-line step separation. Returns ParseResult or ParseError with line/column info.

File: `internal/recipe/cooklang/parser.go`

Parser behavior:
- Split input by blank lines into steps
- Within each step, scan for `@`, `#`, `~` markers
- Parse `{quantity%unit}` blocks; handle `{quantity}` (no unit) and `{}` (no quantity) forms
- Handle multi-word names: `@pecorino romano{100%g}` captures "pecorino romano"
- Handle single-word names without braces: `@salt` captures "salt" (word boundary)
- Strip comments before parsing
- Aggregate ingredients across steps; combine quantities where name and unit match
- Return structured errors with line numbers for malformed syntax

### 2.3 Implement parser validation
**Effort**: M
**Dependencies**: 2.2
**Acceptance**: `Validate(source) → []ParseError` returns descriptive errors for unclosed braces, malformed quantity blocks, empty ingredient names

File: same as 2.2 or `internal/recipe/cooklang/validate.go`

Validation checks:
- Unclosed `{` blocks
- Empty names after `@`, `#`, `~`
- Invalid quantity format (non-numeric where number expected)
- Empty source text

### 2.4 Write comprehensive parser tests
**Effort**: M
**Dependencies**: 2.2, 2.3
**Acceptance**: Tests cover: basic ingredients, multi-word ingredients, ingredients without braces, cookware, timers, comments (line and block), multi-step recipes, ingredient aggregation, error cases, edge cases (empty input, no markers, nested markers)

File: `internal/recipe/cooklang/parser_test.go`

Test cases should use table-driven tests. Cover the carbonara example from the PRD as an integration test.

---

## Phase 3: Recipe CRUD API

Wire the parser, domain model, and provider into HTTP handlers following JSON:API conventions.

### 3.1 Create processor (business logic)
**Effort**: M
**Dependencies**: 1.4, 1.5, 2.2
**Acceptance**: Processor orchestrates: validate Cooklang → build model → persist via provider; parse on read; handle soft delete/restore logic

File: `internal/recipe/processor.go`

Functions:
- `CreateRecipe(db, tenantID, householdID, attrs) → (Recipe, error)` — validates Cooklang, builds model, persists
- `GetRecipe(db, tenantID, householdID, id) → (Recipe, ParseResult, error)` — fetches and parses
- `ListRecipes(db, tenantID, householdID, search, tags, page, pageSize) → ([]Recipe, total, error)`
- `UpdateRecipe(db, tenantID, householdID, id, attrs) → (Recipe, ParseResult, error)` — re-validates if source changed
- `DeleteRecipe(db, tenantID, householdID, id) → error`
- `RestoreRecipe(db, tenantID, householdID, id) → (Recipe, ParseResult, error)`
- `ListTags(db, tenantID, householdID) → ([]TagCount, error)`

### 3.2 Create REST mappings
**Effort**: M
**Dependencies**: 3.1, 2.1
**Acceptance**: JSON:API serialization/deserialization works for all resource types per api-contracts.md; list includes pagination meta; detail includes parsed ingredients and steps

File: `internal/recipe/rest.go`

Mappings:
- Recipe list item → JSON:API resource (no source, no parsed data)
- Recipe detail → JSON:API resource (with source, ingredients, steps)
- Create/Update request body → domain attributes
- Tag count → JSON:API resource
- ParseError → JSON:API error response

### 3.3 Create HTTP resource (handlers + routes)
**Effort**: L
**Dependencies**: 3.1, 3.2
**Acceptance**: All 8 endpoints functional per api-contracts.md; proper HTTP status codes; tenant/household scoping from JWT claims; pagination; search/filter query params

File: `internal/recipe/resource.go`

Handlers:
- `POST /api/v1/recipes/parse` → 200 with parsed ingredients/steps/errors (preview, no persistence)
- `POST /api/v1/recipes` → 201 with detail, or 422 with validation errors
- `GET /api/v1/recipes` → 200 with paginated list, search, tag filter
- `GET /api/v1/recipes/:id` → 200 with detail + parsed data, or 404
- `PATCH /api/v1/recipes/:id` → 200 with updated detail, or 422/404
- `DELETE /api/v1/recipes/:id` → 204
- `POST /api/v1/recipes/restorations` → 200 with restored detail, or 410
- `GET /api/v1/recipes/tags` → 200 with tag list

Route registration in main.go following productivity-service pattern.

### 3.4 Write handler tests
**Effort**: M
**Dependencies**: 3.3
**Acceptance**: Unit tests for processor functions; handler tests verify status codes, response shapes, and error cases

### 3.5 Create service documentation
**Effort**: S
**Dependencies**: 3.3
**Acceptance**: `docs/domain.md`, `docs/rest.md`, `docs/storage.md` exist under recipe-service per the DOCS.md contract

### 3.6 End-to-end backend verification
**Effort**: S
**Dependencies**: All Phase 1–3 tasks
**Acceptance**: Service builds, passes lint (`golangci-lint run ./...`), passes tests (`go test ./... -count=1`), starts cleanly

---

## Phase 4: Frontend

Build the recipe management UI. Can start once Phase 2 types are defined (for the TypeScript parser).

### 4.1 Create TypeScript types for recipes
**Effort**: S
**Dependencies**: API contract defined (PRD)
**Acceptance**: Type file exports `Recipe`, `RecipeAttributes`, `RecipeDetailAttributes`, `RecipeCreateAttributes`, `RecipeUpdateAttributes`, `Ingredient`, `Step`, `Segment`, `RecipeTag`

File: `frontend/src/types/models/recipe.ts`

### 4.2 Create recipe API service
**Effort**: S
**Dependencies**: 4.1
**Acceptance**: RecipeService class extends BaseService with methods for all 7 endpoints; query params support for search, tag filter, pagination

File: `frontend/src/services/api/recipe.ts`

Methods:
- `listRecipes(tenant, params?)` — supports search, tags, pagination
- `getRecipe(tenant, id)`
- `createRecipe(tenant, attrs)`
- `updateRecipe(tenant, id, attrs)`
- `deleteRecipe(tenant, id)`
- `restoreRecipe(tenant, recipeId)`
- `listTags(tenant)`

Export instance and register in `frontend/src/services/api/index.ts`.

### 4.3 Create React Query hooks
**Effort**: M
**Dependencies**: 4.2
**Acceptance**: Hooks with query key factory, tenant scoping, cache invalidation on mutations, error toasts

File: `frontend/src/lib/hooks/api/use-recipes.ts`

Hooks:
- `recipeKeys` factory
- `useRecipes(params?)` — list query with search/tag/pagination params
- `useRecipe(id)` — detail query
- `useCreateRecipe()` — mutation
- `useUpdateRecipe()` — mutation
- `useDeleteRecipe()` — mutation with optimistic update
- `useRestoreRecipe()` — mutation
- `useRecipeTags()` — tags list query
- `useInvalidateRecipes()` — cache invalidation helper

### 4.4 Create Zod schemas for recipe forms
**Effort**: S
**Dependencies**: 4.1
**Acceptance**: Create and update schemas validate title (required, max 255), source (required), description (max 2000), servings (positive int), prep/cook time (non-negative int), source URL (max 2048), tags (array of strings)

File: `frontend/src/lib/schemas/recipe.schema.ts`

### 4.5 Create live preview hook (server-side parse)
**Effort**: M
**Dependencies**: 4.2, 3.3 (parse endpoint)
**Acceptance**: Hook calls `POST /api/v1/recipes/parse` with debounced source text (~300ms); returns parsed ingredients, steps, and errors; handles loading/error states gracefully; cancels in-flight requests when source changes

File: `frontend/src/lib/hooks/use-cooklang-preview.ts`

Uses the recipe API service's parse method. Returns `{ ingredients, steps, errors, isLoading }`. Debounce prevents excessive API calls while typing.

### 4.6 Create recipe list page
**Effort**: L
**Dependencies**: 4.3
**Acceptance**: Card-based list with title, description (truncated), tags (as badges), prep + cook time; search bar; tag multi-select filter; create button; tap navigates to detail; pull-to-refresh on mobile; empty state with create CTA

Files:
- `frontend/src/pages/RecipesPage.tsx` — page component
- `frontend/src/components/features/recipes/recipe-card.tsx` — mobile card

Pattern: follow TasksPage structure. Since recipes have richer metadata than tasks, the card should show tags as Badge components and time as icon + text.

### 4.7 Create recipe detail page
**Effort**: M
**Dependencies**: 4.3
**Acceptance**: Shows title, description, metadata (servings, times, source URL); ingredient list with quantity + unit + name; step-by-step instructions with ingredients/cookware/timers highlighted distinctly; edit button; delete with confirmation dialog

Files:
- `frontend/src/pages/RecipeDetailPage.tsx`
- `frontend/src/components/features/recipes/recipe-ingredients.tsx` — ingredient list component
- `frontend/src/components/features/recipes/recipe-steps.tsx` — step renderer with segment highlighting

This is a new page type (detail view) not yet in the codebase. Route: `/app/recipes/:id`.

### 4.8 Create recipe create/edit page with live preview
**Effort**: XL
**Dependencies**: 4.3, 4.4, 4.5
**Acceptance**: Form with all metadata fields + Cooklang textarea; live preview panel showing parsed ingredients and steps as user types (debounced); side-by-side on desktop, stacked on mobile; syntax errors shown inline; Cooklang cheat sheet toggle; server validation on submit; navigates to detail on success

Files:
- `frontend/src/pages/RecipeFormPage.tsx` — create/edit page (same page, edit pre-populates)
- `frontend/src/components/features/recipes/cooklang-editor.tsx` — textarea with syntax hints
- `frontend/src/components/features/recipes/cooklang-preview.tsx` — live preview renderer
- `frontend/src/components/features/recipes/cooklang-help.tsx` — collapsible syntax reference
- `frontend/src/components/features/recipes/tag-input.tsx` — tag entry component (add/remove tags)

Key implementation details:
- Use `useMemo` or `useEffect` with debounce (150–300ms) to parse on input
- Preview reuses `recipe-ingredients.tsx` and `recipe-steps.tsx` from detail page
- Desktop: `grid grid-cols-2 gap-4`; Mobile: stacked flex column
- Form uses react-hook-form + zod schema from 4.4

### 4.9 Add routing and navigation
**Effort**: S
**Dependencies**: 4.6, 4.7, 4.8
**Acceptance**: "Recipes" nav item appears in sidebar/mobile drawer; routes registered: `/app/recipes`, `/app/recipes/:id`, `/app/recipes/new`, `/app/recipes/:id/edit`

Files to modify:
- `frontend/src/App.tsx` — add routes
- `frontend/src/components/features/navigation/app-shell.tsx` — add nav item with icon (UtensilsCrossed from lucide-react)

### 4.10 Frontend tests
**Effort**: M
**Dependencies**: 4.5–4.8
**Acceptance**: TypeScript Cooklang parser has comprehensive tests; component tests for key interactions (create form submit, search filter, delete confirmation)

---

## Phase 5: Infrastructure Integration

Can be done incrementally. Some tasks (5.1, 5.2) can start as early as Phase 1.

### 5.1 Create Dockerfile
**Effort**: S
**Dependencies**: 1.1
**Acceptance**: `docker build -f services/recipe-service/Dockerfile .` succeeds; image runs and starts the service

File: `services/recipe-service/Dockerfile` — copy from productivity-service, change paths.

### 5.2 Create build script
**Effort**: S
**Dependencies**: 1.1
**Acceptance**: `scripts/build-recipe.sh` builds the service; `scripts/build-all.sh` includes it

Files:
- `scripts/build-recipe.sh` — new
- `scripts/build-all.sh` — add call to build-recipe.sh

### 5.3 Add to Docker Compose
**Effort**: S
**Dependencies**: 5.1
**Acceptance**: `docker compose up` starts recipe-service alongside other services

File: `deploy/compose/docker-compose.yml` — add recipe-service container following productivity-service pattern.

### 5.4 Add nginx routing
**Effort**: S
**Dependencies**: 5.3
**Acceptance**: `/api/v1/recipes` routes to recipe-service through nginx

File: `deploy/compose/nginx.conf` — add upstream and location blocks.

### 5.5 Add K8s manifests
**Effort**: S
**Dependencies**: 5.1
**Acceptance**: `recipe-service.yaml` exists with Deployment + Service; ingress.yaml updated with recipe routing

Files:
- `deploy/k8s/recipe-service.yaml` — new (copy from productivity-service.yaml)
- `deploy/k8s/ingress.yaml` — add `/api/v1/recipes` route

### 5.6 Add CI pipeline
**Effort**: M
**Dependencies**: 5.1
**Acceptance**: PR workflow detects recipe-service changes and runs build/test; main workflow builds and pushes docker image

Files:
- `.github/workflows/pr.yml` — add recipe change detection, build-recipe job, docker matrix entry
- `.github/workflows/main.yml` — add recipe-service to build matrix

### 5.7 Update supporting scripts
**Effort**: S
**Dependencies**: 5.2
**Acceptance**: `test-all.sh`, `lint-all.sh`, `ci-build.sh`, `ci-test.sh` include recipe-service

### 5.8 Create Bruno collection
**Effort**: S
**Dependencies**: 3.3
**Acceptance**: Bruno collection with requests for all 7 endpoints; uses `{{baseUrl}}` variable

Directory: `bruno/recipe/` with .bru files for each endpoint.

---

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Cooklang spec edge cases (metadata blocks, unicode, nested markers) | Parser bugs on unusual recipes | Medium | Implement core subset first; add edge cases via tests |
| Preview latency on slow connections | Noticeable delay between typing and preview update | Low | 300ms debounce masks most latency; show loading indicator; cancel in-flight requests |
| Tag filter performance with many tags/recipes | Slow list queries | Low | Compound indexes on (recipe_id, tag); HAVING clause for AND semantics |
| Large recipe source text | Memory/performance on parse | Low | Cap source at reasonable size (64KB); parser is streaming-style |

## Success Metrics

- All 17 PRD acceptance criteria pass
- Service builds, lints, and tests green
- Docker image builds successfully
- Full CRUD workflow works end-to-end (create via Cooklang editor → view parsed recipe → edit → delete → restore)
- Live preview renders correctly for standard Cooklang recipes
- Search and tag filter return correct results

## Required Resources and Dependencies

**Go dependencies** (new for recipe-service):
- Shared modules: model, tenant, logging, database, server, auth
- `github.com/google/uuid`
- `github.com/gorilla/mux`
- `gorm.io/gorm`
- No external Cooklang library — parser is built in-house

**Frontend dependencies** (no new packages needed):
- Existing: React, TanStack Query, react-hook-form, Zod, shadcn/ui, lucide-react
- No client-side Cooklang parser — preview uses server-side parse endpoint
