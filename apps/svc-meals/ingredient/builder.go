package ingredient

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Builder provides a fluent interface for constructing valid Ingredient models
type Builder struct {
	id          uuid.UUID
	mealId      uuid.UUID
	rawLine     string
	quantity    *float64
	quantityRaw string
	unit        *string
	unitRaw     *string
	ingredient  string
	preparation []string
	notes       []string
	confidence  float64
	createdAt   time.Time
	updatedAt   time.Time
}

// New creates a new Builder
func New() *Builder {
	return &Builder{
		preparation: []string{},
		notes:       []string{},
	}
}

// WithId sets the ingredient ID
func (b *Builder) WithId(id uuid.UUID) *Builder {
	b.id = id
	return b
}

// ForMeal sets the meal ID
func (b *Builder) ForMeal(mealId uuid.UUID) *Builder {
	b.mealId = mealId
	return b
}

// WithRawLine sets the original ingredient line
func (b *Builder) WithRawLine(rawLine string) *Builder {
	b.rawLine = rawLine
	return b
}

// WithQuantity sets the parsed quantity
func (b *Builder) WithQuantity(quantity *float64) *Builder {
	b.quantity = quantity
	return b
}

// WithQuantityRaw sets the raw quantity string
func (b *Builder) WithQuantityRaw(quantityRaw string) *Builder {
	b.quantityRaw = quantityRaw
	return b
}

// WithUnit sets the normalized unit
func (b *Builder) WithUnit(unit *string) *Builder {
	b.unit = unit
	return b
}

// WithUnitRaw sets the raw unit string
func (b *Builder) WithUnitRaw(unitRaw *string) *Builder {
	b.unitRaw = unitRaw
	return b
}

// WithIngredient sets the core ingredient name
func (b *Builder) WithIngredient(ingredient string) *Builder {
	b.ingredient = ingredient
	return b
}

// WithPreparation sets the preparation steps
func (b *Builder) WithPreparation(preparation []string) *Builder {
	if preparation == nil {
		b.preparation = []string{}
	} else {
		b.preparation = preparation
	}
	return b
}

// WithNotes sets the notes
func (b *Builder) WithNotes(notes []string) *Builder {
	if notes == nil {
		b.notes = []string{}
	} else {
		b.notes = notes
	}
	return b
}

// WithConfidence sets the confidence score
func (b *Builder) WithConfidence(confidence float64) *Builder {
	b.confidence = confidence
	return b
}

// WithTimestamps sets both created and updated timestamps
func (b *Builder) WithTimestamps(createdAt, updatedAt time.Time) *Builder {
	b.createdAt = createdAt
	b.updatedAt = updatedAt
	return b
}

// Build constructs and validates the Model
func (b *Builder) Build() (Model, error) {
	// Validate required fields
	if b.mealId == uuid.Nil {
		return Model{}, fmt.Errorf("meal ID is required")
	}
	if b.ingredient == "" {
		return Model{}, fmt.Errorf("ingredient is required")
	}
	if b.confidence < 0.0 || b.confidence > 1.0 {
		return Model{}, fmt.Errorf("confidence must be between 0.0 and 1.0")
	}

	// Generate ID if not provided
	if b.id == uuid.Nil {
		b.id = uuid.New()
	}

	// Set timestamps if not provided
	now := time.Now()
	if b.createdAt.IsZero() {
		b.createdAt = now
	}
	if b.updatedAt.IsZero() {
		b.updatedAt = now
	}

	// Ensure slices are not nil
	if b.preparation == nil {
		b.preparation = []string{}
	}
	if b.notes == nil {
		b.notes = []string{}
	}

	return Model{
		id:          b.id,
		mealId:      b.mealId,
		rawLine:     b.rawLine,
		quantity:    b.quantity,
		quantityRaw: b.quantityRaw,
		unit:        b.unit,
		unitRaw:     b.unitRaw,
		ingredient:  b.ingredient,
		preparation: b.preparation,
		notes:       b.notes,
		confidence:  b.confidence,
		createdAt:   b.createdAt,
		updatedAt:   b.updatedAt,
	}, nil
}
