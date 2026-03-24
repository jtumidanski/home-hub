package reminder

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	db.AutoMigrate(&Entity{})
	return db
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	scheduled := time.Now().UTC().Add(1 * time.Hour)
	m, err := p.Create(uuid.New(), uuid.New(), "Test Reminder", "Notes", scheduled)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Title() != "Test Reminder" {
		t.Errorf("expected title Test Reminder, got %s", m.Title())
	}
}

func TestIsActive_Future(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	future := time.Now().UTC().Add(1 * time.Hour)
	m, _ := p.Create(uuid.New(), uuid.New(), "Future", "", future)
	if m.IsActive() {
		t.Error("expected future reminder to not be active")
	}
}

func TestIsActive_Past(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	past := time.Now().UTC().Add(-1 * time.Hour)
	m, _ := p.Create(uuid.New(), uuid.New(), "Past", "", past)
	if !m.IsActive() {
		t.Error("expected past reminder to be active")
	}
}

func TestSnooze_ValidDuration(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	past := time.Now().UTC().Add(-1 * time.Hour)
	m, _ := p.Create(uuid.New(), uuid.New(), "Snooze Me", "", past)

	snoozedUntil, err := p.Snooze(m.Id(), 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snoozedUntil.Before(time.Now().UTC()) {
		t.Error("expected snoozedUntil to be in the future")
	}
}

func TestSnooze_InvalidDuration(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	past := time.Now().UTC().Add(-1 * time.Hour)
	m, _ := p.Create(uuid.New(), uuid.New(), "Bad Snooze", "", past)

	_, err := p.Snooze(m.Id(), 15)
	if err == nil {
		t.Error("expected error for invalid snooze duration")
	}
}

func TestDismiss(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	past := time.Now().UTC().Add(-1 * time.Hour)
	m, _ := p.Create(uuid.New(), uuid.New(), "Dismiss Me", "", past)

	if err := p.Dismiss(m.Id()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dismissed, _ := p.ByIDProvider(m.Id())()
	if dismissed.IsActive() {
		t.Error("expected dismissed reminder to not be active")
	}
}
