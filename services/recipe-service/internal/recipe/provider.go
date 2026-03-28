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

func getDeletedByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Preload("Tags").Where("id = ? AND deleted_at IS NOT NULL", id)
	})
}

type ListFilters struct {
	Search             string
	Tags               []string
	Page               int
	PageSize           int
	PlannerReady       *bool
	Classification     string
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
