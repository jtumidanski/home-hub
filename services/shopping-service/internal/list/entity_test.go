package list

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMake(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	now := time.Now().Truncate(time.Second)
	archivedAt := now.Add(-time.Hour)

	tests := []struct {
		name     string
		entity   Entity
		wantErr  error
		validate func(t *testing.T, m Model)
	}{
		{
			name: "active list",
			entity: Entity{
				Id: id, TenantId: tenantID, HouseholdId: householdID,
				Name: "Weekly Groceries", Status: "active",
				CreatedBy: userID, CreatedAt: now, UpdatedAt: now,
			},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, id, m.Id())
				assert.Equal(t, tenantID, m.TenantID())
				assert.Equal(t, householdID, m.HouseholdID())
				assert.Equal(t, "Weekly Groceries", m.Name())
				assert.Equal(t, "active", m.Status())
				assert.Nil(t, m.ArchivedAt())
				assert.Equal(t, userID, m.CreatedBy())
				assert.Equal(t, now, m.CreatedAt())
				assert.Equal(t, now, m.UpdatedAt())
			},
		},
		{
			name: "archived list",
			entity: Entity{
				Name: "Old List", Status: "archived", ArchivedAt: &archivedAt,
			},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, "archived", m.Status())
				assert.NotNil(t, m.ArchivedAt())
				assert.Equal(t, archivedAt, *m.ArchivedAt())
			},
		},
		{
			name:   "nil archived at",
			entity: Entity{Name: "Active List", Status: "active"},
			validate: func(t *testing.T, m Model) {
				assert.Nil(t, m.ArchivedAt())
			},
		},
		{
			name:    "empty name returns error",
			entity:  Entity{Name: ""},
			wantErr: ErrNameRequired,
		},
		{
			name:    "name too long returns error",
			entity:  Entity{Name: string(make([]byte, 256))},
			wantErr: ErrNameTooLong,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := Make(tc.entity)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			tc.validate(t, m)
		})
	}
}

func TestToEntity(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	now := time.Now().Truncate(time.Second)
	archivedAt := now.Add(-time.Hour)

	m, err := NewBuilder().
		SetId(id).
		SetTenantID(tenantID).
		SetHouseholdID(householdID).
		SetName("Groceries").
		SetStatus("archived").
		SetArchivedAt(&archivedAt).
		SetCreatedBy(userID).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		Build()
	require.NoError(t, err)

	e := m.ToEntity()
	assert.Equal(t, id, e.Id)
	assert.Equal(t, tenantID, e.TenantId)
	assert.Equal(t, householdID, e.HouseholdId)
	assert.Equal(t, "Groceries", e.Name)
	assert.Equal(t, "archived", e.Status)
	assert.NotNil(t, e.ArchivedAt)
	assert.Equal(t, archivedAt, *e.ArchivedAt)
	assert.Equal(t, userID, e.CreatedBy)
	assert.Equal(t, now, e.CreatedAt)
	assert.Equal(t, now, e.UpdatedAt)
}
