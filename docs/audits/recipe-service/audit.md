# Backend Audit — recipe-service

- **Service Path:** services/recipe-service
- **Guidelines Source:** backend-dev-guidelines skill
- **Date:** 2026-04-12
- **Build:** PASS
- **Tests:** 8 packages passed, 0 failed
- **Overall:** PASS

## Build & Test Results

```
ok  github.com/jtumidanski/home-hub/services/recipe-service/internal/export          0.520s
ok  github.com/jtumidanski/home-hub/services/recipe-service/internal/ingredient      0.827s
ok  github.com/jtumidanski/home-hub/services/recipe-service/internal/normalization   1.079s
ok  github.com/jtumidanski/home-hub/services/recipe-service/internal/plan            1.337s
ok  github.com/jtumidanski/home-hub/services/recipe-service/internal/planitem        3.581s
ok  github.com/jtumidanski/home-hub/services/recipe-service/internal/planner         1.847s
ok  github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe          2.706s
ok  github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe/cooklang 3.328s
ok  github.com/jtumidanski/home-hub/services/recipe-service/internal/retention       3.827s
```

## Domain Checklist Results

### ingredient

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:29 NewBuilder, builder.go:43 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:50 |
| DOM-03 | Make(Entity) function | PASS | entity.go:70 |
| DOM-04 | Transform function | PASS | rest.go:157 |
| DOM-05 | TransformSlice function | PASS | rest.go:149 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:27 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:53,107,175,202,259,347,395,420 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:22-25 |
| DOM-09 | Transform errors handled | PASS | Transform is infallible |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_integration_test.go:22 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:13 database.Query |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Handlers delegate to processor |
| DOM-14 | Handlers don't call providers directly | PASS | All through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No Entity{} in resource.go |
| DOM-16 | administrator.go exists | PASS | administrator.go with createEntity/saveEntity/deleteEntity |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:110-117 (400,422), resource.go:178-185 (404) |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:21-23 |
| DOM-19 | Request models use flat structure | PASS | rest.go:81-94 |
| DOM-20 | Table-driven tests | PASS | builder_test.go, processor_integration_test.go |

### normalization

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:30 NewBuilder, builder.go:46 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:32 |
| DOM-03 | Make(Entity) function | PASS | entity.go:55 |
| DOM-04 | Transform function | PASS | rest.go:37 |
| DOM-05 | TransformSlice function | PASS | rest.go:67 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:34 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:43,67 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:18-19 |
| DOM-09 | Transform errors handled | PASS | Transform is infallible |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_integration_test.go:22 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:9-11 database.SliceQuery |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Delegates to processor |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No Entity{} |
| DOM-16 | administrator.go exists | PASS | administrator.go with bulkCreate/bulkUpdate/deleteByRecipeID |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:46-49 (404) |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:20-22 |
| DOM-19 | Request models use flat structure | PASS | rest.go:24-35 |
| DOM-20 | Table-driven tests | PASS | builder_test.go, processor_test.go, processor_integration_test.go |

### plan

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:27 NewBuilder, builder.go:39 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:42 |
| DOM-03 | Make(Entity) function | PASS | entity.go:28 |
| DOM-04 | Transform function | PASS | rest.go:56 |
| DOM-05 | TransformSlice function | PASS | rest.go:84 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:36 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:80,120,169,189,216,244,280,317,337 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:54-55 /lock and /unlock are body-less action endpoints; RegisterInputHandler requires a JSON:API body and would reject empty requests — RegisterHandler is correct for these |
| DOM-09 | Transform errors handled | PASS | Transform is infallible |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_integration_test.go:23 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:11 database.Query |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Plan item routes moved to planitem.InitializeRoutes; plan/resource.go no longer mounts cross-domain handlers |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No Entity{} |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/update/deleteByID |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:87-93 (409,400), resource.go:192-199 (404,409) |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:20-22 |
| DOM-19 | Request models use flat structure | PASS | rest.go:93-108 |
| DOM-20 | Table-driven tests | PASS | builder_test.go, processor_integration_test.go |

### planitem

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:31 NewBuilder, builder.go:45 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:46 |
| DOM-03 | Make(Entity) function | PASS | entity.go:30 |
| DOM-04 | Transform function | PASS | rest.go:27 |
| DOM-05 | TransformSlice function | PASS | rest.go:42 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:27 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:51,128,171 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | Registered via plan/resource.go:45-46 |
| DOM-09 | Transform errors handled | PASS | Transform is infallible |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_integration_test.go:23 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:9 database.Query |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Cross-domain coupling removed; RecipeValidator interface injected via plan package, processor no longer imports recipe |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No Entity{} |
| DOM-16 | administrator.go exists | PASS | administrator.go with createItem/updateItem/deleteItem |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:68-76 (400), resource.go:131-143 (404) |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:23-25 |
| DOM-19 | Request models use flat structure | PASS | rest.go:51-71 |
| DOM-20 | Table-driven tests | PASS | builder_test.go, processor_integration_test.go |

### planner (internal domain, no HTTP layer)

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:27 NewBuilder, builder.go:38 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:28 |
| DOM-03 | Make(Entity) function | PASS | entity.go:41 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:18 |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_integration_test.go:21 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:9 database.Query |
| DOM-16 | administrator.go exists | PASS | administrator.go with createConfig/updateConfig |
| DOM-20 | Table-driven tests | PASS | builder_test.go, processor_test.go, processor_integration_test.go |

Note: DOM-04/05/07/08/09/12-15/17-19 are N/A — planner is an internal domain with no HTTP layer.

### recipe

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:32 NewBuilder, builder.go:49 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:50 |
| DOM-03 | Make(Entity) function | PASS | entity.go:72 |
| DOM-04 | Transform function | PASS | rest.go:77 |
| DOM-05 | TransformSlice function | PASS | rest.go:93 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:57 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:49,94,160,213,255,305,327,355 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:23-27 |
| DOM-09 | Transform errors handled | PASS | Transform is infallible |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_integration_test.go:24 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:11 database.Query |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Processor orchestrates cross-domain internally |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | No Entity{} |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/update/softDelete/restore/replaceTags |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:172-180 (400,422), resource.go:258-266 (404), resource.go:330-338 (404,400,410) |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:31-33 |
| DOM-19 | Request models use flat structure | PASS | rest.go:154-176 |
| DOM-20 | Table-driven tests | PASS | builder_test.go, processor_test.go, processor_integration_test.go |

## Sub-Domain Checklist Results

### export

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| SUB-01 | Has processor or uses parent | PASS | processor.go:63-71 own Processor |
| SUB-02 | Has administrator for writes | N/A | Read-only — generates markdown/consolidated ingredients |
| SUB-03 | Uses RegisterInputHandler for POST | N/A | No POST — GET endpoints only |
| SUB-04 | No manual JSON parsing | PASS | No json.Unmarshal/NewDecoder in resource.go |

## Summary

### Blocking (must fix)
- None — all resolved

### Non-Blocking (should fix)
- None
