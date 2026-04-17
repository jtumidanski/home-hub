package summary

import (
	"context"
	"time"

	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder"
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

func (p *Processor) TaskSummary(date time.Time) (TaskSummary, error) {
	taskProc := task.NewProcessor(p.l, p.ctx, p.db)

	pending, err := taskProc.PendingCount()
	if err != nil {
		return TaskSummary{}, err
	}
	completed, err := taskProc.CompletedTodayCount(date)
	if err != nil {
		return TaskSummary{}, err
	}
	overdue, err := taskProc.OverdueCount(date)
	if err != nil {
		return TaskSummary{}, err
	}

	return TaskSummary{
		PendingCount:        pending,
		CompletedTodayCount: completed,
		OverdueCount:        overdue,
	}, nil
}

func (p *Processor) ReminderSummary() (ReminderSummary, error) {
	remProc := reminder.NewProcessor(p.l, p.ctx, p.db)

	dueNow, err := remProc.DueNowCount()
	if err != nil {
		return ReminderSummary{}, err
	}
	upcoming, err := remProc.UpcomingCount()
	if err != nil {
		return ReminderSummary{}, err
	}
	snoozed, err := remProc.SnoozedCount()
	if err != nil {
		return ReminderSummary{}, err
	}

	return ReminderSummary{
		DueNowCount:   dueNow,
		UpcomingCount: upcoming,
		SnoozedCount:  snoozed,
	}, nil
}

func (p *Processor) DashboardSummary() (DashboardSummary, error) {
	taskProc := task.NewProcessor(p.l, p.ctx, p.db)
	remProc := reminder.NewProcessor(p.l, p.ctx, p.db)

	pending, err := taskProc.PendingCount()
	if err != nil {
		return DashboardSummary{}, err
	}
	dueNow, err := remProc.DueNowCount()
	if err != nil {
		return DashboardSummary{}, err
	}

	return DashboardSummary{
		PendingTaskCount: pending,
		DueReminderCount: dueNow,
		GeneratedAt:      time.Now().UTC(),
	}, nil
}
