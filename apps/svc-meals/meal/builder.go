package meal

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Builder provides a fluent interface for constructing valid Meal models
type Builder struct {
	id                uuid.UUID
	householdId       uuid.UUID
	userId            uuid.UUID
	title             string
	description       string
	rawIngredientText string
	createdAt         time.Time
	updatedAt         time.Time
}

// New creates a new Builder
func New() *Builder {
	return &Builder{}
}

// WithId sets the meal ID
func (b *Builder) WithId(id uuid.UUID) *Builder {
	b.id = id
	return b
}

// ForHousehold sets the household ID
func (b *Builder) ForHousehold(householdId uuid.UUID) *Builder {
	b.householdId = householdId
	return b
}

// ByUser sets the user ID (creator)
func (b *Builder) ByUser(userId uuid.UUID) *Builder {
	b.userId = userId
	return b
}

// WithTitle sets the meal title
func (b *Builder) WithTitle(title string) *Builder {
	b.title = title
	return b
}

// WithDescription sets the meal description
func (b *Builder) WithDescription(description string) *Builder {
	b.description = description
	return b
}

// WithIngredientText sets the raw ingredient text
func (b *Builder) WithIngredientText(text string) *Builder {
	b.rawIngredientText = text
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
	if b.householdId == uuid.Nil {
		return Model{}, fmt.Errorf("household ID is required")
	}
	if b.userId == uuid.Nil {
		return Model{}, fmt.Errorf("user ID is required")
	}
	if b.title == "" {
		return Model{}, fmt.Errorf("title is required")
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

	return Model{
		id:                b.id,
		householdId:       b.householdId,
		userId:            b.userId,
		title:             b.title,
		description:       b.description,
		rawIngredientText: b.rawIngredientText,
		createdAt:         b.createdAt,
		updatedAt:         b.updatedAt,
	}, nil
}

// FromEntity creates a Builder from an existing Entity
func FromEntity(e Entity) *Builder {
	return &Builder{
		id:                e.Id,
		householdId:       e.HouseholdId,
		userId:            e.UserId,
		title:             e.Title,
		description:       e.Description,
		rawIngredientText: e.RawIngredientText,
		createdAt:         e.CreatedAt,
		updatedAt:         e.UpdatedAt,
	}
}

// FromModel creates a Builder from an existing Model
func FromModel(m Model) *Builder {
	return &Builder{
		id:                m.id,
		householdId:       m.householdId,
		userId:            m.userId,
		title:             m.title,
		description:       m.description,
		rawIngredientText: m.rawIngredientText,
		createdAt:         m.createdAt,
		updatedAt:         m.updatedAt,
	}
}
