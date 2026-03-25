package preference

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID, userID uuid.UUID) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:        uuid.New(),
		TenantId:  tenantID,
		UserId:    userID,
		Theme:     "light",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func updateTheme(db *gorm.DB, id uuid.UUID, theme string) (Entity, error) {
	var e Entity
	if err := db.Where("id = ?", id).First(&e).Error; err != nil {
		return Entity{}, err
	}
	e.Theme = theme
	e.UpdatedAt = time.Now().UTC()
	if err := db.Save(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func setActiveHousehold(db *gorm.DB, id uuid.UUID, householdID uuid.UUID) (Entity, error) {
	var e Entity
	if err := db.Where("id = ?", id).First(&e).Error; err != nil {
		return Entity{}, err
	}
	e.ActiveHouseholdId = &householdID
	e.UpdatedAt = time.Now().UTC()
	if err := db.Save(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}
