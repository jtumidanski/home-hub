package task

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTitleRequired = errors.New("task title is required")
)

type Builder struct {
	id              uuid.UUID
	tenantID        uuid.UUID
	householdID     uuid.UUID
	title           string
	notes           string
	status          string
	dueOn           *time.Time
	rolloverEnabled bool
	ownerUserID     *uuid.UUID
	completedAt     *time.Time
	completedByUID  *uuid.UUID
	deletedAt       *time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

func NewBuilder() *Builder {
	return &Builder{status: "pending"}
}

func (b *Builder) SetId(id uuid.UUID) *Builder              { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder         { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder      { b.householdID = id; return b }
func (b *Builder) SetTitle(title string) *Builder             { b.title = title; return b }
func (b *Builder) SetNotes(notes string) *Builder             { b.notes = notes; return b }
func (b *Builder) SetStatus(status string) *Builder           { b.status = status; return b }
func (b *Builder) SetDueOn(dueOn *time.Time) *Builder         { b.dueOn = dueOn; return b }
func (b *Builder) SetRolloverEnabled(v bool) *Builder         { b.rolloverEnabled = v; return b }
func (b *Builder) SetOwnerUserID(id *uuid.UUID) *Builder       { b.ownerUserID = id; return b }
func (b *Builder) SetCompletedAt(t *time.Time) *Builder       { b.completedAt = t; return b }
func (b *Builder) SetCompletedByUID(id *uuid.UUID) *Builder   { b.completedByUID = id; return b }
func (b *Builder) SetDeletedAt(t *time.Time) *Builder         { b.deletedAt = t; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder          { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder          { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.title == "" {
		return Model{}, ErrTitleRequired
	}
	return Model{
		id:              b.id,
		tenantID:        b.tenantID,
		householdID:     b.householdID,
		title:           b.title,
		notes:           b.notes,
		status:          b.status,
		dueOn:           b.dueOn,
		rolloverEnabled: b.rolloverEnabled,
		ownerUserID:     b.ownerUserID,
		completedAt:     b.completedAt,
		completedByUID:  b.completedByUID,
		deletedAt:       b.deletedAt,
		createdAt:       b.createdAt,
		updatedAt:       b.updatedAt,
	}, nil
}
