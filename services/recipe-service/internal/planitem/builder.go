package planitem

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDayRequired      = errors.New("day is required")
	ErrInvalidSlot      = errors.New("invalid slot value")
	ErrRecipeIDRequired = errors.New("recipe_id is required")
	ErrDayOutOfRange    = errors.New("day must fall within the plan week")
)

type Builder struct {
	id                uuid.UUID
	planWeekID        uuid.UUID
	day               time.Time
	slot              string
	recipeID          uuid.UUID
	servingMultiplier *float64
	plannedServings   *int
	notes             *string
	position          int
	createdAt         time.Time
	updatedAt         time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder                    { b.id = id; return b }
func (b *Builder) SetPlanWeekID(id uuid.UUID) *Builder            { b.planWeekID = id; return b }
func (b *Builder) SetDay(day time.Time) *Builder                  { b.day = day; return b }
func (b *Builder) SetSlot(slot string) *Builder                   { b.slot = slot; return b }
func (b *Builder) SetRecipeID(id uuid.UUID) *Builder              { b.recipeID = id; return b }
func (b *Builder) SetServingMultiplier(v *float64) *Builder       { b.servingMultiplier = v; return b }
func (b *Builder) SetPlannedServings(v *int) *Builder             { b.plannedServings = v; return b }
func (b *Builder) SetNotes(n *string) *Builder                    { b.notes = n; return b }
func (b *Builder) SetPosition(p int) *Builder                     { b.position = p; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder              { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder              { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.day.IsZero() {
		return Model{}, ErrDayRequired
	}
	if !IsValidSlot(b.slot) {
		return Model{}, ErrInvalidSlot
	}
	if b.recipeID == uuid.Nil {
		return Model{}, ErrRecipeIDRequired
	}
	return Model{
		id:                b.id,
		planWeekID:        b.planWeekID,
		day:               b.day,
		slot:              b.slot,
		recipeID:          b.recipeID,
		servingMultiplier: b.servingMultiplier,
		plannedServings:   b.plannedServings,
		notes:             b.notes,
		position:          b.position,
		createdAt:         b.createdAt,
		updatedAt:         b.updatedAt,
	}, nil
}
