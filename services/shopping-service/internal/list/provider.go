package list

import (
	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"gorm.io/gorm"
)

func GetByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func GetByStatus(status string) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status).Order("updated_at DESC")
	})
}

type ItemCounts struct {
	ListId       uuid.UUID `gorm:"column:list_id"`
	ItemCount    int       `gorm:"column:item_count"`
	CheckedCount int       `gorm:"column:checked_count"`
}

func getItemCounts(db *gorm.DB, listIDs []uuid.UUID) (map[uuid.UUID]ItemCounts, error) {
	if len(listIDs) == 0 {
		return map[uuid.UUID]ItemCounts{}, nil
	}
	var counts []ItemCounts
	err := db.Table("shopping_items").
		Select("list_id, COUNT(*) as item_count, SUM(CASE WHEN checked THEN 1 ELSE 0 END) as checked_count").
		Where("list_id IN ?", listIDs).
		Group("list_id").
		Scan(&counts).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]ItemCounts, len(counts))
	for _, c := range counts {
		result[c.ListId] = c
	}
	return result, nil
}

func getItemCountsForList(db *gorm.DB, listID uuid.UUID) (int, int, error) {
	counts, err := getItemCounts(db, []uuid.UUID{listID})
	if err != nil {
		return 0, 0, err
	}
	if c, ok := counts[listID]; ok {
		return c.ItemCount, c.CheckedCount, nil
	}
	return 0, 0, nil
}
