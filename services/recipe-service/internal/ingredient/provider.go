package ingredient

import (
	"strings"
	"time"

	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func GetByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Preload("Aliases").Where("id = ?", id)
	})
}

// GetByIDs fetches canonical ingredients (with their aliases) for a list of
// ids in a single query. GORM 2.x batches the Aliases preload into a single
// follow-up `IN` query, so this issues exactly two SQL statements regardless
// of the number of ids. Empty input short-circuits without hitting the DB.
func GetByIDs(ids []uuid.UUID) func(db *gorm.DB) ([]Entity, error) {
	return func(db *gorm.DB) ([]Entity, error) {
		if len(ids) == 0 {
			return []Entity{}, nil
		}
		var entities []Entity
		err := db.Preload("Aliases").Where("id IN (?)", ids).Find(&entities).Error
		return entities, err
	}
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

type RecipeRef struct {
	RecipeId uuid.UUID `json:"recipeId"`
	RawName  string    `json:"rawName"`
}

func getIngredientRecipes(db *gorm.DB, canonicalIngredientID uuid.UUID, page, pageSize int) ([]RecipeRef, int64, error) {
	query := db.Table("recipe_ingredients").Where("canonical_ingredient_id = ?", canonicalIngredientID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	var refs []RecipeRef
	err := query.Select("recipe_id, raw_name").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&refs).Error
	return refs, total, err
}

func reassignCanonical(db *gorm.DB, fromID, toID uuid.UUID) (int64, error) {
	result := db.Table("recipe_ingredients").
		Where("canonical_ingredient_id = ?", fromID).
		Updates(map[string]interface{}{
			"canonical_ingredient_id": toID,
			"updated_at":             time.Now().UTC(),
		})
	return result.RowsAffected, result.Error
}

type entityWithCategory struct {
	Entity
	CategoryName string
}

type entityWithUsage struct {
	Entity
	UsageCount   int64
	CategoryName string
}

func searchWithUsage(tenantID uuid.UUID, query string, categoryFilter string, page, pageSize int) func(db *gorm.DB) ([]entityWithUsage, int64, error) {
	return func(db *gorm.DB) ([]entityWithUsage, int64, error) {
		q := db.Model(&Entity{}).Where("canonical_ingredients.tenant_id = ?", tenantID)

		if query != "" {
			pattern := "%" + strings.ToLower(query) + "%"
			q = q.Where("canonical_ingredients.name ILIKE ? OR EXISTS (SELECT 1 FROM canonical_ingredient_aliases WHERE canonical_ingredient_aliases.canonical_ingredient_id = canonical_ingredients.id AND canonical_ingredient_aliases.name ILIKE ?)", pattern, pattern)
		}
		if categoryFilter == "null" {
			q = q.Where("canonical_ingredients.category_id IS NULL")
		} else if categoryFilter != "" {
			q = q.Where("canonical_ingredients.category_id = ?", categoryFilter)
		}

		var total int64
		if err := q.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		offset := (page - 1) * pageSize
		var results []entityWithUsage
		dataQ := db.Table("canonical_ingredients").
			Select("canonical_ingredients.*, COALESCE((SELECT COUNT(*) FROM recipe_ingredients WHERE recipe_ingredients.canonical_ingredient_id = canonical_ingredients.id), 0) as usage_count, COALESCE(ingredient_categories.name, '') as category_name").
			Joins("LEFT JOIN ingredient_categories ON ingredient_categories.id = canonical_ingredients.category_id").
			Where("canonical_ingredients.tenant_id = ?", tenantID).
			Scopes(func(tx *gorm.DB) *gorm.DB {
				if query != "" {
					pattern := "%" + strings.ToLower(query) + "%"
					tx = tx.Where("canonical_ingredients.name ILIKE ? OR EXISTS (SELECT 1 FROM canonical_ingredient_aliases WHERE canonical_ingredient_aliases.canonical_ingredient_id = canonical_ingredients.id AND canonical_ingredient_aliases.name ILIKE ?)", pattern, pattern)
				}
				if categoryFilter == "null" {
					tx = tx.Where("canonical_ingredients.category_id IS NULL")
				} else if categoryFilter != "" {
					tx = tx.Where("canonical_ingredients.category_id = ?", categoryFilter)
				}
				return tx
			})
		err := dataQ.Preload("Aliases").
			Order("canonical_ingredients.name ASC").
			Offset(offset).
			Limit(pageSize).
			Find(&results).Error
		return results, total, err
	}
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
