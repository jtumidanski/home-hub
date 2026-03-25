package refresh

import (
	"context"
	"time"

	"github.com/jtumidanski/home-hub/services/weather-service/internal/forecast"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func StartRefreshLoop(ctx context.Context, db *gorm.DB, client *openmeteo.Client, interval time.Duration, l logrus.FieldLogger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	l.WithField("interval", interval.String()).Info("weather refresh loop started")

	for {
		select {
		case <-ctx.Done():
			l.Info("weather refresh loop stopped")
			return
		case <-ticker.C:
			refreshAll(ctx, db, client, l)
		}
	}
}

func refreshAll(ctx context.Context, db *gorm.DB, client *openmeteo.Client, l logrus.FieldLogger) {
	proc := forecast.NewProcessor(l, ctx, db, client)
	entries, err := proc.AllCacheEntries()
	if err != nil {
		l.WithError(err).Error("failed to list weather cache entries for refresh")
		return
	}

	l.WithField("count", len(entries)).Info("refreshing weather cache")

	for _, e := range entries {
		if err := proc.RefreshCache(e); err != nil {
			l.WithError(err).WithField("household_id", e.HouseholdId.String()).Warn("failed to refresh weather cache entry")
		}
	}
}
