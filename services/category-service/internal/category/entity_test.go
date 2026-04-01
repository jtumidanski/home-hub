package category

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMake_Success(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	now := time.Now().Truncate(time.Second)

	e := Entity{
		Id:        id,
		TenantId:  tenantID,
		Name:      "Produce",
		SortOrder: 1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	m, err := Make(e)
	assert.NoError(t, err)
	assert.Equal(t, id, m.Id())
	assert.Equal(t, tenantID, m.TenantID())
	assert.Equal(t, "Produce", m.Name())
	assert.Equal(t, 1, m.SortOrder())
	assert.Equal(t, now, m.CreatedAt())
	assert.Equal(t, now, m.UpdatedAt())
}

func TestMake_ZeroSortOrder(t *testing.T) {
	e := Entity{
		Name:      "Other",
		SortOrder: 0,
	}

	m, err := Make(e)
	assert.NoError(t, err)
	assert.Equal(t, 0, m.SortOrder())
}

func TestMake_EmptyName_ReturnsError(t *testing.T) {
	e := Entity{
		Name: "",
	}

	_, err := Make(e)
	assert.ErrorIs(t, err, ErrNameRequired)
}

func TestMake_NameTooLong_ReturnsError(t *testing.T) {
	e := Entity{
		Name: string(make([]byte, 101)),
	}

	_, err := Make(e)
	assert.ErrorIs(t, err, ErrNameTooLong)
}

func TestMake_NegativeSortOrder_ReturnsError(t *testing.T) {
	e := Entity{
		Name:      "Test",
		SortOrder: -1,
	}

	_, err := Make(e)
	assert.ErrorIs(t, err, ErrInvalidSortOrder)
}

func TestModel_ToEntity_RoundTrip(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	now := time.Now().Truncate(time.Second)

	original := Entity{
		Id:        id,
		TenantId:  tenantID,
		Name:      "Dairy & Eggs",
		SortOrder: 3,
		CreatedAt: now,
		UpdatedAt: now,
	}

	m, err := Make(original)
	assert.NoError(t, err)

	roundTripped := m.ToEntity()
	assert.Equal(t, original.Id, roundTripped.Id)
	assert.Equal(t, original.TenantId, roundTripped.TenantId)
	assert.Equal(t, original.Name, roundTripped.Name)
	assert.Equal(t, original.SortOrder, roundTripped.SortOrder)
	assert.Equal(t, original.CreatedAt, roundTripped.CreatedAt)
	assert.Equal(t, original.UpdatedAt, roundTripped.UpdatedAt)
}
