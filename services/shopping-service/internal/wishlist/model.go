package wishlist

import (
	"time"

	"github.com/google/uuid"
)

const (
	UrgencyMustHave   = "must_have"
	UrgencyNeedToHave = "need_to_have"
	UrgencyWant       = "want"
)

// IsValidUrgency reports whether u is one of the allowed urgency values.
func IsValidUrgency(u string) bool {
	switch u {
	case UrgencyMustHave, UrgencyNeedToHave, UrgencyWant:
		return true
	}
	return false
}

type Model struct {
	id               uuid.UUID
	tenantID         uuid.UUID
	householdID      uuid.UUID
	name             string
	purchaseLocation *string
	urgency          string
	voteCount        int
	createdBy        uuid.UUID
	createdAt        time.Time
	updatedAt        time.Time
}

func (m Model) Id() uuid.UUID                { return m.id }
func (m Model) TenantID() uuid.UUID          { return m.tenantID }
func (m Model) HouseholdID() uuid.UUID       { return m.householdID }
func (m Model) Name() string                 { return m.name }
func (m Model) PurchaseLocation() *string    { return m.purchaseLocation }
func (m Model) Urgency() string              { return m.urgency }
func (m Model) VoteCount() int               { return m.voteCount }
func (m Model) CreatedBy() uuid.UUID         { return m.createdBy }
func (m Model) CreatedAt() time.Time         { return m.createdAt }
func (m Model) UpdatedAt() time.Time         { return m.updatedAt }
