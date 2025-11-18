package meal

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity represents the database model for meals.
// This is separate from the domain Model to isolate persistence concerns.
type Entity struct {
	Id                uuid.UUID `gorm:"type:uuid;primaryKey"`
	HouseholdId       uuid.UUID `gorm:"type:uuid;not null;index:idx_meals_household_created"`
	UserId            uuid.UUID `gorm:"type:uuid;not null"`
	Title             string    `gorm:"type:varchar(255);not null"`
	Description       string    `gorm:"type:text"`
	RawIngredientText string    `gorm:"type:text"`
	CreatedAt         time.Time `gorm:"not null;index:idx_meals_household_created"`
	UpdatedAt         time.Time `gorm:"not null"`
}

// TableName specifies the table name for the meals entity
func (e Entity) TableName() string {
	return "meals"
}

// Make transforms a database Entity into a domain Model
func Make(e Entity) (Model, error) {
	return Model{
		id:                e.Id,
		householdId:       e.HouseholdId,
		userId:            e.UserId,
		title:             e.Title,
		description:       e.Description,
		rawIngredientText: e.RawIngredientText,
		createdAt:         e.CreatedAt,
		updatedAt:         e.UpdatedAt,
	}, nil
}

// ToEntity transforms a domain Model into a database Entity
func (m Model) ToEntity() Entity {
	return Entity{
		Id:                m.id,
		HouseholdId:       m.householdId,
		UserId:            m.userId,
		Title:             m.title,
		Description:       m.description,
		RawIngredientText: m.rawIngredientText,
		CreatedAt:         m.createdAt,
		UpdatedAt:         m.updatedAt,
	}
}

// Migration runs the auto-migration for the meals table
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		return db.AutoMigrate(&Entity{})
	}
}
