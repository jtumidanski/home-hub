package planitem

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func getByPlanWeekID(planWeekID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("plan_week_id = ?", planWeekID).Order("day ASC, position ASC")
	})
}

func getMaxPosition(db *gorm.DB, planWeekID uuid.UUID, day string, slot string) (int, error) {
	var maxPos *int
	err := db.Model(&Entity{}).
		Where("plan_week_id = ? AND day = ? AND slot = ?", planWeekID, day, slot).
		Select("MAX(position)").
		Scan(&maxPos).Error
	if err != nil {
		return 0, err
	}
	if maxPos == nil {
		return -1, nil
	}
	return *maxPos, nil
}

func countByPlanWeekID(db *gorm.DB, planWeekID uuid.UUID) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).Where("plan_week_id = ?", planWeekID).Count(&count).Error
	return count, err
}

type RecipeUsage struct {
	RecipeID    uuid.UUID
	LastUsedDay *string
	UsageCount  int64
}

func getRecipeUsage(db *gorm.DB, recipeIDs []uuid.UUID) ([]RecipeUsage, error) {
	if len(recipeIDs) == 0 {
		return nil, nil
	}
	var results []RecipeUsage
	err := db.Model(&Entity{}).
		Select("recipe_id, MAX(day) as last_used_day, COUNT(*) as usage_count").
		Where("recipe_id IN ?", recipeIDs).
		Group("recipe_id").
		Find(&results).Error
	return results, err
}
