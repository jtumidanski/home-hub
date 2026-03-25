package recipe

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Build(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	tenantID := uuid.New()
	householdID := uuid.New()
	servings := 4
	prepTime := 10
	cookTime := 20

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
					SetTenantID(tenantID).
					SetHouseholdID(householdID).
					SetTitle("Pasta Carbonara").
					SetDescription("Classic Roman dish").
					SetSource("Add @eggs{3}.").
					SetServings(&servings).
					SetPrepTimeMinutes(&prepTime).
					SetCookTimeMinutes(&cookTime).
					SetSourceURL("https://example.com").
					SetTags([]string{"italian", "pasta"}).
					SetCreatedAt(now).
					SetUpdatedAt(now)
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				require.Equal(t, id, m.id)
				require.Equal(t, tenantID, m.tenantID)
				require.Equal(t, householdID, m.householdID)
				require.Equal(t, "Pasta Carbonara", m.title)
				require.Equal(t, "Classic Roman dish", m.description)
				require.Equal(t, "Add @eggs{3}.", m.source)
				require.Equal(t, &servings, m.servings)
				require.Equal(t, &prepTime, m.prepTimeMinutes)
				require.Equal(t, &cookTime, m.cookTimeMinutes)
				require.Equal(t, "https://example.com", m.sourceURL)
				require.Equal(t, []string{"italian", "pasta"}, m.tags)
			},
		},
		{
			name: "minimal valid build",
			setup: func() *Builder {
				return NewBuilder().
					SetTitle("Simple Recipe").
					SetSource("Add @salt{1%tsp}.")
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				require.Equal(t, "Simple Recipe", m.title)
				require.Equal(t, "Add @salt{1%tsp}.", m.source)
				require.Nil(t, m.servings)
				require.Nil(t, m.prepTimeMinutes)
				require.Empty(t, m.description)
			},
		},
		{
			name: "empty title returns ErrTitleRequired",
			setup: func() *Builder {
				return NewBuilder().
					SetTitle("").
					SetSource("Add @salt{1%tsp}.")
			},
			wantErr: ErrTitleRequired,
			assertModel: func(t *testing.T, m Model) {},
		},
		{
			name: "empty source returns ErrSourceRequired",
			setup: func() *Builder {
				return NewBuilder().
					SetTitle("A Recipe").
					SetSource("")
			},
			wantErr: ErrSourceRequired,
			assertModel: func(t *testing.T, m Model) {},
		},
		{
			name: "missing source returns ErrSourceRequired",
			setup: func() *Builder {
				return NewBuilder().
					SetTitle("A Recipe")
			},
			wantErr: ErrSourceRequired,
			assertModel: func(t *testing.T, m Model) {},
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
			tt.assertModel(t, model)
		})
	}
}
