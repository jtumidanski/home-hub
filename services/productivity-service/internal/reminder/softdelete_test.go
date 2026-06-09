package reminder

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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

// seedDeleted inserts a reminder whose deleted_at is set (soft-deleted) and
// returns its id. Uses the raw entity so it bypasses the read-path filters.
func seedDeleted(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	now := time.Now().UTC()
	id := uuid.New()
	require.NoError(t, db.Create(&Entity{
		Id:           id,
		TenantId:     uuid.New(),
		HouseholdId:  uuid.New(),
		Title:        "ghost",
		ScheduledFor: now.Add(-1 * time.Hour),
		DeletedAt:    &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}).Error)
	return id
}

func TestSoftDeletedHiddenFromReads(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	id := seedDeleted(t, db)

	// getByID via the processor returns not-found.
	_, err := p.ByIDProvider(id)()
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound), "ByIDProvider err = %v", err)

	// getAll excludes it.
	all, err := p.AllProvider()()
	require.NoError(t, err)
	require.Empty(t, all)

	// All three counts exclude it (the seeded row is past-due, undismissed,
	// unsnoozed — it would count as due-now if not filtered).
	due, err := p.DueNowCount()
	require.NoError(t, err)
	require.Equal(t, int64(0), due)
	up, err := p.UpcomingCount()
	require.NoError(t, err)
	require.Equal(t, int64(0), up)
	sn, err := p.SnoozedCount()
	require.NoError(t, err)
	require.Equal(t, int64(0), sn)
}

func TestSoftDeletedNotFoundOnMutations(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	id := seedDeleted(t, db)

	_, err := p.Update(id, "new", "", time.Now().UTC(), nil)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound), "Update err = %v", err)

	require.True(t, errors.Is(p.Dismiss(id), gorm.ErrRecordNotFound), "Dismiss")

	_, err = p.Snooze(id, 10)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound), "Snooze err = %v", err)

	require.True(t, errors.Is(p.Delete(id), gorm.ErrRecordNotFound), "Delete")
}
