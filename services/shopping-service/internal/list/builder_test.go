package list

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_Build_RequiresName(t *testing.T) {
	_, err := NewBuilder().Build()
	assert.ErrorIs(t, err, ErrNameRequired)
}

func TestBuilder_Build_RejectsLongName(t *testing.T) {
	_, err := NewBuilder().
		SetName(strings.Repeat("a", 256)).
		Build()
	assert.ErrorIs(t, err, ErrNameTooLong)
}

func TestBuilder_Build_Success(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()

	m, err := NewBuilder().
		SetId(id).
		SetTenantID(tenantID).
		SetHouseholdID(householdID).
		SetName("Weekly Groceries").
		SetCreatedBy(userID).
		Build()

	assert.NoError(t, err)
	assert.Equal(t, id, m.Id())
	assert.Equal(t, tenantID, m.TenantID())
	assert.Equal(t, householdID, m.HouseholdID())
	assert.Equal(t, "Weekly Groceries", m.Name())
	assert.Equal(t, "active", m.Status())
	assert.Equal(t, userID, m.CreatedBy())
	assert.Nil(t, m.ArchivedAt())
	assert.False(t, m.IsArchived())
}

func TestBuilder_Build_ArchivedStatus(t *testing.T) {
	m, err := NewBuilder().
		SetName("Old List").
		SetStatus("archived").
		Build()

	assert.NoError(t, err)
	assert.Equal(t, "archived", m.Status())
	assert.True(t, m.IsArchived())
}
