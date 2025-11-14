package reminder

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired        = errors.New("name is required")
	ErrNameEmpty           = errors.New("name cannot be empty")
	ErrNameTooLong         = errors.New("name exceeds 255 characters")
	ErrUserIdRequired      = errors.New("userId is required")
	ErrHouseholdIdRequired = errors.New("householdId is required")
	ErrRemindAtRequired    = errors.New("remindAt is required")
	ErrStatusInvalid       = errors.New("status must be 'active', 'snoozed', or 'dismissed'")
	ErrDismissedAtInvalid  = errors.New("dismissedAt can only be set when status is dismissed")
)

// Builder provides a fluent API for constructing valid Reminder models
type Builder struct {
	id          *uuid.UUID
	name        *string
	description *string
	userId      *uuid.UUID
	householdId *uuid.UUID
	createdAt   *time.Time
	remindAt    *time.Time
	snoozeCount *int
	status      *Status
	dismissedAt *time.Time
	updatedAt   *time.Time
}

// NewBuilder creates a new reminder builder
func NewBuilder() *Builder {
	return &Builder{}
}

// SetId sets the reminder ID
func (b *Builder) SetId(id uuid.UUID) *Builder {
	b.id = &id
	return b
}

// SetName sets the reminder name/title
func (b *Builder) SetName(name string) *Builder {
	b.name = &name
	return b
}

// SetDescription sets the reminder description
func (b *Builder) SetDescription(description string) *Builder {
	b.description = &description
	return b
}

// SetUserId sets the user ID who owns this reminder
func (b *Builder) SetUserId(userId uuid.UUID) *Builder {
	b.userId = &userId
	return b
}

// SetHouseholdId sets the household ID this reminder belongs to
func (b *Builder) SetHouseholdId(householdId uuid.UUID) *Builder {
	b.householdId = &householdId
	return b
}

// SetCreatedAt sets the creation timestamp
func (b *Builder) SetCreatedAt(createdAt time.Time) *Builder {
	b.createdAt = &createdAt
	return b
}

// SetRemindAt sets the remind-at timestamp
func (b *Builder) SetRemindAt(remindAt time.Time) *Builder {
	b.remindAt = &remindAt
	return b
}

// SetSnoozeCount sets the snooze counter
func (b *Builder) SetSnoozeCount(count int) *Builder {
	b.snoozeCount = &count
	return b
}

// SetStatus sets the reminder status
func (b *Builder) SetStatus(status Status) *Builder {
	b.status = &status
	return b
}

// SetDismissedAt sets the dismissal timestamp
func (b *Builder) SetDismissedAt(dismissedAt time.Time) *Builder {
	b.dismissedAt = &dismissedAt
	return b
}

// ClearDismissedAt removes the dismissal timestamp
func (b *Builder) ClearDismissedAt() *Builder {
	b.dismissedAt = nil
	return b
}

// SetUpdatedAt sets the update timestamp
func (b *Builder) SetUpdatedAt(updatedAt time.Time) *Builder {
	b.updatedAt = &updatedAt
	return b
}

// MarkSnoozed sets the status to snoozed, updates remind-at, and increments snooze count
func (b *Builder) MarkSnoozed(newRemindAt time.Time) *Builder {
	status := StatusSnoozed
	b.status = &status
	b.remindAt = &newRemindAt

	// Increment snooze count
	currentCount := 0
	if b.snoozeCount != nil {
		currentCount = *b.snoozeCount
	}
	newCount := currentCount + 1
	b.snoozeCount = &newCount

	now := time.Now()
	b.updatedAt = &now
	return b
}

// MarkDismissed sets the status to dismissed and sets the dismissal timestamp
func (b *Builder) MarkDismissed() *Builder {
	status := StatusDismissed
	b.status = &status
	now := time.Now()
	b.dismissedAt = &now
	b.updatedAt = &now
	return b
}

// MarkActive sets the status to active and clears the dismissal timestamp
func (b *Builder) MarkActive() *Builder {
	status := StatusActive
	b.status = &status
	b.dismissedAt = nil
	now := time.Now()
	b.updatedAt = &now
	return b
}

// Build validates the builder state and constructs a Reminder Model
func (b *Builder) Build() (Model, error) {
	// Validate name
	if b.name == nil {
		return Model{}, ErrNameRequired
	}
	name := strings.TrimSpace(*b.name)
	if name == "" {
		return Model{}, ErrNameEmpty
	}
	if len(name) > 255 {
		return Model{}, ErrNameTooLong
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

	// Validate remindAt
	if b.remindAt == nil {
		return Model{}, ErrRemindAtRequired
	}
	if b.remindAt.IsZero() {
		return Model{}, ErrRemindAtRequired
	}

	// Validate status and set default
	status := StatusActive
	if b.status != nil {
		if !b.status.IsValid() {
			return Model{}, ErrStatusInvalid
		}
		status = *b.status
	}

	// Validate dismissedAt consistency with status
	if b.dismissedAt != nil && status != StatusDismissed {
		return Model{}, ErrDismissedAtInvalid
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

	// Handle snooze count
	snoozeCount := 0
	if b.snoozeCount != nil {
		snoozeCount = *b.snoozeCount
	}

	return Model{
		id:          id,
		name:        name,
		description: description,
		userId:      *b.userId,
		householdId: *b.householdId,
		createdAt:   createdAt,
		remindAt:    *b.remindAt,
		snoozeCount: snoozeCount,
		status:      status,
		dismissedAt: b.dismissedAt,
		updatedAt:   updatedAt,
	}, nil
}

// Builder creates a builder initialized with the model's current values
// This enables modification flows: model.Builder().SetName(newName).Build()
func (m Model) Builder() *Builder {
	b := &Builder{
		id:          &m.id,
		name:        &m.name,
		description: &m.description,
		userId:      &m.userId,
		householdId: &m.householdId,
		createdAt:   &m.createdAt,
		remindAt:    &m.remindAt,
		snoozeCount: &m.snoozeCount,
		status:      &m.status,
		updatedAt:   &m.updatedAt,
	}

	if m.dismissedAt != nil {
		dismissedAtCopy := *m.dismissedAt
		b.dismissedAt = &dismissedAtCopy
	}

	return b
}
