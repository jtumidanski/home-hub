package membership

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrHouseholdIDRequired = errors.New("membership household ID is required")
	ErrUserIDRequired      = errors.New("membership user ID is required")
	ErrRoleRequired        = errors.New("membership role is required")
)

type Builder struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	userID      uuid.UUID
	role        string
	createdAt   time.Time
	updatedAt   time.Time
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

func (b *Builder) SetHouseholdID(householdID uuid.UUID) *Builder {
	b.householdID = householdID
	return b
}

func (b *Builder) SetUserID(userID uuid.UUID) *Builder {
	b.userID = userID
	return b
}

func (b *Builder) SetRole(role string) *Builder {
	b.role = role
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
	if b.householdID == uuid.Nil {
		return Model{}, ErrHouseholdIDRequired
	}
	if b.userID == uuid.Nil {
		return Model{}, ErrUserIDRequired
	}
	if b.role == "" {
		return Model{}, ErrRoleRequired
	}
	return Model{
		id:          b.id,
		tenantID:    b.tenantID,
		householdID: b.householdID,
		userID:      b.userID,
		role:        b.role,
		createdAt:   b.createdAt,
		updatedAt:   b.updatedAt,
	}, nil
}
