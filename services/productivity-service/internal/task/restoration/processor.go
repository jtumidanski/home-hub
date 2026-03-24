package restoration

import (
	"context"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/task"
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

func (p *Processor) Create(tenantID, householdID, taskID, userID uuid.UUID) (Model, error) {
	taskProc := task.NewProcessor(p.l, p.ctx, p.db)
	if err := taskProc.Restore(taskID); err != nil {
		return Model{}, err
	}
	e, err := create(p.db.WithContext(p.ctx), tenantID, householdID, taskID, userID)
	if err != nil {
		return Model{}, err
	}
	return Make(e)
}
