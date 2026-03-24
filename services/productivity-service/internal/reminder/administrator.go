package reminder

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID, householdID uuid.UUID, title, notes string, scheduledFor time.Time) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:           uuid.New(),
		TenantId:     tenantID,
		HouseholdId:  householdID,
		Title:        title,
		Notes:        notes,
		ScheduledFor: scheduledFor,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func update(db *gorm.DB, id uuid.UUID, title, notes string, scheduledFor time.Time) (Entity, error) {
	var e Entity
	if err := db.Where("id = ?", id).First(&e).Error; err != nil {
		return Entity{}, err
	}
	e.Title = title
	e.Notes = notes
	e.ScheduledFor = scheduledFor
	e.UpdatedAt = time.Now().UTC()
	if err := db.Save(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func dismiss(db *gorm.DB, id uuid.UUID) error {
	now := time.Now().UTC()
	result := db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_dismissed_at": now,
		"updated_at":        now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func snooze(db *gorm.DB, id uuid.UUID, snoozedUntil time.Time) error {
	now := time.Now().UTC()
	return db.Model(&Entity{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_snoozed_until": snoozedUntil,
		"updated_at":         now,
	}).Error
}

func deleteByID(db *gorm.DB, id uuid.UUID) error {
	return db.Where("id = ?", id).Delete(&Entity{}).Error
}
