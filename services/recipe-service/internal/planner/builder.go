package planner

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRecipeIDRequired = errors.New("recipe ID is required")
)

type Builder struct {
	id                 uuid.UUID
	recipeID           uuid.UUID
	classification     string
	servingsYield      *int
	eatWithinDays      *int
	minGapDays         *int
	maxConsecutiveDays *int
	createdAt          time.Time
	updatedAt          time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder                { b.id = id; return b }
func (b *Builder) SetRecipeID(id uuid.UUID) *Builder           { b.recipeID = id; return b }
func (b *Builder) SetClassification(c string) *Builder         { b.classification = c; return b }
func (b *Builder) SetServingsYield(v *int) *Builder            { b.servingsYield = v; return b }
func (b *Builder) SetEatWithinDays(v *int) *Builder            { b.eatWithinDays = v; return b }
func (b *Builder) SetMinGapDays(v *int) *Builder               { b.minGapDays = v; return b }
func (b *Builder) SetMaxConsecutiveDays(v *int) *Builder       { b.maxConsecutiveDays = v; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder           { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder           { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.recipeID == uuid.Nil {
		return Model{}, ErrRecipeIDRequired
	}
	return Model{
		id:                 b.id,
		recipeID:           b.recipeID,
		classification:     b.classification,
		servingsYield:      b.servingsYield,
		eatWithinDays:      b.eatWithinDays,
		minGapDays:         b.minGapDays,
		maxConsecutiveDays: b.maxConsecutiveDays,
		createdAt:          b.createdAt,
		updatedAt:          b.updatedAt,
	}, nil
}
