package performance

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func createPerformance(db *gorm.DB, e *Entity) error {
	if e.Id == uuid.Nil {
		e.Id = uuid.New()
	}
	now := time.Now().UTC()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	e.UpdatedAt = now
	return db.Create(e).Error
}

func updatePerformance(db *gorm.DB, e *Entity) error {
	e.UpdatedAt = time.Now().UTC()
	return db.Save(e).Error
}

func deleteSetsForPerformance(db *gorm.DB, performanceID uuid.UUID) error {
	return db.Where("performance_id = ?", performanceID).Delete(&SetEntity{}).Error
}

func createSet(db *gorm.DB, s *SetEntity) error {
	if s.Id == uuid.Nil {
		s.Id = uuid.New()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}
	return db.Create(s).Error
}

func loadSets(db *gorm.DB, performanceID uuid.UUID) ([]SetEntity, error) {
	var rows []SetEntity
	err := db.Where("performance_id = ?", performanceID).Order("set_number ASC").Find(&rows).Error
	return rows, err
}
