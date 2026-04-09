package trackingitem

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

func GetByIDIncludeDeleted(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func GetAllByUser(userID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND deleted_at IS NULL", userID).Order("sort_order ASC")
	})
}

func GetAllByUserIncludeDeleted(userID uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID).Order("sort_order ASC")
	})
}

func GetByName(userID uuid.UUID, name string) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND name = ? AND deleted_at IS NULL", userID, name)
	})
}

func getMaxSortOrder(db *gorm.DB, userID uuid.UUID) (int, error) {
	var maxOrder *int
	err := db.Model(&Entity{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Select("MAX(sort_order)").
		Scan(&maxOrder).Error
	if err != nil {
		return 0, err
	}
	if maxOrder == nil {
		return 0, nil
	}
	return *maxOrder, nil
}
