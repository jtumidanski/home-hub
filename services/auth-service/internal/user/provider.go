package user

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"gorm.io/gorm"
)

func getByID(id uuid.UUID) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func getByEmail(email string) database.EntityProvider[Entity] {
	return database.Query[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("email = ?", email)
	})
}

func modelFromEntity(e Entity) (Model, error) {
	return Make(e)
}

func byIDProvider(id uuid.UUID) func(db *gorm.DB) model.Provider[Model] {
	return func(db *gorm.DB) model.Provider[Model] {
		return model.Map(modelFromEntity)(getByID(id)(db))
	}
}

func byEmailProvider(email string) func(db *gorm.DB) model.Provider[Model] {
	return func(db *gorm.DB) model.Provider[Model] {
		return model.Map(modelFromEntity)(getByEmail(email)(db))
	}
}

func getByIDs(ids []uuid.UUID) database.EntityProvider[[]Entity] {
	return database.SliceQuery[Entity](func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	})
}
