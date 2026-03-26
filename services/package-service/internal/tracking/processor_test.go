package tracking

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	require.NoError(t, db.AutoMigrate(&Entity{}))
	return db
}

func newTestProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db, 25, nil)
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	m, err := p.Create(uuid.New(), uuid.New(), uuid.New(), CreateAttrs{
		TrackingNumber: "1Z999AA10123456784",
		Carrier:        CarrierUPS,
		Label:          "My Package",
		Notes:          "Notes here",
		Private:        true,
	})
	require.NoError(t, err)
	require.Equal(t, "1Z999AA10123456784", m.TrackingNumber())
	require.Equal(t, CarrierUPS, m.Carrier())
	require.Equal(t, StatusPreTransit, m.Status())
	require.True(t, m.Private())
	require.NotNil(t, m.Label())
	require.Equal(t, "My Package", *m.Label())
}

func TestCreate_DuplicateTrackingNumber(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	householdID := uuid.New()
	tenantID := uuid.New()

	_, err := p.Create(tenantID, householdID, uuid.New(), CreateAttrs{
		TrackingNumber: "DUPE123",
		Carrier:        CarrierFedEx,
	})
	require.NoError(t, err)

	_, err = p.Create(tenantID, householdID, uuid.New(), CreateAttrs{
		TrackingNumber: "DUPE123",
		Carrier:        CarrierFedEx,
	})
	require.ErrorIs(t, err, ErrDuplicate)
}

func TestCreate_HouseholdLimit(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db, 2, nil)
	householdID := uuid.New()
	tenantID := uuid.New()

	_, err := p.Create(tenantID, householdID, uuid.New(), CreateAttrs{
		TrackingNumber: "PKG1",
		Carrier:        CarrierUSPS,
	})
	require.NoError(t, err)

	_, err = p.Create(tenantID, householdID, uuid.New(), CreateAttrs{
		TrackingNumber: "PKG2",
		Carrier:        CarrierUSPS,
	})
	require.NoError(t, err)

	_, err = p.Create(tenantID, householdID, uuid.New(), CreateAttrs{
		TrackingNumber: "PKG3",
		Carrier:        CarrierUSPS,
	})
	require.ErrorIs(t, err, ErrHouseholdLimit)
}

func TestCreate_ValidationErrors(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	tests := []struct {
		name    string
		attrs   CreateAttrs
		wantErr error
	}{
		{
			name:    "empty tracking number",
			attrs:   CreateAttrs{TrackingNumber: "", Carrier: CarrierUPS},
			wantErr: ErrTrackingNumberRequired,
		},
		{
			name:    "empty carrier",
			attrs:   CreateAttrs{TrackingNumber: "ABC123", Carrier: ""},
			wantErr: ErrCarrierRequired,
		},
		{
			name:    "invalid carrier",
			attrs:   CreateAttrs{TrackingNumber: "ABC123", Carrier: "dhl"},
			wantErr: ErrInvalidCarrier,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := p.Create(uuid.New(), uuid.New(), uuid.New(), tc.attrs)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestGet(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	m, err := p.Create(uuid.New(), uuid.New(), uuid.New(), CreateAttrs{
		TrackingNumber: "GET123",
		Carrier:        CarrierUSPS,
	})
	require.NoError(t, err)

	fetched, err := p.Get(m.Id())
	require.NoError(t, err)
	require.Equal(t, m.Id(), fetched.Id())
	require.Equal(t, "GET123", fetched.TrackingNumber())
}

func TestGet_NotFound(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	_, err := p.Get(uuid.New())
	require.ErrorIs(t, err, ErrNotFound)
}

func TestUpdate(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	m, err := p.Create(uuid.New(), uuid.New(), userID, CreateAttrs{
		TrackingNumber: "UPD123",
		Carrier:        CarrierUPS,
	})
	require.NoError(t, err)

	newLabel := "Updated"
	newNotes := "Updated notes"
	newCarrier := CarrierFedEx
	newPrivate := true

	updated, err := p.Update(m.Id(), userID, UpdateAttrs{
		Label:   &newLabel,
		Notes:   &newNotes,
		Carrier: &newCarrier,
		Private: &newPrivate,
	})
	require.NoError(t, err)
	require.Equal(t, "Updated", *updated.Label())
	require.Equal(t, "Updated notes", *updated.Notes())
	require.Equal(t, CarrierFedEx, updated.Carrier())
	require.True(t, updated.Private())
}

func TestUpdate_NotOwner(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	m, err := p.Create(uuid.New(), uuid.New(), uuid.New(), CreateAttrs{
		TrackingNumber: "OWN123",
		Carrier:        CarrierUPS,
	})
	require.NoError(t, err)

	label := "Hacked"
	_, err = p.Update(m.Id(), uuid.New(), UpdateAttrs{Label: &label})
	require.ErrorIs(t, err, ErrNotOwner)
}

func TestUpdate_InvalidCarrier(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	m, err := p.Create(uuid.New(), uuid.New(), userID, CreateAttrs{
		TrackingNumber: "CAR123",
		Carrier:        CarrierUPS,
	})
	require.NoError(t, err)

	bad := "dhl"
	_, err = p.Update(m.Id(), userID, UpdateAttrs{Carrier: &bad})
	require.ErrorIs(t, err, ErrInvalidCarrier)
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	userID := uuid.New()

	m, err := p.Create(uuid.New(), uuid.New(), userID, CreateAttrs{
		TrackingNumber: "DEL123",
		Carrier:        CarrierFedEx,
	})
	require.NoError(t, err)

	err = p.Delete(m.Id(), userID)
	require.NoError(t, err)

	_, err = p.Get(m.Id())
	require.ErrorIs(t, err, ErrNotFound)
}

func TestDelete_NotOwner(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	m, err := p.Create(uuid.New(), uuid.New(), uuid.New(), CreateAttrs{
		TrackingNumber: "DELOWN123",
		Carrier:        CarrierUSPS,
	})
	require.NoError(t, err)

	err = p.Delete(m.Id(), uuid.New())
	require.ErrorIs(t, err, ErrNotOwner)
}

func TestArchive_And_Unarchive(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	m, err := p.Create(uuid.New(), uuid.New(), uuid.New(), CreateAttrs{
		TrackingNumber: "ARC123",
		Carrier:        CarrierUPS,
	})
	require.NoError(t, err)

	archived, err := p.Archive(m.Id())
	require.NoError(t, err)
	require.Equal(t, StatusArchived, archived.Status())
	require.NotNil(t, archived.ArchivedAt())

	unarchived, err := p.Unarchive(m.Id())
	require.NoError(t, err)
	require.Equal(t, StatusDelivered, unarchived.Status())
	require.Nil(t, unarchived.ArchivedAt())
}

func TestUnarchive_NotArchived(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	m, err := p.Create(uuid.New(), uuid.New(), uuid.New(), CreateAttrs{
		TrackingNumber: "NOTARC123",
		Carrier:        CarrierUSPS,
	})
	require.NoError(t, err)

	_, err = p.Unarchive(m.Id())
	require.ErrorIs(t, err, ErrNotArchived)
}

func TestRefresh_Cooldown(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	m, err := p.Create(uuid.New(), uuid.New(), uuid.New(), CreateAttrs{
		TrackingNumber: "REF123",
		Carrier:        CarrierFedEx,
	})
	require.NoError(t, err)

	// Simulate a recent poll by setting last_polled_at directly
	now := time.Now().UTC()
	db.Model(&Entity{}).Where("id = ?", m.Id()).Update("last_polled_at", now)

	// Refresh should fail because last_polled_at is within cooldown
	_, err = p.Refresh(m.Id())
	require.ErrorIs(t, err, ErrRefreshTooSoon)
}

func TestList(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	householdID := uuid.New()
	tenantID := uuid.New()

	_, err := p.Create(tenantID, householdID, uuid.New(), CreateAttrs{
		TrackingNumber: "LIST1",
		Carrier:        CarrierUSPS,
	})
	require.NoError(t, err)

	_, err = p.Create(tenantID, householdID, uuid.New(), CreateAttrs{
		TrackingNumber: "LIST2",
		Carrier:        CarrierUPS,
	})
	require.NoError(t, err)

	models, err := p.List(householdID, false, nil, false, "")
	require.NoError(t, err)
	require.Len(t, models, 2)
}

func TestSummary(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	householdID := uuid.New()
	tenantID := uuid.New()

	_, err := p.Create(tenantID, householdID, uuid.New(), CreateAttrs{
		TrackingNumber: "SUM1",
		Carrier:        CarrierUSPS,
	})
	require.NoError(t, err)

	result, err := p.Summary(householdID)
	require.NoError(t, err)
	require.Equal(t, int64(1), result.InTransitCount)
	require.Equal(t, int64(0), result.ExceptionCount)
}

func TestModel_IsPolling(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{StatusPreTransit, true},
		{StatusInTransit, true},
		{StatusOutForDelivery, true},
		{StatusDelivered, false},
		{StatusException, false},
		{StatusStale, false},
		{StatusArchived, false},
	}

	for _, tc := range tests {
		t.Run(tc.status, func(t *testing.T) {
			m, err := NewBuilder().
				SetTrackingNumber("TEST").
				SetCarrier(CarrierUSPS).
				SetStatus(tc.status).
				Build()
			require.NoError(t, err)
			require.Equal(t, tc.want, m.IsPolling())
		})
	}
}
