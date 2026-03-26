package tracking

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func makeTestModel(t *testing.T, userID uuid.UUID, private bool) Model {
	t.Helper()
	label := "Test Package"
	notes := "Some notes"
	eta := time.Now().Add(24 * time.Hour)
	m, err := NewBuilder().
		SetId(uuid.New()).
		SetTenantID(uuid.New()).
		SetHouseholdID(uuid.New()).
		SetUserID(userID).
		SetTrackingNumber("1Z999AA10123456784").
		SetCarrier(CarrierUPS).
		SetLabel(&label).
		SetNotes(&notes).
		SetStatus(StatusInTransit).
		SetPrivate(private).
		SetEstimatedDelivery(&eta).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)
	return m
}

func TestTransformWithPrivacy_OwnerSeesAll(t *testing.T) {
	userID := uuid.New()
	m := makeTestModel(t, userID, true)

	rm, err := TransformWithPrivacy(m, userID)
	require.NoError(t, err)

	require.True(t, rm.IsOwner)
	require.NotNil(t, rm.TrackingNumber)
	require.Equal(t, "1Z999AA10123456784", *rm.TrackingNumber)
	require.NotNil(t, rm.Label)
	require.Equal(t, "Test Package", *rm.Label)
	require.NotNil(t, rm.Notes)
	require.Equal(t, "Some notes", *rm.Notes)
	require.NotNil(t, rm.Status)
	require.Equal(t, StatusInTransit, *rm.Status)
}

func TestTransformWithPrivacy_NonOwnerRedacted(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	m := makeTestModel(t, ownerID, true)

	rm, err := TransformWithPrivacy(m, otherUserID)
	require.NoError(t, err)

	require.False(t, rm.IsOwner)
	require.Nil(t, rm.TrackingNumber)
	require.Nil(t, rm.Notes)
	require.Nil(t, rm.Status)
	require.Nil(t, rm.LastPolledAt)
	require.NotNil(t, rm.Label)
	require.Equal(t, "Package", *rm.Label)
	// Carrier and ID should still be visible
	require.Equal(t, CarrierUPS, rm.Carrier)
	require.Equal(t, m.Id(), rm.Id)
}

func TestTransformWithPrivacy_NonPrivateVisible(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	m := makeTestModel(t, ownerID, false)

	rm, err := TransformWithPrivacy(m, otherUserID)
	require.NoError(t, err)

	require.False(t, rm.IsOwner)
	require.NotNil(t, rm.TrackingNumber)
	require.NotNil(t, rm.Label)
	require.NotNil(t, rm.Notes)
	require.NotNil(t, rm.Status)
}

func TestTransformSliceWithPrivacy(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()

	m1 := makeTestModel(t, ownerID, true)
	m2 := makeTestModel(t, ownerID, false)

	models := []Model{m1, m2}
	results, err := TransformSliceWithPrivacy(models, otherUserID)
	require.NoError(t, err)
	require.Len(t, results, 2)

	// First is private, should be redacted
	require.Nil(t, results[0].TrackingNumber)
	// Second is not private, should be visible
	require.NotNil(t, results[1].TrackingNumber)
}
