package recipeimage

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRecipeIDRequired = errors.New("recipe ID is required")
	ErrKeysRequired     = errors.New("object keys are required")
)

type Builder struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	householdID uuid.UUID
	recipeID    uuid.UUID
	fullKey     string
	thumbKey    string
	contentType string
	width       int
	height      int
	fullBytes   int
	thumbBytes  int
	createdAt   time.Time
	updatedAt   time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder          { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder    { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder { b.householdID = id; return b }
func (b *Builder) SetRecipeID(id uuid.UUID) *Builder    { b.recipeID = id; return b }
func (b *Builder) SetFullKey(k string) *Builder         { b.fullKey = k; return b }
func (b *Builder) SetThumbKey(k string) *Builder        { b.thumbKey = k; return b }
func (b *Builder) SetContentType(ct string) *Builder    { b.contentType = ct; return b }
func (b *Builder) SetWidth(v int) *Builder              { b.width = v; return b }
func (b *Builder) SetHeight(v int) *Builder             { b.height = v; return b }
func (b *Builder) SetFullBytes(v int) *Builder          { b.fullBytes = v; return b }
func (b *Builder) SetThumbBytes(v int) *Builder         { b.thumbBytes = v; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder    { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder    { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.recipeID == uuid.Nil {
		return Model{}, ErrRecipeIDRequired
	}
	if b.fullKey == "" || b.thumbKey == "" {
		return Model{}, ErrKeysRequired
	}
	return Model{
		id: b.id, tenantID: b.tenantID, householdID: b.householdID, recipeID: b.recipeID,
		fullKey: b.fullKey, thumbKey: b.thumbKey, contentType: b.contentType,
		width: b.width, height: b.height, fullBytes: b.fullBytes, thumbBytes: b.thumbBytes,
		createdAt: b.createdAt, updatedAt: b.updatedAt,
	}, nil
}
