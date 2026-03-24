package user

import (
	"context"
	"testing"

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
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestFindOrCreate_CreatesNewUser(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	m, err := p.FindOrCreate("test@example.com", "Test User", "Test", "User", "https://example.com/avatar.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Email() != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", m.Email())
	}
	if m.DisplayName() != "Test User" {
		t.Errorf("expected display name Test User, got %s", m.DisplayName())
	}
	if m.Id() == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestFindOrCreate_FindsExistingUser(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	first, err := p.FindOrCreate("test@example.com", "Test User", "Test", "User", "")
	if err != nil {
		t.Fatalf("unexpected error creating: %v", err)
	}

	second, err := p.FindOrCreate("test@example.com", "Different Name", "Different", "Name", "")
	if err != nil {
		t.Fatalf("unexpected error finding: %v", err)
	}

	if first.Id() != second.Id() {
		t.Errorf("expected same user ID, got %s and %s", first.Id(), second.Id())
	}
	// Should return original name, not the new one
	if second.DisplayName() != "Test User" {
		t.Errorf("expected original display name, got %s", second.DisplayName())
	}
}

func TestByIDProvider_NotFound(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	_, err := p.ByIDProvider(uuid.New())()
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestByIDProvider_Found(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	created, err := p.FindOrCreate("found@example.com", "Found User", "Found", "User", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, err := p.ByIDProvider(created.Id())()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Email() != "found@example.com" {
		t.Errorf("expected email found@example.com, got %s", found.Email())
	}
}
