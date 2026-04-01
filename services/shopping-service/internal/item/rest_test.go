package item

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransform(t *testing.T) {
	catID := uuid.New()
	qty := "1 cup"
	catName := "Dairy"
	sortOrder := 3
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name     string
		model    func() Model
		validate func(t *testing.T, r RestModel)
	}{
		{
			name: "maps all fields",
			model: func() Model {
				m, _ := NewBuilder().
					SetId(uuid.New()).
					SetListID(uuid.New()).
					SetName("Milk").
					SetQuantity(&qty).
					SetCategoryID(&catID).
					SetCategoryName(&catName).
					SetCategorySortOrder(&sortOrder).
					SetChecked(true).
					SetPosition(2).
					SetCreatedAt(now).
					SetUpdatedAt(now).
					Build()
				return m
			},
			validate: func(t *testing.T, r RestModel) {
				assert.Equal(t, "Milk", r.Name)
				assert.Equal(t, &qty, r.Quantity)
				assert.Equal(t, &catID, r.CategoryId)
				assert.Equal(t, &catName, r.CategoryName)
				assert.Equal(t, &sortOrder, r.CategorySortOrder)
				assert.True(t, r.Checked)
				assert.Equal(t, 2, r.Position)
				assert.Equal(t, now, r.CreatedAt)
				assert.Equal(t, now, r.UpdatedAt)
			},
		},
		{
			name: "nil optional fields",
			model: func() Model {
				m, _ := NewBuilder().SetName("Bread").Build()
				return m
			},
			validate: func(t *testing.T, r RestModel) {
				assert.Equal(t, "Bread", r.Name)
				assert.Nil(t, r.Quantity)
				assert.Nil(t, r.CategoryId)
				assert.Nil(t, r.CategoryName)
				assert.Nil(t, r.CategorySortOrder)
				assert.False(t, r.Checked)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Transform(tc.model())
			require.NoError(t, err)
			tc.validate(t, r)
		})
	}
}

func TestRestModel_JSONAPIInterface(t *testing.T) {
	tests := []struct {
		name   string
		action func() (string, error)
		expect string
	}{
		{
			name: "GetName returns shopping-items",
			action: func() (string, error) {
				return RestModel{}.GetName(), nil
			},
			expect: "shopping-items",
		},
		{
			name: "GetID returns UUID string",
			action: func() (string, error) {
				id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
				return RestModel{Id: id}.GetID(), nil
			},
			expect: "11111111-1111-1111-1111-111111111111",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.action()
			require.NoError(t, err)
			assert.Equal(t, tc.expect, got)
		})
	}
}

func TestRestModel_SetID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid UUID", input: uuid.New().String(), wantErr: false},
		{name: "invalid UUID", input: "bad", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &RestModel{}
			err := r.SetID(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.input, r.Id.String())
			}
		})
	}
}

func TestTransformSlice(t *testing.T) {
	tests := []struct {
		name      string
		models    []Model
		wantCount int
	}{
		{
			name:      "empty slice",
			models:    []Model{},
			wantCount: 0,
		},
		{
			name: "multiple items",
			models: func() []Model {
				models := make([]Model, 2)
				for i := range models {
					m, _ := NewBuilder().
						SetId(uuid.New()).
						SetName("Item " + string(rune('A'+i))).
						Build()
					models[i] = m
				}
				return models
			}(),
			wantCount: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := TransformSlice(tc.models)
			require.NoError(t, err)
			assert.Len(t, result, tc.wantCount)
			for i, r := range result {
				assert.Equal(t, tc.models[i].Id(), r.Id)
				assert.Equal(t, tc.models[i].Name(), r.Name)
			}
		})
	}
}
