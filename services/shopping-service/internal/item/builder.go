package item

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired = errors.New("item name is required")
	ErrNameTooLong  = errors.New("item name must not exceed 255 characters")
)

type Builder struct {
	id                uuid.UUID
	listID            uuid.UUID
	name              string
	quantity          *string
	categoryID        *uuid.UUID
	categoryName      *string
	categorySortOrder *int
	checked           bool
	position          int
	createdAt         time.Time
	updatedAt         time.Time
}

func NewBuilder() *Builder { return &Builder{} }

func (b *Builder) SetId(id uuid.UUID) *Builder                  { b.id = id; return b }
func (b *Builder) SetListID(id uuid.UUID) *Builder              { b.listID = id; return b }
func (b *Builder) SetName(name string) *Builder                 { b.name = name; return b }
func (b *Builder) SetQuantity(q *string) *Builder               { b.quantity = q; return b }
func (b *Builder) SetCategoryID(id *uuid.UUID) *Builder         { b.categoryID = id; return b }
func (b *Builder) SetCategoryName(name *string) *Builder        { b.categoryName = name; return b }
func (b *Builder) SetCategorySortOrder(order *int) *Builder     { b.categorySortOrder = order; return b }
func (b *Builder) SetChecked(checked bool) *Builder             { b.checked = checked; return b }
func (b *Builder) SetPosition(pos int) *Builder                 { b.position = pos; return b }
func (b *Builder) SetCreatedAt(t time.Time) *Builder            { b.createdAt = t; return b }
func (b *Builder) SetUpdatedAt(t time.Time) *Builder            { b.updatedAt = t; return b }

func (b *Builder) Build() (Model, error) {
	if b.name == "" {
		return Model{}, ErrNameRequired
	}
	if len(b.name) > 255 {
		return Model{}, ErrNameTooLong
	}
	return Model{
		id:                b.id,
		listID:            b.listID,
		name:              b.name,
		quantity:          b.quantity,
		categoryID:        b.categoryID,
		categoryName:      b.categoryName,
		categorySortOrder: b.categorySortOrder,
		checked:           b.checked,
		position:          b.position,
		createdAt:         b.createdAt,
		updatedAt:         b.updatedAt,
	}, nil
}
