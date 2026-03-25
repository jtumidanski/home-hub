package restoration

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTaskIDRequired    = errors.New("restoration taskID is required")
	ErrCreatedByRequired = errors.New("restoration createdByUserID is required")
)

type Builder struct {
	id              uuid.UUID
	tenantID        uuid.UUID
	householdID     uuid.UUID
	taskID          uuid.UUID
	createdByUserID uuid.UUID
	createdAt       time.Time
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) SetId(id uuid.UUID) *Builder              { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder         { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder      { b.householdID = id; return b }
func (b *Builder) SetTaskID(id uuid.UUID) *Builder           { b.taskID = id; return b }
func (b *Builder) SetCreatedByUserID(id uuid.UUID) *Builder  { b.createdByUserID = id; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder         { b.createdAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.taskID == uuid.Nil {
		return Model{}, ErrTaskIDRequired
	}
	if b.createdByUserID == uuid.Nil {
		return Model{}, ErrCreatedByRequired
	}
	return Model{
		id:              b.id,
		tenantID:        b.tenantID,
		householdID:     b.householdID,
		taskID:          b.taskID,
		createdByUserID: b.createdByUserID,
		createdAt:       b.createdAt,
	}, nil
}
