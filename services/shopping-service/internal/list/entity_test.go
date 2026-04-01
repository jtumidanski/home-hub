package list

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMake_ActiveList(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	now := time.Now().Truncate(time.Second)

	e := Entity{
		Id:          id,
		TenantId:    tenantID,
		HouseholdId: householdID,
		Name:        "Weekly Groceries",
		Status:      "active",
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	m, err := Make(e)
	assert.NoError(t, err)
	assert.Equal(t, id, m.Id())
	assert.Equal(t, tenantID, m.TenantID())
	assert.Equal(t, householdID, m.HouseholdID())
	assert.Equal(t, "Weekly Groceries", m.Name())
	assert.Equal(t, "active", m.Status())
	assert.Nil(t, m.ArchivedAt())
	assert.Equal(t, userID, m.CreatedBy())
	assert.Equal(t, now, m.CreatedAt())
	assert.Equal(t, now, m.UpdatedAt())
}

func TestMake_ArchivedList(t *testing.T) {
	archivedAt := time.Now().Truncate(time.Second)

	e := Entity{
		Name:       "Old List",
		Status:     "archived",
		ArchivedAt: &archivedAt,
	}

	m, err := Make(e)
	assert.NoError(t, err)
	assert.Equal(t, "archived", m.Status())
	assert.NotNil(t, m.ArchivedAt())
	assert.Equal(t, archivedAt, *m.ArchivedAt())
}

func TestMake_NilArchivedAt(t *testing.T) {
	e := Entity{
		Name:   "Active List",
		Status: "active",
	}

	m, err := Make(e)
	assert.NoError(t, err)
	assert.Nil(t, m.ArchivedAt())
}

func TestMake_EmptyName_ReturnsError(t *testing.T) {
	e := Entity{Name: ""}
	_, err := Make(e)
	assert.ErrorIs(t, err, ErrNameRequired)
}

func TestMake_NameTooLong_ReturnsError(t *testing.T) {
	e := Entity{Name: string(make([]byte, 256))}
	_, err := Make(e)
	assert.ErrorIs(t, err, ErrNameTooLong)
}
