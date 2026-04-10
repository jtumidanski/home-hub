package performance

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func GetByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func GetByPlannedItem(plannedItemID uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("planned_item_id = ?", plannedItemID)
	})
}

// LoadByPlannedItems is the bulk read used by the week-projection assembler.
// It returns two maps keyed by planned_item_id and performance_id respectively
// so the caller can stitch performances back to their parent items in O(1).
func LoadByPlannedItems(db *gorm.DB, plannedItemIDs []uuid.UUID) (map[uuid.UUID]Entity, map[uuid.UUID][]SetEntity, error) {
	perfByItem := make(map[uuid.UUID]Entity)
	setsByPerf := make(map[uuid.UUID][]SetEntity)
	if len(plannedItemIDs) == 0 {
		return perfByItem, setsByPerf, nil
	}

	var perfs []Entity
	if err := db.Where("planned_item_id IN ?", plannedItemIDs).Find(&perfs).Error; err != nil {
		return nil, nil, err
	}
	if len(perfs) == 0 {
		return perfByItem, setsByPerf, nil
	}

	perfIDs := make([]uuid.UUID, 0, len(perfs))
	for _, p := range perfs {
		perfByItem[p.PlannedItemId] = p
		perfIDs = append(perfIDs, p.Id)
	}

	var sets []SetEntity
	if err := db.Where("performance_id IN ?", perfIDs).Order("set_number ASC").Find(&sets).Error; err != nil {
		return nil, nil, err
	}
	for _, s := range sets {
		setsByPerf[s.PerformanceId] = append(setsByPerf[s.PerformanceId], s)
	}
	return perfByItem, setsByPerf, nil
}
