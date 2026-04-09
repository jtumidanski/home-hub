package refresh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/forecast"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRefreshTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	if err := db.AutoMigrate(&forecast.Entity{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func newCountingForecastServer(t *testing.T, count *int64) *httptest.Server {
	t.Helper()
	resp := openmeteo.ForecastResponse{
		Current: openmeteo.CurrentData{Temperature: 22.0, WeatherCode: 1},
		Daily: openmeteo.DailyData{
			Time:           []string{"2026-04-09"},
			TemperatureMax: []float64{25.0},
			TemperatureMin: []float64{15.0},
			WeatherCode:    []int{1},
		},
		Hourly: openmeteo.HourlyData{
			Time:                     []string{"2026-04-09T00:00"},
			Temperature:              []float64{20.0},
			WeatherCode:              []int{1},
			PrecipitationProbability: []int{0},
		},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(count, 1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func seedForecastEntity(t *testing.T, db *gorm.DB, tenantID, householdID uuid.UUID, locationID *uuid.UUID, lat, lon float64, fetchedAt time.Time) forecast.Entity {
	t.Helper()
	now := time.Now().UTC()
	e := forecast.Entity{
		Id:           uuid.New(),
		TenantId:     tenantID,
		HouseholdId:  householdID,
		LocationId:   locationID,
		Latitude:     lat,
		Longitude:    lon,
		Units:        "metric",
		CurrentData:  forecast.JSONCurrentData{Temperature: 0, WeatherCode: 0, Summary: "stale", Icon: "cloud"},
		ForecastData: forecast.JSONForecastData{},
		FetchedAt:    fetchedAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("failed to seed forecast entity: %v", err)
	}
	return e
}

// TestRefreshAll_RefreshesMixedNilAndNonNilLocationRows verifies that the
// refresh loop iterates every cache row and calls the upstream once per row,
// independent of whether the row represents the household primary
// (location_id IS NULL) or a saved location of interest (non-nil location_id).
func TestRefreshAll_RefreshesMixedNilAndNonNilLocationRows(t *testing.T) {
	db := setupRefreshTestDB(t)

	var fetchCount int64
	srv := newCountingForecastServer(t, &fetchCount)
	defer srv.Close()
	client := openmeteo.NewClientWithEndpoints(srv.URL, srv.URL)

	tenantID := uuid.New()
	householdA := uuid.New()
	householdB := uuid.New()
	locA := uuid.New()
	locB := uuid.New()

	stale := time.Now().Add(-2 * time.Hour).UTC()

	// Three rows: one primary (nil), two saved-location rows.
	primary := seedForecastEntity(t, db, tenantID, householdA, nil, 10, 20, stale)
	savedA := seedForecastEntity(t, db, tenantID, householdA, &locA, 30, 40, stale)
	savedB := seedForecastEntity(t, db, tenantID, householdB, &locB, 50, 60, stale)

	l, hook := test.NewNullLogger()
	l.SetLevel(logrus.DebugLevel)
	refreshAll(context.Background(), db, client, l)

	if got := atomic.LoadInt64(&fetchCount); got != 3 {
		t.Errorf("expected 3 upstream fetches (one per row), got %d", got)
	}

	// Each seeded row's FetchedAt should have advanced.
	for _, original := range []forecast.Entity{primary, savedA, savedB} {
		var refreshed forecast.Entity
		if err := db.Where("household_id = ? AND ((location_id IS NULL AND ? IS NULL) OR location_id = ?)",
			original.HouseholdId, original.LocationId, original.LocationId).
			First(&refreshed).Error; err != nil {
			t.Fatalf("failed to read refreshed row for household %v: %v", original.HouseholdId, err)
		}
		if !refreshed.FetchedAt.After(original.FetchedAt) {
			t.Errorf("expected FetchedAt to advance for household %v (loc=%v): before=%v after=%v",
				original.HouseholdId, original.LocationId, original.FetchedAt, refreshed.FetchedAt)
		}
		if refreshed.CurrentData.Summary != "Mostly Clear" {
			t.Errorf("expected refreshed Summary 'Mostly Clear', got %q", refreshed.CurrentData.Summary)
		}
	}

	// No warning logs (per-row error logger should be silent on the happy path).
	for _, e := range hook.AllEntries() {
		if e.Level <= logrus.WarnLevel && e.Level != logrus.InfoLevel {
			t.Errorf("unexpected warn/error log: %s %v", e.Message, e.Data)
		}
	}
}

// TestRefreshAll_LogsLocationFieldOnFailure verifies that per-row error logs
// include a `location_id` field — "primary" for the nil case, the UUID
// otherwise — when the upstream fetch fails.
func TestRefreshAll_LogsLocationFieldOnFailure(t *testing.T) {
	db := setupRefreshTestDB(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	client := openmeteo.NewClientWithEndpoints(srv.URL, srv.URL)

	tenantID := uuid.New()
	householdID := uuid.New()
	locID := uuid.New()
	stale := time.Now().Add(-time.Hour).UTC()

	seedForecastEntity(t, db, tenantID, householdID, nil, 1, 2, stale)
	seedForecastEntity(t, db, tenantID, householdID, &locID, 3, 4, stale)

	l, hook := test.NewNullLogger()
	refreshAll(context.Background(), db, client, l)

	var sawPrimary, sawSaved bool
	for _, e := range hook.AllEntries() {
		if e.Level != logrus.WarnLevel {
			continue
		}
		v, ok := e.Data["location_id"]
		if !ok {
			t.Errorf("warn log missing location_id field: %+v", e.Data)
			continue
		}
		switch v {
		case "primary":
			sawPrimary = true
		case locID.String():
			sawSaved = true
		}
	}
	if !sawPrimary {
		t.Errorf("expected a warn log with location_id=\"primary\"")
	}
	if !sawSaved {
		t.Errorf("expected a warn log with location_id=%s", locID)
	}
}
