package category

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("ingredient category not found")
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
	entities, err := GetByTenantID(tenantID)(p.db.WithContext(p.ctx))
	if err != nil {
		return nil, err
	}

	// Auto-seed defaults if empty
	if len(entities) == 0 {
		if err := p.seedDefaults(tenantID); err != nil {
			return nil, err
		}
		entities, err = GetByTenantID(tenantID)(p.db.WithContext(p.ctx))
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
		count, _ := CountIngredientsByCategory(p.db.WithContext(p.ctx), e.Id)
		models[i] = m.WithIngredientCount(int(count))
	}
	return models, nil
}

func (p *Processor) seedDefaults(tenantID uuid.UUID) error {
	return p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		// Re-check inside transaction to avoid race
		var count int64
		if err := tx.Model(&Entity{}).Where("tenant_id = ?", tenantID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}

		now := time.Now().UTC()
		for _, dc := range defaultCategories {
			e := Entity{
				Id:        uuid.New(),
				TenantId:  tenantID,
				Name:      dc.Name,
				SortOrder: dc.SortOrder,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := tx.Create(&e).Error; err != nil {
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

	// Check for duplicate name
	if _, err := GetByName(tenantID, name)(p.db.WithContext(p.ctx))(); err == nil {
		return Model{}, ErrDuplicateName
	}

	maxOrder, err := GetMaxSortOrder(p.db.WithContext(p.ctx), tenantID)
	if err != nil {
		return Model{}, err
	}

	now := time.Now().UTC()
	e := Entity{
		Id:        uuid.New(),
		TenantId:  tenantID,
		Name:      name,
		SortOrder: maxOrder + 1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := p.db.WithContext(p.ctx).Create(&e).Error; err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Update(id uuid.UUID, tenantID uuid.UUID, name *string, sortOrder *int) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil || e.TenantId != tenantID {
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
		// Check duplicate name (exclude self)
		if existing, err := GetByName(tenantID, trimmed)(p.db.WithContext(p.ctx))(); err == nil && existing.Id != id {
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

	e.UpdatedAt = time.Now().UTC()
	if err := p.db.WithContext(p.ctx).Save(&e).Error; err != nil {
		return Model{}, err
	}

	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	count, _ := CountIngredientsByCategory(p.db.WithContext(p.ctx), e.Id)
	return m.WithIngredientCount(int(count)), nil
}

func (p *Processor) Delete(id uuid.UUID, tenantID uuid.UUID) error {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil || e.TenantId != tenantID {
		return ErrNotFound
	}
	// FK ON DELETE SET NULL handles nullifying ingredient category_id
	return p.db.WithContext(p.ctx).Where("id = ?", id).Delete(&Entity{}).Error
}
