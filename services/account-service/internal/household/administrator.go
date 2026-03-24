package household

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, tenantID uuid.UUID, name, timezone, units string) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:        uuid.New(),
		TenantId:  tenantID,
		Name:      name,
		Timezone:  timezone,
		Units:     units,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}

func update(db *gorm.DB, id uuid.UUID, name, timezone, units string) (Entity, error) {
	var e Entity
	if err := db.Where("id = ?", id).First(&e).Error; err != nil {
		return Entity{}, err
	}
	e.Name = name
	e.Timezone = timezone
	e.Units = units
	e.UpdatedAt = time.Now().UTC()
	if err := db.Save(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}
