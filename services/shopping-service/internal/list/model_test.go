package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsArchived_Active(t *testing.T) {
	m, err := NewBuilder().
		SetName("Groceries").
		SetStatus("active").
		Build()
	assert.NoError(t, err)
	assert.False(t, m.IsArchived())
}

func TestIsArchived_Archived(t *testing.T) {
	m, err := NewBuilder().
		SetName("Old List").
		SetStatus("archived").
		Build()
	assert.NoError(t, err)
	assert.True(t, m.IsArchived())
}

func TestIsArchived_DefaultStatus(t *testing.T) {
	m, err := NewBuilder().
		SetName("New List").
		Build()
	assert.NoError(t, err)
	assert.False(t, m.IsArchived())
}

func TestWithCounts(t *testing.T) {
	m, err := NewBuilder().
		SetName("Groceries").
		Build()
	assert.NoError(t, err)
	assert.Equal(t, 0, m.ItemCount())
	assert.Equal(t, 0, m.CheckedCount())

	updated := m.WithCounts(10, 5)
	assert.Equal(t, 10, updated.ItemCount())
	assert.Equal(t, 5, updated.CheckedCount())
}

func TestWithCounts_DoesNotMutateOriginal(t *testing.T) {
	m, err := NewBuilder().
		SetName("Groceries").
		Build()
	assert.NoError(t, err)

	_ = m.WithCounts(10, 5)
	assert.Equal(t, 0, m.ItemCount())
	assert.Equal(t, 0, m.CheckedCount())
}

func TestWithCounts_ZeroCounts(t *testing.T) {
	m, err := NewBuilder().
		SetName("Empty List").
		Build()
	assert.NoError(t, err)

	updated := m.WithCounts(0, 0)
	assert.Equal(t, 0, updated.ItemCount())
	assert.Equal(t, 0, updated.CheckedCount())
}
