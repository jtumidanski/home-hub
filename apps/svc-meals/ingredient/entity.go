package ingredient

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StringArray is a custom type for storing string arrays in JSON format
type StringArray []string

// Scan implements the sql.Scanner interface
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, s)
}

// Value implements the driver.Valuer interface
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return json.Marshal([]string{})
	}
	return json.Marshal(s)
}

// Entity represents the database model for ingredients
type Entity struct {
	Id          uuid.UUID   `gorm:"type:uuid;primaryKey"`
	MealId      uuid.UUID   `gorm:"type:uuid;not null;index:idx_ingredients_meal"`
	RawLine     string      `gorm:"type:text;not null"`
	Quantity    *float64    `gorm:"type:double precision"`
	QuantityRaw string      `gorm:"type:varchar(100)"`
	Unit        *string     `gorm:"type:varchar(50)"`
	UnitRaw     *string     `gorm:"type:varchar(50)"`
	Ingredient  string      `gorm:"type:varchar(255);not null"`
	Preparation StringArray `gorm:"type:jsonb;default:'[]'"`
	Notes       StringArray `gorm:"type:jsonb;default:'[]'"`
	Confidence  float64     `gorm:"type:double precision;not null"`
	CreatedAt   time.Time   `gorm:"not null"`
	UpdatedAt   time.Time   `gorm:"not null"`
}

// TableName specifies the table name for the ingredients entity
func (e Entity) TableName() string {
	return "ingredients"
}

// Make transforms a database Entity into a domain Model
func Make(e Entity) (Model, error) {
	prep := []string(e.Preparation)
	if prep == nil {
		prep = []string{}
	}

	notes := []string(e.Notes)
	if notes == nil {
		notes = []string{}
	}

	return Model{
		id:          e.Id,
		mealId:      e.MealId,
		rawLine:     e.RawLine,
		quantity:    e.Quantity,
		quantityRaw: e.QuantityRaw,
		unit:        e.Unit,
		unitRaw:     e.UnitRaw,
		ingredient:  e.Ingredient,
		preparation: prep,
		notes:       notes,
		confidence:  e.Confidence,
		createdAt:   e.CreatedAt,
		updatedAt:   e.UpdatedAt,
	}, nil
}

// ToEntity transforms a domain Model into a database Entity
func (m Model) ToEntity() Entity {
	return Entity{
		Id:          m.id,
		MealId:      m.mealId,
		RawLine:     m.rawLine,
		Quantity:    m.quantity,
		QuantityRaw: m.quantityRaw,
		Unit:        m.unit,
		UnitRaw:     m.unitRaw,
		Ingredient:  m.ingredient,
		Preparation: StringArray(m.preparation),
		Notes:       StringArray(m.notes),
		Confidence:  m.confidence,
		CreatedAt:   m.createdAt,
		UpdatedAt:   m.updatedAt,
	}
}

// Migration runs the auto-migration for the ingredients table
func Migration() func(db *gorm.DB) error {
	return func(db *gorm.DB) error {
		return db.AutoMigrate(&Entity{})
	}
}
