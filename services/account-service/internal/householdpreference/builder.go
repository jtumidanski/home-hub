package householdpreference

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTenantRequired    = errors.New("household preference tenant ID is required")
	ErrUserRequired      = errors.New("household preference user ID is required")
	ErrHouseholdRequired = errors.New("household preference household ID is required")
)

type Builder struct {
	id                   uuid.UUID
	tenantID             uuid.UUID
	userID               uuid.UUID
	householdID          uuid.UUID
	defaultDashboardID   *uuid.UUID
	kioskDashboardSeeded bool
	createdAt            time.Time
	updatedAt            time.Time
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) SetId(id uuid.UUID) *Builder {
	b.id = id
	return b
}

func (b *Builder) SetTenantID(tenantID uuid.UUID) *Builder {
	b.tenantID = tenantID
	return b
}

func (b *Builder) SetUserID(userID uuid.UUID) *Builder {
	b.userID = userID
	return b
}

func (b *Builder) SetHouseholdID(householdID uuid.UUID) *Builder {
	b.householdID = householdID
	return b
}

func (b *Builder) SetDefaultDashboardID(id *uuid.UUID) *Builder {
	b.defaultDashboardID = id
	return b
}

func (b *Builder) SetKioskDashboardSeeded(v bool) *Builder {
	b.kioskDashboardSeeded = v
	return b
}

func (b *Builder) SetCreatedAt(t time.Time) *Builder {
	b.createdAt = t
	return b
}

func (b *Builder) SetUpdatedAt(t time.Time) *Builder {
	b.updatedAt = t
	return b
}

func (b *Builder) Build() (Model, error) {
	if b.tenantID == uuid.Nil {
		return Model{}, ErrTenantRequired
	}
	if b.userID == uuid.Nil {
		return Model{}, ErrUserRequired
	}
	if b.householdID == uuid.Nil {
		return Model{}, ErrHouseholdRequired
	}
	return Model{
		id:                   b.id,
		tenantID:             b.tenantID,
		userID:               b.userID,
		householdID:          b.householdID,
		defaultDashboardID:   b.defaultDashboardID,
		kioskDashboardSeeded: b.kioskDashboardSeeded,
		createdAt:            b.createdAt,
		updatedAt:            b.updatedAt,
	}, nil
}
