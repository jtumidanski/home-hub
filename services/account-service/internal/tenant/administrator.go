package tenant

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func create(db *gorm.DB, name string) (Entity, error) {
	now := time.Now().UTC()
	e := Entity{
		Id:        uuid.New(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&e).Error; err != nil {
		return Entity{}, err
	}
	return e, nil
}
