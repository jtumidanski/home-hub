package preference

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
	t.Run("FindOrCreate", func(t *testing.T) {
		tests := []struct {
			name      string
			setup     func(p *Processor, tenantID, userID uuid.UUID)
			wantTheme string
			wantNew   bool
		}{
			{
				name:      "creates new with default theme",
				setup:     func(_ *Processor, _, _ uuid.UUID) {},
				wantTheme: "light",
				wantNew:   true,
			},
			{
				name: "finds existing",
				setup: func(p *Processor, tenantID, userID uuid.UUID) {
					p.FindOrCreate(tenantID, userID)
				},
				wantTheme: "light",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := setupTestDB(t)
				l, _ := test.NewNullLogger()
				p := NewProcessor(l, context.Background(), db)

				tenantID := uuid.New()
				userID := uuid.New()
				tt.setup(p, tenantID, userID)

				m, err := p.FindOrCreate(tenantID, userID)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if m.Theme() != tt.wantTheme {
					t.Errorf("expected theme %s, got %s", tt.wantTheme, m.Theme())
				}
				if tt.wantNew && m.ActiveHouseholdID() != nil {
					t.Error("expected nil active household for new preference")
				}
			})
		}
	})

	t.Run("FindOrCreate_idempotent", func(t *testing.T) {
		db := setupTestDB(t)
		l, _ := test.NewNullLogger()
		p := NewProcessor(l, context.Background(), db)

		tenantID := uuid.New()
		userID := uuid.New()

		first, _ := p.FindOrCreate(tenantID, userID)
		second, err := p.FindOrCreate(tenantID, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if first.Id() != second.Id() {
			t.Errorf("expected same preference ID, got %s and %s", first.Id(), second.Id())
		}
	})

	t.Run("UpdateTheme", func(t *testing.T) {
		tests := []struct {
			name      string
			newTheme  string
			wantTheme string
		}{
			{"switch to dark", "dark", "dark"},
			{"switch to light", "light", "light"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := setupTestDB(t)
				l, _ := test.NewNullLogger()
				p := NewProcessor(l, context.Background(), db)

				m, _ := p.FindOrCreate(uuid.New(), uuid.New())
				updated, err := p.UpdateTheme(m.Id(), tt.newTheme)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if updated.Theme() != tt.wantTheme {
					t.Errorf("expected theme %s, got %s", tt.wantTheme, updated.Theme())
				}
			})
		}
	})

	t.Run("SetActiveHousehold", func(t *testing.T) {
		db := setupTestDB(t)
		l, _ := test.NewNullLogger()
		p := NewProcessor(l, context.Background(), db)

		m, _ := p.FindOrCreate(uuid.New(), uuid.New())
		hhID := uuid.New()
		updated, err := p.SetActiveHousehold(m.Id(), hhID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.ActiveHouseholdID() == nil {
			t.Fatal("expected non-nil active household")
		}
		if *updated.ActiveHouseholdID() != hhID {
			t.Errorf("expected household %s, got %s", hhID, *updated.ActiveHouseholdID())
		}
	})
}
