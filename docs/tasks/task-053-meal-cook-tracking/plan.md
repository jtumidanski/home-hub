# Meal Cook Tracking & Frequency Sort — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the existing read-only per-recipe cook count (number of `plan_items` referencing a recipe) into a sortable dimension on the recipe list endpoint, and expose ascending/descending frequency sort in the meal-planner recipe selector and the main Recipes page.

**Architecture:** The cook count stays **computed on the fly** — no persisted counter, no migration. The core backend change moves the `plan_items` aggregation out of the post-pagination enrichment and into the ordered, paginated list query via a `LEFT JOIN` on a grouped derived table (`COUNT(*)` + `MAX(day)`, `COALESCE`d to 0), scoped to the requesting tenant **and** household through a `plan_weeks` join. The frontend adds a shared sort `<Select>` to both surfaces and threads a `sort` query parameter through the existing React Query hook.

**Tech Stack:** Go (recipe-service, GORM, JSON:API via api2go), SQLite (in-memory test DB), React + TypeScript, TanStack React Query, shadcn/ui `Select`, Vitest + React Testing Library.

---

## File Structure

**Backend — `services/recipe-service/internal/recipe/`**
- `provider.go` — add `UsageSort` type + `parseUsageSort`; add `TenantID`/`HouseholdID`/`UsageSort` to `ListFilters`; add `recipeWithUsage` scan struct; rework `getAll` to return a usage map and support the frequency-sort join; re-scope `getRecipeUsageFromPlanItems` via a `plan_weeks` join.
- `processor.go` — change `Processor.List` to return the usage map; change `Processor.GetRecipeUsage` to take tenant/household.
- `resource.go` — parse the `sort` param, pass tenant/household into `ListFilters`, and merge sort-path usage into list enrichments regardless of `include_usage`.
- `plan_test_fixtures_test.go` *(new)* — minimal local `plan_weeks`/`plan_items` structs + seed helpers (avoids the `plan`→`recipe` import cycle).
- `usage_sort_test.go` *(new)* — unit test for `parseUsageSort`.
- `usage_scope_test.go` *(new)* — integration test for tenant/household scoping of `GetRecipeUsage`.
- `frequency_sort_test.go` *(new)* — integration tests for ordering, zero-count inclusion, tie-breaker stability across pages, filter composition, and default-order preservation.

**Frontend — `frontend/src/`**
- `types/models/recipe.ts` — add `RecipeSort` type (the `usageCount`/`lastUsedDate` attributes already exist).
- `services/api/recipe.ts` — add `sort` to `RecipeListParams` + `listRecipes` query building.
- `lib/hooks/api/use-recipes.ts` — add `sort` to `UseRecipesParams`.
- `components/features/recipes/recipe-sort-select.tsx` *(new)* — shared sort control used by both surfaces.
- `components/features/meals/recipe-selector.tsx` — add sort state + control.
- `pages/RecipesPage.tsx` — add sort state + control.
- `components/features/recipes/recipe-card.tsx` — add `cooked Nx` indicator when `usageCount > 0`.
- Test files alongside each (`__tests__/`).

---

## Backend

### Task 1: `UsageSort` type and `parseUsageSort` parser

**Files:**
- Modify: `services/recipe-service/internal/recipe/provider.go`
- Test: `services/recipe-service/internal/recipe/usage_sort_test.go` (create)

- [ ] **Step 1: Write the failing test**

Create `services/recipe-service/internal/recipe/usage_sort_test.go`:

```go
package recipe

import "testing"

func TestParseUsageSort(t *testing.T) {
	cases := []struct {
		in   string
		want UsageSort
	}{
		{"usageCount", UsageSortAsc},
		{"-usageCount", UsageSortDesc},
		{"", UsageSortNone},
		{"title", UsageSortNone},
		{"cookCount", UsageSortNone},
	}
	for _, c := range cases {
		if got := parseUsageSort(c.in); got != c.want {
			t.Errorf("parseUsageSort(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd services/recipe-service && go test ./internal/recipe/ -run TestParseUsageSort`
Expected: FAIL — compile error `undefined: UsageSort` / `undefined: parseUsageSort`.

- [ ] **Step 3: Add the type and parser**

In `services/recipe-service/internal/recipe/provider.go`, add directly above `type ListFilters struct`:

```go
// UsageSort selects whether and how the recipe list is ordered by cook frequency.
type UsageSort int

const (
	UsageSortNone UsageSort = iota // default order (created_at DESC)
	UsageSortAsc                   // least cooked first
	UsageSortDesc                  // most cooked first
)

// parseUsageSort maps the JSON:API `sort` query value to a UsageSort.
// Unknown values fall back to the default order (lenient, per PRD §5.3).
func parseUsageSort(v string) UsageSort {
	switch v {
	case "usageCount":
		return UsageSortAsc
	case "-usageCount":
		return UsageSortDesc
	default:
		return UsageSortNone
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd services/recipe-service && go test ./internal/recipe/ -run TestParseUsageSort`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add services/recipe-service/internal/recipe/provider.go services/recipe-service/internal/recipe/usage_sort_test.go
git commit -m "feat(recipe-service): add UsageSort type and sort param parser"
```

---

### Task 2: Plan-table test fixtures + scope `GetRecipeUsage` by household

This task re-scopes the existing `include_usage` aggregation to the requesting tenant **and** household via a `plan_weeks` join (FR-17), and introduces the shared test fixtures that later tasks reuse.

**Files:**
- Create: `services/recipe-service/internal/recipe/plan_test_fixtures_test.go`
- Modify: `services/recipe-service/internal/recipe/provider.go` (`getRecipeUsageFromPlanItems`)
- Modify: `services/recipe-service/internal/recipe/processor.go` (`GetRecipeUsage`)
- Modify: `services/recipe-service/internal/recipe/resource.go` (call site)
- Test: `services/recipe-service/internal/recipe/usage_scope_test.go` (create)

- [ ] **Step 1: Create the shared plan-table test fixtures**

Create `services/recipe-service/internal/recipe/plan_test_fixtures_test.go`. We redeclare minimal `plan_weeks`/`plan_items` structs here because importing the `plan` / `planitem` packages would create an import cycle (`plan` imports `recipe`):

```go
package recipe

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// testPlanWeek mirrors the columns of the plan_weeks table that the cook-count
// aggregation reads. It is redeclared locally (rather than imported from the
// plan package) to avoid an import cycle: plan imports recipe.
type testPlanWeek struct {
	Id          uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId    uuid.UUID `gorm:"type:uuid;not null"`
	HouseholdId uuid.UUID `gorm:"type:uuid;not null"`
}

func (testPlanWeek) TableName() string { return "plan_weeks" }

// testPlanItem mirrors the columns of the plan_items table that the cook-count
// aggregation reads.
type testPlanItem struct {
	Id         uuid.UUID `gorm:"type:uuid;primaryKey"`
	PlanWeekId uuid.UUID `gorm:"type:uuid;not null"`
	RecipeId   uuid.UUID `gorm:"type:uuid;not null"`
	Day        time.Time `gorm:"type:date;not null"`
}

func (testPlanItem) TableName() string { return "plan_items" }

func migratePlanTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&testPlanWeek{}, &testPlanItem{}); err != nil {
		t.Fatalf("failed to migrate plan tables: %v", err)
	}
}

func seedPlanWeek(t *testing.T, db *gorm.DB, tenantID, householdID uuid.UUID) uuid.UUID {
	t.Helper()
	pw := testPlanWeek{Id: uuid.New(), TenantId: tenantID, HouseholdId: householdID}
	if err := db.Create(&pw).Error; err != nil {
		t.Fatalf("failed to create plan week: %v", err)
	}
	return pw.Id
}

func seedPlanItem(t *testing.T, db *gorm.DB, planWeekID, recipeID uuid.UUID, day time.Time) {
	t.Helper()
	pi := testPlanItem{Id: uuid.New(), PlanWeekId: planWeekID, RecipeId: recipeID, Day: day}
	if err := db.Create(&pi).Error; err != nil {
		t.Fatalf("failed to create plan item: %v", err)
	}
}
```

- [ ] **Step 2: Write the failing scoping test**

Create `services/recipe-service/internal/recipe/usage_scope_test.go`:

```go
package recipe

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestGetRecipeUsageScopedToHousehold(t *testing.T) {
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()

	db := setupTestDB(t)
	migratePlanTables(t, db)

	ctx := tenantctx.WithContext(context.Background(), tenantctx.New(tenantID, householdID, userID))
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, ctx, db)

	created, _, err := p.Create(tenantID, householdID, CreateAttrs{Title: "Soup", Source: "Boil water."})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	recipeID := created.Id()

	// Requesting household: 2 plan items (these count).
	pw := seedPlanWeek(t, db, tenantID, householdID)
	seedPlanItem(t, db, pw, recipeID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))
	seedPlanItem(t, db, pw, recipeID, time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC))

	// Same tenant, DIFFERENT household: 3 plan items (must NOT count).
	pwOtherHH := seedPlanWeek(t, db, tenantID, uuid.New())
	for _, d := range []int{2, 9, 16} {
		seedPlanItem(t, db, pwOtherHH, recipeID, time.Date(2026, 5, d, 0, 0, 0, 0, time.UTC))
	}

	// DIFFERENT tenant: 4 plan items (must NOT count).
	pwOtherTenant := seedPlanWeek(t, db, uuid.New(), uuid.New())
	for _, d := range []int{3, 10, 17, 24} {
		seedPlanItem(t, db, pwOtherTenant, recipeID, time.Date(2026, 5, d, 0, 0, 0, 0, time.UTC))
	}

	usage := p.GetRecipeUsage([]uuid.UUID{recipeID}, tenantID, householdID)
	got, ok := usage[recipeID]
	if !ok {
		t.Fatalf("expected usage entry for recipe %s", recipeID)
	}
	if got.usageCount != 2 {
		t.Fatalf("expected usageCount 2 (requesting household only), got %d", got.usageCount)
	}
	if got.lastUsedDay == nil || *got.lastUsedDay == "" {
		t.Fatalf("expected lastUsedDay to be set, got %v", got.lastUsedDay)
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd services/recipe-service && go test ./internal/recipe/ -run TestGetRecipeUsageScopedToHousehold`
Expected: FAIL — compile error: `GetRecipeUsage` is called with 3 args but is defined with 1.

- [ ] **Step 4: Re-scope `getRecipeUsageFromPlanItems`**

In `services/recipe-service/internal/recipe/provider.go`, replace the whole `getRecipeUsageFromPlanItems` function with:

```go
func getRecipeUsageFromPlanItems(db *gorm.DB, recipeIDs []uuid.UUID, tenantID, householdID uuid.UUID) map[uuid.UUID]recipeUsageResult {
	if len(recipeIDs) == 0 {
		return nil
	}
	var rows []recipeUsageRow
	db.Table("plan_items AS pi").
		Select("pi.recipe_id AS recipe_id, MAX(pi.day) AS last_used_day, COUNT(*) AS usage_count").
		Joins("JOIN plan_weeks AS pw ON pw.id = pi.plan_week_id").
		Where("pw.tenant_id = ? AND pw.household_id = ? AND pi.recipe_id IN ?", tenantID, householdID, recipeIDs).
		Group("pi.recipe_id").
		Find(&rows)
	result := make(map[uuid.UUID]recipeUsageResult, len(rows))
	for _, r := range rows {
		result[r.RecipeID] = recipeUsageResult{
			recipeID:    r.RecipeID,
			lastUsedDay: r.LastUsedDay,
			usageCount:  r.UsageCount,
		}
	}
	return result
}
```

- [ ] **Step 5: Update `GetRecipeUsage` signature**

In `services/recipe-service/internal/recipe/processor.go`, replace the `GetRecipeUsage` method with:

```go
func (p *Processor) GetRecipeUsage(recipeIDs []uuid.UUID, tenantID, householdID uuid.UUID) map[uuid.UUID]recipeUsageResult {
	return getRecipeUsageFromPlanItems(p.db.WithContext(p.ctx), recipeIDs, tenantID, householdID)
}
```

- [ ] **Step 6: Update the handler call site**

In `services/recipe-service/internal/recipe/resource.go`, inside `listHandler`'s returned function, add a tenant lookup at the top of the handler (just before `filters := ListFilters{`):

```go
				t := tenantctx.MustFromContext(r.Context())
```

Then update the `include_usage` block so the `GetRecipeUsage` call passes tenant/household:

```go
				if includeUsage && len(models) > 0 {
					recipeIDs := make([]uuid.UUID, len(models))
					for i, m := range models {
						recipeIDs[i] = m.Id()
					}
					usageMap = proc.GetRecipeUsage(recipeIDs, t.Id(), t.HouseholdId())
				}
```

(`tenantctx` is already imported in `resource.go`.)

- [ ] **Step 7: Run the scoping test and the full package**

Run: `cd services/recipe-service && go test ./internal/recipe/ -run TestGetRecipeUsageScopedToHousehold && go test ./internal/recipe/`
Expected: PASS (all tests).

- [ ] **Step 8: Commit**

```bash
git add services/recipe-service/internal/recipe/provider.go services/recipe-service/internal/recipe/processor.go services/recipe-service/internal/recipe/resource.go services/recipe-service/internal/recipe/plan_test_fixtures_test.go services/recipe-service/internal/recipe/usage_scope_test.go
git commit -m "feat(recipe-service): scope cook-count aggregation to tenant and household"
```

---

### Task 3: Frequency-sorted, paginated list query

Moves the aggregation into the ordered, paginated query (FR-5). Because `getAll`, `Processor.List`, and `listHandler` form one compile unit, they change together so the package always builds.

**Files:**
- Modify: `services/recipe-service/internal/recipe/provider.go` (`ListFilters`, `recipeWithUsage`, `getAll`)
- Modify: `services/recipe-service/internal/recipe/processor.go` (`Processor.List`)
- Modify: `services/recipe-service/internal/recipe/resource.go` (parse `sort`, pass tenant/household, merge usage)
- Test: `services/recipe-service/internal/recipe/frequency_sort_test.go` (create)

- [ ] **Step 1: Write the failing frequency-sort tests**

Create `services/recipe-service/internal/recipe/frequency_sort_test.go`:

```go
package recipe

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus/hooks/test"
)

func newFreqProcessor(t *testing.T) (*Processor, *gorm.DB, uuid.UUID, uuid.UUID) {
	t.Helper()
	tenantID := uuid.New()
	householdID := uuid.New()
	db := setupTestDB(t)
	migratePlanTables(t, db)
	ctx := tenantctx.WithContext(context.Background(), tenantctx.New(tenantID, householdID, uuid.New()))
	l, _ := test.NewNullLogger()
	return NewProcessor(l, ctx, db), db, tenantID, householdID
}

func mkRecipe(t *testing.T, p *Processor, tenantID, householdID uuid.UUID, title string) uuid.UUID {
	t.Helper()
	c, _, err := p.Create(tenantID, householdID, CreateAttrs{Title: title, Source: "Boil water."})
	if err != nil {
		t.Fatalf("create %s: %v", title, err)
	}
	return c.Id()
}

func ids(models []Model) []uuid.UUID {
	out := make([]uuid.UUID, len(models))
	for i, m := range models {
		out[i] = m.Id()
	}
	return out
}

func TestListFrequencySortOrder(t *testing.T) {
	p, db, tenantID, householdID := newFreqProcessor(t)

	apple := mkRecipe(t, p, tenantID, householdID, "Apple")   // 3 cooks
	banana := mkRecipe(t, p, tenantID, householdID, "Banana") // 1 cook
	cherry := mkRecipe(t, p, tenantID, householdID, "Cherry") // 0 cooks

	pw := seedPlanWeek(t, db, tenantID, householdID)
	for i := 0; i < 3; i++ {
		seedPlanItem(t, db, pw, apple, time.Date(2026, 5, 1+i, 0, 0, 0, 0, time.UTC))
	}
	seedPlanItem(t, db, pw, banana, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))

	// Descending: Apple(3), Banana(1), Cherry(0)
	desc, usageMap, total, err := p.List(ListFilters{Page: 1, PageSize: 20, TenantID: tenantID, HouseholdID: householdID, UsageSort: UsageSortDesc})
	if err != nil {
		t.Fatalf("list desc: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
	wantDesc := []uuid.UUID{apple, banana, cherry}
	if got := ids(desc); !equalIDs(got, wantDesc) {
		t.Fatalf("desc order = %v, want %v", got, wantDesc)
	}
	if usageMap[apple].usageCount != 3 || usageMap[banana].usageCount != 1 {
		t.Fatalf("unexpected usage counts: %+v", usageMap)
	}
	// Zero-count recipe is present in the usage map with count 0.
	if c, ok := usageMap[cherry]; !ok || c.usageCount != 0 {
		t.Fatalf("expected cherry usageCount 0 present, got %+v ok=%v", c, ok)
	}

	// Ascending: Cherry(0), Banana(1), Apple(3)
	asc, _, _, err := p.List(ListFilters{Page: 1, PageSize: 20, TenantID: tenantID, HouseholdID: householdID, UsageSort: UsageSortAsc})
	if err != nil {
		t.Fatalf("list asc: %v", err)
	}
	wantAsc := []uuid.UUID{cherry, banana, apple}
	if got := ids(asc); !equalIDs(got, wantAsc) {
		t.Fatalf("asc order = %v, want %v", got, wantAsc)
	}
}

func TestListFrequencySortTieBreakerAcrossPages(t *testing.T) {
	p, _, tenantID, householdID := newFreqProcessor(t)

	// Four never-scheduled recipes (all count 0); created in non-alphabetical order.
	mkRecipe(t, p, tenantID, householdID, "Delta")
	mkRecipe(t, p, tenantID, householdID, "Charlie")
	mkRecipe(t, p, tenantID, householdID, "Bravo")
	mkRecipe(t, p, tenantID, householdID, "Alpha")

	// Ascending sort: all tied at 0, so tie-broken by title ASC -> Alpha, Bravo, Charlie, Delta.
	page1, _, _, err := p.List(ListFilters{Page: 1, PageSize: 2, TenantID: tenantID, HouseholdID: householdID, UsageSort: UsageSortAsc})
	if err != nil {
		t.Fatalf("page1: %v", err)
	}
	page2, _, _, err := p.List(ListFilters{Page: 2, PageSize: 2, TenantID: tenantID, HouseholdID: householdID, UsageSort: UsageSortAsc})
	if err != nil {
		t.Fatalf("page2: %v", err)
	}
	titles := []string{page1[0].Title(), page1[1].Title(), page2[0].Title(), page2[1].Title()}
	want := []string{"Alpha", "Bravo", "Charlie", "Delta"}
	for i := range want {
		if titles[i] != want[i] {
			t.Fatalf("paged titles = %v, want %v", titles, want)
		}
	}
}

func TestListFrequencySortComposesWithSearch(t *testing.T) {
	p, db, tenantID, householdID := newFreqProcessor(t)

	applePie := mkRecipe(t, p, tenantID, householdID, "Apple Pie")   // matches "pie", 2 cooks
	peachPie := mkRecipe(t, p, tenantID, householdID, "Peach Pie")   // matches "pie", 1 cook
	mkRecipe(t, p, tenantID, householdID, "Garden Salad")            // does NOT match "pie"

	pw := seedPlanWeek(t, db, tenantID, householdID)
	for i := 0; i < 2; i++ {
		seedPlanItem(t, db, pw, applePie, time.Date(2026, 5, 1+i, 0, 0, 0, 0, time.UTC))
	}
	seedPlanItem(t, db, pw, peachPie, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))

	models, _, total, err := p.List(ListFilters{Search: "pie", Page: 1, PageSize: 20, TenantID: tenantID, HouseholdID: householdID, UsageSort: UsageSortDesc})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2 (search-filtered), got %d", total)
	}
	want := []uuid.UUID{applePie, peachPie}
	if got := ids(models); !equalIDs(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestListDefaultOrderUnchanged(t *testing.T) {
	p, _, tenantID, householdID := newFreqProcessor(t)

	mkRecipe(t, p, tenantID, householdID, "First")
	mkRecipe(t, p, tenantID, householdID, "Second")

	// No UsageSort -> default path, usageMap must be nil.
	models, usageMap, total, err := p.List(ListFilters{Page: 1, PageSize: 20, TenantID: tenantID, HouseholdID: householdID})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if usageMap != nil {
		t.Fatalf("expected nil usageMap on default path, got %v", usageMap)
	}
	// Default order is created_at DESC -> most recently created ("Second") first.
	if models[0].Title() != "Second" {
		t.Fatalf("expected default order Second first, got %s", models[0].Title())
	}
}

func equalIDs(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
```

> Note: this file references `gorm.DB` in `newFreqProcessor`. Add the import `"gorm.io/gorm"` to the import block.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd services/recipe-service && go test ./internal/recipe/ -run TestListFrequencySort`
Expected: FAIL — compile errors: `ListFilters` has no field `TenantID`/`HouseholdID`/`UsageSort`; `p.List` returns 3 values, not 4.

- [ ] **Step 3: Extend `ListFilters` and add the scan struct**

In `services/recipe-service/internal/recipe/provider.go`, replace the `ListFilters` struct with:

```go
type ListFilters struct {
	Search              string
	Tags                []string
	Page                int
	PageSize            int
	PlannerReady        *bool
	Classification      string
	NormalizationStatus string
	TenantID            uuid.UUID
	HouseholdID         uuid.UUID
	UsageSort           UsageSort
}
```

Then add the scan struct directly below `type recipeUsageRow struct { ... }`:

```go
// recipeWithUsage scans a recipe row joined with its cook-count aggregate.
// It embeds Entity so Preload("Tags") still resolves via recipes.id.
type recipeWithUsage struct {
	Entity
	UsageCount  int64   `gorm:"column:usage_count"`
	LastUsedDay *string `gorm:"column:last_used_day"`
}
```

- [ ] **Step 4: Add `fmt` import and rework `getAll`**

In `services/recipe-service/internal/recipe/provider.go`, add `"fmt"` to the import block. Then replace the entire `getAll` function (the closure that currently returns `([]Entity, int64, error)`) with:

```go
func getAll(filters ListFilters) func(db *gorm.DB) ([]Entity, map[uuid.UUID]recipeUsageResult, int64, error) {
	return func(db *gorm.DB) ([]Entity, map[uuid.UUID]recipeUsageResult, int64, error) {
		query := db.Model(&Entity{}).Where("deleted_at IS NULL")

		if filters.Search != "" {
			query = query.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(filters.Search)+"%")
		}

		if len(filters.Tags) > 0 {
			for _, tag := range filters.Tags {
				query = query.Where("id IN (?)",
					db.Model(&TagEntity{}).Select("recipe_id").Where("tag = ?", strings.ToLower(strings.TrimSpace(tag))),
				)
			}
		}

		if filters.Classification != "" {
			query = query.Where("id IN (?)",
				db.Model(&TagEntity{}).Select("recipe_id").Where("tag = ?", strings.ToLower(strings.TrimSpace(filters.Classification))),
			)
		}

		if filters.PlannerReady != nil {
			if *filters.PlannerReady {
				query = query.Where("id IN (?)",
					db.Table("recipe_planner_configs").Select("recipe_id").Where("classification IS NOT NULL AND classification != ''"),
				).Where("servings IS NOT NULL")
			} else {
				query = query.Where("(id NOT IN (?) OR servings IS NULL)",
					db.Table("recipe_planner_configs").Select("recipe_id").Where("classification IS NOT NULL AND classification != ''"),
				)
			}
		}

		if filters.NormalizationStatus == "complete" {
			query = query.Where("id IN (?)",
				db.Table("recipe_ingredients").Select("DISTINCT recipe_id"),
			).Where("id NOT IN (?)",
				db.Table("recipe_ingredients").Select("DISTINCT recipe_id").Where("normalization_status = 'unresolved'"),
			)
		} else if filters.NormalizationStatus == "incomplete" {
			query = query.Where("id IN (?)",
				db.Table("recipe_ingredients").Select("DISTINCT recipe_id").Where("normalization_status = 'unresolved'"),
			)
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			return nil, nil, 0, err
		}

		offset := (filters.Page - 1) * filters.PageSize

		if filters.UsageSort == UsageSortNone {
			// Existing default path — unchanged behavior, no join, nil usage map.
			var entities []Entity
			err := query.Preload("Tags").
				Order("created_at DESC").
				Offset(offset).
				Limit(filters.PageSize).
				Find(&entities).Error
			return entities, nil, total, err
		}

		// Frequency-sort path: aggregate plan_items once (scoped to the
		// requesting tenant + household via plan_weeks), LEFT JOIN it 1:1, and
		// order before LIMIT/OFFSET.
		usageSub := db.Table("plan_items AS pi").
			Select("pi.recipe_id AS recipe_id, COUNT(*) AS usage_count, MAX(pi.day) AS last_used_day").
			Joins("JOIN plan_weeks AS pw ON pw.id = pi.plan_week_id").
			Where("pw.tenant_id = ? AND pw.household_id = ?", filters.TenantID, filters.HouseholdID).
			Group("pi.recipe_id")

		dir := "ASC"
		if filters.UsageSort == UsageSortDesc {
			dir = "DESC"
		}
		// Deterministic tie-breaker (FR-7): equal counts ordered by title then id.
		order := fmt.Sprintf("COALESCE(u.usage_count, 0) %s, recipes.title ASC, recipes.id ASC", dir)

		var rows []recipeWithUsage
		err := query.
			Joins("LEFT JOIN (?) AS u ON u.recipe_id = recipes.id", usageSub).
			Select("recipes.*, COALESCE(u.usage_count, 0) AS usage_count, u.last_used_day").
			Preload("Tags").
			Order(order).
			Offset(offset).
			Limit(filters.PageSize).
			Find(&rows).Error
		if err != nil {
			return nil, nil, 0, err
		}

		entities := make([]Entity, len(rows))
		usageMap := make(map[uuid.UUID]recipeUsageResult, len(rows))
		for i, row := range rows {
			entities[i] = row.Entity
			usageMap[row.Entity.Id] = recipeUsageResult{
				recipeID:    row.Entity.Id,
				lastUsedDay: row.LastUsedDay,
				usageCount:  row.UsageCount,
			}
		}
		return entities, usageMap, total, nil
	}
}
```

- [ ] **Step 5: Update `Processor.List`**

In `services/recipe-service/internal/recipe/processor.go`, replace the `List` method with:

```go
func (p *Processor) List(filters ListFilters) ([]Model, map[uuid.UUID]recipeUsageResult, int64, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 20
	}

	entities, usageMap, total, err := getAll(filters)(p.db.WithContext(p.ctx))
	if err != nil {
		return nil, nil, 0, err
	}

	models := make([]Model, 0, len(entities))
	for _, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, nil, 0, err
		}
		models = append(models, m)
	}
	return models, usageMap, total, nil
}
```

- [ ] **Step 6: Update `listHandler` to parse `sort` and merge usage**

In `services/recipe-service/internal/recipe/resource.go`, update `listHandler`. The `t := tenantctx.MustFromContext(r.Context())` line was added in Task 2; ensure it sits above the `filters` literal. Set the new filter fields:

```go
				filters := ListFilters{
					Search:              r.URL.Query().Get("search"),
					Tags:                r.URL.Query()["tag"],
					Page:                queryInt(r, "page[number]", 1),
					PageSize:            queryInt(r, "page[size]", 20),
					Classification:      r.URL.Query().Get("classification"),
					NormalizationStatus: r.URL.Query().Get("normalizationStatus"),
					TenantID:            t.Id(),
					HouseholdID:         t.HouseholdId(),
					UsageSort:           parseUsageSort(r.URL.Query().Get("sort")),
				}
```

Update the `proc.List` call and the usage-merge block to:

```go
				models, usageMap, total, err := proc.List(filters)
				if err != nil {
					d.Logger().WithError(err).Error("Failed to list recipes")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				// Frequency sort auto-populates usageMap. Otherwise honor include_usage.
				includeUsage := r.URL.Query().Get("include_usage") == "true"
				if usageMap == nil && includeUsage && len(models) > 0 {
					recipeIDs := make([]uuid.UUID, len(models))
					for i, m := range models {
						recipeIDs[i] = m.Id()
					}
					usageMap = proc.GetRecipeUsage(recipeIDs, t.Id(), t.HouseholdId())
				}

				enrichments := proc.BuildListEnrichments(models)
				if usageMap != nil {
					for i, m := range models {
						if usage, ok := usageMap[m.Id()]; ok {
							enrichments[i].LastUsedDate = usage.lastUsedDay
							enrichments[i].UsageCount = usage.usageCount
						}
					}
				}
```

(Leave the rest of `listHandler` — `TransformSlice`, marshaling, meta — unchanged.)

> **Implementation guard (per design §4.3):** the tenant callback injects a bare `tenant_id = ?`; in the joined query only `recipes` has `tenant_id` and only the derived table `u` exposes `usage_count`/`last_used_day`/`recipe_id`, so no column is ambiguous. If any DB engine complains about an ambiguous column when running the tests, qualify the offending reference (e.g. `recipes.tenant_id`). The SQLite test run in Step 7 is the first place this would surface.

- [ ] **Step 7: Run the frequency-sort tests and the full package**

Run: `cd services/recipe-service && go test ./internal/recipe/`
Expected: PASS — all tests including `TestListFrequencySortOrder`, `TestListFrequencySortTieBreakerAcrossPages`, `TestListFrequencySortComposesWithSearch`, `TestListDefaultOrderUnchanged`, and the pre-existing suite.

- [ ] **Step 8: Build the whole service**

Run: `cd services/recipe-service && go build ./...`
Expected: no output (success).

- [ ] **Step 9: Commit**

```bash
git add services/recipe-service/internal/recipe/provider.go services/recipe-service/internal/recipe/processor.go services/recipe-service/internal/recipe/resource.go services/recipe-service/internal/recipe/frequency_sort_test.go
git commit -m "feat(recipe-service): order recipe list by cook frequency before pagination"
```

---

### Task 4: Verify the recipe-service Docker build

No shared-library changes are expected (the design only reads `shared/go/database`), so only `recipe-service` needs a Docker build check.

**Files:** none modified.

- [ ] **Step 1: Confirm no shared-library files changed**

Run: `git diff --name-only main...HEAD -- shared/`
Expected: empty output. If any `shared/` file changed, a cross-service rebuild is required — flag it.

- [ ] **Step 2: Build the recipe-service Docker image**

Run (from the worktree root): `docker build -f services/recipe-service/Dockerfile -t recipe-service:task-053 .`
Expected: build succeeds. If the Dockerfile path or build context differs, follow `scripts/local-up.sh` for the canonical build invocation.

- [ ] **Step 3: No commit needed** (verification only).

---

## Frontend

### Task 5: Add the `RecipeSort` type

**Files:**
- Modify: `frontend/src/types/models/recipe.ts`

- [ ] **Step 1: Add the type**

In `frontend/src/types/models/recipe.ts`, add directly above `export interface RecipeListAttributes {`:

```ts
export type RecipeSort = "usageCount" | "-usageCount";
```

- [ ] **Step 2: Type-check**

Run: `cd frontend && npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/types/models/recipe.ts
git commit -m "feat(frontend): add RecipeSort type"
```

---

### Task 6: Thread `sort` through the recipe service

**Files:**
- Modify: `frontend/src/services/api/recipe.ts`
- Test: `frontend/src/services/api/__tests__/recipe.test.ts` (create)

- [ ] **Step 1: Write the failing service test**

Create `frontend/src/services/api/__tests__/recipe.test.ts`:

```ts
import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/api/client", () => ({
  api: {
    setTenant: vi.fn(),
    get: vi.fn().mockResolvedValue({ data: [], meta: { total: 0, page: 1, pageSize: 20 } }),
  },
}));

const tenant = { id: "tenant-1", type: "tenants" as const, attributes: { name: "T", createdAt: "", updatedAt: "" } };

describe("recipeService.listRecipes", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("sends sort=-usageCount when sort is descending", async () => {
    const { api } = await import("@/lib/api/client");
    const { recipeService } = await import("../recipe");
    await recipeService.listRecipes(tenant, { sort: "-usageCount" });
    const url = (api.get as ReturnType<typeof vi.fn>).mock.calls[0]![0] as string;
    expect(url).toContain("sort=-usageCount");
  });

  it("sends sort=usageCount when sort is ascending", async () => {
    const { api } = await import("@/lib/api/client");
    const { recipeService } = await import("../recipe");
    await recipeService.listRecipes(tenant, { sort: "usageCount" });
    const url = (api.get as ReturnType<typeof vi.fn>).mock.calls[0]![0] as string;
    expect(url).toContain("sort=usageCount");
  });

  it("omits sort when not provided", async () => {
    const { api } = await import("@/lib/api/client");
    const { recipeService } = await import("../recipe");
    await recipeService.listRecipes(tenant, {});
    const url = (api.get as ReturnType<typeof vi.fn>).mock.calls[0]![0] as string;
    expect(url).not.toContain("sort=");
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && npx vitest run src/services/api/__tests__/recipe.test.ts`
Expected: FAIL — `sort=-usageCount` is not in the URL (and TS error: `sort` not in `RecipeListParams`).

- [ ] **Step 3: Add `sort` to params and query building**

In `frontend/src/services/api/recipe.ts`:

1. Add `RecipeSort` to the type import:

```ts
import type {
  RecipeListItem,
  RecipeDetail,
  RecipeCreateAttributes,
  RecipeUpdateAttributes,
  RecipeTag,
  RecipeParseResult,
  RecipeIngredient,
  RecipeSort,
} from "@/types/models/recipe";
```

2. Add `sort` to `RecipeListParams`:

```ts
interface RecipeListParams {
  search?: string | undefined;
  tags?: string[] | undefined;
  page?: number | undefined;
  pageSize?: number | undefined;
  plannerReady?: boolean | undefined;
  classification?: string | undefined;
  normalizationStatus?: string | undefined;
  sort?: RecipeSort | undefined;
}
```

3. In `listRecipes`, add the query line just after the `normalizationStatus` line (do **not** send `include_usage` — frequency sort auto-populates `usageCount`):

```ts
    if (params?.sort) query.set("sort", params.sort);
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd frontend && npx vitest run src/services/api/__tests__/recipe.test.ts`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/services/api/recipe.ts frontend/src/services/api/__tests__/recipe.test.ts
git commit -m "feat(frontend): thread sort param through recipe list service"
```

---

### Task 7: Add `sort` to the `useRecipes` hook params

**Files:**
- Modify: `frontend/src/lib/hooks/api/use-recipes.ts`

- [ ] **Step 1: Add the type import and param**

In `frontend/src/lib/hooks/api/use-recipes.ts`:

1. Extend the existing recipe-types import to include `RecipeSort`:

```ts
import type { RecipeCreateAttributes, RecipeUpdateAttributes, RecipeSort } from "@/types/models/recipe";
```

2. Add `sort` to `UseRecipesParams`:

```ts
interface UseRecipesParams {
  search?: string | undefined;
  tags?: string[] | undefined;
  page?: number | undefined;
  pageSize?: number | undefined;
  plannerReady?: boolean | undefined;
  classification?: string | undefined;
  normalizationStatus?: string | undefined;
  sort?: RecipeSort | undefined;
}
```

(`useRecipes` already passes `params` to `recipeService.listRecipes` and spreads `params` into the query key, so changing `sort` produces a distinct cache entry and re-queries — no other change needed.)

- [ ] **Step 2: Type-check + existing hook tests**

Run: `cd frontend && npx tsc --noEmit && npx vitest run src/lib/hooks/api/__tests__/use-recipes.test.tsx`
Expected: no TS errors; existing hook tests PASS.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/hooks/api/use-recipes.ts
git commit -m "feat(frontend): accept sort param in useRecipes hook"
```

---

### Task 8: Shared `RecipeSortSelect` control

A single labeled `Select` with direction-explicit labels (FR-18), reused by both surfaces. Uses a `"default"` sentinel value (Radix `SelectItem` must not use an empty string).

**Files:**
- Create: `frontend/src/components/features/recipes/recipe-sort-select.tsx`
- Test: `frontend/src/components/features/recipes/__tests__/recipe-sort-select.test.tsx` (create)

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/features/recipes/__tests__/recipe-sort-select.test.tsx`:

```tsx
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { RecipeSortSelect } from "../recipe-sort-select";

describe("RecipeSortSelect", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls onChange with -usageCount when 'Most cooked' is picked", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<RecipeSortSelect value={undefined} onChange={onChange} />);

    await user.click(screen.getByRole("combobox", { name: /sort/i }));
    await user.click(await screen.findByRole("option", { name: /most cooked/i }));

    expect(onChange).toHaveBeenCalledWith("-usageCount");
  });

  it("calls onChange with usageCount when 'Least cooked' is picked", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<RecipeSortSelect value={undefined} onChange={onChange} />);

    await user.click(screen.getByRole("combobox", { name: /sort/i }));
    await user.click(await screen.findByRole("option", { name: /least cooked/i }));

    expect(onChange).toHaveBeenCalledWith("usageCount");
  });

  it("calls onChange with undefined when 'Default' is picked", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<RecipeSortSelect value="-usageCount" onChange={onChange} />);

    await user.click(screen.getByRole("combobox", { name: /sort/i }));
    await user.click(await screen.findByRole("option", { name: /default/i }));

    expect(onChange).toHaveBeenCalledWith(undefined);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && npx vitest run src/components/features/recipes/__tests__/recipe-sort-select.test.tsx`
Expected: FAIL — cannot resolve `../recipe-sort-select`.

- [ ] **Step 3: Create the component**

Create `frontend/src/components/features/recipes/recipe-sort-select.tsx`:

```tsx
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { RecipeSort } from "@/types/models/recipe";

const DEFAULT_VALUE = "default";

interface RecipeSortSelectProps {
  value: RecipeSort | undefined;
  onChange: (value: RecipeSort | undefined) => void;
  className?: string;
}

export function RecipeSortSelect({ value, onChange, className }: RecipeSortSelectProps) {
  return (
    <Select
      value={value ?? DEFAULT_VALUE}
      onValueChange={(v) => onChange(v === DEFAULT_VALUE ? undefined : (v as RecipeSort))}
    >
      <SelectTrigger className={className ?? "w-[140px]"} aria-label="Sort recipes">
        <SelectValue placeholder="Sort" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value={DEFAULT_VALUE}>Default</SelectItem>
        <SelectItem value="-usageCount">Most cooked</SelectItem>
        <SelectItem value="usageCount">Least cooked</SelectItem>
      </SelectContent>
    </Select>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd frontend && npx vitest run src/components/features/recipes/__tests__/recipe-sort-select.test.tsx`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/features/recipes/recipe-sort-select.tsx frontend/src/components/features/recipes/__tests__/recipe-sort-select.test.tsx
git commit -m "feat(frontend): add shared RecipeSortSelect control"
```

---

### Task 9: Wire the sort control into the recipe selector

**Files:**
- Modify: `frontend/src/components/features/meals/recipe-selector.tsx`
- Test: `frontend/src/components/features/meals/__tests__/recipe-selector.test.tsx` (create)

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/features/meals/__tests__/recipe-selector.test.tsx`:

```tsx
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const useRecipesMock = vi.fn();
vi.mock("@/lib/hooks/api/use-recipes", () => ({
  useRecipes: (params: unknown) => useRecipesMock(params),
}));

import { RecipeSelector } from "../recipe-selector";

describe("RecipeSelector sort control", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useRecipesMock.mockReturnValue({ data: { data: [] }, isLoading: false });
  });

  it("defaults to no sort param", () => {
    render(<RecipeSelector onSelectRecipe={vi.fn()} />);
    expect(useRecipesMock).toHaveBeenCalledWith(expect.objectContaining({ sort: undefined }));
  });

  it("passes sort=-usageCount after picking 'Most cooked'", async () => {
    const user = userEvent.setup();
    render(<RecipeSelector onSelectRecipe={vi.fn()} />);

    await user.click(screen.getByRole("combobox", { name: /sort/i }));
    await user.click(await screen.findByRole("option", { name: /most cooked/i }));

    expect(useRecipesMock).toHaveBeenLastCalledWith(expect.objectContaining({ sort: "-usageCount" }));
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && npx vitest run src/components/features/meals/__tests__/recipe-selector.test.tsx`
Expected: FAIL — no `combobox` named `sort`; `sort` not passed to `useRecipes`.

- [ ] **Step 3: Add sort state, control, and thread it into the query**

In `frontend/src/components/features/meals/recipe-selector.tsx`:

1. Add imports:

```tsx
import { RecipeSortSelect } from "@/components/features/recipes/recipe-sort-select";
import type { RecipeListItem, RecipeSort } from "@/types/models/recipe";
```

(Remove the now-duplicated `import type { RecipeListItem } from "@/types/models/recipe";` line — fold it into the line above.)

2. Add state next to the existing `classification` state:

```tsx
  const [sort, setSort] = useState<RecipeSort | undefined>(undefined);
```

3. Pass `sort` into the `useRecipes` call:

```tsx
  const { data, isLoading } = useRecipes({
    search: search || undefined,
    classification: classification || undefined,
    plannerReady: true,
    pageSize: 50,
    sort,
  });
```

4. Render the control inside the filter row, immediately after the classification `</Select>` (before the closing `</div>` of the `flex gap-2` row):

```tsx
        <RecipeSortSelect value={sort} onChange={setSort} className="w-[130px]" />
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd frontend && npx vitest run src/components/features/meals/__tests__/recipe-selector.test.tsx`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/features/meals/recipe-selector.tsx frontend/src/components/features/meals/__tests__/recipe-selector.test.tsx
git commit -m "feat(frontend): add cook-frequency sort to meal-planner recipe selector"
```

---

### Task 10: Wire the sort control into the Recipes page

**Files:**
- Modify: `frontend/src/pages/RecipesPage.tsx`
- Test: `frontend/src/pages/__tests__/RecipesPage.test.tsx` (create)

- [ ] **Step 1: Write the failing test**

Create `frontend/src/pages/__tests__/RecipesPage.test.tsx`:

```tsx
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";

const useRecipesMock = vi.fn();
vi.mock("@/lib/hooks/api/use-recipes", () => ({
  useRecipes: (params: unknown) => useRecipesMock(params),
  useRecipeTags: () => ({ data: { data: [] } }),
  useDeleteRecipe: () => ({ mutateAsync: vi.fn() }),
}));
vi.mock("react-router-dom", () => ({ useNavigate: () => vi.fn() }));
vi.mock("@/lib/hooks/use-mobile", () => ({ useMobile: () => false }));
vi.mock("@/components/common/pull-to-refresh", () => ({
  PullToRefresh: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

import { RecipesPage } from "../RecipesPage";

describe("RecipesPage sort control", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useRecipesMock.mockReturnValue({ data: { data: [] }, isLoading: false, refetch: vi.fn() });
  });

  it("passes sort=usageCount after picking 'Least cooked'", async () => {
    const user = userEvent.setup();
    render(<RecipesPage />);

    await user.click(screen.getByRole("combobox", { name: /sort/i }));
    await user.click(await screen.findByRole("option", { name: /least cooked/i }));

    expect(useRecipesMock).toHaveBeenLastCalledWith(expect.objectContaining({ sort: "usageCount" }));
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && npx vitest run src/pages/__tests__/RecipesPage.test.tsx`
Expected: FAIL — no `combobox` named `sort`.

- [ ] **Step 3: Add sort state and control**

In `frontend/src/pages/RecipesPage.tsx`:

1. Add imports:

```tsx
import { RecipeSortSelect } from "@/components/features/recipes/recipe-sort-select";
import type { RecipeListItem, RecipeSort } from "@/types/models/recipe";
```

(Fold the existing `import type { RecipeListItem } from "@/types/models/recipe";` into the line above.)

2. Add state next to the existing `plannerFilter` state:

```tsx
  const [sort, setSort] = useState<RecipeSort | undefined>(undefined);
```

3. Pass `sort` into `useRecipes`:

```tsx
  const { data, isLoading, refetch } = useRecipes({
    search: search || undefined,
    tags: selectedNonClassTags.length > 0 ? selectedNonClassTags : undefined,
    classification: selectedClassification,
    plannerReady: plannerFilter === "ready" ? true : plannerFilter === "not-ready" ? false : undefined,
    sort,
  });
```

4. Render the control as a top-level control next to the search box. Wrap the existing search `<div className="relative">…</div>` and the new control in a flex row so the sort sits beside search. Replace the search block's outer element:

```tsx
        {/* Search + sort */}
        <div className="flex gap-2">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search recipes..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-9"
            />
            {search && (
              <button
                type="button"
                onClick={() => setSearch("")}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            )}
          </div>
          <RecipeSortSelect value={sort} onChange={setSort} />
        </div>
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd frontend && npx vitest run src/pages/__tests__/RecipesPage.test.tsx`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/RecipesPage.tsx frontend/src/pages/__tests__/RecipesPage.test.tsx
git commit -m "feat(frontend): add cook-frequency sort to Recipes page"
```

---

### Task 11: Show the cook count on the recipe card

The list endpoint only returns `usageCount` when frequency sort is active, so a "show when present and > 0" indicator naturally appears exactly when sort is in effect (FR-15).

**Files:**
- Modify: `frontend/src/components/features/recipes/recipe-card.tsx`
- Test: `frontend/src/components/features/recipes/__tests__/recipe-card.test.tsx` (extend)

- [ ] **Step 1: Write the failing tests**

Add to `frontend/src/components/features/recipes/__tests__/recipe-card.test.tsx`, inside the existing `describe("RecipeCard", …)` block:

```tsx
  it("shows cooked count when usageCount is present and > 0", () => {
    render(<RecipeCard recipe={makeRecipe({ usageCount: 4 })} onDelete={vi.fn()} />);
    expect(screen.getByText(/cooked 4x/i)).toBeInTheDocument();
  });

  it("does not show cooked count when usageCount is 0", () => {
    render(<RecipeCard recipe={makeRecipe({ usageCount: 0 })} onDelete={vi.fn()} />);
    expect(screen.queryByText(/cooked/i)).not.toBeInTheDocument();
  });

  it("does not show cooked count when usageCount is absent", () => {
    render(<RecipeCard recipe={makeRecipe()} onDelete={vi.fn()} />);
    expect(screen.queryByText(/cooked/i)).not.toBeInTheDocument();
  });
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && npx vitest run src/components/features/recipes/__tests__/recipe-card.test.tsx`
Expected: FAIL — `cooked 4x` not found.

- [ ] **Step 3: Add the indicator**

In `frontend/src/components/features/recipes/recipe-card.tsx`, inside the metadata row (`<div className="flex flex-wrap items-center gap-2">`), add directly after the total-time `<span>…</span>` (before the `{allTags.map(...)}`):

```tsx
          {(attributes.usageCount ?? 0) > 0 && (
            <span className="text-xs text-muted-foreground">
              cooked {attributes.usageCount}x
            </span>
          )}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd frontend && npx vitest run src/components/features/recipes/__tests__/recipe-card.test.tsx`
Expected: PASS (existing + 3 new).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/features/recipes/recipe-card.tsx frontend/src/components/features/recipes/__tests__/recipe-card.test.tsx
git commit -m "feat(frontend): show cook count on recipe card when present"
```

---

### Task 12: Full frontend verification (including TZ=UTC)

**Files:** none modified.

- [ ] **Step 1: Type-check the whole frontend**

Run: `cd frontend && npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 2: Lint (if the project lints in CI)**

Run: `cd frontend && npm run lint`
Expected: no errors. (If `lint` is not a script, skip.)

- [ ] **Step 3: Run the full frontend test suite under TZ=UTC**

CI runs UTC while the dev machine is EDT, so any `lastUsedDate`/date assertions must pass under UTC.

Run: `cd frontend && TZ=UTC npx vitest run`
Expected: all tests PASS.

- [ ] **Step 4: Build the frontend**

Run: `cd frontend && npm run build`
Expected: build succeeds.

- [ ] **Step 5: No commit needed** (verification only).

---

## Final verification checklist (run before code review)

- [ ] `cd services/recipe-service && go build ./... && go test ./internal/recipe/` — all PASS.
- [ ] `cd frontend && npx tsc --noEmit && TZ=UTC npx vitest run && npm run build` — all PASS.
- [ ] Docker build for `recipe-service` succeeds (Task 4).
- [ ] `git diff --name-only main...HEAD -- shared/` is empty (no shared-library changes; if not, rebuild dependent services).

After all tasks pass, run the project's code-review step (`superpowers:requesting-code-review`) before opening a PR, per CLAUDE.md.

---

## Spec coverage map (plan self-review)

| Requirement | Task(s) |
| --- | --- |
| FR-1 cook count = COUNT(plan_items) | Task 2, Task 3 |
| FR-2 zero-count recipes sortable/included | Task 3 (`TestListFrequencySortOrder` cherry=0; `COALESCE`) |
| FR-3 computed at query time (no bookkeeping) | Task 2, Task 3 (no persisted counter) |
| FR-4 sort param both directions | Task 1, Task 3 |
| FR-5 order full set before pagination | Task 3 (`getAll` join + order before LIMIT) |
| FR-6 composes with filters | Task 3 (`TestListFrequencySortComposesWithSearch`) |
| FR-7 deterministic tie-breaker, stable across pages | Task 3 (`TestListFrequencySortTieBreakerAcrossPages`) |
| FR-8 default order unchanged | Task 3 (`TestListDefaultOrderUnchanged`) |
| FR-9 usageCount present on sort without include_usage | Task 3 (`getAll` returns usageMap; handler merges when `usageMap != nil`) |
| FR-10 selector sort control | Task 8, Task 9 |
| FR-11 selector "used Nx" matches sort value | Task 9 (existing render lights up; same usageMap source) |
| FR-12 selector re-queries full set | Task 7 (query key spreads params), Task 9 |
| FR-13 selector default unchanged | Task 9 (default `undefined` sort) |
| FR-14 recipes page sort control | Task 8, Task 10 |
| FR-15 recipes page surfaces count | Task 11 |
| FR-16 page sort composes with filters | Task 10 (sort added alongside existing filters) |
| FR-17 counts scoped to household | Task 2 (`TestGetRecipeUsageScopedToHousehold`), Task 3 (join scope) |
| FR-18 direction-explicit labels | Task 8 ("Most cooked" / "Least cooked") |
| Sort token `usageCount`/`-usageCount` | Task 1, Task 6 |
| No migration / no denormalized column | (none — verified by design; no entity/migration changes) |
| Backend + frontend builds & tests, TZ=UTC | Task 3, Task 12 |
| Docker build verified | Task 4 |

> **Handler-level HTTP test note:** FR-9's HTTP behavior (response carries `usageCount` on the sort path without `include_usage`) is exercised at the `Processor.List` level (Task 3 proves the usage map is produced) and wired into `listHandler` by code. The recipe package has no existing HTTP handler test harness, so no `httptest`-based handler test is added; the merge logic is a thin, reviewed glue layer over the tested `List` result.
