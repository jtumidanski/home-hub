package planitem

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id                uuid.UUID `gorm:"type:uuid;primaryKey"`
	PlanWeekId        uuid.UUID `gorm:"type:uuid;not null;index:idx_plan_item_plan_week"`
	Day               time.Time `gorm:"type:date;not null"`
	Slot              string    `gorm:"type:varchar(20);not null"`
	RecipeId          uuid.UUID `gorm:"type:uuid;not null;index:idx_plan_item_recipe"`
	ServingMultiplier *float64  `gorm:"type:decimal(5,2)"`
	PlannedServings   *int      `gorm:"type:integer"`
	Notes             *string   `gorm:"type:text"`
	Position          int       `gorm:"not null;default:0"`
	CreatedAt         time.Time `gorm:"not null"`
	UpdatedAt         time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "plan_items" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetPlanWeekID(e.PlanWeekId).
		SetDay(e.Day).
		SetSlot(e.Slot).
		SetRecipeID(e.RecipeId).
		SetServingMultiplier(e.ServingMultiplier).
		SetPlannedServings(e.PlannedServings).
		SetNotes(e.Notes).
		SetPosition(e.Position).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:                m.id,
		PlanWeekId:        m.planWeekID,
		Day:               m.day,
		Slot:              m.slot,
		RecipeId:          m.recipeID,
		ServingMultiplier: m.servingMultiplier,
		PlannedServings:   m.plannedServings,
		Notes:             m.notes,
		Position:          m.position,
		CreatedAt:         m.createdAt,
		UpdatedAt:         m.updatedAt,
	}
}
