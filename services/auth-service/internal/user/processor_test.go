package user

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

func TestFindOrCreate(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, p *Processor)
		email       string
		displayName string
		wantNew     bool
	}{
		{
			name:        "creates new user",
			email:       "new@example.com",
			displayName: "New User",
			wantNew:     true,
		},
		{
			name: "finds existing user",
			setup: func(t *testing.T, p *Processor) {
				_, err := p.FindOrCreate("existing@example.com", "Original Name", "Original", "Name", "")
				if err != nil {
					t.Fatalf("setup: %v", err)
				}
			},
			email:       "existing@example.com",
			displayName: "Different Name",
			wantNew:     false,
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

			m, err := p.FindOrCreate(tt.email, tt.displayName, "Given", "Family", "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.Email() != tt.email {
				t.Errorf("expected email %s, got %s", tt.email, m.Email())
			}
			if m.Id() == uuid.Nil {
				t.Error("expected non-nil UUID")
			}
			if !tt.wantNew && m.DisplayName() != "Original Name" {
				t.Errorf("expected original display name, got %s", m.DisplayName())
			}
		})
	}
}

func TestByIDProvider(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, p *Processor) uuid.UUID
		wantErr bool
	}{
		{
			name: "found",
			setup: func(t *testing.T, p *Processor) uuid.UUID {
				m, err := p.FindOrCreate("found@example.com", "Found User", "Found", "User", "")
				if err != nil {
					t.Fatalf("setup: %v", err)
				}
				return m.Id()
			},
			wantErr: false,
		},
		{
			name: "not found",
			setup: func(t *testing.T, p *Processor) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			l, _ := test.NewNullLogger()
			p := NewProcessor(l, context.Background(), db)

			id := tt.setup(t, p)
			_, err := p.ByIDProvider(id)()
			if (err != nil) != tt.wantErr {
				t.Errorf("ByIDProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
