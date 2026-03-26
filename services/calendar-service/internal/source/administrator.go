package source

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID, householdID, connectionID uuid.UUID, externalID, name string, primary bool, color string) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:           uuid.New(),
		TenantId:     tenantID,
		HouseholdId:  householdID,
		ConnectionId: connectionID,
		ExternalId:   externalID,
		Name:         name,
		Primary:      primary,
		Visible:      true,
		Color:        color,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func updateVisibility(db *gorm.DB, id uuid.UUID, visible bool) error {
	return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"visible":    visible,
		"updated_at": time.Now().UTC(),
	}).Error
}

func updateSyncToken(db *gorm.DB, id uuid.UUID, syncToken string) error {
	return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"sync_token": syncToken,
		"updated_at": time.Now().UTC(),
	}).Error
}

func updateNameAndColor(db *gorm.DB, id uuid.UUID, name, color string, primary bool) error {
	return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":       name,
		"color":      color,
		"primary":    primary,
		"updated_at": time.Now().UTC(),
	}).Error
}

func deleteByConnection(db *gorm.DB, connectionID uuid.UUID) error {
	return db.Where("connection_id = ?", connectionID).Delete(&Entity{}).Error
}

func deleteByID(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&Entity{}).Error
}
