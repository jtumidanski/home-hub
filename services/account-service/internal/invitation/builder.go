package invitation

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmailRequired       = errors.New("invitation email is required")
	ErrRoleRequired        = errors.New("invitation role is required")
	ErrInvalidRole         = errors.New("invitation role must be admin, editor, or viewer")
	ErrHouseholdIDRequired = errors.New("invitation household ID is required")
	ErrInvitedByRequired   = errors.New("invitation invitedBy is required")
)

var validRoles = map[string]bool{
	"admin":  true,
	"editor": true,
	"viewer": true,
}

type Builder struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	email       string
	role        string
	status      string
	invitedBy   uuid.UUID
	expiresAt   time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

func NewBuilder() *Builder {
	return &Builder{
		role:   "viewer",
		status: "pending",
	}
}

func (b *Builder) SetId(id uuid.UUID) *Builder          { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder     { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder  { b.householdID = id; return b }
func (b *Builder) SetEmail(email string) *Builder        { b.email = email; return b }
func (b *Builder) SetRole(role string) *Builder          { b.role = role; return b }
func (b *Builder) SetStatus(status string) *Builder      { b.status = status; return b }
func (b *Builder) SetInvitedBy(id uuid.UUID) *Builder    { b.invitedBy = id; return b }
func (b *Builder) SetExpiresAt(t time.Time) *Builder     { b.expiresAt = t; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder     { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder     { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.email == "" {
		return Model{}, ErrEmailRequired
	}
	if b.role == "" {
		return Model{}, ErrRoleRequired
	}
	if !validRoles[b.role] {
		return Model{}, ErrInvalidRole
	}
	if b.householdID == uuid.Nil {
		return Model{}, ErrHouseholdIDRequired
	}
	if b.invitedBy == uuid.Nil {
		return Model{}, ErrInvitedByRequired
	}
	return Model{
		id:          b.id,
		tenantID:    b.tenantID,
		householdID: b.householdID,
		email:       b.email,
		role:        b.role,
		status:      b.status,
		invitedBy:   b.invitedBy,
		expiresAt:   b.expiresAt,
		createdAt:   b.createdAt,
		updatedAt:   b.updatedAt,
	}, nil
}
