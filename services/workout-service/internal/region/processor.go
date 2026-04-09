package region

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("region not found")
	ErrDuplicateName = errors.New("region name already exists for this user")
)

// DefaultRegions is the seeded list installed for a brand-new user. The slice
// order doubles as `sort_order` so the body-region picker renders in this
// canonical sequence on first use.
var DefaultRegions = []string{
	"Chest", "Shoulders", "Back", "Biceps", "Triceps",
	"Core", "Legs", "Glutes", "Full Body", "Other",
}

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) EnsureSeeded(tenantID, userID uuid.UUID) error {
	return p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		count, err := CountForUser(tx, userID)
		if err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		for i, name := range DefaultRegions {
			e := Entity{
				TenantId:  tenantID,
				UserId:    userID,
				Name:      name,
				SortOrder: i,
			}
			if err := createRegion(tx, &e); err != nil {
				return err
			}
		}
		return nil
	})
}

func (p *Processor) List(tenantID, userID uuid.UUID) ([]Model, error) {
	if err := p.EnsureSeeded(tenantID, userID); err != nil {
		return nil, err
	}
	entities, err := GetAllByUser(userID)(p.db.WithContext(p.ctx))()
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

func (p *Processor) Get(id uuid.UUID) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	return Make(e)
}

func (p *Processor) Create(tenantID, userID uuid.UUID, name string, sortOrder int) (Model, error) {
	name = strings.TrimSpace(name)
	if _, err := NewBuilder().SetName(name).SetSortOrder(sortOrder).Build(); err != nil {
		return Model{}, err
	}
	if _, err := GetByName(userID, name)(p.db.WithContext(p.ctx))(); err == nil {
		return Model{}, ErrDuplicateName
	}
	e := Entity{
		TenantId:  tenantID,
		UserId:    userID,
		Name:      name,
		SortOrder: sortOrder,
	}
	if err := createRegion(p.db.WithContext(p.ctx), &e); err != nil {
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
		if len(trimmed) > 50 {
			return Model{}, ErrNameTooLong
		}
		if existing, err := GetByName(e.UserId, trimmed)(p.db.WithContext(p.ctx))(); err == nil && existing.Id != id {
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
	if err := updateRegion(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Delete(id uuid.UUID) error {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return ErrNotFound
	}
	return softDeleteRegion(p.db.WithContext(p.ctx), &e)
}
