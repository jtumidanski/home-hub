package wishlist

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("wish list item not found")

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// CreateInput captures fields a caller may set when creating an item.
// Pointers distinguish "not provided" from "empty".
type CreateInput struct {
	Name             string
	PurchaseLocation *string
	Urgency          *string
}

// UpdateInput captures fields a caller may modify on an existing item.
// vote_count is intentionally absent — it can only be modified via Vote.
type UpdateInput struct {
	Name             *string
	PurchaseLocation *string
	Urgency          *string
}

func (p *Processor) List(householdID uuid.UUID) ([]Model, error) {
	entities, err := ListByHousehold(householdID)(p.db.WithContext(p.ctx))()
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

func (p *Processor) Get(householdID uuid.UUID, id uuid.UUID) (Model, error) {
	e, err := GetByID(id, householdID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	return Make(e)
}

func (p *Processor) Create(tenantID, householdID, userID uuid.UUID, input CreateInput) (Model, error) {
	name := strings.TrimSpace(input.Name)
	urgency := UrgencyWant
	if input.Urgency != nil && *input.Urgency != "" {
		urgency = *input.Urgency
	}
	var loc *string
	if input.PurchaseLocation != nil {
		trimmed := strings.TrimSpace(*input.PurchaseLocation)
		if trimmed != "" {
			loc = &trimmed
		}
	}

	if _, err := NewBuilder().
		SetName(name).
		SetPurchaseLocation(loc).
		SetUrgency(urgency).
		Build(); err != nil {
		return Model{}, err
	}

	e := Entity{
		TenantId:         tenantID,
		HouseholdId:      householdID,
		Name:             name,
		PurchaseLocation: loc,
		Urgency:          urgency,
		VoteCount:        0,
		CreatedBy:        userID,
	}
	if err := createItem(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) Update(householdID uuid.UUID, id uuid.UUID, input UpdateInput) (Model, error) {
	e, err := GetByID(id, householdID)(p.db.WithContext(p.ctx))()
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
	if input.PurchaseLocation != nil {
		trimmed := strings.TrimSpace(*input.PurchaseLocation)
		if trimmed == "" {
			e.PurchaseLocation = nil
		} else {
			if len(trimmed) > 255 {
				return Model{}, ErrPurchaseLocationTooLong
			}
			e.PurchaseLocation = &trimmed
		}
	}
	if input.Urgency != nil {
		if !IsValidUrgency(*input.Urgency) {
			return Model{}, ErrInvalidUrgency
		}
		e.Urgency = *input.Urgency
	}

	if err := updateItem(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}
	// Re-read to capture the actual stored values (e.g. updated_at).
	fresh, err := GetByID(id, householdID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, err
	}
	return Make(fresh)
}

func (p *Processor) Delete(householdID uuid.UUID, id uuid.UUID) error {
	if _, err := GetByID(id, householdID)(p.db.WithContext(p.ctx))(); err != nil {
		return ErrNotFound
	}
	return deleteItem(p.db.WithContext(p.ctx), id, householdID)
}

func (p *Processor) Vote(householdID uuid.UUID, id uuid.UUID) (Model, error) {
	if err := incrementVote(p.db.WithContext(p.ctx), id, householdID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Model{}, ErrNotFound
		}
		return Model{}, err
	}
	e, err := GetByID(id, householdID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	return Make(e)
}
