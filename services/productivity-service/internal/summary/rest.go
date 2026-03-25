package summary

import (
	"time"

	"github.com/google/uuid"
)

type TaskSummary struct {
	Id                  uuid.UUID `json:"-"`
	PendingCount        int64     `json:"pendingCount"`
	CompletedTodayCount int64     `json:"completedTodayCount"`
	OverdueCount        int64     `json:"overdueCount"`
}

func (r TaskSummary) GetName() string      { return "task-summaries" }
func (r TaskSummary) GetID() string         { return "current" }
func (r *TaskSummary) SetID(_ string) error { return nil }

type ReminderSummary struct {
	Id            uuid.UUID `json:"-"`
	DueNowCount   int64     `json:"dueNowCount"`
	UpcomingCount int64     `json:"upcomingCount"`
	SnoozedCount  int64     `json:"snoozedCount"`
}

func (r ReminderSummary) GetName() string      { return "reminder-summaries" }
func (r ReminderSummary) GetID() string         { return "current" }
func (r *ReminderSummary) SetID(_ string) error { return nil }

type DashboardSummary struct {
	Id               uuid.UUID `json:"-"`
	PendingTaskCount int64     `json:"pendingTaskCount"`
	DueReminderCount int64     `json:"dueReminderCount"`
	GeneratedAt      time.Time `json:"generatedAt"`
}

func (r DashboardSummary) GetName() string      { return "dashboard-summaries" }
func (r DashboardSummary) GetID() string         { return "current" }
func (r *DashboardSummary) SetID(_ string) error { return nil }
