package category

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("category not found")
	ErrDuplicateName = errors.New("category name already exists for this tenant")
)

var defaultCategories = []struct {
	Name      string
	SortOrder int
}{
	{"Produce", 1},
	{"Meats & Seafood", 2},
	{"Dairy & Eggs", 3},
	{"Bakery & Bread", 4},
	{"Pantry & Dry Goods", 5},
	{"Frozen", 6},
	{"Beverages", 7},
	{"Snacks & Sweets", 8},
	{"Condiments & Sauces", 9},
	{"Spices & Seasonings", 10},
	{"Other", 11},
	{"Household", 12},
	{"Personal Care", 13},
	{"Baby & Kids", 14},
	{"Pet Supplies", 15},
}

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) List(tenantID uuid.UUID) ([]Model, error) {
	entities, err := GetAll()(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		if err := p.seedDefaults(tenantID); err != nil {
			return nil, err
		}
		entities, err = GetAll()(p.db.WithContext(p.ctx))()
		if err != nil {
			return nil, err
		}
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

// seedDefaults inserts default categories for a new tenant within a transaction.
func (p *Processor) seedDefaults(tenantID uuid.UUID) error {
	return p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		count, err := countAll(tx)
		if err != nil {
			return err
		}
		if count > 0 {
			return nil
		}

		for _, dc := range defaultCategories {
			e := Entity{
				TenantId:  tenantID,
				Name:      dc.Name,
				SortOrder: dc.SortOrder,
			}
			if err := createCategory(tx, &e); err != nil {
				return err
			}
		}
		return nil
	})
}

func (p *Processor) Create(tenantID uuid.UUID, name string) (Model, error) {
	name = strings.TrimSpace(name)
	if _, err := NewBuilder().SetName(name).Build(); err != nil {
		return Model{}, err
	}

	if _, err := GetByName(name)(p.db.WithContext(p.ctx))(); err == nil {
		return Model{}, ErrDuplicateName
	}

	maxOrder, err := getMaxSortOrder(p.db.WithContext(p.ctx))
	if err != nil {
		return Model{}, err
	}

	e := Entity{
		TenantId:  tenantID,
		Name:      name,
		SortOrder: maxOrder + 1,
	}
	if err := createCategory(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Update(id uuid.UUID, name *string, sortOrder *int) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}

	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return Model{}, ErrNameRequired
		}
		if len(trimmed) > 100 {
			return Model{}, ErrNameTooLong
		}
		if existing, err := GetByName(trimmed)(p.db.WithContext(p.ctx))(); err == nil && existing.Id != id {
			return Model{}, ErrDuplicateName
		}
		e.Name = trimmed
	}
	if sortOrder != nil {
		if *sortOrder < 0 {
			return Model{}, ErrInvalidSortOrder
		}
		e.SortOrder = *sortOrder
	}

	if err := updateCategory(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}

	return Make(e)
}

func (p *Processor) Delete(id uuid.UUID) error {
	if _, err := GetByID(id)(p.db.WithContext(p.ctx))(); err != nil {
		return ErrNotFound
	}
	return deleteCategory(p.db.WithContext(p.ctx), id)
}
