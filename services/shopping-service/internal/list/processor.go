package list

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
	ErrNotFound       = errors.New("shopping list not found")
	ErrAlreadyArchived = errors.New("shopping list is already archived")
	ErrNotArchived    = errors.New("shopping list is not archived")
	ErrArchived       = errors.New("shopping list is archived")
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) List(status string) ([]Model, error) {
	if status == "" {
		status = "active"
	}
	entities, err := GetByStatus(status)(p.db.WithContext(p.ctx))()
	if err != nil {
		return nil, err
	}

	listIDs := make([]uuid.UUID, len(entities))
	for i, e := range entities {
		listIDs[i] = e.Id
	}
	counts, err := getItemCounts(p.db.WithContext(p.ctx), listIDs)
	if err != nil {
		return nil, err
	}

	models := make([]Model, len(entities))
	for i, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		if c, ok := counts[e.Id]; ok {
			m = m.WithCounts(c.ItemCount, c.CheckedCount)
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
	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	itemCount, checkedCount, _ := getItemCountsForList(p.db.WithContext(p.ctx), id)
	return m.WithCounts(itemCount, checkedCount), nil
}

func (p *Processor) Create(tenantID, householdID, userID uuid.UUID, name string) (Model, error) {
	name = strings.TrimSpace(name)
	if _, err := NewBuilder().SetName(name).Build(); err != nil {
		return Model{}, err
	}

	e := Entity{
		TenantId:    tenantID,
		HouseholdId: householdID,
		Name:        name,
		Status:      "active",
		CreatedBy:   userID,
	}
	if err := createList(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Update(id uuid.UUID, name string) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if e.Status == "archived" {
		return Model{}, ErrArchived
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return Model{}, ErrNameRequired
	}
	if len(name) > 255 {
		return Model{}, ErrNameTooLong
	}
	e.Name = name

	if err := updateList(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	itemCount, checkedCount, _ := getItemCountsForList(p.db.WithContext(p.ctx), id)
	return m.WithCounts(itemCount, checkedCount), nil
}

func (p *Processor) Delete(id uuid.UUID) error {
	if _, err := GetByID(id)(p.db.WithContext(p.ctx))(); err != nil {
		return ErrNotFound
	}
	return deleteList(p.db.WithContext(p.ctx), id)
}

func (p *Processor) Archive(id uuid.UUID) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if e.Status == "archived" {
		return Model{}, ErrAlreadyArchived
	}

	now := time.Now().UTC()
	e.Status = "archived"
	e.ArchivedAt = &now

	if err := updateList(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	itemCount, checkedCount, _ := getItemCountsForList(p.db.WithContext(p.ctx), id)
	return m.WithCounts(itemCount, checkedCount), nil
}

func (p *Processor) Unarchive(id uuid.UUID) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if e.Status != "archived" {
		return Model{}, ErrNotArchived
	}

	e.Status = "active"
	e.ArchivedAt = nil

	if err := updateList(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	itemCount, checkedCount, _ := getItemCountsForList(p.db.WithContext(p.ctx), id)
	return m.WithCounts(itemCount, checkedCount), nil
}
