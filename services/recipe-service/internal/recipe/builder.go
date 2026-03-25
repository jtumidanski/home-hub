package recipe

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTitleRequired  = errors.New("recipe title is required")
	ErrSourceRequired = errors.New("recipe source is required")
)

type Builder struct {
	id              uuid.UUID
	tenantID        uuid.UUID
	householdID     uuid.UUID
	title           string
	description     string
	source          string
	servings        *int
	prepTimeMinutes *int
	cookTimeMinutes *int
	sourceURL       string
	tags            []string
	deletedAt       *time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder              { b.id = id; return b }
func (b *Builder) SetTenantID(id uuid.UUID) *Builder         { b.tenantID = id; return b }
func (b *Builder) SetHouseholdID(id uuid.UUID) *Builder      { b.householdID = id; return b }
func (b *Builder) SetTitle(title string) *Builder             { b.title = title; return b }
func (b *Builder) SetDescription(desc string) *Builder        { b.description = desc; return b }
func (b *Builder) SetSource(source string) *Builder           { b.source = source; return b }
func (b *Builder) SetServings(v *int) *Builder                { b.servings = v; return b }
func (b *Builder) SetPrepTimeMinutes(v *int) *Builder         { b.prepTimeMinutes = v; return b }
func (b *Builder) SetCookTimeMinutes(v *int) *Builder         { b.cookTimeMinutes = v; return b }
func (b *Builder) SetSourceURL(url string) *Builder           { b.sourceURL = url; return b }
func (b *Builder) SetTags(tags []string) *Builder             { b.tags = tags; return b }
func (b *Builder) SetDeletedAt(t *time.Time) *Builder         { b.deletedAt = t; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder          { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder          { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.title == "" {
		return Model{}, ErrTitleRequired
	}
	if b.source == "" {
		return Model{}, ErrSourceRequired
	}
	return Model{
		id:              b.id,
		tenantID:        b.tenantID,
		householdID:     b.householdID,
		title:           b.title,
		description:     b.description,
		source:          b.source,
		servings:        b.servings,
		prepTimeMinutes: b.prepTimeMinutes,
		cookTimeMinutes: b.cookTimeMinutes,
		sourceURL:       b.sourceURL,
		tags:            b.tags,
		deletedAt:       b.deletedAt,
		createdAt:       b.createdAt,
		updatedAt:       b.updatedAt,
	}, nil
}
