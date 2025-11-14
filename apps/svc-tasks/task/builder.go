package task

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTitleRequired       = errors.New("title is required")
	ErrTitleEmpty          = errors.New("title cannot be empty")
	ErrTitleTooLong        = errors.New("title exceeds 255 characters")
	ErrUserIdRequired      = errors.New("userId is required")
	ErrHouseholdIdRequired = errors.New("householdId is required")
	ErrDayRequired         = errors.New("day is required")
	ErrStatusInvalid       = errors.New("status must be 'incomplete' or 'complete'")
	ErrCompletedAtInvalid  = errors.New("completedAt can only be set when status is complete")
)

// Builder provides a fluent API for constructing valid Task models
type Builder struct {
	id          *uuid.UUID
	userId      *uuid.UUID
	householdId *uuid.UUID
	day         *time.Time
	title       *string
	description *string
	status      *Status
	createdAt   *time.Time
	completedAt *time.Time
	updatedAt   *time.Time
}

// NewBuilder creates a new task builder
func NewBuilder() *Builder {
	return &Builder{}
}

// SetId sets the task ID
func (b *Builder) SetId(id uuid.UUID) *Builder {
	b.id = &id
	return b
}

// SetUserId sets the user ID who owns this task
func (b *Builder) SetUserId(userId uuid.UUID) *Builder {
	b.userId = &userId
	return b
}

// SetHouseholdId sets the household ID this task belongs to
func (b *Builder) SetHouseholdId(householdId uuid.UUID) *Builder {
	b.householdId = &householdId
	return b
}

// SetDay sets the date this task is scheduled for
func (b *Builder) SetDay(day time.Time) *Builder {
	// Normalize to date only (strip time component)
	normalized := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	b.day = &normalized
	return b
}

// SetTitle sets the task title
func (b *Builder) SetTitle(title string) *Builder {
	b.title = &title
	return b
}

// SetDescription sets the task description
func (b *Builder) SetDescription(description string) *Builder {
	b.description = &description
	return b
}

// SetStatus sets the task status
func (b *Builder) SetStatus(status Status) *Builder {
	b.status = &status
	return b
}

// SetCreatedAt sets the creation timestamp
func (b *Builder) SetCreatedAt(createdAt time.Time) *Builder {
	b.createdAt = &createdAt
	return b
}

// SetCompletedAt sets the completion timestamp
func (b *Builder) SetCompletedAt(completedAt time.Time) *Builder {
	b.completedAt = &completedAt
	return b
}

// ClearCompletedAt removes the completion timestamp
func (b *Builder) ClearCompletedAt() *Builder {
	b.completedAt = nil
	return b
}

// SetUpdatedAt sets the update timestamp
func (b *Builder) SetUpdatedAt(updatedAt time.Time) *Builder {
	b.updatedAt = &updatedAt
	return b
}

// MarkComplete sets the status to complete and sets the completion timestamp
func (b *Builder) MarkComplete() *Builder {
	status := StatusComplete
	b.status = &status
	now := time.Now()
	b.completedAt = &now
	b.updatedAt = &now
	return b
}

// MarkIncomplete sets the status to incomplete and clears the completion timestamp
func (b *Builder) MarkIncomplete() *Builder {
	status := StatusIncomplete
	b.status = &status
	b.completedAt = nil
	now := time.Now()
	b.updatedAt = &now
	return b
}

// Build validates the builder state and constructs a Task Model
func (b *Builder) Build() (Model, error) {
	// Validate title
	if b.title == nil {
		return Model{}, ErrTitleRequired
	}
	title := strings.TrimSpace(*b.title)
	if title == "" {
		return Model{}, ErrTitleEmpty
	}
	if len(title) > 255 {
		return Model{}, ErrTitleTooLong
	}

	// Validate userId
	if b.userId == nil {
		return Model{}, ErrUserIdRequired
	}
	if *b.userId == uuid.Nil {
		return Model{}, ErrUserIdRequired
	}

	// Validate householdId
	if b.householdId == nil {
		return Model{}, ErrHouseholdIdRequired
	}
	if *b.householdId == uuid.Nil {
		return Model{}, ErrHouseholdIdRequired
	}

	// Validate day
	if b.day == nil {
		return Model{}, ErrDayRequired
	}

	// Validate status and set default
	status := StatusIncomplete
	if b.status != nil {
		if !b.status.IsValid() {
			return Model{}, ErrStatusInvalid
		}
		status = *b.status
	}

	// Validate completedAt consistency with status
	if b.completedAt != nil && status != StatusComplete {
		return Model{}, ErrCompletedAtInvalid
	}

	// Generate ID if not provided
	id := uuid.New()
	if b.id != nil {
		id = *b.id
	}

	// Generate timestamps if not provided
	now := time.Now()
	createdAt := now
	if b.createdAt != nil {
		createdAt = *b.createdAt
	}
	updatedAt := now
	if b.updatedAt != nil {
		updatedAt = *b.updatedAt
	}

	// Handle description
	description := ""
	if b.description != nil {
		description = strings.TrimSpace(*b.description)
	}

	return Model{
		id:          id,
		userId:      *b.userId,
		householdId: *b.householdId,
		day:         *b.day,
		title:       title,
		description: description,
		status:      status,
		createdAt:   createdAt,
		completedAt: b.completedAt,
		updatedAt:   updatedAt,
	}, nil
}

// Builder creates a builder initialized with the model's current values
// This enables modification flows: model.Builder().SetTitle(newTitle).Build()
func (m Model) Builder() *Builder {
	b := &Builder{
		id:          &m.id,
		userId:      &m.userId,
		householdId: &m.householdId,
		day:         &m.day,
		title:       &m.title,
		description: &m.description,
		status:      &m.status,
		createdAt:   &m.createdAt,
		updatedAt:   &m.updatedAt,
	}

	if m.completedAt != nil {
		completedAtCopy := *m.completedAt
		b.completedAt = &completedAtCopy
	}

	return b
}
