package recipe

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/gorm"
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
