package dismissal

import (
	"context"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder"
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

func (p *Processor) Create(tenantID, householdID, reminderID, userID uuid.UUID) (Entity, error) {
	remProc := reminder.NewProcessor(p.l, p.ctx, p.db)
	if err := remProc.Dismiss(reminderID); err != nil {
		return Entity{}, err
	}
	return create(p.db.WithContext(p.ctx), tenantID, householdID, reminderID, userID)
}
