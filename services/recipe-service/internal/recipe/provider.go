package recipe

import (
	"fmt"
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
	TenantID            uuid.UUID
	HouseholdID         uuid.UUID
	UsageSort           UsageSort
}

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

// recipeWithUsage scans a recipe row joined with its cook-count aggregate.
// It embeds Entity so Preload("Tags") still resolves via recipes.id.
type recipeWithUsage struct {
	Entity
	UsageCount  int64   `gorm:"column:usage_count"`
	LastUsedDay *string `gorm:"column:last_used_day"`
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
