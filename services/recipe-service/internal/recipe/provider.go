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

func getAll(search string, tags []string, page, pageSize int) func(db *gorm.DB) ([]Entity, int64, error) {
	return func(db *gorm.DB) ([]Entity, int64, error) {
		query := db.Model(&Entity{}).Where("deleted_at IS NULL")

		if search != "" {
			query = query.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(search)+"%")
		}

		if len(tags) > 0 {
			for _, tag := range tags {
				query = query.Where("id IN (?)",
					db.Model(&TagEntity{}).Select("recipe_id").Where("tag = ?", strings.ToLower(strings.TrimSpace(tag))),
				)
			}
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		offset := (page - 1) * pageSize
		var entities []Entity
		err := query.Preload("Tags").
			Order("created_at DESC").
			Offset(offset).
			Limit(pageSize).
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

func create(db *gorm.DB, e *Entity) error {
	return db.Create(e).Error
}

func save(db *gorm.DB, e *Entity) error {
	return db.Save(e).Error
}

func softDelete(db *gorm.DB, id uuid.UUID) error {
	return db.Exec("UPDATE recipes SET deleted_at = NOW(), updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", id).Error
}

func restoreByID(db *gorm.DB, id uuid.UUID) error {
	return db.Exec("UPDATE recipes SET deleted_at = NULL, updated_at = NOW() WHERE id = ?", id).Error
}

func getDeletedByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Preload("Tags").Where("id = ? AND deleted_at IS NOT NULL", id)
	})
}

func replaceTags(db *gorm.DB, recipeID uuid.UUID, tags []string) error {
	if err := db.Where("recipe_id = ?", recipeID).Delete(&TagEntity{}).Error; err != nil {
		return err
	}
	for _, tag := range tags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		te := TagEntity{Id: uuid.New(), RecipeId: recipeID, Tag: normalized}
		if err := db.Create(&te).Error; err != nil {
			return err
		}
	}
	return nil
}
