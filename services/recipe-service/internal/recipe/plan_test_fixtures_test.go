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
