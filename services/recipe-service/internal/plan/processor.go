package plan

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/audit"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/export"
	"github.com/jtumidanski/home-hub/shared/go/model"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound       = errors.New("plan not found")
	ErrAlreadyExists  = errors.New("plan already exists for this week")
	ErrLocked         = errors.New("plan is locked")
	ErrAlreadyLocked  = errors.New("plan is already locked")
	ErrNotLocked      = errors.New("plan is not locked")
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) ByIDProvider(id uuid.UUID) model.Provider[Model] {
	return model.Map(Make)(getByID(id)(p.db.WithContext(p.ctx)))
}

type CreateAttrs struct {
	StartsOn time.Time
	Name     string
}

func (p *Processor) Create(tenantID, householdID, userID uuid.UUID, attrs CreateAttrs) (Model, error) {
	if attrs.StartsOn.IsZero() {
		return Model{}, ErrStartsOnRequired
	}

	name := attrs.Name
	if name == "" {
		name = fmt.Sprintf("Week of %s", attrs.StartsOn.Format("January 2, 2006"))
	}

	// Check uniqueness
	_, err := getByHouseholdAndStartsOn(p.db.WithContext(p.ctx), householdID, attrs.StartsOn)
	if err == nil {
		return Model{}, ErrAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Model{}, err
	}

	e, err := create(p.db.WithContext(p.ctx), tenantID, householdID, attrs.StartsOn, name, userID)
	if err != nil {
		return Model{}, err
	}

	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	p.emitAudit(m.Id(), "plan.created", nil)
	return m, nil
}

func (p *Processor) Get(id uuid.UUID) (Model, error) {
	m, err := p.ByIDProvider(id)()
	if err != nil {
		return Model{}, ErrNotFound
	}
	return m, nil
}

func (p *Processor) List(filters ListFilters) ([]Model, int64, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 20
	}

	entities, total, err := getAll(filters)(p.db.WithContext(p.ctx))
	if err != nil {
		return nil, 0, err
	}

	models := make([]Model, 0, len(entities))
	for _, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, 0, err
		}
		models = append(models, m)
	}
	return models, total, nil
}

func (p *Processor) UpdateName(id uuid.UUID, name string) (Model, error) {
	e, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if e.Locked {
		return Model{}, ErrLocked
	}

	e.Name = name
	if err := update(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}

	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	p.emitAudit(m.Id(), "plan.updated", nil)
	return m, nil
}

func (p *Processor) Lock(id uuid.UUID) (Model, error) {
	e, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if e.Locked {
		return Model{}, ErrAlreadyLocked
	}

	e.Locked = true
	if err := update(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}

	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	p.emitAudit(m.Id(), "plan.locked", nil)
	return m, nil
}

func (p *Processor) Unlock(id uuid.UUID) (Model, error) {
	e, err := getByID(id)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, ErrNotFound
	}
	if !e.Locked {
		return Model{}, ErrNotLocked
	}

	e.Locked = false
	if err := update(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}

	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	p.emitAudit(m.Id(), "plan.unlocked", nil)
	return m, nil
}

func (p *Processor) Duplicate(sourceID uuid.UUID, tenantID, householdID, userID uuid.UUID, targetStartsOn time.Time) (Model, error) {
	source, err := p.Get(sourceID)
	if err != nil {
		return Model{}, err
	}

	newPlan, err := p.Create(tenantID, householdID, userID, CreateAttrs{
		StartsOn: targetStartsOn,
	})
	if err != nil {
		return Model{}, err
	}

	p.emitAudit(newPlan.Id(), "plan.duplicated", map[string]interface{}{
		"source_plan_id": source.Id().String(),
	})
	return newPlan, nil
}

func (p *Processor) ExportMarkdown(id uuid.UUID, authHeader string, exportProc interface{ GenerateMarkdown(export.PlanData) string }) (string, error) {
	m, err := p.Get(id)
	if err != nil {
		return "", err
	}

	markdown := exportProc.GenerateMarkdown(export.PlanData{
		ID: m.Id(), TenantID: m.TenantID(), Name: m.Name(), StartsOn: m.StartsOn(),
		AuthHeader: authHeader,
	})

	p.emitAudit(m.Id(), "plan.exported", map[string]interface{}{
		"format": "markdown",
	})
	return markdown, nil
}

func (p *Processor) Delete(id uuid.UUID) error {
	_, err := p.Get(id)
	if err != nil {
		return ErrNotFound
	}
	return deleteByID(p.db.WithContext(p.ctx), id)
}

func (p *Processor) emitAudit(entityID uuid.UUID, action string, metadata map[string]interface{}) {
	t, ok := tenantctx.FromContext(p.ctx)
	if !ok {
		return
	}
	audit.Emit(p.l, p.db.WithContext(p.ctx), t.Id(), "plan", entityID, action, t.UserId(), metadata)
}
