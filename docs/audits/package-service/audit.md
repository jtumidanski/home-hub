# Backend Audit — package-service

- **Service Path:** services/package-service
- **Guidelines Source:** backend-dev-guidelines skill
- **Date:** 2026-04-12
- **Build:** PASS
- **Tests:** 3 packages passed, 0 failed
- **Overall:** PASS

## Build & Test Results

```
ok  github.com/jtumidanski/home-hub/services/package-service/internal/carrier        0.188s
ok  github.com/jtumidanski/home-hub/services/package-service/internal/tracking       0.457s
ok  github.com/jtumidanski/home-hub/services/package-service/internal/trackingevent  0.215s
```

## Domain Checklist Results

### tracking (Domain)

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS | builder.go:67 NewBuilder(), builder.go:91-107 fluent setters, builder.go:109 Build() |
| DOM-02 | ToEntity() method | PASS | entity.go:45 `func (m Model) ToEntity() Entity` |
| DOM-03 | Make(Entity) function | PASS | entity.go:67 `func Make(e Entity) (Model, error)` |
| DOM-04 | Transform function | PASS | rest.go:165 `func Transform(m Model) (RestModel, error)` |
| DOM-05 | TransformSlice function | PASS | rest.go:193 `func TransformSlice(models []Model) ([]RestModel, error)` |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go:44 `l logrus.FieldLogger`, processor.go:51 `NewProcessor(l logrus.FieldLogger, ...)` |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:39-41 `newProc(d.Logger(), r, db, ...)` used in all handlers |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS | resource.go:23 `RegisterInputHandler[CreateRequest]`, resource.go:24 `RegisterInputHandler[UpdateRequest]`; bodyless POST endpoints (archive/unarchive/refresh) correctly use RegisterHandler |
| DOM-09 | Transform errors handled | PASS | resource.go:342 uses `server.WriteError()` for the 429 response, consistent with all other error paths |
| DOM-10 | Test DB has tenant callbacks | PASS | processor_test.go:21 `database.RegisterTenantCallbacks(l, db)` |
| DOM-11 | Providers use lazy evaluation | PASS | provider.go:12 `database.Query[Entity]`, provider.go:18,24,31 `database.SliceQuery[Entity]` |
| DOM-12 | No os.Getenv() in resource.go | PASS | No `os.Getenv` in tracking package; only in config/config.go |
| DOM-13 | No cross-domain logic in handlers | PASS | All handlers delegate to Processor methods; detectCarrierHandler delegates to proc.DetectCarrier() |
| DOM-14 | Handlers don't call providers directly | PASS | All handler functions call Processor methods |
| DOM-15 | No db.Create/db.Save/db.Delete in resource.go | PASS | All GORM write operations in administrator.go |
| DOM-16 | administrator.go exists | PASS | administrator.go:8 create(), administrator.go:12 update(), administrator.go:16 deleteByID() |
| DOM-17 | Domain error to HTTP status mapping | PASS | resource.go:57-59 validation->400, resource.go:61-63 duplicate->409, resource.go:65-67 limit->422, resource.go:170-172 notFound->404, resource.go:174-176 notOwner->403, resource.go:341-356 refreshTooSoon->429, resource.go:268-270 notArchived->400 |
| DOM-18 | JSON:API interface on REST models | PASS | rest.go:30-32 RestModel, rest.go:41-43 RestSummaryModel, rest.go:55-57 RestTrackingEventModel, rest.go:79-81 RestDetailModel, rest.go:141-150 CreateRequest, rest.go:161-163 UpdateRequest |
| DOM-19 | Request models use flat structure | PASS | rest.go:132-139 CreateRequest, rest.go:153-159 UpdateRequest — no nested Data/Type/Attributes |
| DOM-20 | Table-driven tests | PASS | builder_test.go:22-158, processor_test.go:97-129 TestCreate_ValidationErrors, processor_test.go:342-367 TestModel_IsPolling |

### trackingevent (Sub-Domain)

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| SUB-01 | Has processor or uses parent processor | PASS | processor.go:13-16 Processor struct, processor.go:19 NewProcessor(l logrus.FieldLogger, ...) |
| SUB-02 | Has administrator for writes | PASS | administrator.go:10 `func Create(db *gorm.DB, ...)` |
| SUB-03 | Uses RegisterInputHandler[T] for POST | N/A | No HTTP endpoints; events created programmatically via Processor.CreateEvent() |
| SUB-04 | No manual JSON parsing | PASS | No json.NewDecoder, json.Unmarshal, or io.ReadAll in trackingevent package |

## Summary

### Blocking (must fix)
- None.

### Non-Blocking
- None.

### Totals

| Status | Count |
|--------|-------|
| PASS | 23 |
| FAIL | 0 |
| N/A | 1 |
