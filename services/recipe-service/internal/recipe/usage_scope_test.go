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
