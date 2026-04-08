package trackingitem

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/schedule"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound           = errors.New("tracking item not found")
	ErrDuplicateName      = errors.New("tracking item name already exists for this user")
	ErrScaleTypeImmutable = errors.New("scale type cannot be changed after creation")
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) List(userID uuid.UUID) ([]Model, error) {
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

// ListWithSchedules returns the user's tracking items paired with their
// current effective schedule. The list handler uses this so the per-item
// schedule lookup happens inside the processor instead of being threaded
// through the REST layer one Model at a time.
func (p *Processor) ListWithSchedules(userID uuid.UUID) ([]ItemWithSchedule, error) {
	models, err := p.List(userID)
	if err != nil {
		return nil, err
	}
	results := make([]ItemWithSchedule, len(models))
	for i, m := range models {
		sched, _ := p.GetCurrentSchedule(m.Id())
		results[i] = ItemWithSchedule{Item: m, Schedule: sched}
	}
	return results, nil
}

func (p *Processor) Get(id uuid.UUID) (Model, error) {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	return Make(e)
}

func (p *Processor) Create(tenantID, userID uuid.UUID, name, scaleType, color string, scaleConfig json.RawMessage, sched []int, sortOrder int) (Model, error) {
	name = strings.TrimSpace(name)

	if err := ValidateSchedule(sched); err != nil {
		return Model{}, err
	}

	if _, err := NewBuilder().
		SetName(name).
		SetScaleType(scaleType).
		SetColor(color).
		SetScaleConfig(scaleConfig).
		SetSortOrder(sortOrder).
		Build(); err != nil {
		return Model{}, err
	}

	if _, err := GetByName(userID, name)(p.db.WithContext(p.ctx))(); err == nil {
		return Model{}, ErrDuplicateName
	}

	if sortOrder == 0 {
		maxOrder, err := getMaxSortOrder(p.db.WithContext(p.ctx), userID)
		if err != nil {
			return Model{}, err
		}
		sortOrder = maxOrder + 1
	}

	var m Model
	err := p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		e := Entity{
			TenantId:    tenantID,
			UserId:      userID,
			Name:        name,
			ScaleType:   scaleType,
			ScaleConfig: scaleConfig,
			Color:       color,
			SortOrder:   sortOrder,
		}
		if err := createTrackingItem(tx, &e); err != nil {
			return err
		}

		today := time.Now().UTC().Truncate(24 * time.Hour)
		if _, err := schedule.NewProcessor(p.l, p.ctx, tx).CreateSnapshot(e.Id, sched, today); err != nil {
			return err
		}

		var err error
		m, err = Make(e)
		return err
	})
	if err != nil {
		return Model{}, err
	}
	return m, nil
}

func (p *Processor) Update(id uuid.UUID, name *string, color *string, sched *[]int, sortOrder *int, scaleConfig *json.RawMessage) (Model, error) {
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
		if existing, err := GetByName(e.UserId, trimmed)(p.db.WithContext(p.ctx))(); err == nil && existing.Id != id {
			return Model{}, ErrDuplicateName
		}
		e.Name = trimmed
	}
	if color != nil {
		if !validColors[*color] {
			return Model{}, ErrInvalidColor
		}
		e.Color = *color
	}
	if sortOrder != nil {
		if *sortOrder < 0 {
			return Model{}, ErrInvalidSortOrder
		}
		e.SortOrder = *sortOrder
	}
	if scaleConfig != nil {
		if e.ScaleType == "range" {
			var rc RangeConfig
			if err := json.Unmarshal(*scaleConfig, &rc); err != nil {
				return Model{}, ErrRangeConfigRequired
			}
			if rc.Min >= rc.Max {
				return Model{}, ErrInvalidRangeConfig
			}
		}
		e.ScaleConfig = *scaleConfig
	}

	var m Model
	err = p.db.WithContext(p.ctx).Transaction(func(tx *gorm.DB) error {
		if err := updateTrackingItem(tx, &e); err != nil {
			return err
		}

		if sched != nil {
			if err := ValidateSchedule(*sched); err != nil {
				return err
			}
			today := time.Now().UTC().Truncate(24 * time.Hour)
			if _, err := schedule.NewProcessor(p.l, p.ctx, tx).CreateSnapshot(e.Id, *sched, today); err != nil {
				return err
			}
		}

		var err error
		m, err = Make(e)
		return err
	})
	if err != nil {
		return Model{}, err
	}
	return m, nil
}

func (p *Processor) Delete(id uuid.UUID) error {
	e, err := GetByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return ErrNotFound
	}
	return softDeleteTrackingItem(p.db.WithContext(p.ctx), &e)
}

func (p *Processor) GetScheduleHistory(itemID uuid.UUID) ([]schedule.Model, error) {
	return schedule.NewProcessor(p.l, p.ctx, p.db).GetHistory(itemID)
}

func (p *Processor) GetCurrentSchedule(itemID uuid.UUID) ([]int, error) {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	m, err := schedule.NewProcessor(p.l, p.ctx, p.db).GetEffective(itemID, now)
	if err != nil {
		return []int{}, nil
	}
	return m.Schedule(), nil
}
