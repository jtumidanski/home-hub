package reminder

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeletedAtRoundTrip(t *testing.T) {
	now := time.Now().UTC()
	e := Entity{
		Id:           uuid.New(),
		TenantId:     uuid.New(),
		HouseholdId:  uuid.New(),
		Title:        "x",
		ScheduledFor: now,
		DeletedAt:    &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	m, err := Make(e)
	require.NoError(t, err)
	require.NotNil(t, m.DeletedAt())
	require.True(t, m.IsDeleted())
	require.Equal(t, e, m.ToEntity())

	// Nil case: a live reminder is not deleted.
	e2 := e
	e2.DeletedAt = nil
	m2, err := Make(e2)
	require.NoError(t, err)
	require.Nil(t, m2.DeletedAt())
	require.False(t, m2.IsDeleted())
}
