# Meals v2 — Implementation Context

Last Updated: 2026-03-29

---

## Key Files

### Backend — Existing Patterns to Follow

| File | Purpose |
|------|---------|
| `services/recipe-service/cmd/main.go` | Entry point; register new migrations and route initializers here |
| `services/recipe-service/internal/recipe/model.go` | Reference: immutable domain model pattern (private fields, public getters) |
| `services/recipe-service/internal/recipe/entity.go` | Reference: GORM entity, Make(), ToEntity(), Migration() |
| `services/recipe-service/internal/recipe/builder.go` | Reference: fluent builder with validation in Build() |
| `services/recipe-service/internal/recipe/processor.go` | Reference: Processor struct with (logger, context, db), business logic |
| `services/recipe-service/internal/recipe/provider.go` | Reference: functional query builders, pagination, filtering |
| `services/recipe-service/internal/recipe/administrator.go` | Reference: low-level CRUD operations |
| `services/recipe-service/internal/recipe/resource.go` | Reference: JSON:API resource types, list vs detail models, transformations |
| `services/recipe-service/internal/recipe/rest.go` | Reference: InitializeRoutes pattern, handler factories, query param parsing |
| `services/recipe-service/internal/audit/emitter.go` | Audit.Emit() — fire-and-forget event logging |
| `services/recipe-service/internal/normalization/entity.go` | recipe_ingredients table — used for consolidation reads |
| `services/recipe-service/internal/normalization/processor.go` | GetByRecipeID() — fetch normalized ingredients per recipe |
| `services/recipe-service/internal/normalization/unit_registry.go` | LookupUnit() — canonical unit resolution, unit families |
| `services/recipe-service/internal/ingredient/model.go` | Canonical ingredient model — DisplayName(), UnitFamily() |
| `services/recipe-service/internal/ingredient/provider.go` | Ingredient lookup by ID |
| `services/recipe-service/internal/planner/model.go` | Planner config — Classification(), ServingsYield() |
| `services/recipe-service/internal/planner/processor.go` | GetByRecipeID() — fetch planner config per recipe |

### Backend — New Files to Create

| File | Purpose |
|------|---------|
| `services/recipe-service/internal/plan/model.go` | PlanWeek immutable domain model |
| `services/recipe-service/internal/plan/entity.go` | PlanWeek GORM entity + migration |
| `services/recipe-service/internal/plan/builder.go` | PlanWeek builder with validation |
| `services/recipe-service/internal/plan/processor.go` | Plan business logic (create, update, lock, unlock, duplicate) |
| `services/recipe-service/internal/plan/provider.go` | Plan query functions (getByID, getAll with filters) |
| `services/recipe-service/internal/plan/administrator.go` | Plan CRUD operations |
| `services/recipe-service/internal/plan/resource.go` | JSON:API resource types + transformations |
| `services/recipe-service/internal/plan/rest.go` | HTTP handlers + route registration |
| `services/recipe-service/internal/planitem/model.go` | PlanItem immutable domain model |
| `services/recipe-service/internal/planitem/entity.go` | PlanItem GORM entity + migration |
| `services/recipe-service/internal/planitem/builder.go` | PlanItem builder with validation |
| `services/recipe-service/internal/planitem/processor.go` | Item business logic (add, update, remove, validate) |
| `services/recipe-service/internal/planitem/provider.go` | Item query functions (getByPlanID, getByRecipeID for usage) |
| `services/recipe-service/internal/planitem/administrator.go` | Item CRUD operations |
| `services/recipe-service/internal/planitem/resource.go` | JSON:API resource types |
| `services/recipe-service/internal/planitem/rest.go` | HTTP handlers (nested under plan routes) |
| `services/recipe-service/internal/export/processor.go` | Cross-domain orchestration for consolidation |
| `services/recipe-service/internal/export/markdown.go` | Markdown template generation |

### Frontend — Existing Patterns to Follow

| File | Purpose |
|------|---------|
| `frontend/src/App.tsx` | Route registration — add `/app/meals` |
| `frontend/src/components/features/navigation/nav-config.ts` | Navigation config — add Meals entry |
| `frontend/src/services/api/base.ts` | BaseService class pattern |
| `frontend/src/services/api/recipe.ts` | Reference: service implementation |
| `frontend/src/lib/hooks/api/use-recipes.ts` | Reference: TanStack Query hooks with cache key factories |
| `frontend/src/lib/schemas/recipe.schema.ts` | Reference: Zod validation schemas |
| `frontend/src/components/features/recipes/` | Reference: feature component organization |
| `frontend/src/context/tenant-context.tsx` | Tenant context for multi-tenancy |

### Frontend — New Files to Create

| File | Purpose |
|------|---------|
| `frontend/src/services/api/meals.ts` | MealsService — plan and item API calls |
| `frontend/src/lib/hooks/api/use-meals.ts` | TanStack Query hooks for plans, items, export |
| `frontend/src/lib/schemas/meals.schema.ts` | Zod schemas for plan and item forms |
| `frontend/src/pages/MealsPage.tsx` | Planner page component |
| `frontend/src/components/features/meals/` | Feature components directory |
| `frontend/src/components/features/meals/week-grid.tsx` | Weekly planning grid |
| `frontend/src/components/features/meals/recipe-selector.tsx` | Recipe search/select panel |
| `frontend/src/components/features/meals/ingredient-preview.tsx` | Consolidated ingredient preview |
| `frontend/src/components/features/meals/export-modal.tsx` | Markdown export modal |
| `frontend/src/components/features/meals/plan-item-popover.tsx` | Add/edit item popover |

### Infrastructure

| File | Purpose |
|------|---------|
| `deploy/compose/nginx.conf` | Add `/api/v1/meals` location block |
| `docs/architecture.md` | Update routing table |

---

## Key Decisions

1. **Two domains, not one:** `plan` and `planitem` are separate packages (consistent with existing recipe/normalization split) — plan handles week-level concerns, planitem handles item-level concerns.

2. **Export as separate package:** The `export` package performs read-only cross-domain aggregation (plan + planitem + normalization + ingredient). This follows the architecture guideline for dashboard-style aggregations.

3. **No FK cascade on recipe_id:** Plan items survive recipe soft-deletion. The `recipe_deleted` flag is computed at read time by checking the recipe's `deleted_at` field.

4. **Conservative ingredient consolidation:** Only merge when canonical_ingredient_id AND canonical_unit match exactly. Unit family conversion is a future concern.

5. **Plan created eagerly on save:** The frontend creates the plan_week via `POST /plans` when the user first saves or adds an item, then adds items via separate `POST /plans/{id}/items` calls. Empty plans are valid. No compound endpoint.

6. **Markdown export not persisted:** Generated on-demand from current plan state. No storage of exported artifacts.

7. **Serving scaling precedence:** planned_servings (integer) > serving_multiplier (decimal) > default 1.0. The effective multiplier is `planned_servings / recipe.servings_yield` when planned_servings is set.

8. **Day validation in processor, not DB:** The `day` field is validated against the plan week range in the processor/builder, not via database constraints.

9. **Slot enum in code, not DB:** The slot constraint is enforced in the builder validation, not as a database CHECK constraint.

10. **Routes under /api/v1/meals:** All plan endpoints use `/api/v1/meals/plans` prefix, routed to recipe-service via nginx.

11. **Serving multiplier uses `float64`, not shopspring/decimal:** The DB stores `DECIMAL(5,2)` but Go uses `*float64`. This is a scaling multiplier (values like 1.5, 2.0), not financial math — `float64` precision is more than sufficient and avoids adding a third-party dependency. GORM handles the `float64` <-> `DECIMAL` mapping natively.

12. **Cooklang fallback requires recipe+cooklang dependency in export:** When a recipe has no normalized ingredients, the export processor calls `recipe.Processor.ParseSource()` to get raw Cooklang ingredients. This means the `export` package depends on `recipe` (already needed for title/servings) and `cooklang` (for parsing).

13. **Plan creation is eager via separate API calls:** The frontend creates the plan via `POST /plans` first, then adds items via `POST /plans/{id}/items`. Empty plans are valid per the API contract. The frontend manages this with `if (!plan) createPlan() then addItem()`. No compound "create plan with items" endpoint.

14. **No plan delete in v2:** Plans are retained indefinitely per the PRD. Deleting would break recipe usage history. Soft-delete can be added later without schema changes if needed.

15. **Position auto-assigned by backend:** When adding a plan item without an explicit `position`, the backend defaults to `MAX(position) + 1` for the given day+slot. The frontend can optionally pass a position for explicit reordering but doesn't need to track positions for basic add operations.

---

## Dependencies Between Phases

```
Phase 1 (Domain Models) ──→ Phase 2 (REST API) ──→ Phase 3 (Export)
                                                          │
Phase 4 (Nginx) ──────────────────────────────────────────┤
                                                          ▼
                                                   Phase 5 (Frontend)
```

Phase 4 (Nginx) can be done in parallel with Phase 1-3. Phase 5 depends on all other phases.
