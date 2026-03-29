package sync

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/connection"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/crypto"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/event"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/googlecal"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/oauthstate"
	"github.com/jtumidanski/home-hub/services/calendar-service/internal/source"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	maxJitter     = 60 * time.Second
	syncWindowPast   = 7 * 24 * time.Hour
	syncWindowFuture = 30 * 24 * time.Hour
)

type Engine struct {
	db       *gorm.DB
	gcClient *googlecal.Client
	enc      *crypto.Encryptor
	l        logrus.FieldLogger
}

func NewEngine(db *gorm.DB, gcClient *googlecal.Client, enc *crypto.Encryptor, l logrus.FieldLogger) *Engine {
	return &Engine{db: db, gcClient: gcClient, enc: enc, l: l}
}

func (e *Engine) StartLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	e.l.WithField("interval", interval.String()).Info("calendar sync loop started")

	for {
		select {
		case <-ctx.Done():
			e.l.Info("calendar sync loop stopped")
			return
		case <-ticker.C:
			e.syncAll(ctx)
		}
	}
}

func (e *Engine) SyncConnection(conn connection.Model) {
	ctx := context.Background()
	e.syncOne(ctx, conn)
}

func (e *Engine) syncAll(ctx context.Context) {
	stateProc := oauthstate.NewProcessor(e.l, ctx, e.db)
	stateProc.CleanupExpired()

	connProc := connection.NewProcessor(e.l, ctx, e.db)
	connections, err := connProc.AllConnected()
	if err != nil {
		e.l.WithError(err).Error("failed to list connected connections for sync")
		return
	}

	e.l.WithField("count", len(connections)).Info("starting calendar sync cycle")

	for _, conn := range connections {
		jitter := time.Duration(rand.Int63n(int64(maxJitter)))
		select {
		case <-ctx.Done():
			return
		case <-time.After(jitter):
			e.syncOne(ctx, conn)
		}
	}
}

func (e *Engine) syncOne(ctx context.Context, conn connection.Model) {
	l := e.l.WithFields(logrus.Fields{
		"connection_id": conn.Id().String(),
		"user_id":       conn.UserID().String(),
		"household_id":  conn.HouseholdID().String(),
	})

	l.Info("syncing calendar connection")

	accessToken, err := e.getValidAccessToken(ctx, conn)
	if err != nil {
		l.WithError(err).Warn("failed to get valid access token, marking disconnected")
		connProc := connection.NewProcessor(l, ctx, e.db)
		_ = connProc.UpdateStatus(conn.Id(), "disconnected")
		return
	}

	e.refreshCalendarList(ctx, conn, accessToken, l)

	srcProc := source.NewProcessor(l, ctx, e.db)
	sources, err := srcProc.ListByConnection(conn.Id())
	if err != nil {
		l.WithError(err).Error("failed to list sources")
		return
	}

	totalEvents := 0
	for _, src := range sources {
		count := e.syncSource(ctx, conn, src, accessToken, l)
		totalEvents += count
	}

	connProc := connection.NewProcessor(l, ctx, e.db)
	_ = connProc.UpdateSyncInfo(conn.Id(), totalEvents)
	l.WithField("total_events", totalEvents).Info("calendar sync completed")
}

func (e *Engine) getValidAccessToken(ctx context.Context, conn connection.Model) (string, error) {
	accessToken, err := e.enc.Decrypt(conn.AccessToken())
	if err != nil {
		return "", err
	}

	if !conn.IsTokenExpired() {
		return accessToken, nil
	}

	refreshToken, err := e.enc.Decrypt(conn.RefreshToken())
	if err != nil {
		return "", err
	}

	tokenResp, err := e.gcClient.RefreshToken(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	encAccess, err := e.enc.Encrypt(tokenResp.AccessToken)
	if err != nil {
		return "", err
	}

	tokenExpiry := time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	connProc := connection.NewProcessor(e.l, ctx, e.db)
	_ = connProc.UpdateTokens(conn.Id(), encAccess, tokenExpiry)

	return tokenResp.AccessToken, nil
}

func (e *Engine) refreshCalendarList(ctx context.Context, conn connection.Model, accessToken string, l logrus.FieldLogger) {
	calList, err := e.gcClient.ListCalendars(ctx, accessToken)
	if err != nil {
		l.WithError(err).Warn("failed to refresh calendar list")
		return
	}

	srcProc := source.NewProcessor(l, ctx, e.db)
	for _, cal := range calList.Items {
		_, err := srcProc.CreateOrUpdate(
			conn.TenantID(), conn.HouseholdID(), conn.Id(),
			cal.ID, cal.Summary, cal.Primary, cal.BackgroundColor,
		)
		if err != nil {
			l.WithError(err).WithField("calendar_id", cal.ID).Warn("failed to upsert source calendar")
		}
	}
}

func (e *Engine) syncSource(ctx context.Context, conn connection.Model, src source.Model, accessToken string, l logrus.FieldLogger) int {
	l = l.WithField("source_id", src.Id().String()).WithField("source_name", src.Name())

	now := time.Now().UTC()
	timeMin := now.Add(-syncWindowPast)
	timeMax := now.Add(syncWindowFuture)

	eventsResp, err := e.gcClient.ListEvents(ctx, accessToken, src.ExternalID(), timeMin, timeMax, src.SyncToken())
	if err != nil {
		if googlecal.IsSyncTokenInvalid(err) {
			l.Info("sync token invalid, performing full sync")
			srcProc := source.NewProcessor(l, ctx, e.db)
			_ = srcProc.ClearSyncToken(src.Id())

			eventsResp, err = e.gcClient.ListEvents(ctx, accessToken, src.ExternalID(), timeMin, timeMax, "")
			if err != nil {
				l.WithError(err).Error("full sync failed")
				return 0
			}
		} else {
			l.WithError(err).Error("failed to list events")
			return 0
		}
	}

	isFullSync := src.SyncToken() == ""

	evtProc := event.NewProcessor(l, ctx, e.db)
	var cancelledIDs []string
	var seenIDs []string
	eventCount := 0

	for _, ge := range eventsResp.Items {
		if ge.Status == "cancelled" {
			cancelledIDs = append(cancelledIDs, ge.ID)
			continue
		}

		seenIDs = append(seenIDs, ge.ID)

		allDay := false
		startTime := time.Time{}
		endTime := time.Time{}

		if ge.Start != nil {
			allDay = ge.Start.IsAllDay()
			startTime = ge.Start.Time()
		}
		if ge.End != nil {
			endTime = ge.End.Time()
			// Google Calendar API end dates are exclusive for all-day events; subtract one day to store inclusive.
			if allDay {
				endTime = endTime.AddDate(0, 0, -1)
			}
		}

		title := ge.Summary
		if title == "" {
			title = "(No title)"
		}

		visibility := ge.Visibility
		if visibility == "" {
			visibility = "default"
		}

		entity := event.Entity{
			TenantId:         conn.TenantID(),
			HouseholdId:      conn.HouseholdID(),
			ConnectionId:     conn.Id(),
			SourceId:         src.Id(),
			UserId:           conn.UserID(),
			ExternalId:       ge.ID,
			GoogleCalendarId: src.ExternalID(),
			Title:           title,
			Description:     ge.Description,
			StartTime:       startTime,
			EndTime:         endTime,
			AllDay:          allDay,
			Location:        ge.Location,
			Visibility:      visibility,
			UserDisplayName: conn.UserDisplayName(),
			UserColor:       conn.UserColor(),
		}

		if err := evtProc.Upsert(entity); err != nil {
			l.WithError(err).WithField("event_id", ge.ID).Warn("failed to upsert event")
		}
		eventCount++
	}

	if len(cancelledIDs) > 0 {
		if err := evtProc.DeleteBySourceAndExternalIDs(src.Id(), cancelledIDs); err != nil {
			l.WithError(err).Warn("failed to delete cancelled events")
		}
	}

	// During a full sync, remove any local events not present in Google's response.
	if isFullSync {
		if err := evtProc.DeleteBySourceExcludingExternalIDs(src.Id(), seenIDs); err != nil {
			l.WithError(err).Warn("failed to reconcile stale events")
		}
	}

	if eventsResp.NextSyncToken != "" {
		srcProc := source.NewProcessor(l, ctx, e.db)
		_ = srcProc.UpdateSyncToken(src.Id(), eventsResp.NextSyncToken)
	}

	l.WithField("event_count", eventCount).WithField("cancelled_count", len(cancelledIDs)).Debug("source sync completed")
	return eventCount
}

func (e *Engine) DeleteConnectionData(ctx context.Context, connectionID uuid.UUID) {
	l := e.l.WithField("connection_id", connectionID.String())

	evtProc := event.NewProcessor(l, ctx, e.db)
	if err := evtProc.DeleteByConnection(connectionID); err != nil {
		l.WithError(err).Error("failed to delete events for connection")
	}

	srcProc := source.NewProcessor(l, ctx, e.db)
	if err := srcProc.DeleteByConnection(connectionID); err != nil {
		l.WithError(err).Error("failed to delete sources for connection")
	}
}
