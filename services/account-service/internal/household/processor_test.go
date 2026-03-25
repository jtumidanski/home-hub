package household

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/account-service/internal/membership"
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
	db.AutoMigrate(&membership.Entity{})
	return db
}

func TestProcessor(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name         string
			hhName       string
			timezone     string
			units        string
			wantName     string
			wantTimezone string
			wantUnits    string
		}{
			{"imperial household", "Main Home", "America/Detroit", "imperial", "Main Home", "America/Detroit", "imperial"},
			{"metric household", "Beach House", "UTC", "metric", "Beach House", "UTC", "metric"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := setupTestDB(t)
				l, _ := test.NewNullLogger()
				p := NewProcessor(l, context.Background(), db)

				m, err := p.Create(uuid.New(), tt.hhName, tt.timezone, tt.units)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if m.Name() != tt.wantName {
					t.Errorf("expected name %s, got %s", tt.wantName, m.Name())
				}
				if m.Timezone() != tt.wantTimezone {
					t.Errorf("expected timezone %s, got %s", tt.wantTimezone, m.Timezone())
				}
				if m.Units() != tt.wantUnits {
					t.Errorf("expected units %s, got %s", tt.wantUnits, m.Units())
				}
			})
		}
	})

	t.Run("AllProvider", func(t *testing.T) {
		db := setupTestDB(t)
		l, _ := test.NewNullLogger()
		p := NewProcessor(l, context.Background(), db)

		tenantID := uuid.New()
		p.Create(tenantID, "Home 1", "UTC", "metric")
		p.Create(tenantID, "Home 2", "UTC", "metric")

		models, err := p.AllProvider()()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(models) != 2 {
			t.Errorf("expected 2 households, got %d", len(models))
		}
	})

	t.Run("Update", func(t *testing.T) {
		db := setupTestDB(t)
		l, _ := test.NewNullLogger()
		p := NewProcessor(l, context.Background(), db)

		m, _ := p.Create(uuid.New(), "Old Name", "UTC", "metric")
		updated, err := p.Update(m.Id(), "New Name", "America/Chicago", "imperial", nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name() != "New Name" {
			t.Errorf("expected name New Name, got %s", updated.Name())
		}
		if updated.Timezone() != "America/Chicago" {
			t.Errorf("expected timezone America/Chicago, got %s", updated.Timezone())
		}
	})

	t.Run("Update with location", func(t *testing.T) {
		db := setupTestDB(t)
		l, _ := test.NewNullLogger()
		p := NewProcessor(l, context.Background(), db)

		m, _ := p.Create(uuid.New(), "Home", "UTC", "metric")
		lat := 40.7128
		lon := -74.006
		locName := "New York, NY, United States"
		updated, err := p.Update(m.Id(), "Home", "UTC", "metric", &lat, &lon, &locName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !updated.HasLocation() {
			t.Error("expected household to have location")
		}
		if *updated.Latitude() != lat {
			t.Errorf("expected latitude %f, got %f", lat, *updated.Latitude())
		}
		if *updated.Longitude() != lon {
			t.Errorf("expected longitude %f, got %f", lon, *updated.Longitude())
		}
		if *updated.LocationName() != locName {
			t.Errorf("expected location name %s, got %s", locName, *updated.LocationName())
		}
	})

	t.Run("Update clear location", func(t *testing.T) {
		db := setupTestDB(t)
		l, _ := test.NewNullLogger()
		p := NewProcessor(l, context.Background(), db)

		m, _ := p.Create(uuid.New(), "Home", "UTC", "metric")
		lat := 40.7128
		lon := -74.006
		locName := "New York"
		m, _ = p.Update(m.Id(), "Home", "UTC", "metric", &lat, &lon, &locName)

		updated, err := p.Update(m.Id(), "Home", "UTC", "metric", nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.HasLocation() {
			t.Error("expected household to not have location after clearing")
		}
	})

	t.Run("CreateWithOwner", func(t *testing.T) {
		db := setupTestDB(t)
		l, _ := test.NewNullLogger()
		p := NewProcessor(l, context.Background(), db)

		tenantID := uuid.New()
		userID := uuid.New()
		m, err := p.CreateWithOwner(tenantID, userID, "My House", "UTC", "metric")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m.Name() != "My House" {
			t.Errorf("expected name My House, got %s", m.Name())
		}

		// Verify owner membership was created
		memProc := membership.NewProcessor(l, context.Background(), db)
		mem, err := memProc.ByHouseholdAndUserProvider(m.Id(), userID)()
		if err != nil {
			t.Fatalf("expected owner membership, got error: %v", err)
		}
		if mem.Role() != "owner" {
			t.Errorf("expected role owner, got %s", mem.Role())
		}
	})
}
