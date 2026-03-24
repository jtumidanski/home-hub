package task

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
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
	db.AutoMigrate(&Entity{})
	return db
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	m, err := p.Create(uuid.New(), uuid.New(), "Test Task", "Some notes", nil, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Title() != "Test Task" {
		t.Errorf("expected title Test Task, got %s", m.Title())
	}
	if m.Status() != "pending" {
		t.Errorf("expected status pending, got %s", m.Status())
	}
}

func TestUpdate_Complete(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)
	userID := uuid.New()

	m, _ := p.Create(uuid.New(), uuid.New(), "Task", "", nil, false)
	updated, err := p.Update(m.Id(), "Task", "", "completed", nil, false, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated.IsCompleted() {
		t.Error("expected task to be completed")
	}
	if updated.CompletedAt() == nil {
		t.Error("expected completedAt to be set")
	}
}

func TestUpdate_Reopen(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)
	userID := uuid.New()

	m, _ := p.Create(uuid.New(), uuid.New(), "Task", "", nil, false)
	completed, _ := p.Update(m.Id(), "Task", "", "completed", nil, false, userID)
	reopened, err := p.Update(completed.Id(), "Task", "", "pending", nil, false, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reopened.IsCompleted() {
		t.Error("expected task to be pending")
	}
	if reopened.CompletedAt() != nil {
		t.Error("expected completedAt to be nil after reopen")
	}
}

func TestSoftDelete_And_Restore(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	m, _ := p.Create(uuid.New(), uuid.New(), "Delete Me", "", nil, false)
	if err := p.Delete(m.Id()); err != nil {
		t.Fatalf("unexpected error deleting: %v", err)
	}

	deleted, _ := p.ByIDProvider(m.Id())()
	if !deleted.IsDeleted() {
		t.Error("expected task to be deleted")
	}

	if err := p.Restore(m.Id()); err != nil {
		t.Fatalf("unexpected error restoring: %v", err)
	}

	restored, _ := p.ByIDProvider(m.Id())()
	if restored.IsDeleted() {
		t.Error("expected task to be restored")
	}
}

func TestRestore_NotDeleted(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	m, _ := p.Create(uuid.New(), uuid.New(), "Not Deleted", "", nil, false)
	err := p.Restore(m.Id())
	if err != ErrNotDeleted {
		t.Errorf("expected ErrNotDeleted, got %v", err)
	}
}

func TestRestore_WindowExpired(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	m, _ := p.Create(uuid.New(), uuid.New(), "Old Delete", "", nil, false)

	// Manually set deleted_at to 4 days ago
	oldTime := time.Now().UTC().Add(-4 * 24 * time.Hour)
	db.Model(&Entity{}).Where("id = ?", m.Id()).Update("deleted_at", oldTime)

	err := p.Restore(m.Id())
	if err != ErrRestoreWindow {
		t.Errorf("expected ErrRestoreWindow, got %v", err)
	}
}
