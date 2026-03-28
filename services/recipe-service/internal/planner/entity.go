package planner

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	RecipeId           uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_planner_config_recipe"`
	Classification     *string   `gorm:"type:varchar(50)"`
	ServingsYield      *int      `gorm:"type:int"`
	EatWithinDays      *int      `gorm:"type:int"`
	MinGapDays         *int      `gorm:"type:int"`
	MaxConsecutiveDays *int      `gorm:"type:int"`
	CreatedAt          time.Time `gorm:"not null"`
	UpdatedAt          time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "recipe_planner_configs" }

func Migration(db *gorm.DB) error {
	return db.AutoMigrate(&Entity{})
}

func Make(e Entity) (Model, error) {
	classification := ""
	if e.Classification != nil {
		classification = *e.Classification
	}
	return NewBuilder().
		SetId(e.Id).
		SetRecipeID(e.RecipeId).
		SetClassification(classification).
		SetServingsYield(e.ServingsYield).
		SetEatWithinDays(e.EatWithinDays).
		SetMinGapDays(e.MinGapDays).
		SetMaxConsecutiveDays(e.MaxConsecutiveDays).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
