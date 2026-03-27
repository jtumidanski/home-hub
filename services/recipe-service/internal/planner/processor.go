package planner

import (
	"context"

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
	e, err := upsert(p.db.WithContext(p.ctx), recipeID, attrs.Classification, attrs.ServingsYield, attrs.EatWithinDays, attrs.MinGapDays, attrs.MaxConsecutiveDays)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}

func (p *Processor) GetByRecipeID(recipeID uuid.UUID) (Model, error) {
	e, err := getByRecipeID(p.db.WithContext(p.ctx), recipeID)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
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
