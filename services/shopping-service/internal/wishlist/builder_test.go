package wishlist

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Build(t *testing.T) {
	longString := strings.Repeat("a", 256)

	tests := []struct {
		name     string
		build    func() (Model, error)
		wantErr  error
		validate func(t *testing.T, m Model)
	}{
		{
			name:    "requires name",
			build:   func() (Model, error) { return NewBuilder().Build() },
			wantErr: ErrNameRequired,
		},
		{
			name: "rejects long name",
			build: func() (Model, error) {
				return NewBuilder().SetName(longString).Build()
			},
			wantErr: ErrNameTooLong,
		},
		{
			name: "rejects long purchase location",
			build: func() (Model, error) {
				loc := longString
				return NewBuilder().SetName("Thing").SetPurchaseLocation(&loc).Build()
			},
			wantErr: ErrPurchaseLocationTooLong,
		},
		{
			name: "rejects invalid urgency",
			build: func() (Model, error) {
				return NewBuilder().SetName("Thing").SetUrgency("nope").Build()
			},
			wantErr: ErrInvalidUrgency,
		},
		{
			name: "defaults urgency to want",
			build: func() (Model, error) {
				return NewBuilder().SetName("Thing").Build()
			},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, UrgencyWant, m.Urgency())
				assert.Equal(t, 0, m.VoteCount())
				assert.Nil(t, m.PurchaseLocation())
			},
		},
		{
			name: "happy path with all fields",
			build: func() (Model, error) {
				loc := "Costco"
				return NewBuilder().
					SetId(uuid.New()).
					SetTenantID(uuid.New()).
					SetHouseholdID(uuid.New()).
					SetName("Standing desk").
					SetPurchaseLocation(&loc).
					SetUrgency(UrgencyMustHave).
					SetVoteCount(3).
					SetCreatedBy(uuid.New()).
					Build()
			},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, "Standing desk", m.Name())
				assert.Equal(t, UrgencyMustHave, m.Urgency())
				assert.Equal(t, 3, m.VoteCount())
				require.NotNil(t, m.PurchaseLocation())
				assert.Equal(t, "Costco", *m.PurchaseLocation())
			},
		},
		{
			name: "rejects negative vote count",
			build: func() (Model, error) {
				return NewBuilder().SetName("Thing").SetVoteCount(-1).Build()
			},
			wantErr: ErrVoteCountNegative,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := tc.build()
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			if tc.validate != nil {
				tc.validate(t, m)
			}
		})
	}
}
