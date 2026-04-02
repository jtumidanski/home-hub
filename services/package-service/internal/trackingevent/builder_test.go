package trackingevent

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Build(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	pkgID := uuid.New()
	loc := "New York, NY"
	raw := "DELIVERED"

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
					SetPackageID(pkgID).
					SetTimestamp(now).
					SetStatus("delivered").
					SetDescription("Package delivered").
					SetLocation(&loc).
					SetRawStatus(&raw).
					SetCreatedAt(now)
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				require.Equal(t, id, m.Id())
				require.Equal(t, pkgID, m.PackageID())
				require.Equal(t, now, m.Timestamp())
				require.Equal(t, "delivered", m.Status())
				require.Equal(t, "Package delivered", m.Description())
				require.NotNil(t, m.Location())
				require.Equal(t, "New York, NY", *m.Location())
				require.NotNil(t, m.RawStatus())
				require.Equal(t, "DELIVERED", *m.RawStatus())
				require.Equal(t, now, m.CreatedAt())
			},
		},
		{
			name: "valid build with required fields only",
			setup: func() *Builder {
				return NewBuilder().
					SetStatus("in_transit").
					SetDescription("In transit")
			},
			wantErr: nil,
			assertModel: func(t *testing.T, m Model) {
				t.Helper()
				require.Equal(t, "in_transit", m.Status())
				require.Equal(t, "In transit", m.Description())
				require.Nil(t, m.Location())
				require.Nil(t, m.RawStatus())
			},
		},
		{
			name: "empty description returns error",
			setup: func() *Builder {
				return NewBuilder().
					SetStatus("delivered").
					SetDescription("")
			},
			wantErr: ErrDescriptionRequired,
			assertModel: func(t *testing.T, m Model) {},
		},
		{
			name: "empty status returns error",
			setup: func() *Builder {
				return NewBuilder().
					SetStatus("").
					SetDescription("Something happened")
			},
			wantErr: ErrStatusRequired,
			assertModel: func(t *testing.T, m Model) {},
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

func TestMake_RoundTrip(t *testing.T) {
	now := time.Now()
	loc := "Chicago, IL"
	raw := "IN_TRANSIT"

	e := Entity{
		Id:          uuid.New(),
		PackageId:   uuid.New(),
		Timestamp:   now,
		Status:      "in_transit",
		Description: "Package in transit",
		Location:    &loc,
		RawStatus:   &raw,
		CreatedAt:   now,
	}

	m, err := Make(e)
	require.NoError(t, err)

	require.Equal(t, e.Id, m.Id())
	require.Equal(t, e.PackageId, m.PackageID())
	require.Equal(t, e.Timestamp, m.Timestamp())
	require.Equal(t, e.Status, m.Status())
	require.Equal(t, e.Description, m.Description())
	require.Equal(t, e.Location, m.Location())
	require.Equal(t, e.RawStatus, m.RawStatus())
	require.Equal(t, e.CreatedAt, m.CreatedAt())

	roundTripped := m.ToEntity()
	require.Equal(t, e.Id, roundTripped.Id)
	require.Equal(t, e.PackageId, roundTripped.PackageId)
	require.Equal(t, e.Status, roundTripped.Status)
	require.Equal(t, e.Description, roundTripped.Description)
}
