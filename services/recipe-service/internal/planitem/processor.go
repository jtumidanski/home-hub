package planitem

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/audit"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("plan item not found")
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

type AddAttrs struct {
	Day               time.Time
	Slot              string
	RecipeID          uuid.UUID
	ServingMultiplier *float64
	PlannedServings   *int
	Notes             *string
	Position          *int
}

func (p *Processor) AddItem(planWeekID uuid.UUID, planStartsOn time.Time, attrs AddAttrs) (Model, error) {
	// Validate day within week range
	if err := validateDayInWeek(attrs.Day, planStartsOn); err != nil {
		return Model{}, err
	}
	if !IsValidSlot(attrs.Slot) {
		return Model{}, ErrInvalidSlot
	}
	if attrs.RecipeID == uuid.Nil {
		return Model{}, ErrRecipeIDRequired
	}

	position := 0
	if attrs.Position != nil {
		position = *attrs.Position
	} else {
		maxPos, err := getMaxPosition(p.db.WithContext(p.ctx), planWeekID, attrs.Day.Format("2006-01-02"), attrs.Slot)
		if err != nil {
			return Model{}, err
		}
		position = maxPos + 1
	}

	e := Entity{
		PlanWeekId:        planWeekID,
		Day:               attrs.Day,
		Slot:              attrs.Slot,
		RecipeId:          attrs.RecipeID,
		ServingMultiplier: attrs.ServingMultiplier,
		PlannedServings:   attrs.PlannedServings,
		Notes:             attrs.Notes,
		Position:          position,
	}
	if err := createItem(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}

	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	p.emitAudit(planWeekID, "plan.item_added", map[string]interface{}{
		"recipe_id":    attrs.RecipeID.String(),
		"plan_item_id": m.Id().String(),
	})
	return m, nil
}

type UpdateAttrs struct {
	Day               *time.Time
	Slot              *string
	ServingMultiplier **float64
	PlannedServings   **int
	Notes             **string
	Position          *int
}

func (p *Processor) UpdateItem(itemID uuid.UUID, planStartsOn time.Time, attrs UpdateAttrs) (Model, error) {
	e, err := getByID(p.db.WithContext(p.ctx), itemID)
	if err != nil {
		return Model{}, ErrNotFound
	}

	if attrs.Day != nil {
		if err := validateDayInWeek(*attrs.Day, planStartsOn); err != nil {
			return Model{}, err
		}
		e.Day = *attrs.Day
	}
	if attrs.Slot != nil {
		if !IsValidSlot(*attrs.Slot) {
			return Model{}, ErrInvalidSlot
		}
		e.Slot = *attrs.Slot
	}
	if attrs.ServingMultiplier != nil {
		e.ServingMultiplier = *attrs.ServingMultiplier
	}
	if attrs.PlannedServings != nil {
		e.PlannedServings = *attrs.PlannedServings
	}
	if attrs.Notes != nil {
		e.Notes = *attrs.Notes
	}
	if attrs.Position != nil {
		e.Position = *attrs.Position
	}

	if err := updateItem(p.db.WithContext(p.ctx), &e); err != nil {
		return Model{}, err
	}

	m, err := Make(e)
	if err != nil {
		return Model{}, err
	}
	p.emitAudit(e.PlanWeekId, "plan.item_updated", map[string]interface{}{
		"plan_item_id": m.Id().String(),
	})
	return m, nil
}

func (p *Processor) RemoveItem(itemID, planWeekID uuid.UUID) error {
	e, err := getByID(p.db.WithContext(p.ctx), itemID)
	if err != nil {
		return ErrNotFound
	}

	if err := deleteItem(p.db.WithContext(p.ctx), itemID); err != nil {
		return err
	}

	p.emitAudit(planWeekID, "plan.item_removed", map[string]interface{}{
		"plan_item_id": itemID.String(),
		"recipe_id":    e.RecipeId.String(),
	})
	return nil
}

func (p *Processor) GetByPlanWeekID(planWeekID uuid.UUID) ([]Model, error) {
	entities, err := getByPlanWeekID(p.db.WithContext(p.ctx), planWeekID)
	if err != nil {
		return nil, err
	}

	models := make([]Model, 0, len(entities))
	for _, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, nil
}

func (p *Processor) CountByPlanWeekID(planWeekID uuid.UUID) (int64, error) {
	return countByPlanWeekID(p.db.WithContext(p.ctx), planWeekID)
}

func (p *Processor) GetRecipeUsage(recipeIDs []uuid.UUID) (map[uuid.UUID]RecipeUsage, error) {
	usages, err := getRecipeUsage(p.db.WithContext(p.ctx), recipeIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]RecipeUsage, len(usages))
	for _, u := range usages {
		result[u.RecipeID] = u
	}
	return result, nil
}

func (p *Processor) CopyItems(sourcePlanWeekID, targetPlanWeekID uuid.UUID, dayOffset int) error {
	entities, err := getByPlanWeekID(p.db.WithContext(p.ctx), sourcePlanWeekID)
	if err != nil {
		return err
	}

	for _, e := range entities {
		newEntity := Entity{
			PlanWeekId:        targetPlanWeekID,
			Day:               e.Day.AddDate(0, 0, dayOffset),
			Slot:              e.Slot,
			RecipeId:          e.RecipeId,
			ServingMultiplier: e.ServingMultiplier,
			PlannedServings:   e.PlannedServings,
			Notes:             nil, // Notes are not copied per PRD
			Position:          e.Position,
		}
		if err := createItem(p.db.WithContext(p.ctx), &newEntity); err != nil {
			return err
		}
	}
	return nil
}

func (p *Processor) emitAudit(planWeekID uuid.UUID, action string, metadata map[string]interface{}) {
	t, ok := tenantctx.FromContext(p.ctx)
	if !ok {
		return
	}
	audit.Emit(p.l, p.db.WithContext(p.ctx), t.Id(), "plan", planWeekID, action, t.UserId(), metadata)
}

func validateDayInWeek(day, startsOn time.Time) error {
	dayDate := day.Truncate(24 * time.Hour)
	startDate := startsOn.Truncate(24 * time.Hour)
	endDate := startDate.AddDate(0, 0, 6)
	if dayDate.Before(startDate) || dayDate.After(endDate) {
		return ErrDayOutOfRange
	}
	return nil
}
