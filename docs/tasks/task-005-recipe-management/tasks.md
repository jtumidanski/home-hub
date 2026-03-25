# Recipe Management — Task Checklist

Last Updated: 2026-03-25

## Phase 1: Service Scaffolding

- [ ] **1.1** Initialize Go module and directory structure (S)
- [ ] **1.2** Add recipe-service to Go workspace (S)
- [ ] **1.3** Create GORM entities and migration (M)
- [ ] **1.4** Create domain model and builder (M)
- [ ] **1.5** Create provider — database access layer (M)
- [ ] **1.6** Verify service builds and starts (S)

## Phase 2: Cooklang Parser

- [ ] **2.1** Define parser types (S)
- [ ] **2.2** Implement core parser — @ingredients, #cookware, ~timers, comments, steps (L)
- [ ] **2.3** Implement parser validation — error reporting with line/column (M)
- [ ] **2.4** Write comprehensive parser tests (M)

## Phase 3: Recipe CRUD API

- [ ] **3.1** Create processor — business logic orchestration (M)
- [ ] **3.2** Create REST mappings — JSON:API serialization (M)
- [ ] **3.3** Create HTTP resource — handlers and route registration (L)
- [ ] **3.4** Write handler tests (M)
- [ ] **3.5** Create service documentation — domain.md, rest.md, storage.md (S)
- [ ] **3.6** End-to-end backend verification — build, lint, test (S)

## Phase 4: Frontend

- [ ] **4.1** Create TypeScript types for recipes (S)
- [ ] **4.2** Create recipe API service (S)
- [ ] **4.3** Create React Query hooks (M)
- [ ] **4.4** Create Zod schemas for recipe forms (S)
- [ ] **4.5** Create live preview hook — debounced server-side parse calls (M)
- [ ] **4.6** Create recipe list page — cards, search, tag filter (L)
- [ ] **4.7** Create recipe detail page — ingredients, steps, metadata (M)
- [ ] **4.8** Create recipe create/edit page with live Cooklang preview (XL)
- [ ] **4.9** Add routing and navigation (S)
- [ ] **4.10** Frontend tests — components (M)

## Phase 5: Infrastructure Integration

- [ ] **5.1** Create Dockerfile (S)
- [ ] **5.2** Create build script + update build-all.sh (S)
- [ ] **5.3** Add to Docker Compose (S)
- [ ] **5.4** Add nginx routing for /api/v1/recipes (S)
- [ ] **5.5** Add K8s manifests — deployment, service, ingress (S)
- [ ] **5.6** Add CI pipeline — PR detection + main build (M)
- [ ] **5.7** Update supporting scripts — test-all, lint-all, ci-build, ci-test (S)
- [ ] **5.8** Create Bruno collection (S)

---

## Summary

| Phase | Tasks | Effort |
|-------|-------|--------|
| 1. Service Scaffolding | 6 | 2S + 3M + 1S |
| 2. Cooklang Parser | 4 | 1S + 1L + 2M |
| 3. Recipe CRUD API | 6 | 2M + 1L + 1M + 2S |
| 4. Frontend | 10 | 3S + 2M + 1L + 1M + 1XL + 1S + 1M |
| 5. Infrastructure | 8 | 6S + 1M + 1S |
| **Total** | **34** | |
