package category

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTransform(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	now := time.Now().Truncate(time.Second)

	m, err := NewBuilder().
		SetId(id).
		SetTenantID(tenantID).
		SetName("Produce").
		SetSortOrder(1).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Build()
	assert.NoError(t, err)

	r, err := Transform(m)
	assert.NoError(t, err)
	assert.Equal(t, id, r.Id)
	assert.Equal(t, "Produce", r.Name)
	assert.Equal(t, 1, r.SortOrder)
	assert.Equal(t, now, r.CreatedAt)
	assert.Equal(t, now, r.UpdatedAt)
}

func TestRestModel_JSONAPIInterface(t *testing.T) {
	tests := []struct {
		name     string
		run      func(t *testing.T)
	}{
		{
			name: "GetName returns categories",
			run: func(t *testing.T) {
				r := RestModel{}
				assert.Equal(t, "categories", r.GetName())
			},
		},
		{
			name: "GetID returns UUID string",
			run: func(t *testing.T) {
				id := uuid.New()
				r := RestModel{Id: id}
				assert.Equal(t, id.String(), r.GetID())
			},
		},
		{
			name: "SetID parses valid UUID",
			run: func(t *testing.T) {
				id := uuid.New()
				r := &RestModel{}
				err := r.SetID(id.String())
				assert.NoError(t, err)
				assert.Equal(t, id, r.Id)
			},
		},
		{
			name: "SetID rejects invalid UUID",
			run: func(t *testing.T) {
				r := &RestModel{}
				err := r.SetID("not-a-uuid")
				assert.Error(t, err)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.run(t)
		})
	}
}

func TestTransformSlice(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name     string
		models   []Model
		wantLen  int
	}{
		{
			name:    "empty slice",
			models:  []Model{},
			wantLen: 0,
		},
		{
			name: "multiple models",
			models: func() []Model {
				models := make([]Model, 3)
				for i := range models {
					m, _ := NewBuilder().
						SetId(uuid.New()).
						SetName("Category " + string(rune('A'+i))).
						SetSortOrder(i).
						SetCreatedAt(now).
						SetUpdatedAt(now).
						Build()
					models[i] = m
				}
				return models
			}(),
			wantLen: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := TransformSlice(tc.models)
			assert.NoError(t, err)
			assert.Len(t, result, tc.wantLen)
			for i, r := range result {
				assert.Equal(t, tc.models[i].Id(), r.Id)
				assert.Equal(t, tc.models[i].Name(), r.Name)
				assert.Equal(t, tc.models[i].SortOrder(), r.SortOrder)
			}
		})
	}
}
