# Backend Audit — recipe-service (task-053 cook-count sort)

- **Service Path:** services/recipe-service
- **Scope:** changed Go files in range f8aff3a..a0b53fc (recipe package only)
- **Guidelines Source:** backend-dev-guidelines skill
- **Date:** 2026-06-14
- **Build:** PASS
- **Tests:** PASS (recipe package + full service; all green)
- **Overall:** PASS

## Build & Test Results

```
cd services/recipe-service && go build ./...        -> exit 0
cd services/recipe-service && go test ./... -count=1 -> all packages ok
cd services/recipe-service && go vet ./internal/recipe/ -> clean
```

New tests added by the change all pass:
- TestListFrequencySortOrder, TestListFrequencySortTieBreakerAcrossPages,
  TestListFrequencySortComposesWithSearch, TestListDefaultOrderUnchanged
- TestGetRecipeUsageScopedToHousehold
- TestParseUsageSort

## Domain Package: recipe

The only domain package in scope (`model.go` present). This is a pre-existing
domain; the diff touches the list/usage **read path** only (provider.go,
processor.go, resource.go) plus new tests. No new domain, sub-domain, builder,
entity, or rest.go surface was introduced, so most DOM-* items are unchanged by
this diff and pass on existing code.

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| DOM-01 | builder.go exists | PASS (pre-existing) | internal/recipe/builder.go present |
| DOM-02 | ToEntity() method | PASS (pre-existing) | entity.go:64 `ToEntity` on Model |
| DOM-03 | Make(Entity) | PASS (pre-existing) | entity.go (Make used at processor.go list/get) |
| DOM-04 | Transform function | PASS (pre-existing) | rest.go TransformDetail/TransformSlice |
| DOM-05 | TransformSlice used in list | PASS | resource.go:130 `rest := TransformSlice(models, enrichments)` — no inline transform loop |
| DOM-06 | Processor accepts FieldLogger | PASS | processor.go constructor uses logrus.FieldLogger (unchanged) |
| DOM-07 | Handlers pass d.Logger() | PASS | resource.go:102 `NewProcessor(d.Logger(), r.Context(), db)`; no StandardLogger |
| DOM-08 | POST/PATCH use RegisterInputHandler | PASS (unaffected) | resource.go:24-27,33,35 RegisterInputHandler for create/update/parse/restore. listHandler is GET (RegisterHandler), correct |
| DOM-09 | Transform errors handled | PASS | resource.go:136-147 MarshalWithURLs/json.Unmarshal errors checked & logged |
| DOM-10 | Test DB registers tenant callbacks | PASS | processor_integration_test.go:24 `database.RegisterTenantCallbacks(l, db)` |
| DOM-11 | Providers lazy / read-only | PARTIAL/ACCEPTED | getByID/getDeletedByID use database.Query (lazy). getAll and getRecipeUsageFromPlanItems are eager imperative funcs — pre-existing style, not introduced by this change; both are read-only |
| DOM-12 | No os.Getenv in handlers | PASS | grep: 0 matches in resource.go/provider.go/processor.go |
| DOM-13 | No cross-domain logic in handlers | PASS | listHandler only orchestrates recipe processor methods |
| DOM-14 | Handlers don't call providers directly | PASS | resource.go calls only proc.* methods |
| DOM-15 | No direct entity create/save/delete in handlers | PASS | grep: 0 db.Create/Save/Delete in resource.go (only in *_test fixtures) |
| DOM-16 | administrator.go for writes | PASS (pre-existing, unaffected) | internal/recipe/administrator.go present; no writes added by diff |
| DOM-17 | Error → HTTP status mapping | PASS (unaffected) | list errors → 500; create/update/restore map 400/404/410/422 |
| DOM-18 | JSON:API interface on REST models | PASS (pre-existing) | rest.go RestModel methods |
| DOM-19 | Flat request models | PASS (pre-existing) | CreateRequest/UpdateRequest flat |
| DOM-20 | Table-driven tests | PASS | usage_sort_test.go cases table; frequency_sort tests structured |

## Security Review (SEC) — multi-tenancy / SQL focus

| ID | Check | Status | Evidence |
|----|-------|--------|----------|
| SEC-tenant-recipes | Recipe list rows scoped to caller tenant in frequency-sort path | PASS | Outer query is `db.Model(&Entity{})` (provider.go:72) so Statement.Model=*Entity → tenant callback injects `tenant_id = ?` (tenant_callbacks.go:39-46 checks Model first). Confirmed empirically: an injected cross-tenant isolation probe returned only the caller's recipe. recipeWithUsage embeds Entity but the join keeps tenant scoping on `recipes`. |
| SEC-count-scope | Cook count scoped to tenant AND household | PASS | provider.go:140 `Where("pw.tenant_id = ? AND pw.household_id = ?", filters.TenantID, filters.HouseholdID)`; legacy path provider.go:210 same. Tenant/household sourced from context: resource.go:82,90-91 `t := tenantctx.MustFromContext(...)`. TestGetRecipeUsageScopedToHousehold proves other-household (3) and other-tenant (4) plan items are excluded; count = 2. |
| SEC-raw-table-callback | Raw db.Table aggregation does not silently bypass scoping | PASS | plan_items/plan_weeks have no Model-level TenantId on the subquery, so the callback is a no-op there (tenant_callbacks.go:60-93) — which is exactly why the join scopes plan_weeks explicitly. This matches the documented raw-query exception in anti-patterns.md. |
| SEC-injection | No untrusted input in raw SQL / ORDER BY | PASS | ORDER BY built via fmt.Sprintf (provider.go:148) but `dir` is from controlled enum ("ASC"/"DESC" only) and all column names are literals; user `sort` is mapped through parseUsageSort to an enum (provider.go:46-55). All WHERE values are bound parameters. |
| SEC-ambiguous-cols | No ambiguous columns in joined query | PASS | Only `recipes` has title/id/deleted_at/tenant_id; derived `u` exposes only recipe_id/usage_count/last_used_day. Order qualifies `recipes.title`/`recipes.id`. Confirmed: frequency-sort path executes against SQLite in tests without ambiguity errors. |
| SEC-secrets | No hardcoded secrets | PASS | none introduced |

## Summary

### Blocking (must fix)
- None.

### Non-Blocking (observations only)
- DOM-11: `getAll` / `getRecipeUsageFromPlanItems` are eager imperative read
  functions rather than `database.SliceQuery`-style lazy providers. This is the
  pre-existing style of this file and is not introduced by task-053; both are
  read-only and tenant/household-scoped. Not a regression.
- The recipe **list itself** is tenant-scoped but not household-scoped (the
  tenant callback ignores household). A recipe shared within a tenant can show
  `usageCount: 0` to a non-owning household. This is the explicit, documented
  FR-17 decision (context.md:36) and strictly improves on the prior unscoped
  aggregation — accepted as designed, not a finding.
- `last_used_day` is scanned into `*string` for a Postgres `date` column. This
  matches the pre-existing `recipeUsageRow` pattern; tests exercise SQLite only.
  Worth a one-time manual check against Postgres, but unchanged by this diff.

### Verdict
PASS — no Critical or Important issues. Build and tests green; multi-tenancy
and household scoping of the cook count are correct and test-proven; the raw
aggregation scopes plan_weeks explicitly; no SQL-injection or ambiguous-column
risk.
