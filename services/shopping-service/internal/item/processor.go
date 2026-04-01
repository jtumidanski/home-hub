package item

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("shopping item not found")
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) GetByListID(listID uuid.UUID) ([]Model, error) {
	entities, err := GetByListID(listID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}
	models := make([]Model, len(entities))
	for i, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		models[i] = m
	}
	return models, nil
}

type AddInput struct {
	ListID            uuid.UUID
	Name              string
	Quantity          *string
	CategoryID        *uuid.UUID
	CategoryName      *string
	CategorySortOrder *int
	Position          *int
}

func (p *Processor) Add(input AddInput) (Model, error) {
	name := strings.TrimSpace(input.Name)
	if _, err := NewBuilder().SetName(name).Build(); err != nil {
		return Model{}, err
	}

	pos := 0
	if input.Position != nil {
		pos = *input.Position
	} else {
		maxPos, err := getMaxPosition(p.db.WithContext(p.ctx), input.ListID)
		if err == nil {
			pos = maxPos + 1
		}
	}

	e := Entity{
		ListId:            input.ListID,
		Name:              name,
		Quantity:          input.Quantity,
		CategoryId:        input.CategoryID,
		CategoryName:      input.CategoryName,
		CategorySortOrder: input.CategorySortOrder,
		Checked:           false,
		Position:          pos,
	}
	if err := createItem(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

type UpdateInput struct {
	Name              *string
	Quantity          *string
	CategoryID        *uuid.UUID
	CategoryName      *string
	CategorySortOrder *int
	Position          *int
	ClearCategory     bool
}

func (p *Processor) Update(id uuid.UUID, input UpdateInput) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return Model{}, ErrNameRequired
		}
		if len(name) > 255 {
			return Model{}, ErrNameTooLong
		}
		e.Name = name
	}
	if input.Quantity != nil {
		e.Quantity = input.Quantity
	}
	if input.ClearCategory {
		e.CategoryId = nil
		e.CategoryName = nil
		e.CategorySortOrder = nil
	} else if input.CategoryID != nil {
		e.CategoryId = input.CategoryID
		e.CategoryName = input.CategoryName
		e.CategorySortOrder = input.CategorySortOrder
	}
	if input.Position != nil {
		e.Position = *input.Position
	}

	if err := updateItem(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Delete(id uuid.UUID) error {
	if _, err := GetByID(id)(p.db.WithContext(p.ctx))(); err != nil {
		return ErrNotFound
	}
	return deleteItem(p.db.WithContext(p.ctx), id)
}

func (p *Processor) Check(id uuid.UUID, checked bool) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	e.Checked = checked
	if err := updateItem(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) UncheckAll(listID uuid.UUID) error {
	return uncheckAll(p.db.WithContext(p.ctx), listID)
}
