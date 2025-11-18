package meal

import (
	"time"

	"github.com/google/uuid"
)

// Create creates a new meal model with generated ID and timestamps
func Create(householdId, userId uuid.UUID, title, description, ingredientText string) (Model, error) {
	return New().
		ForHousehold(householdId).
		ByUser(userId).
		WithTitle(title).
		WithDescription(description).
		WithIngredientText(ingredientText).
		Build()
}

// Update updates a meal's mutable fields and returns a new model
func Update(m Model, title, description *string) Model {
	builder := FromModel(m)

	if title != nil {
		builder.WithTitle(*title)
	}
	if description != nil {
		builder.WithDescription(*description)
	}

	// Update the updatedAt timestamp
	builder.updatedAt = time.Now()

	// Build should not fail for updates since we're starting from a valid model
	updated, _ := builder.Build()
	return updated
}
