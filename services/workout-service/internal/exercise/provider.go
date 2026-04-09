package exercise

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func GetByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ? AND deleted_at IS NULL", id)
	})
}

// GetByIDIncludeDeleted resolves an exercise even if it has been soft-deleted.
// Used by the planned-item / week / summary read paths so historical references
// to a removed exercise still display the original name.
func GetByIDIncludeDeleted(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func GetByName(userID uuid.UUID, name string) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND name = ? AND deleted_at IS NULL", userID, name)
	})
}

// ListByUser fetches active exercises optionally filtered by theme and region.
// `regionID` matches if the row's primary `region_id` equals the filter OR the
// region appears in the `secondary_region_ids` jsonb array. The `@>` containment
// operator is the cheapest way to express this on jsonb.
func ListByUser(db *gorm.DB, userID uuid.UUID, themeID, regionID *uuid.UUID) ([]Entity, error) {
	q := db.Where("user_id = ? AND deleted_at IS NULL", userID)
	if themeID != nil {
		q = q.Where("theme_id = ?", *themeID)
	}
	if regionID != nil {
		q = q.Where("region_id = ? OR secondary_region_ids @> ?::jsonb", *regionID, "[\""+regionID.String()+"\"]")
	}
	var rows []Entity
	if err := q.Order("name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
