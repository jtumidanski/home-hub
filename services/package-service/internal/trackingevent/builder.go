package trackingevent

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDescriptionRequired = errors.New("description is required")
	ErrStatusRequired      = errors.New("status is required")
)

type Builder struct {
	id          uuid.UUID
	packageID   uuid.UUID
	timestamp   time.Time
	status      string
	description string
	location    *string
	rawStatus   *string
	createdAt   time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder            { b.id = id; return b }
func (b *Builder) SetPackageID(id uuid.UUID) *Builder      { b.packageID = id; return b }
func (b *Builder) SetTimestamp(t time.Time) *Builder        { b.timestamp = t; return b }
func (b *Builder) SetStatus(s string) *Builder              { b.status = s; return b }
func (b *Builder) SetDescription(d string) *Builder         { b.description = d; return b }
func (b *Builder) SetLocation(l *string) *Builder           { b.location = l; return b }
func (b *Builder) SetRawStatus(s *string) *Builder          { b.rawStatus = s; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder        { b.createdAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.description == "" {
		return Model{}, ErrDescriptionRequired
	}
	if b.status == "" {
		return Model{}, ErrStatusRequired
	}
	return Model{
		id:          b.id,
		packageID:   b.packageID,
		timestamp:   b.timestamp,
		status:      b.status,
		description: b.description,
		location:    b.location,
		rawStatus:   b.rawStatus,
		createdAt:   b.createdAt,
	}, nil
}
