package task

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuilder_Build_Success(t *testing.T) {
	tests := []struct {
		name        string
		buildFunc   func() (*Builder, error)
		expectError bool
	}{
		{
			name: "valid task with all fields",
			buildFunc: func() (*Builder, error) {
				userId := uuid.New()
				householdId := uuid.New()
				day := time.Now()

				_, err := NewBuilder().
					SetUserId(userId).
					SetHouseholdId(householdId).
					SetDay(day).
					SetTitle("Buy groceries").
					SetDescription("Milk, eggs, bread").
					Build()

				return nil, err
			},
			expectError: false,
		},
		{
			name: "valid task with minimum fields",
			buildFunc: func() (*Builder, error) {
				userId := uuid.New()
				householdId := uuid.New()
				day := time.Now()

				_, err := NewBuilder().
					SetUserId(userId).
					SetHouseholdId(householdId).
					SetDay(day).
					SetTitle("Simple task").
					Build()

				return nil, err
			},
			expectError: false,
		},
		{
			name: "valid task with completed status",
			buildFunc: func() (*Builder, error) {
				userId := uuid.New()
				householdId := uuid.New()
				day := time.Now()

				_, err := NewBuilder().
					SetUserId(userId).
					SetHouseholdId(householdId).
					SetDay(day).
					SetTitle("Completed task").
					MarkComplete().
					Build()

				return nil, err
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.buildFunc()
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBuilder_Build_Validation_Errors(t *testing.T) {
	userId := uuid.New()
	householdId := uuid.New()
	day := time.Now()

	tests := []struct {
		name        string
		buildFunc   func() (*Builder, error)
		expectedErr error
	}{
		{
			name: "missing title",
			buildFunc: func() (*Builder, error) {
				_, err := NewBuilder().
					SetUserId(userId).
					SetHouseholdId(householdId).
					SetDay(day).
					Build()
				return nil, err
			},
			expectedErr: ErrTitleRequired,
		},
		{
			name: "empty title",
			buildFunc: func() (*Builder, error) {
				_, err := NewBuilder().
					SetUserId(userId).
					SetHouseholdId(householdId).
					SetDay(day).
					SetTitle("   ").
					Build()
				return nil, err
			},
			expectedErr: ErrTitleEmpty,
		},
		{
			name: "title too long",
			buildFunc: func() (*Builder, error) {
				longTitle := string(make([]byte, 256))
				_, err := NewBuilder().
					SetUserId(userId).
					SetHouseholdId(householdId).
					SetDay(day).
					SetTitle(longTitle).
					Build()
				return nil, err
			},
			expectedErr: ErrTitleTooLong,
		},
		{
			name: "missing userId",
			buildFunc: func() (*Builder, error) {
				_, err := NewBuilder().
					SetHouseholdId(householdId).
					SetDay(day).
					SetTitle("Task").
					Build()
				return nil, err
			},
			expectedErr: ErrUserIdRequired,
		},
		{
			name: "nil userId",
			buildFunc: func() (*Builder, error) {
				_, err := NewBuilder().
					SetUserId(uuid.Nil).
					SetHouseholdId(householdId).
					SetDay(day).
					SetTitle("Task").
					Build()
				return nil, err
			},
			expectedErr: ErrUserIdRequired,
		},
		{
			name: "missing householdId",
			buildFunc: func() (*Builder, error) {
				_, err := NewBuilder().
					SetUserId(userId).
					SetDay(day).
					SetTitle("Task").
					Build()
				return nil, err
			},
			expectedErr: ErrHouseholdIdRequired,
		},
		{
			name: "missing day",
			buildFunc: func() (*Builder, error) {
				_, err := NewBuilder().
					SetUserId(userId).
					SetHouseholdId(householdId).
					SetTitle("Task").
					Build()
				return nil, err
			},
			expectedErr: ErrDayRequired,
		},
		{
			name: "invalid status",
			buildFunc: func() (*Builder, error) {
				_, err := NewBuilder().
					SetUserId(userId).
					SetHouseholdId(householdId).
					SetDay(day).
					SetTitle("Task").
					SetStatus(Status("invalid")).
					Build()
				return nil, err
			},
			expectedErr: ErrStatusInvalid,
		},
		{
			name: "completedAt without complete status",
			buildFunc: func() (*Builder, error) {
				_, err := NewBuilder().
					SetUserId(userId).
					SetHouseholdId(householdId).
					SetDay(day).
					SetTitle("Task").
					SetStatus(StatusIncomplete).
					SetCompletedAt(time.Now()).
					Build()
				return nil, err
			},
			expectedErr: ErrCompletedAtInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.buildFunc()
			if err == nil {
				t.Errorf("expected error but got none")
				return
			}
			if err != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestBuilder_MarkComplete(t *testing.T) {
	userId := uuid.New()
	householdId := uuid.New()
	day := time.Now()

	model, err := NewBuilder().
		SetUserId(userId).
		SetHouseholdId(householdId).
		SetDay(day).
		SetTitle("Task to complete").
		MarkComplete().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if model.Status() != StatusComplete {
		t.Errorf("expected status %v, got %v", StatusComplete, model.Status())
	}

	if model.CompletedAt() == nil {
		t.Error("expected completedAt to be set, got nil")
	}
}

func TestBuilder_MarkIncomplete(t *testing.T) {
	userId := uuid.New()
	householdId := uuid.New()
	day := time.Now()

	model, err := NewBuilder().
		SetUserId(userId).
		SetHouseholdId(householdId).
		SetDay(day).
		SetTitle("Task to mark incomplete").
		MarkComplete().
		Build()

	if err != nil {
		t.Fatalf("unexpected error building completed task: %v", err)
	}

	// Now reopen it
	reopenedModel, err := model.Builder().
		MarkIncomplete().
		Build()

	if err != nil {
		t.Fatalf("unexpected error marking incomplete: %v", err)
	}

	if reopenedModel.Status() != StatusIncomplete {
		t.Errorf("expected status %v, got %v", StatusIncomplete, reopenedModel.Status())
	}

	if reopenedModel.CompletedAt() != nil {
		t.Error("expected completedAt to be nil after marking incomplete")
	}
}

func TestModel_Immutability(t *testing.T) {
	userId := uuid.New()
	householdId := uuid.New()
	day := time.Now()

	model, err := NewBuilder().
		SetUserId(userId).
		SetHouseholdId(householdId).
		SetDay(day).
		SetTitle("Original title").
		SetDescription("Original description").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Modify via builder should create new model
	modifiedModel, err := model.Builder().
		SetTitle("Modified title").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Original should be unchanged
	if model.Title() != "Original title" {
		t.Errorf("original model was mutated: expected 'Original title', got '%s'", model.Title())
	}

	// New model should have the change
	if modifiedModel.Title() != "Modified title" {
		t.Errorf("modified model does not have new title: expected 'Modified title', got '%s'", modifiedModel.Title())
	}
}

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{"incomplete is valid", StatusIncomplete, true},
		{"complete is valid", StatusComplete, true},
		{"invalid status", Status("invalid"), false},
		{"empty status", Status(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
