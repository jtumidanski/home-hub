package planner

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

type ConfigAttrs struct {
	Classification     *string
	ServingsYield      *int
	EatWithinDays      *int
	MinGapDays         *int
	MaxConsecutiveDays *int
}

func (p *Processor) CreateOrUpdate(recipeID uuid.UUID, attrs ConfigAttrs) (Model, error) {
	existing, err := getByRecipeID(recipeID)(p.db.WithContext(p.ctx))()
	now := time.Now().UTC()

	if err == gorm.ErrRecordNotFound {
		e := Entity{
			Id:                 uuid.New(),
			RecipeId:           recipeID,
			Classification:     attrs.Classification,
			ServingsYield:      attrs.ServingsYield,
			EatWithinDays:      attrs.EatWithinDays,
			MinGapDays:         attrs.MinGapDays,
			MaxConsecutiveDays: attrs.MaxConsecutiveDays,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := createConfig(p.db.WithContext(p.ctx), &e); err != nil {
			return Model{}, err
		}
		return Make(e)
	}
	if err != nil {
		return Model{}, err
	}

	existing.Classification = attrs.Classification
	existing.ServingsYield = attrs.ServingsYield
	existing.EatWithinDays = attrs.EatWithinDays
	existing.MinGapDays = attrs.MinGapDays
	existing.MaxConsecutiveDays = attrs.MaxConsecutiveDays
	if err := updateConfig(p.db.WithContext(p.ctx), &existing); err != nil {
		return Model{}, err
	}
	return Make(existing)
}

func (p *Processor) GetByRecipeID(recipeID uuid.UUID) (Model, error) {
	e, err := getByRecipeID(recipeID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

// GetByRecipeIDs fetches planner configs for many recipes in a single query and
// returns them keyed by recipe_id. Empty input returns an empty map without
// hitting the database.
func (p *Processor) GetByRecipeIDs(recipeIDs []uuid.UUID) (map[uuid.UUID]Model, error) {
	result := make(map[uuid.UUID]Model)
	if len(recipeIDs) == 0 {
		return result, nil
	}
	entities, err := getByRecipeIDs(recipeIDs)(p.db.WithContext(p.ctx))
	if err != nil {
		return nil, err
	}
	for _, e := range entities {
		m, err := Make(e)
		if err != nil {
			return nil, err
		}
		result[e.RecipeId] = m
	}
	return result, nil
}

type Readiness struct {
	Ready  bool     `json:"plannerReady"`
	Issues []string `json:"plannerIssues"`
}

func ComputeReadiness(config *Model, recipeServings *int) Readiness {
	r := Readiness{Ready: true}

	if config == nil {
		r.Ready = false
		r.Issues = append(r.Issues, "planner configuration is missing")
		return r
	}

	if config.Classification() == "" {
		r.Ready = false
		r.Issues = append(r.Issues, "classification is not set")
	}

	hasServings := false
	if config.ServingsYield() != nil {
		hasServings = true
	} else if recipeServings != nil {
		hasServings = true
	}
	if !hasServings {
		r.Ready = false
		r.Issues = append(r.Issues, "servings is not set")
	}

	if r.Issues == nil {
		r.Issues = []string{}
	}
	return r
}
