package recipe

import (
	"strings"

	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Preload("Tags").Where("id = ? AND deleted_at IS NULL", id)
	})
}

func getByIDs(ids []uuid.UUID) func(db *gorm.DB) ([]Entity, error) {
	return func(db *gorm.DB) ([]Entity, error) {
		if len(ids) == 0 {
			return []Entity{}, nil
		}
		var entities []Entity
		err := db.Preload("Tags").Where("id IN (?) AND deleted_at IS NULL", ids).Find(&entities).Error
		return entities, err
	}
}

func getDeletedByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Preload("Tags").Where("id = ? AND deleted_at IS NOT NULL", id)
	})
}

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

type ListFilters struct {
	Search              string
	Tags                []string
	Page                int
	PageSize            int
	PlannerReady        *bool
	Classification      string
	NormalizationStatus string
}

func getAll(filters ListFilters) func(db *gorm.DB) ([]Entity, int64, error) {
	return func(db *gorm.DB) ([]Entity, int64, error) {
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
				// Planner ready = has classification tag AND has servings
				query = query.Where("id IN (?)",
					db.Table("recipe_planner_configs").Select("recipe_id").Where("classification IS NOT NULL AND classification != ''"),
				).Where("servings IS NOT NULL")
			} else {
				// Not planner ready = missing classification OR missing servings
				query = query.Where("(id NOT IN (?) OR servings IS NULL)",
					db.Table("recipe_planner_configs").Select("recipe_id").Where("classification IS NOT NULL AND classification != ''"),
				)
			}
		}

		if filters.NormalizationStatus == "complete" {
			// All ingredients resolved: no unresolved recipe_ingredients for this recipe
			query = query.Where("id IN (?)",
				db.Table("recipe_ingredients").Select("DISTINCT recipe_id"),
			).Where("id NOT IN (?)",
				db.Table("recipe_ingredients").Select("DISTINCT recipe_id").Where("normalization_status = 'unresolved'"),
			)
		} else if filters.NormalizationStatus == "incomplete" {
			// Has at least one unresolved ingredient
			query = query.Where("id IN (?)",
				db.Table("recipe_ingredients").Select("DISTINCT recipe_id").Where("normalization_status = 'unresolved'"),
			)
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		offset := (filters.Page - 1) * filters.PageSize
		var entities []Entity
		err := query.Preload("Tags").
			Order("created_at DESC").
			Offset(offset).
			Limit(filters.PageSize).
			Find(&entities).Error
		return entities, total, err
	}
}

type TagCount struct {
	Tag   string
	Count int64
}

type recipeUsageResult struct {
	recipeID    uuid.UUID
	lastUsedDay *string
	usageCount  int64
}

type recipeUsageRow struct {
	RecipeID    uuid.UUID `gorm:"column:recipe_id"`
	LastUsedDay *string   `gorm:"column:last_used_day"`
	UsageCount  int64     `gorm:"column:usage_count"`
}

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

func getAllTags(db *gorm.DB) ([]TagCount, error) {
	var results []TagCount
	err := db.Model(&TagEntity{}).
		Select("tag, COUNT(*) as count").
		Joins("JOIN recipes ON recipes.id = recipe_tags.recipe_id AND recipes.deleted_at IS NULL").
		Group("tag").
		Order("count DESC").
		Find(&results).Error
	return results, err
}
