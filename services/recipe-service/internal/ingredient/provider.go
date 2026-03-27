package ingredient

import (
	"strings"

	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func GetByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Preload("Aliases").Where("id = ?", id)
	})
}

func GetByName(tenantID uuid.UUID, name string) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Preload("Aliases").Where("tenant_id = ? AND name = ?", tenantID, strings.ToLower(strings.TrimSpace(name)))
	})
}

func GetByAlias(tenantID uuid.UUID, aliasName string) func(db *gorm.DB) (*Entity, *AliasEntity, error) {
	return func(db *gorm.DB) (*Entity, *AliasEntity, error) {
		var alias AliasEntity
		err := db.Where("tenant_id = ? AND name = ?", tenantID, strings.ToLower(strings.TrimSpace(aliasName))).First(&alias).Error
		if err != nil {
			return nil, nil, err
		}
		var entity Entity
		err = db.Preload("Aliases").Where("id = ?", alias.CanonicalIngredientId).First(&entity).Error
		if err != nil {
			return nil, nil, err
		}
		return &entity, &alias, nil
	}
}

func search(tenantID uuid.UUID, query string, page, pageSize int) func(db *gorm.DB) ([]Entity, int64, error) {
	return func(db *gorm.DB) ([]Entity, int64, error) {
		q := db.Model(&Entity{}).Where("tenant_id = ?", tenantID)

		if query != "" {
			pattern := "%" + strings.ToLower(query) + "%"
			q = q.Where("name ILIKE ? OR EXISTS (SELECT 1 FROM canonical_ingredient_aliases WHERE canonical_ingredient_aliases.canonical_ingredient_id = canonical_ingredients.id AND canonical_ingredient_aliases.name ILIKE ?)", pattern, pattern)
		}

		var total int64
		if err := q.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		offset := (page - 1) * pageSize
		var entities []Entity
		err := q.Preload("Aliases").
			Order("name ASC").
			Offset(offset).
			Limit(pageSize).
			Find(&entities).Error
		return entities, total, err
	}
}

func getUsageCount(db *gorm.DB, canonicalIngredientID uuid.UUID) (int64, error) {
	var count int64
	err := db.Table("recipe_ingredients").
		Where("canonical_ingredient_id = ?", canonicalIngredientID).
		Count(&count).Error
	return count, err
}

func suggestByPrefix(tenantID uuid.UUID, prefix string, limit int) func(db *gorm.DB) ([]Entity, error) {
	return func(db *gorm.DB) ([]Entity, error) {
		pattern := strings.ToLower(prefix) + "%"
		var entities []Entity
		err := db.Preload("Aliases").
			Where("tenant_id = ? AND (name ILIKE ? OR EXISTS (SELECT 1 FROM canonical_ingredient_aliases WHERE canonical_ingredient_aliases.canonical_ingredient_id = canonical_ingredients.id AND canonical_ingredient_aliases.name ILIKE ?))", tenantID, pattern, pattern).
			Order("(SELECT COUNT(*) FROM recipe_ingredients WHERE recipe_ingredients.canonical_ingredient_id = canonical_ingredients.id) DESC").
			Limit(limit).
			Find(&entities).Error
		return entities, err
	}
}
