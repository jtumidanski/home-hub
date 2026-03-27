# Ingredient Normalization — Context

Last Updated: 2026-03-27

## Key Files — Backend

### Existing (to modify)
| File | Purpose | Changes Needed |
|------|---------|----------------|
| `services/recipe-service/cmd/main.go` | Service entry point | Add migrations + route registration for 4 new domains |
| `services/recipe-service/internal/recipe/processor.go` | Recipe business logic | Wire normalization pipeline into Create/Update, include planner config in responses |
| `services/recipe-service/internal/recipe/resource.go` | HTTP handlers | Add filter params to list, wire normalization/planner into create/update/get handlers |
| `services/recipe-service/internal/recipe/rest.go` | JSON:API models | Extend RestDetailModel (ingredients, plannerConfig, plannerReady, plannerIssues), extend RestModel (plannerReady, classification, counts) |
| `services/recipe-service/internal/recipe/provider.go` | DB queries | Add filter support (plannerReady, classification, normalizationStatus), join for ingredient counts |
| `services/recipe-service/internal/recipe/model.go` | Domain model | No changes needed — planner config is separate domain |
| `services/recipe-service/internal/recipe/cooklang/parser.go` | Cooklang parser | No changes needed — produces Ingredient structs already |

### New Domains
| Directory | Files | Purpose |
|-----------|-------|---------|
| `internal/ingredient/` | model.go, entity.go, builder.go, processor.go, provider.go, resource.go, rest.go | Canonical ingredient registry CRUD |
| `internal/normalization/` | model.go, entity.go, builder.go, processor.go, provider.go, resource.go, rest.go, unit_registry.go | Recipe ingredient persistence, normalization pipeline, manual correction |
| `internal/planner/` | model.go, entity.go, builder.go, processor.go, provider.go | Planner config, readiness computation |
| `internal/audit/` | entity.go, emitter.go | Audit event persistence |

## Key Files — Frontend

### Existing (to modify)
| File | Changes Needed |
|------|----------------|
| `services/frontend/src/pages/RecipeDetailPage.tsx` | Add normalization panel, planner badge, planner config display, re-normalize button |
| `services/frontend/src/pages/RecipeFormPage.tsx` | Add planner settings section, normalization review panel, wire normalization into preview |
| `services/frontend/src/pages/RecipesPage.tsx` | Add planner badges to cards, normalization summary, filter controls |
| `services/frontend/src/hooks/use-recipes.ts` | Add filter params to useRecipes, extend query key factory |
| `services/frontend/src/services/api/recipe.ts` | Add resolve, renormalize endpoints; extend types |
| `services/frontend/src/components/features/recipes/recipe-card.tsx` | Add planner badge, normalization summary |
| `services/frontend/src/components/features/recipes/recipe-ingredients.tsx` | Add normalization status indicators |
| `services/frontend/src/components/features/navigation/nav-config.ts` | Add "Ingredients" nav item |
| `services/frontend/src/App.tsx` (or router config) | Add routes for ingredient pages |

### New Files
| File | Purpose |
|------|---------|
| `services/frontend/src/pages/IngredientsPage.tsx` | Canonical ingredient list |
| `services/frontend/src/pages/IngredientDetailPage.tsx` | Canonical ingredient detail/edit |
| `services/frontend/src/services/api/ingredient.ts` | API service for canonical ingredients |
| `services/frontend/src/hooks/use-ingredients.ts` | React Query hooks for ingredients + aliases |
| `services/frontend/src/hooks/use-ingredient-normalization.ts` | Hooks for resolve + renormalize |
| `components/features/recipes/ingredient-normalization-panel.tsx` | Normalization status display + actions |
| `components/features/recipes/ingredient-resolver.tsx` | Search/select for resolving ingredients |
| `components/features/recipes/planner-config-form.tsx` | Planner settings form fields |
| `components/features/recipes/planner-ready-badge.tsx` | Readiness badge |

## Key Decisions

1. **Normalization is synchronous** — runs within recipe create/update request. No async workers needed. Pipeline is DB-only lookups, target <50ms.

2. **4 separate domains, not 1 mega-domain** — ingredient, normalization, planner, audit are separate packages with own files. Cross-domain calls go through processor interfaces, not direct DB access.

3. **Reconciliation on source update** — match existing recipe_ingredients by raw_name, preserve manually_confirmed, re-normalize others. Simple and deterministic.

4. **Planner readiness computed on read** — not stored. Avoids staleness issues. Just checks classification + servings presence.

5. **Unresolved ingredients don't block planner readiness** — they block shopping-list generation (future task), not meal planning.

6. **Suggestion ranking via ILIKE + usage count** — no full-text search extensions needed. Simple, effective for the expected registry size.

7. **Audit events are write-only** — no read API in this task. Just persist for future use.

8. **Parse preview includes normalization** — the parse endpoint gets extended to show normalization status, giving users feedback before saving.

9. **Unit registry is static in-memory** — no database table. Hardcoded map of common unit strings to canonical identities.

10. **Canonical ingredients are tenant-scoped, not household-scoped** — shared across all households in a tenant. Recipe ingredients inherit household scope from their parent recipe.

## Dependencies Between Tasks

```
Phase 1 (Foundations) → Phase 2 (Pipeline) → Phase 5.2 (Normalization endpoints)
Phase 1 (Foundations) → Phase 3 (Planner)  → Phase 5.3 (Recipe endpoint mods)
Phase 1 (Foundations) → Phase 4 (Audit)    → Phase 5.3 (Recipe endpoint mods)
Phase 1 (Foundations) → Phase 5.1 (Ingredient endpoints)
Phase 5 (All API)     → Phase 6 (Frontend)
Phase 6.1 (API/Types) → Phase 6.2 (Hooks) → Phase 6.3-6.9 (Components/Pages)
```

Phase 1 tasks are independent of each other. Phase 5 tasks can mostly proceed in parallel once their domain foundations are complete. Frontend phases 6.3-6.9 are mostly independent once hooks (6.2) are done.

## Routing Impact

New routes added to recipe-service (no infrastructure changes needed):
- `/api/v1/ingredients` — handled by recipe-service (same service, new route group)
- `/api/v1/recipes/:id/ingredients/:ingredientId/resolve` — nested under existing recipe routes
- `/api/v1/recipes/:id/renormalize` — nested under existing recipe routes

**Note**: The `/api/v1/ingredients` path is NEW and needs to be routed to recipe-service in nginx/ingress config. Check `docker-compose.yml` and `k8s/` manifests.
