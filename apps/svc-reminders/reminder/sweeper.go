package reminder

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PerformSweep finds and dismisses all overdue reminders (more than 24 hours past remind_at)
func PerformSweep(db *gorm.DB, log logrus.FieldLogger) error {
	log.Debug("Starting reminder sweep for overdue reminders")

	// Query overdue reminders
	overdue, err := GetOverdueForDismissal(db)()
	if err != nil {
		log.WithError(err).Error("Failed to query overdue reminders")
		return err
	}

	if len(overdue) == 0 {
		log.Debug("No overdue reminders found")
		return nil
	}

	log.WithField("count", len(overdue)).Info("Found overdue reminders, dismissing...")

	// Dismiss each overdue reminder
	dismissed := 0
	for _, reminder := range overdue {
		// Build dismissed version
		updated, err := reminder.Builder().MarkDismissed().Build()
		if err != nil {
			log.WithError(err).WithField("reminderId", reminder.Id()).Error("Failed to build dismissed reminder")
			continue
		}

		// Save to database
		entity := updated.ToEntity()
		if err := db.Save(&entity).Error; err != nil {
			log.WithError(err).WithField("reminderId", reminder.Id()).Error("Failed to save dismissed reminder")
			continue
		}

		dismissed++
		log.WithFields(logrus.Fields{
			"reminderId": reminder.Id(),
			"userId":     reminder.UserId(),
			"name":       reminder.Name(),
			"remindAt":   reminder.RemindAt().Format(time.RFC3339),
		}).Debug("Dismissed overdue reminder")
	}

	log.WithFields(logrus.Fields{
		"total":     len(overdue),
		"dismissed": dismissed,
		"failed":    len(overdue) - dismissed,
	}).Info("Reminder sweep completed")

	return nil
}

// StartSweeper starts the background sweeper goroutine
// Runs PerformSweep every 1 minute until the context is cancelled
func StartSweeper(db *gorm.DB, log logrus.FieldLogger, ctx context.Context) {
	log.Info("Starting reminder sweeper (1-minute interval)")

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Run immediately on startup
	if err := PerformSweep(db, log); err != nil {
		log.WithError(err).Error("Initial sweep failed")
	}

	for {
		select {
		case <-ticker.C:
			if err := PerformSweep(db, log); err != nil {
				log.WithError(err).Error("Sweep failed")
			}
		case <-ctx.Done():
			log.Info("Reminder sweeper shutting down")
			return
		}
	}
}
