package wishlist

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired             = errors.New("wish list item name is required")
	ErrNameTooLong              = errors.New("wish list item name must not exceed 255 characters")
	ErrPurchaseLocationTooLong  = errors.New("purchase location must not exceed 255 characters")
	ErrInvalidUrgency           = errors.New("urgency must be one of must_have, need_to_have, want")
	ErrVoteCountNegative        = errors.New("vote count cannot be negative")
)

type Builder struct {
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

func NewBuilder() *Builder {
	return &Builder{urgency: UrgencyWant}
}

func (b *Builder) SetId(id uuid.UUID) *Builder            { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder       { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder    { b.householdID = id; return b }
func (b *Builder) SetName(name string) *Builder            { b.name = name; return b }
func (b *Builder) SetPurchaseLocation(loc *string) *Builder {
	b.purchaseLocation = loc
	return b
}
func (b *Builder) SetUrgency(u string) *Builder       { b.urgency = u; return b }
func (b *Builder) SetVoteCount(v int) *Builder        { b.voteCount = v; return b }
func (b *Builder) SetCreatedBy(id uuid.UUID) *Builder { b.createdBy = id; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder  { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder  { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.name == "" {
		return Model{}, ErrNameRequired
	}
	if len(b.name) > 255 {
		return Model{}, ErrNameTooLong
	}
	if b.purchaseLocation != nil && len(*b.purchaseLocation) > 255 {
		return Model{}, ErrPurchaseLocationTooLong
	}
	if b.urgency == "" {
		b.urgency = UrgencyWant
	}
	if !IsValidUrgency(b.urgency) {
		return Model{}, ErrInvalidUrgency
	}
	if b.voteCount < 0 {
		return Model{}, ErrVoteCountNegative
	}
	return Model{
		id:               b.id,
		tenantID:         b.tenantID,
		householdID:      b.householdID,
		name:             b.name,
		purchaseLocation: b.purchaseLocation,
		urgency:          b.urgency,
		voteCount:        b.voteCount,
		createdBy:        b.createdBy,
		createdAt:        b.createdAt,
		updatedAt:        b.updatedAt,
	}, nil
}
