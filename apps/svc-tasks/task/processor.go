package task

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
	ErrTaskNotFound    = errors.New("task not found")
	ErrUnauthorized    = errors.New("user does not have access to this task")
	ErrAlreadyComplete = errors.New("task is already completed")
)

// CreateInput contains the data needed to create a new task
type CreateInput struct {
	UserId      uuid.UUID
	HouseholdId uuid.UUID
	Day         time.Time
	Title       string
	Description string
}

// UpdateInput contains the data to update an existing task
type UpdateInput struct {
	Title       *string
	Description *string
	Day         *time.Time
	Status      *Status
}

// Processor handles business logic for task operations
type Processor struct {
	log logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

// NewProcessor creates a new task processor with dependencies
func NewProcessor(log logrus.FieldLogger, ctx context.Context, db *gorm.DB) Processor {
	return Processor{
		log: log,
		ctx: ctx,
		db:  db,
	}
}

// Create creates a new task with the provided input
func (p Processor) Create(input CreateInput) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithFields(logrus.Fields{
			"userId": input.UserId,
			"title":  input.Title,
			"day":    input.Day.Format("2006-01-02"),
		}).Info("Creating new task")

		// Build the model
		model, err := NewBuilder().
			SetUserId(input.UserId).
			SetHouseholdId(input.HouseholdId).
			SetDay(input.Day).
			SetTitle(input.Title).
			SetDescription(input.Description).
			SetStatus(StatusIncomplete).
			Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build task model")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Create(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to create task in database")
			return Model{}, err
		}

		p.log.WithField("taskId", model.Id()).Info("Task created successfully")
		return Make(entity)
	}
}

// GetById retrieves a task by ID and verifies user authorization
func (p Processor) GetById(id uuid.UUID, userId uuid.UUID) ops.Provider[Model] {
	return func() (Model, error) {
		model, err := GetById(p.db)(id)()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return Model{}, ErrTaskNotFound
			}
			return Model{}, err
		}

		// Verify user owns this task
		if model.UserId() != userId {
			p.log.WithFields(logrus.Fields{
				"taskId":       id,
				"taskUserId":   model.UserId(),
				"requesterId":  userId,
			}).Warn("Unauthorized task access attempt")
			return Model{}, ErrUnauthorized
		}

		return model, nil
	}
}

// List retrieves all tasks for a user
func (p Processor) List(userId uuid.UUID) ops.Provider[[]Model] {
	return func() ([]Model, error) {
		p.log.WithField("userId", userId).Debug("Listing tasks for user")
		return GetByUserId(p.db)(userId)()
	}
}

// ListByDay retrieves all tasks for a user on a specific day
func (p Processor) ListByDay(userId uuid.UUID, day time.Time) ops.Provider[[]Model] {
	return func() ([]Model, error) {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"day":    day.Format("2006-01-02"),
		}).Debug("Listing tasks for user by day")
		return GetByUserIdAndDay(p.db)(userId, day)()
	}
}

// ListByStatus retrieves all tasks for a user with a specific status
func (p Processor) ListByStatus(userId uuid.UUID, status Status) ops.Provider[[]Model] {
	return func() ([]Model, error) {
		p.log.WithFields(logrus.Fields{
			"userId": userId,
			"status": status,
		}).Debug("Listing tasks for user by status")
		return GetByUserIdAndStatus(p.db)(userId, status)()
	}
}

// Update updates an existing task
func (p Processor) Update(id uuid.UUID, userId uuid.UUID, input UpdateInput) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithField("taskId", id).Info("Updating task")

		// Fetch existing task and verify authorization
		existing, err := p.GetById(id, userId)()
		if err != nil {
			return Model{}, err
		}

		// Build updated model
		builder := existing.Builder().
			SetUpdatedAt(time.Now())

		if input.Title != nil {
			builder.SetTitle(*input.Title)
		}

		if input.Description != nil {
			builder.SetDescription(*input.Description)
		}

		if input.Day != nil {
			builder.SetDay(*input.Day)
		}

		if input.Status != nil {
			builder.SetStatus(*input.Status)
			// If marking as complete, set completion timestamp
			if *input.Status == StatusComplete && existing.Status() != StatusComplete {
				now := time.Now()
				builder.SetCompletedAt(now)
			}
			// If marking as incomplete, clear completion timestamp
			if *input.Status == StatusIncomplete && existing.Status() == StatusComplete {
				builder.ClearCompletedAt()
			}
		}

		model, err := builder.Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build updated task model")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Save(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to update task in database")
			return Model{}, err
		}

		p.log.WithField("taskId", id).Info("Task updated successfully")
		return Make(entity)
	}
}

// Complete marks a task as completed
func (p Processor) Complete(id uuid.UUID, userId uuid.UUID) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithField("taskId", id).Info("Completing task")

		// Fetch existing task and verify authorization
		existing, err := p.GetById(id, userId)()
		if err != nil {
			return Model{}, err
		}

		// Check if already complete
		if existing.IsComplete() {
			p.log.WithField("taskId", id).Warn("Task is already completed")
			return existing, nil // Return existing model, not an error
		}

		// Mark as complete
		model, err := existing.Builder().
			MarkComplete().
			Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build completed task model")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Save(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to complete task in database")
			return Model{}, err
		}

		p.log.WithField("taskId", id).Info("Task completed successfully")
		return Make(entity)
	}
}

// Uncomplete marks a task as incomplete (reopen)
func (p Processor) Uncomplete(id uuid.UUID, userId uuid.UUID) ops.Provider[Model] {
	return func() (Model, error) {
		p.log.WithField("taskId", id).Info("Marking task as incomplete")

		// Fetch existing task and verify authorization
		existing, err := p.GetById(id, userId)()
		if err != nil {
			return Model{}, err
		}

		// Check if already incomplete
		if existing.IsIncomplete() {
			p.log.WithField("taskId", id).Warn("Task is already incomplete")
			return existing, nil
		}

		// Mark as incomplete
		model, err := existing.Builder().
			MarkIncomplete().
			Build()
		if err != nil {
			p.log.WithError(err).Error("Failed to build incomplete task model")
			return Model{}, err
		}

		// Save to database
		entity := model.ToEntity()
		if err := p.db.Save(&entity).Error; err != nil {
			p.log.WithError(err).Error("Failed to mark task as incomplete in database")
			return Model{}, err
		}

		p.log.WithField("taskId", id).Info("Task marked as incomplete successfully")
		return Make(entity)
	}
}

// Delete removes a task by ID
func (p Processor) Delete(id uuid.UUID, userId uuid.UUID) error {
	p.log.WithField("taskId", id).Info("Deleting task")

	// First verify the task exists and user has access
	_, err := p.GetById(id, userId)()
	if err != nil {
		return err
	}

	// Delete the task
	result := p.db.Delete(&Entity{}, "id = ?", id)
	if result.Error != nil {
		p.log.WithError(result.Error).Error("Failed to delete task")
		return result.Error
	}

	if result.RowsAffected == 0 {
		p.log.WithField("taskId", id).Warn("Task not found for deletion")
		return ErrTaskNotFound
	}

	p.log.WithField("taskId", id).Info("Task deleted successfully")
	return nil
}
