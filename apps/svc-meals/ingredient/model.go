package ingredient

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Model represents an immutable ingredient in the domain.
// All fields are private to enforce immutability.
type Model struct {
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

// Id returns the ingredient's unique identifier
func (m Model) Id() uuid.UUID {
	return m.id
}

// MealId returns the ID of the meal this ingredient belongs to
func (m Model) MealId() uuid.UUID {
	return m.mealId
}

// RawLine returns the original ingredient line
func (m Model) RawLine() string {
	return m.rawLine
}

// Quantity returns the parsed quantity (may be nil)
func (m Model) Quantity() *float64 {
	return m.quantity
}

// QuantityRaw returns the raw quantity string
func (m Model) QuantityRaw() string {
	return m.quantityRaw
}

// Unit returns the normalized unit (may be nil)
func (m Model) Unit() *string {
	return m.unit
}

// UnitRaw returns the raw unit string (may be nil)
func (m Model) UnitRaw() *string {
	return m.unitRaw
}

// Ingredient returns the core ingredient name
func (m Model) Ingredient() string {
	return m.ingredient
}

// Preparation returns preparation steps
func (m Model) Preparation() []string {
	return m.preparation
}

// Notes returns additional notes
func (m Model) Notes() []string {
	return m.notes
}

// Confidence returns the parsing confidence score (0.0 to 1.0)
func (m Model) Confidence() float64 {
	return m.confidence
}

// CreatedAt returns when the ingredient was created
func (m Model) CreatedAt() time.Time {
	return m.createdAt
}

// UpdatedAt returns when the ingredient was last updated
func (m Model) UpdatedAt() time.Time {
	return m.updatedAt
}

// IsLowConfidence returns true if confidence is below 0.7
func (m Model) IsLowConfidence() bool {
	return m.confidence < 0.7
}

// String returns a string representation of the ingredient for debugging
func (m Model) String() string {
	return fmt.Sprintf("Ingredient[id=%s, meal=%s, ingredient=%s, confidence=%.2f]",
		m.id.String(), m.mealId.String(), m.ingredient, m.confidence)
}

// MarshalJSON implements json.Marshaler for the Model
func (m Model) MarshalJSON() ([]byte, error) {
	type alias struct {
		Id          uuid.UUID `json:"id"`
		MealId      uuid.UUID `json:"mealId"`
		RawLine     string    `json:"rawLine"`
		Quantity    *float64  `json:"quantity"`
		QuantityRaw string    `json:"quantityRaw"`
		Unit        *string   `json:"unit"`
		UnitRaw     *string   `json:"unitRaw"`
		Ingredient  string    `json:"ingredient"`
		Preparation []string  `json:"preparation"`
		Notes       []string  `json:"notes"`
		Confidence  float64   `json:"confidence"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}

	return json.Marshal(&alias{
		Id:          m.id,
		MealId:      m.mealId,
		RawLine:     m.rawLine,
		Quantity:    m.quantity,
		QuantityRaw: m.quantityRaw,
		Unit:        m.unit,
		UnitRaw:     m.unitRaw,
		Ingredient:  m.ingredient,
		Preparation: m.preparation,
		Notes:       m.notes,
		Confidence:  m.confidence,
		CreatedAt:   m.createdAt,
		UpdatedAt:   m.updatedAt,
	})
}

// Is returns true if the given model represents the same ingredient
func (m Model) Is(other Model) bool {
	return m.id == other.id
}
