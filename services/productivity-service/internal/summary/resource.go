package summary

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/manyminds/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/reminder"
	"github.com/jtumidanski/home-hub/services/productivity-service/internal/task"
	"github.com/jtumidanski/home-hub/shared/go/server"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TaskSummary struct {
	Id                uuid.UUID `json:"-"`
	PendingCount      int64     `json:"pendingCount"`
	CompletedTodayCount int64   `json:"completedTodayCount"`
	OverdueCount      int64     `json:"overdueCount"`
}

func (r TaskSummary) GetName() string       { return "task-summaries" }
func (r TaskSummary) GetID() string          { return "current" }
func (r *TaskSummary) SetID(_ string) error  { return nil }

type ReminderSummary struct {
	Id            uuid.UUID `json:"-"`
	DueNowCount   int64     `json:"dueNowCount"`
	UpcomingCount int64     `json:"upcomingCount"`
	SnoozedCount  int64     `json:"snoozedCount"`
}

func (r ReminderSummary) GetName() string       { return "reminder-summaries" }
func (r ReminderSummary) GetID() string          { return "current" }
func (r *ReminderSummary) SetID(_ string) error  { return nil }

type DashboardSummary struct {
	Id               uuid.UUID `json:"-"`
	HouseholdName    string    `json:"householdName"`
	Timezone         string    `json:"timezone"`
	PendingTaskCount int64     `json:"pendingTaskCount"`
	DueReminderCount int64     `json:"dueReminderCount"`
	GeneratedAt      time.Time `json:"generatedAt"`
}

func (r DashboardSummary) GetName() string       { return "dashboard-summaries" }
func (r DashboardSummary) GetID() string          { return "current" }
func (r *DashboardSummary) SetID(_ string) error  { return nil }

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)

		api.HandleFunc("/summary/tasks", rh("GetTaskSummary", taskSummaryHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/summary/reminders", rh("GetReminderSummary", reminderSummaryHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/summary/dashboard", rh("GetDashboardSummary", dashboardSummaryHandler(db))).Methods(http.MethodGet)
	}
}

func taskSummaryHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			proc := task.NewProcessor(logrus.StandardLogger(), r.Context(), db)
			pending, _ := proc.PendingCount()
			completed, _ := proc.CompletedTodayCount()
			overdue, _ := proc.OverdueCount()

			s := TaskSummary{PendingCount: pending, CompletedTodayCount: completed, OverdueCount: overdue}
			server.MarshalResponse[TaskSummary](d.Logger())(w)(c.ServerInformation())(r.URL.Query())(s)
		}
	}
}

func reminderSummaryHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			proc := reminder.NewProcessor(logrus.StandardLogger(), r.Context(), db)
			dueNow, _ := proc.DueNowCount()
			upcoming, _ := proc.UpcomingCount()
			snoozed, _ := proc.SnoozedCount()

			s := ReminderSummary{DueNowCount: dueNow, UpcomingCount: upcoming, SnoozedCount: snoozed}
			server.MarshalResponse[ReminderSummary](d.Logger())(w)(c.ServerInformation())(r.URL.Query())(s)
		}
	}
}

func dashboardSummaryHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			taskProc := task.NewProcessor(logrus.StandardLogger(), r.Context(), db)
			remProc := reminder.NewProcessor(logrus.StandardLogger(), r.Context(), db)

			pending, _ := taskProc.PendingCount()
			dueNow, _ := remProc.DueNowCount()

			s := DashboardSummary{
				PendingTaskCount: pending,
				DueReminderCount: dueNow,
				GeneratedAt:      time.Now().UTC(),
			}
			server.MarshalResponse[DashboardSummary](d.Logger())(w)(c.ServerInformation())(r.URL.Query())(s)
		}
	}
}
