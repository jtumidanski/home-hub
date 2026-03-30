package planitem

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidSlot(t *testing.T) {
	tests := []struct {
		name string
		slot string
		want bool
	}{
		{"breakfast is valid", SlotBreakfast, true},
		{"lunch is valid", SlotLunch, true},
		{"dinner is valid", SlotDinner, true},
		{"snack is valid", SlotSnack, true},
		{"side is valid", SlotSide, true},
		{"empty is invalid", "", false},
		{"unknown is invalid", "brunch", false},
		{"capitalized is invalid", "Breakfast", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidSlot(tt.slot))
		})
	}
}

func TestBuilder_Build(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	planWeekID := uuid.New()
	recipeID := uuid.New()
	day := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	multiplier := 1.5
	servings := 4
	notes := "Extra spicy"

	tests := []struct {
		name        string
		setup       func() *Builder
		wantErr     error
		assertModel func(t *testing.T, m Model)
	}{
		{
			name: "valid build with all fields",
			setup: func() *Builder {
				return NewBuilder().
					SetId(id).
					SetPlanWeekID(planWeekID).
					SetDay(day).
					SetSlot(SlotDinner).
					SetRecipeID(recipeID).
					SetServingMultiplier(&multiplier).
					SetPlannedServings(&servings).
					SetNotes(&notes).
					SetPosition(2).
					SetCreatedAt(now).
					SetUpdatedAt(now)
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				assert.Equal(t, id, m.Id())
				assert.Equal(t, planWeekID, m.PlanWeekID())
				assert.Equal(t, day, m.Day())
				assert.Equal(t, SlotDinner, m.Slot())
				assert.Equal(t, recipeID, m.RecipeID())
				require.NotNil(t, m.ServingMultiplier())
				assert.Equal(t, 1.5, *m.ServingMultiplier())
				require.NotNil(t, m.PlannedServings())
				assert.Equal(t, 4, *m.PlannedServings())
				require.NotNil(t, m.Notes())
				assert.Equal(t, "Extra spicy", *m.Notes())
				assert.Equal(t, 2, m.Position())
			},
		},
		{
			name: "minimal valid build",
			setup: func() *Builder {
				return NewBuilder().
					SetDay(day).
					SetSlot(SlotBreakfast).
					SetRecipeID(recipeID)
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				assert.Equal(t, day, m.Day())
				assert.Equal(t, SlotBreakfast, m.Slot())
				assert.Equal(t, recipeID, m.RecipeID())
				assert.Nil(t, m.ServingMultiplier())
				assert.Nil(t, m.PlannedServings())
				assert.Nil(t, m.Notes())
				assert.Equal(t, 0, m.Position())
			},
		},
		{
			name: "zero day returns ErrDayRequired",
			setup: func() *Builder {
				return NewBuilder().
					SetSlot(SlotDinner).
					SetRecipeID(recipeID)
			},
			wantErr: ErrDayRequired,
		},
		{
			name: "invalid slot returns ErrInvalidSlot",
			setup: func() *Builder {
				return NewBuilder().
					SetDay(day).
					SetSlot("brunch").
					SetRecipeID(recipeID)
			},
			wantErr: ErrInvalidSlot,
		},
		{
			name: "empty slot returns ErrInvalidSlot",
			setup: func() *Builder {
				return NewBuilder().
					SetDay(day).
					SetSlot("").
					SetRecipeID(recipeID)
			},
			wantErr: ErrInvalidSlot,
		},
		{
			name: "nil recipe ID returns ErrRecipeIDRequired",
			setup: func() *Builder {
				return NewBuilder().
					SetDay(day).
					SetSlot(SlotLunch)
			},
			wantErr: ErrRecipeIDRequired,
		},
		{
			name: "each valid slot accepted",
			setup: func() *Builder {
				return NewBuilder().
					SetDay(day).
					SetSlot(SlotSnack).
					SetRecipeID(recipeID)
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				assert.Equal(t, SlotSnack, m.Slot())
			},
		},
		{
			name: "side slot accepted",
			setup: func() *Builder {
				return NewBuilder().
					SetDay(day).
					SetSlot(SlotSide).
					SetRecipeID(recipeID)
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				assert.Equal(t, SlotSide, m.Slot())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()
			model, err := builder.Build()

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			if tt.assertModel != nil {
				tt.assertModel(t, model)
			}
		})
	}
}
