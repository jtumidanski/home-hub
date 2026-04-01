package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsArchived(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{name: "active is not archived", status: "active", want: false},
		{name: "archived is archived", status: "archived", want: true},
		{name: "default status is not archived", status: "", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := NewBuilder().SetName("Test")
			if tc.status != "" {
				b.SetStatus(tc.status)
			}
			m, err := b.Build()
			require.NoError(t, err)
			assert.Equal(t, tc.want, m.IsArchived())
		})
	}
}

func TestWithCounts(t *testing.T) {
	tests := []struct {
		name         string
		itemCount    int
		checkedCount int
	}{
		{name: "nonzero counts", itemCount: 10, checkedCount: 5},
		{name: "zero counts", itemCount: 0, checkedCount: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := NewBuilder().SetName("Groceries").Build()
			require.NoError(t, err)

			updated := m.WithCounts(tc.itemCount, tc.checkedCount)
			assert.Equal(t, tc.itemCount, updated.ItemCount())
			assert.Equal(t, tc.checkedCount, updated.CheckedCount())
		})
	}
}

func TestWithCounts_DoesNotMutateOriginal(t *testing.T) {
	m, err := NewBuilder().SetName("Groceries").Build()
	require.NoError(t, err)

	_ = m.WithCounts(10, 5)
	assert.Equal(t, 0, m.ItemCount())
	assert.Equal(t, 0, m.CheckedCount())
}
