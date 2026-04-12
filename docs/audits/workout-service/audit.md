# Backend Audit — workout-service

- **Service Path:** services/workout-service
- **Guidelines Source:** backend-dev-guidelines skill
- **Date:** 2026-04-12
- **Build:** PASS
- **Tests:** 10 packages passed, 0 failed
- **Overall:** NEEDS-WORK

## Build & Test Results

```
ok  github.com/jtumidanski/home-hub/services/workout-service/internal            3.869s
ok  github.com/jtumidanski/home-hub/services/workout-service/internal/exercise    2.805s
ok  github.com/jtumidanski/home-hub/services/workout-service/internal/performance 1.811s
ok  github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem 0.604s
ok  github.com/jtumidanski/home-hub/services/workout-service/internal/region      4.930s
ok  github.com/jtumidanski/home-hub/services/workout-service/internal/retention   4.538s
ok  github.com/jtumidanski/home-hub/services/workout-service/internal/theme       5.432s
ok  github.com/jtumidanski/home-hub/services/workout-service/internal/today       6.090s
ok  github.com/jtumidanski/home-hub/services/workout-service/internal/tz          5.795s
ok  github.com/jtumidanski/home-hub/services/workout-service/internal/week        5.628s
```

## Domain Checklist Results

### exercise

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:63 NewBuilder, builder.go:86 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:51 |
| DOM-03 | Make(Entity) function | PASS | entity.go:77 |
| DOM-04 | Transform function | PASS | rest.go:79 |
| DOM-05 | TransformSlice function | PASS | rest.go:107 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:30 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:33 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:19-20 |
| DOM-09 | Transform errors handled | FAIL | entity.go:52 json.Marshal error discarded with _ in ToEntity() |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:29 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:9 EntityProvider |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Handlers delegate to processor |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | Through processor |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/update/softDelete |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:162-182 writeExerciseError |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:36-38, rest.go:52-61, rest.go:75-77 |
| DOM-19 | Request models use flat structure | FAIL | rest.go:48-49 CreateRequest has nested Defaults DefaultsRest; rest.go:73 UpdateRequest has nested *DefaultsRest |
| DOM-20 | Table-driven tests | FAIL | No table-driven t.Run tests — individual test functions only |

### performance

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:47 NewBuilder, builder.go:65 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:66 |
| DOM-03 | Make(Entity) function | PASS | entity.go:87 |
| DOM-04 | Transform function | PASS | rest.go:56 |
| DOM-05 | TransformSlice function | PASS | rest.go:85 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:28 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:79 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:22-23 |
| DOM-09 | Transform errors handled | FAIL | processor.go:134,277 Make(e) error discarded with _ |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:91 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:9 EntityProvider |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Handlers delegate to processor |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | Through processor |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/update/deleteSets/createSet/loadSets |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:39-53 writePerformanceError |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:43-50 |
| DOM-19 | Request models use flat structure | FAIL | rest.go:107-113 PatchPerformanceRequest has nested Actuals; rest.go:135-150 PutPerformanceSetsRequest has nested Sets |
| DOM-20 | Table-driven tests | PASS | processor_test.go:24-49 TestApplyExplicitStatus with t.Run |

### planneditem

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:37 NewBuilder, builder.go:56 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:56 |
| DOM-03 | Make(Entity) function | PASS | entity.go:78 |
| DOM-04 | Transform function | PASS | rest.go:49 |
| DOM-05 | TransformSlice function | PASS | rest.go:71 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:27 |
| DOM-07 | Handlers pass d.Logger() | PASS | Via weekview/resource.go:191 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | Via weekview/resource.go:31-34 |
| DOM-09 | Transform errors handled | PASS | Transform is infallible |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:29 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:9 EntityProvider |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Handlers in weekview delegate to processor |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | Through processor |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/update/delete/Clone |
| DOM-17 | Domain error to HTTP status mapping | PASS | weekview/resource.go:351-365 |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:38-46 |
| DOM-19 | Request models use flat structure | FAIL | weekview/rest.go:156 has nested Planned *PlannedAttrs |
| DOM-20 | Table-driven tests | FAIL | Individual test functions only, no table-driven t.Run |

### region

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:27 NewBuilder, builder.go:38 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:31 |
| DOM-03 | Make(Entity) function | PASS | entity.go:44 |
| DOM-04 | Transform function | PASS | rest.go:48 |
| DOM-05 | TransformSlice function | PASS | rest.go:58 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:32 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:33 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:19-20 |
| DOM-09 | Transform errors handled | PASS | Transform is infallible |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:25 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:9 EntityProvider |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Delegates to processor |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | Through processor |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/update/softDelete |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:52-56,80-91,105-106 |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:17-19 |
| DOM-19 | Request models use flat structure | PASS | rest.go:22-25, rest.go:38-41 |
| DOM-20 | Table-driven tests | PASS | processor_test.go:51-71 with t.Run |

### theme

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:27 NewBuilder, builder.go:38 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:33 |
| DOM-03 | Make(Entity) function | PASS | entity.go:46 |
| DOM-04 | Transform function | PASS | rest.go:48 |
| DOM-05 | TransformSlice function | PASS | rest.go:57 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:29 |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:33 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:19-20 |
| DOM-09 | Transform errors handled | PASS | Transform is infallible |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:25 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:9 EntityProvider |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Delegates to processor |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | Through processor |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/update/softDelete |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:52-56,80-91,105-106 |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:17-19 |
| DOM-19 | Request models use flat structure | PASS | rest.go flat structures |
| DOM-20 | Table-driven tests | PASS | processor_test.go:51-71 with t.Run |

### week

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:23 NewBuilder, builder.go:41 Build |
| DOM-02 | ToEntity() method | PASS | entity.go:32 |
| DOM-03 | Make(Entity) function | PASS | entity.go:45 |
| DOM-04 | Transform function | PASS | rest.go:32 |
| DOM-05 | TransformSlice function | PASS | rest.go:45 |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:23 |
| DOM-07 | Handlers pass d.Logger() | PASS | Via weekview/resource.go:105 |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | Via weekview/resource.go:29 |
| DOM-09 | Transform errors handled | FAIL | entity.go:33 json.Marshal error discarded with _ in ToEntity() |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:25 |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:12 EntityProvider |
| DOM-12 | No os.Getenv() in handlers | PASS | No matches |
| DOM-13 | No cross-domain logic in handlers | PASS | Delegates to processor |
| DOM-14 | Handlers don't call providers directly | PASS | Through processor |
| DOM-15 | No direct entity creation in handlers | PASS | Through processor |
| DOM-16 | administrator.go exists | PASS | administrator.go with create/update |
| DOM-17 | Domain error to HTTP status mapping | PASS | weekview/resource.go:139 (400), weekview/resource.go:109 (404) |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:21-29 |
| DOM-19 | Request models use flat structure | PASS | weekview/rest.go:105-108 |
| DOM-20 | Table-driven tests | PASS | processor_test.go:77-98 with t.Run |

## Sub-Domain Checklist Results

### summary

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| SUB-01 | Has processor or uses parent | PASS | processor.go:22-29 own Processor |
| SUB-02 | Has administrator for writes | N/A | Read-only projection |
| SUB-03 | Uses RegisterInputHandler for POST | N/A | GET-only endpoint |
| SUB-04 | No manual JSON parsing | PASS | No manual parsing |

### today

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| SUB-01 | Has processor or uses parent | PASS | processor.go:19-26 own Processor |
| SUB-02 | Has administrator for writes | N/A | Read-only projection |
| SUB-03 | Uses RegisterInputHandler for POST | N/A | GET-only endpoint |
| SUB-04 | No manual JSON parsing | PASS | No manual parsing |

### weekview

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| SUB-01 | Has processor or uses parent | PASS | processor.go:34-42 own Processor |
| SUB-02 | Has administrator for writes | PASS | administrator.go with cloneItems |
| SUB-03 | Uses RegisterInputHandler for POST | PASS | resource.go:30-34 |
| SUB-04 | No manual JSON parsing | PASS | No manual parsing |

## Summary

### Blocking (must fix)
- **exercise DOM-09**: entity.go:52 json.Marshal error discarded with _ in ToEntity()
- **exercise DOM-19**: rest.go:48-49 CreateRequest has nested Defaults DefaultsRest struct
- **exercise DOM-20**: No table-driven t.Run tests
- **performance DOM-09**: processor.go:134,277 Make(e) error discarded with _
- **performance DOM-19**: rest.go:107-113 PatchPerformanceRequest has nested Actuals; rest.go:135-150 PutPerformanceSetsRequest has nested Sets
- **planneditem DOM-19**: weekview/rest.go:156 has nested Planned *PlannedAttrs
- **planneditem DOM-20**: No table-driven t.Run tests
- **week DOM-09**: entity.go:33 json.Marshal error discarded with _ in ToEntity()

### Non-Blocking (should fix)
- region and theme are fully compliant reference implementations (20/20)

---

## Remediation Notes (2026-04-12)

| Item | Status | Implementation |
|------|--------|----------------|
| exercise DOM-09 | DONE | `entity.go:51` ToEntity() now returns `(Entity, error)`, propagates json.Marshal error |
| exercise DOM-19 | DONE | `rest.go` CreateRequest/UpdateRequest flattened — `DefaultSets`, `DefaultReps`, etc. replace nested `Defaults DefaultsRest`. Handler in `resource.go` updated. Frontend `workout.ts` and `use-workouts.ts` updated to match flat wire format. |
| exercise DOM-20 | DONE | `processor_test.go` Create rejection tests consolidated into table-driven `TestProcessor_Create_Rejections` with `t.Run` |
| performance DOM-09 | DONE | `processor.go:134,280` Make(e) errors now propagated instead of discarded |
| performance DOM-19 | DONE | `rest.go` PatchPerformanceRequest flattened — `ActualSets`, `ActualReps`, etc. replace nested `Actuals *PerformanceActualsAttrs`. Handler in `resource.go` updated. Frontend `WorkoutTodayPage.tsx` and `use-workouts.ts` updated. |
| planneditem DOM-19 | DONE | `weekview/rest.go` AddPlannedItemRequest, BulkAddPlannedItemEntry, UpdatePlannedItemRequest flattened — `PlannedSets`, `PlannedReps`, etc. replace nested `Planned *PlannedAttrs`. Handler in `weekview/resource.go` updated. |
| planneditem DOM-20 | DONE | `processor_test.go` Add rejections consolidated into table-driven `TestProcessor_Add_Rejections`; BulkAdd cases into `TestProcessor_BulkAdd`; Reorder validation into `TestProcessor_Reorder_Validation` — all with `t.Run` |
| week DOM-09 | DONE | `entity.go:32` ToEntity() now returns `(Entity, error)`, propagates json.Marshal error |

### Verification
- `go build ./...` — clean
- `go test ./... -count=1` — all 10 packages pass
- `npx tsc --noEmit` — frontend clean
