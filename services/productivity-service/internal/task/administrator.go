package task

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID, householdID uuid.UUID, title, notes, status string, dueOn *time.Time, rolloverEnabled bool) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:              uuid.New(),
		TenantId:        tenantID,
		HouseholdId:     householdID,
		Title:           title,
		Notes:           notes,
		Status:          status,
		DueOn:           dueOn,
		RolloverEnabled: rolloverEnabled,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func update(db *gorm.DB, id uuid.UUID, title, notes, status string, dueOn *time.Time, rolloverEnabled bool, userID uuid.UUID) (Entity, error) {
	var e Entity
	if err := db.Where("id = ?", id).First(&e).Error; err != nil {
		return Entity{}, err
	}
	e.Title = title
	e.Notes = notes
	e.RolloverEnabled = rolloverEnabled
	if dueOn != nil {
		e.DueOn = dueOn
	}
	e.UpdatedAt = time.Now().UTC()

	// Handle status transitions
	if status == "completed" && e.Status != "completed" {
		now := time.Now().UTC()
		e.CompletedAt = &now
		e.CompletedByUserId = &userID
	} else if status == "pending" && e.Status == "completed" {
		e.CompletedAt = nil
		e.CompletedByUserId = nil
	}
	e.Status = status

	if err := db.Save(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
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
