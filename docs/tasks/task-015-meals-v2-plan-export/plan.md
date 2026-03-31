# Meals v2: Plan Generation + Markdown Export — Implementation Plan

Last Updated: 2026-03-29

---

## Executive Summary

Add weekly meal planning to the recipe-service. Users select existing recipes, assign them to days/slots on a weekly grid, save plans, and export consolidated ingredient lists as markdown. This builds on recipe management (task-005) and ingredient normalization (task-014).

The implementation spans 5 phases: backend domain modeling, backend API/export logic, nginx routing, frontend planner UI, and integration testing. The backend introduces two new domain packages (`plan`, `planitem`) plus an `export` package for ingredient consolidation and markdown generation. The frontend adds a new planner page with a weekly grid, recipe selector, ingredient preview, and export modal.

---

## Current State Analysis

### Backend (recipe-service)

- **Existing domains:** `recipe`, `ingredient`, `normalization`, `planner`, `audit`
- **Established patterns:** Immutable models, builder pattern, processor/provider/administrator separation, JSON:API resources, audit event emission
- **Ingredient normalization complete:** `recipe_ingredients` table with canonical ingredient resolution, unit registry with family-based matching
- **Planner config exists:** `recipe_planner_configs` table with classification, servings_yield per recipe
- **Audit infrastructure ready:** `recipe_audit_events` table with `audit.Emit()` pattern

### Frontend

- **Stack:** React 19, TanStack Query, react-hook-form + Zod, shadcn/ui, Tailwind CSS
- **Recipe management complete:** List/detail/create/edit pages with filtering, search, pagination
- **API patterns established:** BaseService, TanStack Query hooks with cache key factories, tenant context
- **Navigation config:** NavGroups in `nav-config.ts`, routes in `App.tsx`

### Infrastructure

- **Nginx:** Path-prefix routing to services, recipe-service handles `/api/v1/recipes` and `/api/v1/ingredients`
- **Need:** Add `/api/v1/meals` route to recipe-service

---

## Proposed Future State

### New Backend Packages

| Package | Purpose |
|---------|---------|
| `internal/plan` | Plan week domain: model, entity, builder, processor, provider, administrator, resource, rest |
| `internal/planitem` | Plan item domain: model, entity, builder, processor, provider, administrator, resource, rest |
| `internal/export` | Read-only aggregation: ingredient consolidation + markdown generation |

### New Database Tables

- `plan_weeks` — Weekly plan header (tenant-scoped, unique per household+starts_on)
- `plan_items` — Recipe assignments to day/slot within a plan week

### New API Endpoints (under `/api/v1/meals`)

- Plan CRUD: create, list, get, update name
- Plan actions: lock, unlock, duplicate
- Plan items: add, update, remove
- Export: markdown generation, consolidated ingredients JSON

### New Frontend Pages

- `/app/meals` — Planner page with week grid, recipe selector, ingredient preview
- Export modal with markdown preview, copy, and download

---

## Implementation Phases

### Phase 1: Backend Domain Models (Backend Foundation)

Build the `plan` and `planitem` domain packages following existing patterns. This phase establishes the data layer without REST endpoints.

**Deliverables:**
- Plan week: model, entity, builder, processor, provider, administrator
- Plan item: model, entity, builder, processor, provider, administrator
- GORM migrations for both tables
- Registration in `main.go`

### Phase 2: Backend REST API (Backend Endpoints)

Add JSON:API resources and HTTP handlers for all plan and plan item endpoints. Wire audit event emission.

**Deliverables:**
- Plan REST: resource types, list/detail handlers, create/update/lock/unlock/duplicate handlers
- Plan item REST: resource types, add/update/remove handlers
- Audit events for all state-changing operations
- Recipe usage query support (`include_usage=true` on recipe list)

### Phase 3: Backend Export (Ingredient Consolidation + Markdown)

Build the `export` package for cross-domain read aggregation.

**Deliverables:**
- Ingredient consolidation logic (canonical matching, unit matching, serving scaling)
- Markdown generation with plan/day/slot/ingredient formatting
- JSON endpoint for consolidated ingredients
- Export audit event

### Phase 4: Infrastructure (Nginx Routing)

Add the `/api/v1/meals` route to nginx configuration.

**Deliverables:**
- Nginx route: `/api/v1/meals` -> recipe-service
- Architecture docs update

### Phase 5: Frontend (Planner UI)

Build the planner page with all UI components.

**Deliverables:**
- Planner page with week navigation
- Weekly grid component (day rows, slot columns)
- Recipe selector panel with search, classification auto-filter, usage metadata
- Plan item management (add/edit/remove via popovers)
- Serving multiplier/planned servings input
- Ingredient preview panel
- Lock/unlock/duplicate/export actions
- Export modal with markdown preview, copy-to-clipboard, download
- Deleted recipe warning indicators
- Navigation config and routing

---

## Risk Assessment and Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Ingredient consolidation complexity | Medium | Conservative matching rules (exact canonical_id + exact unit only); unresolved ingredients listed individually |
| Performance of consolidation for large plans | Low | Typical plan is <= 21 items; synchronous processing sufficient for v2 |
| Recipe deletion after plan creation | Medium | No FK cascade; `recipe_deleted` computed at read time; UI warning indicators |
| Duplicate unique constraint races | Low | Database unique index on (tenant, household, starts_on); 409 response |
| Frontend grid complexity | Medium | Start with table-based layout; defer drag-and-drop to stretch goal |

---

## Success Metrics

- All PRD acceptance criteria met (31 items)
- Plan CRUD operations complete within performance targets (list < 200ms, detail < 300ms, export < 500ms)
- Recipe-service Docker build passes
- Frontend build passes
- Existing recipe and ingredient functionality unaffected

---

## Required Resources and Dependencies

### Dependencies (Existing)

- `recipe` domain — recipe validation, title/servings lookup
- `normalization` domain — recipe_ingredients data for consolidation
- `ingredient` domain — canonical ingredient display names
- `planner` domain — classification and servings_yield for display
- `audit` domain — event emission

### Dependencies (New)

- None — `serving_multiplier` uses native `float64` (no third-party decimal library needed)

### Infrastructure

- Nginx config change (one location block addition)

---

## Timeline Estimates

| Phase | Effort | Dependencies |
|-------|--------|--------------|
| Phase 1: Backend Domain Models | L | None |
| Phase 2: Backend REST API | L | Phase 1 |
| Phase 3: Backend Export | M | Phase 1, Phase 2 |
| Phase 4: Infrastructure | S | None (can parallel) |
| Phase 5: Frontend | XL | Phase 2, Phase 3, Phase 4 |
