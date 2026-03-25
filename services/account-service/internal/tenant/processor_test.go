package tenant

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
	db.AutoMigrate(&Entity{})
	return db
}

func TestProcessor(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			wantName string
		}{
			{"valid tenant", "Test Tenant", "Test Tenant"},
			{"unicode name", "Haus Müller", "Haus Müller"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := setupTestDB(t)
				l, _ := test.NewNullLogger()
				p := NewProcessor(l, context.Background(), db)

				m, err := p.Create(tt.input)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if m.Name() != tt.wantName {
					t.Errorf("expected name %s, got %s", tt.wantName, m.Name())
				}
				if m.Id() == uuid.Nil {
					t.Error("expected non-nil UUID")
				}
			})
		}
	})

	t.Run("ByIDProvider", func(t *testing.T) {
		tests := []struct {
			name    string
			setup   func(p *Processor) uuid.UUID
			wantErr bool
		}{
			{
				name: "existing tenant",
				setup: func(p *Processor) uuid.UUID {
					m, _ := p.Create("Lookup Tenant")
					return m.Id()
				},
			},
			{
				name: "non-existent tenant",
				setup: func(_ *Processor) uuid.UUID {
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

				id := tt.setup(p)
				_, err := p.ByIDProvider(id)()
				if tt.wantErr && err == nil {
					t.Error("expected error")
				}
				if !tt.wantErr && err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})
		}
	})
}
