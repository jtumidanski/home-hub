package scheduler

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Scheduler manages periodic background tasks
type Scheduler interface {
	Start(ctx context.Context) error
	Stop() error
}

// RefreshFunc defines a function that refreshes weather data for a household
type RefreshFunc func(ctx context.Context, householdID uuid.UUID) error

// WeatherScheduler implements background refresh workers
type WeatherScheduler struct {
	currentRefresh  RefreshFunc
	forecastRefresh RefreshFunc
	currentTTL      time.Duration
	forecastTTL     time.Duration
	jitter          float64
	households      *HouseholdTracker
	logger          logrus.FieldLogger
	wg              sync.WaitGroup
	stopChan        chan struct{}
	started         bool
	mu              sync.Mutex
}

// NewWeatherScheduler creates a new scheduler
func NewWeatherScheduler(
	currentRefresh RefreshFunc,
	forecastRefresh RefreshFunc,
	currentTTL time.Duration,
	forecastTTL time.Duration,
	jitter float64,
	logger logrus.FieldLogger,
) *WeatherScheduler {
	return &WeatherScheduler{
		currentRefresh:  currentRefresh,
		forecastRefresh: forecastRefresh,
		currentTTL:      currentTTL,
		forecastTTL:     forecastTTL,
		jitter:          jitter,
		households:      NewHouseholdTracker(),
		logger:          logger,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the background refresh workers
func (s *WeatherScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return nil
	}

	s.started = true
	s.logger.Info("Starting weather scheduler")

	// Start current weather refresh worker
	s.wg.Add(1)
	go s.runCurrentRefresh(ctx)

	// Start forecast weather refresh worker
	s.wg.Add(1)
	go s.runForecastRefresh(ctx)

	return nil
}

// Stop gracefully shuts down the scheduler
func (s *WeatherScheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	s.logger.Info("Stopping weather scheduler")
	close(s.stopChan)
	s.wg.Wait()
	s.started = false

	return nil
}

// TrackHousehold adds a household to be refreshed
func (s *WeatherScheduler) TrackHousehold(householdID uuid.UUID) {
	s.households.Track(householdID)
}

// runCurrentRefresh periodically refreshes current weather for tracked households
func (s *WeatherScheduler) runCurrentRefresh(ctx context.Context) {
	defer s.wg.Done()

	interval := s.currentTTL
	ticker := time.NewTicker(s.addJitter(interval))
	defer ticker.Stop()

	s.logger.WithField("interval", interval).Info("Current weather refresh worker started")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Current weather refresh worker stopped (context cancelled)")
			return
		case <-s.stopChan:
			s.logger.Info("Current weather refresh worker stopped")
			return
		case <-ticker.C:
			s.refreshAllCurrent(ctx)
			// Reset ticker with new jitter
			ticker.Reset(s.addJitter(interval))
		}
	}
}

// runForecastRefresh periodically refreshes forecast weather for tracked households
func (s *WeatherScheduler) runForecastRefresh(ctx context.Context) {
	defer s.wg.Done()

	interval := s.forecastTTL
	ticker := time.NewTicker(s.addJitter(interval))
	defer ticker.Stop()

	s.logger.WithField("interval", interval).Info("Forecast weather refresh worker started")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Forecast weather refresh worker stopped (context cancelled)")
			return
		case <-s.stopChan:
			s.logger.Info("Forecast weather refresh worker stopped")
			return
		case <-ticker.C:
			s.refreshAllForecast(ctx)
			// Reset ticker with new jitter
			ticker.Reset(s.addJitter(interval))
		}
	}
}

// refreshAllCurrent refreshes current weather for all tracked households
func (s *WeatherScheduler) refreshAllCurrent(ctx context.Context) {
	households := s.households.GetAll()
	if len(households) == 0 {
		s.logger.Debug("No households tracked for current weather refresh")
		return
	}

	s.logger.WithField("count", len(households)).Info("Refreshing current weather for tracked households")

	var successCount, errorCount int

	for _, householdID := range households {
		if err := s.currentRefresh(ctx, householdID); err != nil {
			s.logger.WithError(err).WithField("household_id", householdID).Warn("Failed to refresh current weather")
			errorCount++
		} else {
			successCount++
		}
	}

	s.logger.WithFields(logrus.Fields{
		"success": successCount,
		"errors":  errorCount,
		"total":   len(households),
	}).Info("Current weather refresh completed")
}

// refreshAllForecast refreshes forecast weather for all tracked households
func (s *WeatherScheduler) refreshAllForecast(ctx context.Context) {
	households := s.households.GetAll()
	if len(households) == 0 {
		s.logger.Debug("No households tracked for forecast weather refresh")
		return
	}

	s.logger.WithField("count", len(households)).Info("Refreshing forecast weather for tracked households")

	var successCount, errorCount int

	for _, householdID := range households {
		if err := s.forecastRefresh(ctx, householdID); err != nil {
			s.logger.WithError(err).WithField("household_id", householdID).Warn("Failed to refresh forecast weather")
			errorCount++
		} else {
			successCount++
		}
	}

	s.logger.WithFields(logrus.Fields{
		"success": successCount,
		"errors":  errorCount,
		"total":   len(households),
	}).Info("Forecast weather refresh completed")
}

// addJitter adds random jitter to an interval to prevent thundering herd
func (s *WeatherScheduler) addJitter(interval time.Duration) time.Duration {
	if s.jitter == 0 {
		return interval
	}

	// Calculate jitter range: ±jitter%
	jitterRange := float64(interval) * s.jitter
	jitterAmount := (rand.Float64()*2 - 1) * jitterRange // Random value in [-jitterRange, +jitterRange]

	jittered := time.Duration(float64(interval) + jitterAmount)

	s.logger.WithFields(logrus.Fields{
		"original": interval,
		"jittered": jittered,
		"delta":    jittered - interval,
	}).Debug("Applied jitter to interval")

	return jittered
}
