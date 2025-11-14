package reminder

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
)

var (
	ErrReminderNotFound = errors.New("reminder not found")
	ErrUnauthorized     = errors.New("user does not have access to this reminder")
)

// CreateInput contains the data needed to create a new reminder
type CreateInput struct {
	UserId      uuid.UUID
	HouseholdId uuid.UUID
	Name        string
	Description string
	RemindAt    time.Time
}

// UpdateInput contains the data to update an existing reminder
type UpdateInput struct {
	Name        *string
	Description *string
	RemindAt    *time.Time
	Status      *Status
}

// Processor handles business logic for reminder operations
type Processor struct {
	log logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

// NewProcessor creates a new reminder processor with dependencies
func NewProcessor(log logrus.FieldLogger, ctx context.Context, db *gorm.DB) Processor {
	return Processor{
		log: log,
		ctx: ctx,
		db:  db,
	}
}

// Create creates a new reminder with the provided input
func (p Processor) Create(input CreateInput) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithFields(logrus.Fields{
			"userId":   input.UserId,
			"name":     input.Name,
			"remindAt": input.RemindAt.Format(time.RFC3339),
		}).Info("Creating new reminder")

		// Build the model
		model, err := NewBuilder().
			SetUserId(input.UserId).
			SetHouseholdId(input.HouseholdId).
			SetName(input.Name).
			SetDescription(input.Description).
			SetRemindAt(input.RemindAt).
			SetStatus(StatusActive).
			SetSnoozeCount(0).
			Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build reminder model")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Create(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to create reminder in database")
			return Model{}, err
		}

		p.log.WithField("reminderId", model.Id()).Info("Reminder created successfully")
		return Make(entity)
	}
}

// GetById retrieves a reminder by ID and verifies user authorization
func (p Processor) GetById(id uuid.UUID, userId uuid.UUID) ops.Provider[Model] {
	return func() (Model, error) {
		model, err := GetById(p.db)(id)()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return Model{}, ErrReminderNotFound
			}
			return Model{}, err
		}

		// Verify user owns this reminder
		if model.UserId() != userId {
			p.log.WithFields(logrus.Fields{
				"reminderId":   id,
				"reminderUserId": model.UserId(),
				"requesterId":    userId,
			}).Warn("Unauthorized reminder access attempt")
			return Model{}, ErrUnauthorized
		}

		return model, nil
	}
}

// List retrieves all reminders for a user
func (p Processor) List(userId uuid.UUID) ops.Provider[[]Model] {
	return func() ([]Model, error) {
		p.log.WithField("userId", userId).Debug("Listing reminders for user")
		return GetByUserId(p.db)(userId)()
	}
}

// ListByStatus retrieves all reminders for a user with a specific status
func (p Processor) ListByStatus(userId uuid.UUID, status Status) ops.Provider[[]Model] {
	return func() ([]Model, error) {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"status": status,
		}).Debug("Listing reminders for user by status")
		return GetByUserIdAndStatus(p.db)(userId, status)()
	}
}

// ListByHousehold retrieves all reminders for a household
func (p Processor) ListByHousehold(householdId uuid.UUID) ops.Provider[[]Model] {
	return func() ([]Model, error) {
		p.log.WithField("householdId", householdId).Debug("Listing reminders for household")
		return GetByHouseholdId(p.db)(householdId)()
	}
}

// Update updates an existing reminder
func (p Processor) Update(id uuid.UUID, userId uuid.UUID, input UpdateInput) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithField("reminderId", id).Info("Updating reminder")

		// Fetch existing reminder and verify authorization
		existing, err := p.GetById(id, userId)()
		if err != nil {
			return Model{}, err
		}

		// Build updated model
		builder := existing.Builder().
			SetUpdatedAt(time.Now())

		if input.Name != nil {
			builder.SetName(*input.Name)
		}

		if input.Description != nil {
			builder.SetDescription(*input.Description)
		}

		if input.RemindAt != nil {
			builder.SetRemindAt(*input.RemindAt)
		}

		if input.Status != nil {
			builder.SetStatus(*input.Status)
			// If marking as dismissed, set dismissal timestamp
			if *input.Status == StatusDismissed && existing.Status() != StatusDismissed {
				now := time.Now()
				builder.SetDismissedAt(now)
			}
			// If marking as not dismissed, clear dismissal timestamp
			if *input.Status != StatusDismissed && existing.Status() == StatusDismissed {
				builder.ClearDismissedAt()
			}
		}

		model, err := builder.Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build updated reminder model")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Save(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to update reminder in database")
			return Model{}, err
		}

		p.log.WithField("reminderId", id).Info("Reminder updated successfully")
		return Make(entity)
	}
}

// Snooze snoozes a reminder to a new time and increments the snooze counter
func (p Processor) Snooze(id uuid.UUID, userId uuid.UUID, newRemindAt time.Time) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithFields(logrus.Fields{
			"reminderId":   id,
			"newRemindAt": newRemindAt.Format(time.RFC3339),
		}).Info("Snoozing reminder")

		// Fetch existing reminder and verify authorization
		existing, err := p.GetById(id, userId)()
		if err != nil {
			return Model{}, err
		}

		// Mark as snoozed
		model, err := existing.Builder().
			MarkSnoozed(newRemindAt).
			Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build snoozed reminder model")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Save(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to snooze reminder in database")
			return Model{}, err
		}

		p.log.WithFields(logrus.Fields{
			"reminderId":   id,
			"snoozeCount": model.SnoozeCount(),
		}).Info("Reminder snoozed successfully")
		return Make(entity)
	}
}

// Dismiss dismisses a reminder
func (p Processor) Dismiss(id uuid.UUID, userId uuid.UUID) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithField("reminderId", id).Info("Dismissing reminder")

		// Fetch existing reminder and verify authorization
		existing, err := p.GetById(id, userId)()
		if err != nil {
			return Model{}, err
		}

		// Check if already dismissed
		if existing.IsDismissed() {
			p.log.WithField("reminderId", id).Warn("Reminder is already dismissed")
			return existing, nil
		}

		// Mark as dismissed
		model, err := existing.Builder().
			MarkDismissed().
			Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build dismissed reminder model")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Save(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to dismiss reminder in database")
			return Model{}, err
		}

		p.log.WithField("reminderId", id).Info("Reminder dismissed successfully")
		return Make(entity)
	}
}

// Delete removes a reminder by ID
func (p Processor) Delete(id uuid.UUID, userId uuid.UUID) error {
	p.log.WithField("reminderId", id).Info("Deleting reminder")

	// First verify the reminder exists and user has access
	_, err := p.GetById(id, userId)()
	if err != nil {
		return err
	}

	// Delete the reminder
	result := p.db.Delete(&Entity{}, "id = ?", id)
	if result.Error != nil {
		p.log.WithError(result.Error).Error("Failed to delete reminder")
		return result.Error
	}

	if result.RowsAffected == 0 {
		p.log.WithField("reminderId", id).Warn("Reminder not found for deletion")
		return ErrReminderNotFound
	}

	p.log.WithField("reminderId", id).Info("Reminder deleted successfully")
	return nil
}
