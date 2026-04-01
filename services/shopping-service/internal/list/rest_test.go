package list

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/shopping-service/internal/item"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransform(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	archivedAt := now.Add(-time.Hour)

	tests := []struct {
		name     string
		model    func() Model
		validate func(t *testing.T, r RestModel)
	}{
		{
			name: "maps all fields for archived list",
			model: func() Model {
				m, _ := NewBuilder().
					SetId(uuid.New()).
					SetTenantID(uuid.New()).
					SetHouseholdID(uuid.New()).
					SetName("Weekly Groceries").
					SetStatus("archived").
					SetArchivedAt(&archivedAt).
					SetCreatedBy(uuid.New()).
					SetCreatedAt(now).
					SetUpdatedAt(now).
					Build()
				return m.WithCounts(8, 3)
			},
			validate: func(t *testing.T, r RestModel) {
				assert.Equal(t, "Weekly Groceries", r.Name)
				assert.Equal(t, "archived", r.Status)
				assert.Equal(t, 8, r.ItemCount)
				assert.Equal(t, 3, r.CheckedCount)
				assert.NotNil(t, r.ArchivedAt)
				assert.Equal(t, archivedAt, *r.ArchivedAt)
				assert.Equal(t, now, r.CreatedAt)
				assert.Equal(t, now, r.UpdatedAt)
			},
		},
		{
			name: "active list with defaults",
			model: func() Model {
				m, _ := NewBuilder().SetName("Shopping").Build()
				return m
			},
			validate: func(t *testing.T, r RestModel) {
				assert.Equal(t, "active", r.Status)
				assert.Nil(t, r.ArchivedAt)
				assert.Equal(t, 0, r.ItemCount)
				assert.Equal(t, 0, r.CheckedCount)
				assert.Empty(t, r.Items)
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
		action func() string
		expect string
	}{
		{
			name:   "GetName returns shopping-lists",
			action: func() string { return RestModel{}.GetName() },
			expect: "shopping-lists",
		},
		{
			name: "GetID returns UUID string",
			action: func() string {
				id := uuid.MustParse("22222222-2222-2222-2222-222222222222")
				return RestModel{Id: id}.GetID()
			},
			expect: "22222222-2222-2222-2222-222222222222",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.action())
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
		{name: "invalid UUID", input: "invalid", wantErr: true},
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

func TestTransformWithItems(t *testing.T) {
	tests := []struct {
		name      string
		items     []item.RestModel
		wantCount int
	}{
		{
			name: "with items",
			items: []item.RestModel{
				{Id: uuid.New(), Name: "Milk"},
				{Id: uuid.New(), Name: "Bread"},
			},
			wantCount: 2,
		},
		{
			name:      "empty items",
			items:     []item.RestModel{},
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, _ := NewBuilder().SetName("Groceries").Build()
			r, err := TransformWithItems(m, tc.items)
			require.NoError(t, err)
			assert.Equal(t, "Groceries", r.Name)
			assert.Len(t, r.Items, tc.wantCount)
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
			name: "multiple lists",
			models: func() []Model {
				models := make([]Model, 3)
				for i := range models {
					m, _ := NewBuilder().
						SetId(uuid.New()).
						SetName("List " + string(rune('A'+i))).
						Build()
					models[i] = m
				}
				return models
			}(),
			wantCount: 3,
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
