package event

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/googlecal"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrRangeTooLarge = errors.New("query range exceeds 90 days")
)

const maxQueryRange = 90 * 24 * time.Hour

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func (p *Processor) ByID(id uuid.UUID) (Model, error) {
	return model.Map(Make)(getByID(id)(p.db.WithContext(p.ctx)))()
}

func (p *Processor) QueryByHouseholdAndTimeRange(householdID uuid.UUID, start, end time.Time) ([]Model, error) {
	if end.Sub(start) > maxQueryRange {
		return nil, ErrRangeTooLarge
	}
	return model.SliceMap(Make)(getVisibleByHouseholdAndTimeRange(householdID, start, end)(p.db.WithContext(p.ctx)))()
}

func (p *Processor) Upsert(e Entity) error {
	return upsert(p.noTenantDB(), e)
}

func (p *Processor) DeleteBySourceAndExternalIDs(sourceID uuid.UUID, externalIDs []string) error {
	return deleteBySourceAndExternalIDs(p.noTenantDB(), sourceID, externalIDs)
}

func (p *Processor) DeleteBySourceExcludingExternalIDs(sourceID uuid.UUID, keepIDs []string) error {
	return deleteBySourceExcludingExternalIDs(p.noTenantDB(), sourceID, keepIDs)
}

func (p *Processor) DeleteByConnection(connectionID uuid.UUID) error {
	return deleteByConnection(p.noTenantDB(), connectionID)
}

func (p *Processor) DeleteBySource(sourceID uuid.UUID) error {
	return deleteBySource(p.noTenantDB(), sourceID)
}

func (p *Processor) CountByConnection(connectionID uuid.UUID) (int64, error) {
	return countByConnection(p.noTenantDB(), connectionID)
}

func (p *Processor) CreateOnGoogle(gcClient *googlecal.Client, accessToken, calendarID string, input CreateEventRequest) error {
	gcEvent := googlecal.InsertEventRequest{
		Summary:     input.Title,
		Location:    input.Location,
		Description: input.Description,
		Recurrence:  input.Recurrence,
	}

	if input.AllDay {
		startDate := parseDate(input.Start)
		endDate := parseDate(input.End)
		if endDate == "" {
			endDate = startDate
		}
		// Google Calendar API treats end date as exclusive, so add one day.
		gcEvent.Start = &googlecal.EventTime{Date: startDate}
		gcEvent.End = &googlecal.EventTime{Date: addDay(endDate)}
	} else {
		startTime, _ := time.Parse(time.RFC3339, input.Start)
		endTime, _ := time.Parse(time.RFC3339, input.End)
		if endTime.IsZero() {
			endTime = startTime.Add(time.Hour)
		}
		gcEvent.Start = &googlecal.EventTime{DateTime: &startTime}
		gcEvent.End = &googlecal.EventTime{DateTime: &endTime}
	}

	_, err := gcClient.InsertEvent(p.ctx, accessToken, calendarID, gcEvent)
	return err
}

func (p *Processor) UpdateOnGoogle(gcClient *googlecal.Client, accessToken string, evt Model, input UpdateEventRequest) error {
	gcUpdate := googlecal.UpdateEventRequest{}
	if input.Title != nil {
		gcUpdate.Summary = input.Title
	}
	if input.Location != nil {
		gcUpdate.Location = input.Location
	}
	if input.Description != nil {
		gcUpdate.Description = input.Description
	}
	if input.Start != nil {
		if input.AllDay != nil && *input.AllDay {
			gcUpdate.Start = &googlecal.EventTime{Date: parseDate(*input.Start)}
		} else {
			st, _ := time.Parse(time.RFC3339, *input.Start)
			gcUpdate.Start = &googlecal.EventTime{DateTime: &st}
		}
	}
	if input.End != nil {
		if input.AllDay != nil && *input.AllDay {
			// Google Calendar API treats end date as exclusive, so add one day.
			gcUpdate.End = &googlecal.EventTime{Date: addDay(parseDate(*input.End))}
		} else {
			et, _ := time.Parse(time.RFC3339, *input.End)
			gcUpdate.End = &googlecal.EventTime{DateTime: &et}
		}
	}

	googleEventID := evt.ExternalID()
	if input.Scope == "all" {
		googleEventID = extractBaseEventID(googleEventID)
	}

	_, err := gcClient.UpdateEvent(p.ctx, accessToken, evt.GoogleCalendarID(), googleEventID, gcUpdate)
	return err
}

func (p *Processor) DeleteOnGoogle(gcClient *googlecal.Client, accessToken string, evt Model, scope string) error {
	googleEventID := evt.ExternalID()
	if scope == "all" {
		googleEventID = extractBaseEventID(googleEventID)
	}
	return gcClient.DeleteEvent(p.ctx, accessToken, evt.GoogleCalendarID(), googleEventID)
}

func extractBaseEventID(instanceID string) string {
	if idx := strings.LastIndex(instanceID, "_"); idx > 0 {
		suffix := instanceID[idx+1:]
		if len(suffix) >= 8 && isDigits(suffix[:8]) {
			return instanceID[:idx]
		}
	}
	return instanceID
}

func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func parseDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func addDay(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.AddDate(0, 0, 1).Format("2006-01-02")
}

func (p *Processor) noTenantDB() *gorm.DB {
	return p.db.WithContext(database.WithoutTenantFilter(p.ctx))
}
