package event

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/connection"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/crypto"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/googlecal"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/source"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ConnectionProcessor interface {
	ByIDProvider(id uuid.UUID) model.Provider[connection.Model]
	GetOrRefreshAccessToken(conn connection.Model, gcClient *googlecal.Client, enc *crypto.Encryptor) (string, error)
}

type SourceProcessor interface {
	ByIDProvider(id uuid.UUID) model.Provider[source.Model]
}

var (
	ErrRangeTooLarge      = errors.New("query range exceeds 90 days")
	ErrConnectionNotFound = errors.New("connection not found")
	ErrNotOwner           = errors.New("not the connection owner")
	ErrNoWriteAccess      = errors.New("connection does not have write access")
	ErrSourceNotFound     = errors.New("source not found")
	ErrSourceMismatch     = errors.New("source does not belong to connection")
	ErrEventNotFound      = errors.New("event not found")
	ErrEventMismatch      = errors.New("event does not belong to connection")
	ErrGoogleWriteDenied  = errors.New("google calendar write denied")
	ErrAuthFailed         = errors.New("failed to authenticate with Google")
)

type SyncConnectionFunc func(conn connection.Model)

const maxQueryRange = 90 * 24 * time.Hour

type Processor struct {
	l        logrus.FieldLogger
	ctx      context.Context
	db       *gorm.DB
	connProc ConnectionProcessor
	srcProc  SourceProcessor
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

func NewMutationProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB, connProc ConnectionProcessor, srcProc SourceProcessor) *Processor {
	return &Processor{l: l, ctx: ctx, db: db, connProc: connProc, srcProc: srcProc}
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
		// Google Calendar requires an IANA time zone on start/end whenever the
		// event has a recurrence rule. Set it unconditionally so non-recurring
		// edits later (which may add recurrence) round-trip safely.
		tz := resolveTimeZone(input.TimeZone)
		gcEvent.Start = &googlecal.EventTime{DateTime: &startTime, TimeZone: tz}
		gcEvent.End = &googlecal.EventTime{DateTime: &endTime, TimeZone: tz}
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
	tz := ""
	if input.TimeZone != nil {
		tz = *input.TimeZone
	}
	tz = resolveTimeZone(tz)
	if input.Start != nil {
		if input.AllDay != nil && *input.AllDay {
			gcUpdate.Start = &googlecal.EventTime{Date: parseDate(*input.Start)}
		} else {
			st, _ := time.Parse(time.RFC3339, *input.Start)
			gcUpdate.Start = &googlecal.EventTime{DateTime: &st, TimeZone: tz}
		}
	}
	if input.End != nil {
		if input.AllDay != nil && *input.AllDay {
			// Google Calendar API treats end date as exclusive, so add one day.
			gcUpdate.End = &googlecal.EventTime{Date: addDay(parseDate(*input.End))}
		} else {
			et, _ := time.Parse(time.RFC3339, *input.End)
			gcUpdate.End = &googlecal.EventTime{DateTime: &et, TimeZone: tz}
		}
	}

	googleEventID := evt.ExternalID()
	if input.Scope == "all" {
		googleEventID = extractBaseEventID(googleEventID)
	}

	_, err := gcClient.UpdateEvent(p.ctx, accessToken, evt.GoogleCalendarID(), googleEventID, gcUpdate)
	return err
}

func (p *Processor) validateConnectionForWrite(connID, userID uuid.UUID, gcClient *googlecal.Client, enc *crypto.Encryptor) (connection.Model, string, error) {
	conn, err := p.connProc.ByIDProvider(connID)()
	if err != nil {
		return connection.Model{}, "", ErrConnectionNotFound
	}
	if conn.UserID() != userID {
		return connection.Model{}, "", ErrNotOwner
	}
	if !conn.WriteAccess() {
		return connection.Model{}, "", ErrNoWriteAccess
	}
	accessToken, err := p.connProc.GetOrRefreshAccessToken(conn, gcClient, enc)
	if err != nil {
		return connection.Model{}, "", ErrAuthFailed
	}
	return conn, accessToken, nil
}

func (p *Processor) CreateEventOnGoogle(connID, calID, userID uuid.UUID, input CreateEventRequest, gcClient *googlecal.Client, enc *crypto.Encryptor, syncConn SyncConnectionFunc) error {
	conn, accessToken, err := p.validateConnectionForWrite(connID, userID, gcClient, enc)
	if err != nil {
		return err
	}

	src, err := p.srcProc.ByIDProvider(calID)()
	if err != nil {
		return ErrSourceNotFound
	}
	if src.ConnectionID() != connID {
		return ErrSourceMismatch
	}

	err = p.CreateOnGoogle(gcClient, accessToken, src.ExternalID(), input)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			return ErrGoogleWriteDenied
		}
		return err
	}

	if syncConn != nil {
		go syncConn(conn)
	}
	return nil
}

func (p *Processor) UpdateEventOnGoogle(connID, eventID, userID uuid.UUID, input UpdateEventRequest, gcClient *googlecal.Client, enc *crypto.Encryptor, syncConn SyncConnectionFunc) (Model, error) {
	conn, accessToken, err := p.validateConnectionForWrite(connID, userID, gcClient, enc)
	if err != nil {
		return Model{}, err
	}

	evt, err := p.ByID(eventID)
	if err != nil {
		return Model{}, ErrEventNotFound
	}
	if evt.ConnectionID() != connID {
		return Model{}, ErrEventMismatch
	}

	err = p.UpdateOnGoogle(gcClient, accessToken, evt, input)
	if err != nil {
		return Model{}, err
	}

	if syncConn != nil {
		go syncConn(conn)
	}

	updated, err := p.ByID(eventID)
	if err != nil {
		return Model{}, err
	}
	return updated, nil
}

func (p *Processor) DeleteEventOnGoogle(connID, eventID, userID uuid.UUID, scope string, gcClient *googlecal.Client, enc *crypto.Encryptor, syncConn SyncConnectionFunc) error {
	conn, accessToken, err := p.validateConnectionForWrite(connID, userID, gcClient, enc)
	if err != nil {
		return err
	}

	evt, err := p.ByID(eventID)
	if err != nil {
		return ErrEventNotFound
	}
	if evt.ConnectionID() != connID {
		return ErrEventMismatch
	}

	err = p.DeleteOnGoogle(gcClient, accessToken, evt, scope)
	if err != nil {
		return err
	}

	if syncConn != nil {
		go syncConn(conn)
	}
	return nil
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

// resolveTimeZone returns an IANA time zone for a Google Calendar event payload.
// Google rejects recurring-event creates without start.timeZone/end.timeZone
// (a UTC offset on the dateTime is not sufficient), so we always send one and
// fall back to UTC when the client omits the field.
func resolveTimeZone(tz string) string {
	if tz == "" {
		return "UTC"
	}
	return tz
}

func (p *Processor) noTenantDB() *gorm.DB {
	return p.db.WithContext(database.WithoutTenantFilter(p.ctx))
}
