package recipe

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID, householdID uuid.UUID, title, source string, description, sourceURL *string, servings, prepTimeMinutes, cookTimeMinutes *int, tags []string) (Entity, error) {
	now := time.Now().UTC()
	id := uuid.New()

	tagEntities := make([]TagEntity, len(tags))
	for i, t := range tags {
		tagEntities[i] = TagEntity{Id: uuid.New(), RecipeId: id, Tag: t}
	}

	e := Entity{
		Id: id, TenantId: tenantID, HouseholdId: householdID,
		Title: title, Description: description, Source: source,
		Servings: servings, PrepTimeMinutes: prepTimeMinutes, CookTimeMinutes: cookTimeMinutes,
		SourceURL: sourceURL, Tags: tagEntities,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func update(db *gorm.DB, e *Entity) (Entity, error) {
	e.UpdatedAt = time.Now().UTC()
	if err := db.Omit("Tags").Save(e).Error; err != nil {
		return Entity{}, err
	}
	return *e, nil
}

func softDelete(db *gorm.DB, id uuid.UUID) error {
	now := time.Now().UTC()
	return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at": now,
		"updated_at": now,
	}).Error
}

func restore(db *gorm.DB, id uuid.UUID) error {
	return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at": nil,
		"updated_at": time.Now().UTC(),
	}).Error
}

func createRestoration(db *gorm.DB, recipeID uuid.UUID) error {
	re := RestorationEntity{Id: uuid.New(), RecipeId: recipeID, RestoredAt: time.Now().UTC()}
	return db.Create(&re).Error
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
