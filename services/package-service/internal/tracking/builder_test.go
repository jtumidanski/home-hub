package tracking

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Build(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	label := "My Package"
	notes := "Some notes"
	eta := now.Add(48 * time.Hour)

	tests := []struct {
		name        string
		setup       func() *Builder
		wantErr     error
		assertModel func(t *testing.T, m Model)
	}{
		{
			name: "valid build with all fields",
			setup: func() *Builder {
				return NewBuilder().
					SetId(id).
					SetTenantID(tenantID).
					SetHouseholdID(householdID).
					SetUserID(userID).
					SetTrackingNumber("1Z999AA10123456784").
					SetCarrier(CarrierUPS).
					SetLabel(&label).
					SetNotes(&notes).
					SetStatus(StatusInTransit).
					SetPrivate(true).
					SetEstimatedDelivery(&eta).
					SetCreatedAt(now).
					SetUpdatedAt(now)
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				require.Equal(t, id, m.id)
				require.Equal(t, tenantID, m.tenantID)
				require.Equal(t, householdID, m.householdID)
				require.Equal(t, userID, m.userID)
				require.Equal(t, "1Z999AA10123456784", m.trackingNumber)
				require.Equal(t, CarrierUPS, m.carrier)
				require.Equal(t, &label, m.label)
				require.Equal(t, &notes, m.notes)
				require.Equal(t, StatusInTransit, m.status)
				require.True(t, m.private)
				require.Equal(t, &eta, m.estimatedDelivery)
			},
		},
		{
			name: "default status is pre_transit",
			setup: func() *Builder {
				return NewBuilder().
					SetTrackingNumber("1Z999AA10123456784").
					SetCarrier(CarrierUPS)
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				require.Equal(t, StatusPreTransit, m.status)
			},
		},
		{
			name: "empty tracking number returns error",
			setup: func() *Builder {
				return NewBuilder().
					SetTrackingNumber("").
					SetCarrier(CarrierUPS)
			},
			wantErr: ErrTrackingNumberRequired,
			assertModel: func(t *testing.T, m Model) {},
		},
		{
			name: "empty carrier returns error",
			setup: func() *Builder {
				return NewBuilder().
					SetTrackingNumber("1Z999AA10123456784").
					SetCarrier("")
			},
			wantErr: ErrCarrierRequired,
			assertModel: func(t *testing.T, m Model) {},
		},
		{
			name: "invalid carrier returns error",
			setup: func() *Builder {
				return NewBuilder().
					SetTrackingNumber("1Z999AA10123456784").
					SetCarrier("dhl")
			},
			wantErr: ErrInvalidCarrier,
			assertModel: func(t *testing.T, m Model) {},
		},
		{
			name: "invalid status returns error",
			setup: func() *Builder {
				return NewBuilder().
					SetTrackingNumber("1Z999AA10123456784").
					SetCarrier(CarrierFedEx).
					SetStatus("unknown_status")
			},
			wantErr: ErrInvalidStatus,
			assertModel: func(t *testing.T, m Model) {},
		},
		{
			name: "all valid carriers accepted",
			setup: func() *Builder {
				return NewBuilder().
					SetTrackingNumber("1234").
					SetCarrier(CarrierUSPS)
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				require.Equal(t, CarrierUSPS, m.carrier)
			},
		},
		{
			name: "all valid statuses accepted",
			setup: func() *Builder {
				return NewBuilder().
					SetTrackingNumber("1234").
					SetCarrier(CarrierFedEx).
					SetStatus(StatusArchived)
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				require.Equal(t, StatusArchived, m.status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()
			model, err := builder.Build()

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			tt.assertModel(t, model)
		})
	}
}
