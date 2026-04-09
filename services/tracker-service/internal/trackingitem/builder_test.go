package trackingitem

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Build_Valid(t *testing.T) {
	m, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(uuid.New()).
		SetUserID(uuid.New()).
		SetName("Running").
		SetScaleType("sentiment").
		SetColor("blue").
		SetSortOrder(1).
		Build()
	require.NoError(t, err)
	assert.Equal(t, "Running", m.Name())
	assert.Equal(t, "sentiment", m.ScaleType())
	assert.Equal(t, "blue", m.Color())
}

func TestBuilder_Build_NameRequired(t *testing.T) {
	_, err := NewBuilder().
		SetScaleType("sentiment").
		SetColor("blue").
		Build()
	assert.ErrorIs(t, err, ErrNameRequired)
}

func TestBuilder_Build_NameTooLong(t *testing.T) {
	longName := make([]byte, 101)
	for i := range longName {
		longName[i] = 'a'
	}
	_, err := NewBuilder().
		SetName(string(longName)).
		SetScaleType("sentiment").
		SetColor("blue").
		Build()
	assert.ErrorIs(t, err, ErrNameTooLong)
}

func TestBuilder_Build_InvalidScaleType(t *testing.T) {
	_, err := NewBuilder().
		SetName("Test").
		SetScaleType("invalid").
		SetColor("blue").
		Build()
	assert.ErrorIs(t, err, ErrInvalidScaleType)
}

func TestBuilder_Build_InvalidColor(t *testing.T) {
	_, err := NewBuilder().
		SetName("Test").
		SetScaleType("sentiment").
		SetColor("neon").
		Build()
	assert.ErrorIs(t, err, ErrInvalidColor)
}

func TestBuilder_Build_NegativeSortOrder(t *testing.T) {
	_, err := NewBuilder().
		SetName("Test").
		SetScaleType("sentiment").
		SetColor("blue").
		SetSortOrder(-1).
		Build()
	assert.ErrorIs(t, err, ErrInvalidSortOrder)
}

func TestBuilder_Build_RangeRequiresConfig(t *testing.T) {
	_, err := NewBuilder().
		SetName("Test").
		SetScaleType("range").
		SetColor("blue").
		Build()
	assert.ErrorIs(t, err, ErrRangeConfigRequired)
}

func TestBuilder_Build_RangeInvalidConfig(t *testing.T) {
	cfg, _ := json.Marshal(RangeConfig{Min: 100, Max: 50})
	_, err := NewBuilder().
		SetName("Test").
		SetScaleType("range").
		SetColor("blue").
		SetScaleConfig(cfg).
		Build()
	assert.ErrorIs(t, err, ErrInvalidRangeConfig)
}

func TestBuilder_Build_RangeValidConfig(t *testing.T) {
	cfg, _ := json.Marshal(RangeConfig{Min: 0, Max: 100})
	m, err := NewBuilder().
		SetName("Sleep Quality").
		SetScaleType("range").
		SetColor("green").
		SetScaleConfig(cfg).
		Build()
	require.NoError(t, err)
	assert.Equal(t, "range", m.ScaleType())
}

func TestValidateSchedule(t *testing.T) {
	assert.NoError(t, ValidateSchedule([]int{0, 1, 2, 3, 4, 5, 6}))
	assert.NoError(t, ValidateSchedule([]int{}))
	assert.ErrorIs(t, ValidateSchedule([]int{7}), ErrInvalidScheduleDay)
	assert.ErrorIs(t, ValidateSchedule([]int{-1}), ErrInvalidScheduleDay)
}
