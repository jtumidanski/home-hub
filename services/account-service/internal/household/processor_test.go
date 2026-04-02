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
		tests := []struct {
			name      string
			count     int
			wantCount int
		}{
			{"two households", 2, 2},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := setupTestDB(t)
				l, _ := test.NewNullLogger()
				p := NewProcessor(l, context.Background(), db)

				tenantID := uuid.New()
				for i := 0; i < tt.count; i++ {
					p.Create(tenantID, "Home", "UTC", "metric")
				}

				models, err := p.AllProvider()()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(models) != tt.wantCount {
					t.Errorf("expected %d households, got %d", tt.wantCount, len(models))
				}
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		tests := []struct {
			name         string
			initName     string
			initTimezone string
			initUnits    string
			newName      string
			newTimezone  string
			newUnits     string
			lat          *float64
			lon          *float64
			locName      *string
			wantName     string
			wantTimezone string
			wantLocation bool
		}{
			{
				name:         "update name and timezone",
				initName:     "Old Name",
				initTimezone: "UTC",
				initUnits:    "metric",
				newName:      "New Name",
				newTimezone:  "America/Chicago",
				newUnits:     "imperial",
				wantName:     "New Name",
				wantTimezone: "America/Chicago",
				wantLocation: false,
			},
			{
				name:         "update with location",
				initName:     "Home",
				initTimezone: "UTC",
				initUnits:    "metric",
				newName:      "Home",
				newTimezone:  "UTC",
				newUnits:     "metric",
				lat:          ptrFloat(40.7128),
				lon:          ptrFloat(-74.006),
				locName:      ptrString("New York, NY, United States"),
				wantName:     "Home",
				wantTimezone: "UTC",
				wantLocation: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := setupTestDB(t)
				l, _ := test.NewNullLogger()
				p := NewProcessor(l, context.Background(), db)

				m, _ := p.Create(uuid.New(), tt.initName, tt.initTimezone, tt.initUnits)
				updated, err := p.Update(m.Id(), tt.newName, tt.newTimezone, tt.newUnits, tt.lat, tt.lon, tt.locName)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if updated.Name() != tt.wantName {
					t.Errorf("expected name %s, got %s", tt.wantName, updated.Name())
				}
				if updated.Timezone() != tt.wantTimezone {
					t.Errorf("expected timezone %s, got %s", tt.wantTimezone, updated.Timezone())
				}
				if updated.HasLocation() != tt.wantLocation {
					t.Errorf("expected hasLocation %v, got %v", tt.wantLocation, updated.HasLocation())
				}
				if tt.lat != nil && updated.Latitude() != nil && *updated.Latitude() != *tt.lat {
					t.Errorf("expected latitude %f, got %f", *tt.lat, *updated.Latitude())
				}
				if tt.lon != nil && updated.Longitude() != nil && *updated.Longitude() != *tt.lon {
					t.Errorf("expected longitude %f, got %f", *tt.lon, *updated.Longitude())
				}
				if tt.locName != nil && updated.LocationName() != nil && *updated.LocationName() != *tt.locName {
					t.Errorf("expected location name %s, got %s", *tt.locName, *updated.LocationName())
				}
			})
		}
	})

	t.Run("Update clear location", func(t *testing.T) {
		tests := []struct {
			name string
		}{
			{"clear previously set location"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
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
		}
	})

	t.Run("CreateWithOwner", func(t *testing.T) {
		tests := []struct {
			name     string
			hhName   string
			wantRole string
		}{
			{"creates household with owner membership", "My House", "owner"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := setupTestDB(t)
				l, _ := test.NewNullLogger()
				p := NewProcessor(l, context.Background(), db)

				tenantID := uuid.New()
				userID := uuid.New()
				m, err := p.CreateWithOwner(tenantID, userID, tt.hhName, "UTC", "metric")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if m.Name() != tt.hhName {
					t.Errorf("expected name %s, got %s", tt.hhName, m.Name())
				}

				memProc := membership.NewProcessor(l, context.Background(), db)
				mem, err := memProc.ByHouseholdAndUserProvider(m.Id(), userID)()
				if err != nil {
					t.Fatalf("expected owner membership, got error: %v", err)
				}
				if mem.Role() != tt.wantRole {
					t.Errorf("expected role %s, got %s", tt.wantRole, mem.Role())
				}
			})
		}
	})
}

func ptrFloat(f float64) *float64 { return &f }
func ptrString(s string) *string  { return &s }
