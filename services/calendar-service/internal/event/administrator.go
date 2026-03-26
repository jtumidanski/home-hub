package event

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func upsert(db *gorm.DB, e Entity) error {
	now := time.Now().UTC()
	e.UpdatedAt = now

	result := db.Where("source_id = ? AND external_id = ?", e.SourceId, e.ExternalId).First(&Entity{})
	if result.Error == nil {
		return db.Model(&Entity{}).
			Where("source_id = ? AND external_id = ?", e.SourceId, e.ExternalId).
			Updates(map[string]interface{}{
				"title":             e.Title,
				"description":       e.Description,
				"start_time":        e.StartTime,
				"end_time":          e.EndTime,
				"all_day":           e.AllDay,
				"location":          e.Location,
				"visibility":        e.Visibility,
				"user_display_name": e.UserDisplayName,
				"user_color":        e.UserColor,
				"updated_at":        now,
			}).Error
	}

	e.Id = uuid.New()
	e.CreatedAt = now
	return db.Create(&e).Error
}

func deleteBySourceAndExternalIDs(db *gorm.DB, sourceID uuid.UUID, externalIDs []string) error {
	if len(externalIDs) == 0 {
		return nil
	}
	return db.Where("source_id = ? AND external_id IN ?", sourceID, externalIDs).Delete(&Entity{}).Error
}

func deleteByConnection(db *gorm.DB, connectionID uuid.UUID) error {
	return db.Where("connection_id = ?", connectionID).Delete(&Entity{}).Error
}

func deleteBySource(db *gorm.DB, sourceID uuid.UUID) error {
	return db.Where("source_id = ?", sourceID).Delete(&Entity{}).Error
}

func countByConnection(db *gorm.DB, connectionID uuid.UUID) (int64, error) {
	var count int64
	err := db.Model(&Entity{}).Where("connection_id = ?", connectionID).Count(&count).Error
	return count, err
}
