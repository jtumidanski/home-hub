package tracking

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPreTransit      = "pre_transit"
	StatusInTransit       = "in_transit"
	StatusOutForDelivery  = "out_for_delivery"
	StatusDelivered       = "delivered"
	StatusException       = "exception"
	StatusStale           = "stale"
	StatusArchived        = "archived"

	CarrierUSPS  = "usps"
	CarrierUPS   = "ups"
	CarrierFedEx = "fedex"
)

var (
	ErrTrackingNumberRequired = errors.New("tracking number is required")
	ErrCarrierRequired        = errors.New("carrier is required")
	ErrInvalidCarrier         = errors.New("carrier must be usps, ups, or fedex")
	ErrInvalidStatus          = errors.New("invalid package status")
)

var validCarriers = map[string]bool{
	CarrierUSPS:  true,
	CarrierUPS:   true,
	CarrierFedEx: true,
}

var validStatuses = map[string]bool{
	StatusPreTransit:     true,
	StatusInTransit:      true,
	StatusOutForDelivery: true,
	StatusDelivered:      true,
	StatusException:      true,
	StatusStale:          true,
	StatusArchived:       true,
}

type Builder struct {
	id                 uuid.UUID
	tenantID           uuid.UUID
	householdID        uuid.UUID
	userID             uuid.UUID
	trackingNumber     string
	carrier            string
	label              *string
	notes              *string
	status             string
	private            bool
	estimatedDelivery  *time.Time
	actualDelivery     *time.Time
	lastPolledAt       *time.Time
	lastStatusChangeAt *time.Time
	archivedAt         *time.Time
	createdAt          time.Time
	updatedAt          time.Time
}

func NewBuilder() *Builder { return &Builder{status: StatusPreTransit} }

func (b *Builder) SetId(id uuid.UUID) *Builder                         { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder                   { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder                { b.householdID = id; return b }
func (b *Builder) SetUserID(id uuid.UUID) *Builder                     { b.userID = id; return b }
func (b *Builder) SetTrackingNumber(tn string) *Builder                { b.trackingNumber = tn; return b }
func (b *Builder) SetCarrier(c string) *Builder                        { b.carrier = c; return b }
func (b *Builder) SetLabel(l *string) *Builder                         { b.label = l; return b }
func (b *Builder) SetNotes(n *string) *Builder                         { b.notes = n; return b }
func (b *Builder) SetStatus(s string) *Builder                         { b.status = s; return b }
func (b *Builder) SetPrivate(p bool) *Builder                          { b.private = p; return b }
func (b *Builder) SetEstimatedDelivery(t *time.Time) *Builder          { b.estimatedDelivery = t; return b }
func (b *Builder) SetActualDelivery(t *time.Time) *Builder             { b.actualDelivery = t; return b }
func (b *Builder) SetLastPolledAt(t *time.Time) *Builder               { b.lastPolledAt = t; return b }
func (b *Builder) SetLastStatusChangeAt(t *time.Time) *Builder         { b.lastStatusChangeAt = t; return b }
func (b *Builder) SetArchivedAt(t *time.Time) *Builder                 { b.archivedAt = t; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder                   { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder                   { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.trackingNumber == "" {
		return Model{}, ErrTrackingNumberRequired
	}
	if b.carrier == "" {
		return Model{}, ErrCarrierRequired
	}
	if !validCarriers[b.carrier] {
		return Model{}, ErrInvalidCarrier
	}
	if b.status != "" && !validStatuses[b.status] {
		return Model{}, ErrInvalidStatus
	}
	return Model{
		id:                 b.id,
		tenantID:           b.tenantID,
		householdID:        b.householdID,
		userID:             b.userID,
		trackingNumber:     b.trackingNumber,
		carrier:            b.carrier,
		label:              b.label,
		notes:              b.notes,
		status:             b.status,
		private:            b.private,
		estimatedDelivery:  b.estimatedDelivery,
		actualDelivery:     b.actualDelivery,
		lastPolledAt:       b.lastPolledAt,
		lastStatusChangeAt: b.lastStatusChangeAt,
		archivedAt:         b.archivedAt,
		createdAt:          b.createdAt,
		updatedAt:          b.updatedAt,
	}, nil
}
