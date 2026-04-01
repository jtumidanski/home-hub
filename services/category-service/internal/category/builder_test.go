package category

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
		SetName(strings.Repeat("a", 101)).
		Build()
	assert.ErrorIs(t, err, ErrNameTooLong)
}

func TestBuilder_Build_RejectsNegativeSortOrder(t *testing.T) {
	_, err := NewBuilder().
		SetName("Test").
		SetSortOrder(-1).
		Build()
	assert.ErrorIs(t, err, ErrInvalidSortOrder)
}

func TestBuilder_Build_Success(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()

	m, err := NewBuilder().
		SetId(id).
		SetTenantID(tenantID).
		SetName("Produce").
		SetSortOrder(1).
		Build()

	assert.NoError(t, err)
	assert.Equal(t, id, m.Id())
	assert.Equal(t, tenantID, m.TenantID())
	assert.Equal(t, "Produce", m.Name())
	assert.Equal(t, 1, m.SortOrder())
}

func TestBuilder_Build_MaxLengthName(t *testing.T) {
	m, err := NewBuilder().
		SetName(strings.Repeat("a", 100)).
		Build()
	assert.NoError(t, err)
	assert.Len(t, m.Name(), 100)
}

func TestBuilder_Build_ZeroSortOrder(t *testing.T) {
	m, err := NewBuilder().
		SetName("Test").
		SetSortOrder(0).
		Build()
	assert.NoError(t, err)
	assert.Equal(t, 0, m.SortOrder())
}
