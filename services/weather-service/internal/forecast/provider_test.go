package forecast

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupForecastTestDB(t *testing.T) *gorm.DB {
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

func seedEntity(t *testing.T, db *gorm.DB, tenantID, householdID uuid.UUID, locationID *uuid.UUID, lat, lon float64) Entity {
	t.Helper()
	now := time.Now().UTC()
	e := Entity{
		Id:           uuid.New(),
		TenantId:     tenantID,
		HouseholdId:  householdID,
		LocationId:   locationID,
		Latitude:     lat,
		Longitude:    lon,
		Units:        "metric",
		CurrentData:  JSONCurrentData{Temperature: 20, WeatherCode: 1, Summary: "Mostly Clear", Icon: "cloud-sun"},
		ForecastData: JSONForecastData{},
		FetchedAt:    now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("failed to seed entity: %v", err)
	}
	return e
}

func TestGetByHouseholdAndLocation_PrimaryRow(t *testing.T) {
	db := setupForecastTestDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	otherLoc := uuid.New()

	primary := seedEntity(t, db, tenantID, householdID, nil, 10, 20)
	saved := seedEntity(t, db, tenantID, householdID, &otherLoc, 30, 40)

	got, err := getByHouseholdAndLocation(householdID, nil)(db)()
	if err != nil {
		t.Fatalf("primary lookup failed: %v", err)
	}
	if got.Id != primary.Id {
		t.Errorf("expected primary row %v, got %v", primary.Id, got.Id)
	}
	if got.LocationId != nil {
		t.Errorf("primary row should have nil LocationId, got %v", *got.LocationId)
	}
	// Sanity: the saved row exists, but is not what was returned.
	if got.Id == saved.Id {
		t.Errorf("primary lookup unexpectedly returned the saved-location row")
	}
}

func TestGetByHouseholdAndLocation_SavedLocationRow(t *testing.T) {
	db := setupForecastTestDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	locationID := uuid.New()

	primary := seedEntity(t, db, tenantID, householdID, nil, 10, 20)
	saved := seedEntity(t, db, tenantID, householdID, &locationID, 30, 40)

	got, err := getByHouseholdAndLocation(householdID, &locationID)(db)()
	if err != nil {
		t.Fatalf("saved lookup failed: %v", err)
	}
	if got.Id != saved.Id {
		t.Errorf("expected saved row %v, got %v", saved.Id, got.Id)
	}
	if got.LocationId == nil || *got.LocationId != locationID {
		t.Errorf("expected location id %v, got %v", locationID, got.LocationId)
	}
	if got.Id == primary.Id {
		t.Errorf("saved lookup unexpectedly returned the primary row")
	}
}

func TestGetByHouseholdAndLocation_OtherLocationDoesNotMatchPrimary(t *testing.T) {
	db := setupForecastTestDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	locationID := uuid.New()

	// Only a saved-location row exists; a primary lookup should miss.
	seedEntity(t, db, tenantID, householdID, &locationID, 30, 40)

	_, err := getByHouseholdAndLocation(householdID, nil)(db)()
	if err == nil {
		t.Errorf("expected error for primary lookup with no primary row, got nil")
	}
}

func TestGetByHouseholdAndLocation_PrimaryDoesNotMatchSavedLookup(t *testing.T) {
	db := setupForecastTestDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	locationID := uuid.New()

	// Only a primary row exists; a saved-location lookup should miss.
	seedEntity(t, db, tenantID, householdID, nil, 10, 20)

	_, err := getByHouseholdAndLocation(householdID, &locationID)(db)()
	if err == nil {
		t.Errorf("expected error for saved lookup with no matching row, got nil")
	}
}

func TestGetAll_ReturnsBothPrimaryAndSavedRows(t *testing.T) {
	db := setupForecastTestDB(t)
	tenantID := uuid.New()
	householdID := uuid.New()
	locA := uuid.New()
	locB := uuid.New()

	seedEntity(t, db, tenantID, householdID, nil, 10, 20)
	seedEntity(t, db, tenantID, householdID, &locA, 30, 40)
	seedEntity(t, db, tenantID, householdID, &locB, 50, 60)

	rows, err := getAll()(db)()
	if err != nil {
		t.Fatalf("getAll failed: %v", err)
	}
	if len(rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(rows))
	}
	var primary, saved int
	for _, r := range rows {
		if r.LocationId == nil {
			primary++
		} else {
			saved++
		}
	}
	if primary != 1 || saved != 2 {
		t.Errorf("expected 1 primary + 2 saved, got %d primary + %d saved", primary, saved)
	}
}
