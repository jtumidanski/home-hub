package externalidentity

import (
	"context"
	"testing"

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
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	userID := uuid.New()
	m, err := p.Create(userID, "google", "sub-12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Id() == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if m.UserId() != userID {
		t.Errorf("expected user ID %s, got %s", userID, m.UserId())
	}
	if m.Provider() != "google" {
		t.Errorf("expected provider google, got %s", m.Provider())
	}
	if m.ProviderSubject() != "sub-12345" {
		t.Errorf("expected subject sub-12345, got %s", m.ProviderSubject())
	}
}

func TestFindByProviderSubject(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, p *Processor)
		subject string
		wantErr bool
	}{
		{
			name: "found",
			setup: func(t *testing.T, p *Processor) {
				if _, err := p.Create(uuid.New(), "google", "sub-found"); err != nil {
					t.Fatalf("setup: %v", err)
				}
			},
			subject: "sub-found",
			wantErr: false,
		},
		{
			name:    "not found",
			subject: "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			p := NewProcessor(l, context.Background(), db)

			if tt.setup != nil {
				tt.setup(t, p)
			}

			_, err := p.FindByProviderSubject("google", tt.subject)()
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByProviderSubject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreate_DuplicateSubject(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, context.Background(), db)

	if _, err := p.Create(uuid.New(), "google", "dup-subject"); err != nil {
		t.Fatalf("unexpected error on first create: %v", err)
	}

	if _, err := p.Create(uuid.New(), "google", "dup-subject"); err == nil {
		t.Error("expected error for duplicate provider+subject")
	}
}
