package locationofinterest

import (
	"context"
	"strings"
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

func newTestProcessor(t *testing.T, db *gorm.DB, warmer CacheWarmer) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db, warmer)
}

type recordingWarmer struct {
	calls []struct {
		tenantID, householdID, locationID uuid.UUID
		lat, lon                          float64
	}
	err error
}

func (r *recordingWarmer) WarmLocationCache(tenantID, householdID, locationID uuid.UUID, lat, lon float64) error {
	r.calls = append(r.calls, struct {
		tenantID, householdID, locationID uuid.UUID
		lat, lon                          float64
	}{tenantID, householdID, locationID, lat, lon})
	return r.err
}

var errBoom = &boomError{}

type boomError struct{}

func (b *boomError) Error() string { return "boom" }

func TestCreate(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "normalizes coordinates to four decimals and forwards to warmer",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				w := &recordingWarmer{}
				p := newTestProcessor(t, db, w)

				m, err := p.Create(uuid.New(), uuid.New(), CreateInput{
					PlaceName: "Somewhere",
					Latitude:  40.123456789,
					Longitude: -74.987654321,
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if m.Latitude() != 40.1235 {
					t.Errorf("expected lat 40.1235, got %v", m.Latitude())
				}
				if m.Longitude() != -74.9877 {
					t.Errorf("expected lon -74.9877, got %v", m.Longitude())
				}
				if len(w.calls) != 1 {
					t.Fatalf("expected warmer to be called once, got %d", len(w.calls))
				}
				if w.calls[0].lat != 40.1235 || w.calls[0].lon != -74.9877 {
					t.Errorf("warmer received unrounded coords: %+v", w.calls[0])
				}
			},
		},
		{
			name: "trims and stores label",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				p := newTestProcessor(t, db, nil)

				label := "  Mom's House  "
				m, err := p.Create(uuid.New(), uuid.New(), CreateInput{
					Label:     &label,
					PlaceName: "Detroit, MI",
					Latitude:  42.0,
					Longitude: -83.0,
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if m.Label() == nil || *m.Label() != "Mom's House" {
					t.Errorf("expected trimmed label 'Mom's House', got %v", m.Label())
				}
			},
		},
		{
			name: "rejects label over 64 chars",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				p := newTestProcessor(t, db, nil)

				label := strings.Repeat("a", 66)
				_, err := p.Create(uuid.New(), uuid.New(), CreateInput{
					Label:     &label,
					PlaceName: "X",
					Latitude:  0,
					Longitude: 0,
				})
				if err != ErrLabelTooLong {
					t.Errorf("expected ErrLabelTooLong, got %v", err)
				}
			},
		},
		{
			name: "empty label stored as nil",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				p := newTestProcessor(t, db, nil)

				label := "   "
				m, err := p.Create(uuid.New(), uuid.New(), CreateInput{
					Label:     &label,
					PlaceName: "X",
					Latitude:  0,
					Longitude: 0,
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if m.Label() != nil {
					t.Errorf("expected nil label, got %v", *m.Label())
				}
			},
		},
		{
			name: "warmer error is swallowed",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				w := &recordingWarmer{err: errBoom}
				p := newTestProcessor(t, db, w)

				if _, err := p.Create(uuid.New(), uuid.New(), CreateInput{
					PlaceName: "X", Latitude: 0, Longitude: 0,
				}); err != nil {
					t.Errorf("warmer error should not bubble up, got %v", err)
				}
				if len(w.calls) != 1 {
					t.Errorf("expected warmer to be called once, got %d", len(w.calls))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestCreate_Cap(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "cap at ten returns ErrCapReached with PRD-mandated message",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				p := newTestProcessor(t, db, nil)

				tenantID := uuid.New()
				householdID := uuid.New()
				for i := 0; i < MaxLocations; i++ {
					if _, err := p.Create(tenantID, householdID, CreateInput{
						PlaceName: "Place",
						Latitude:  float64(i),
						Longitude: float64(i),
					}); err != nil {
						t.Fatalf("unexpected error on row %d: %v", i, err)
					}
				}

				_, err := p.Create(tenantID, householdID, CreateInput{
					PlaceName: "Eleventh", Latitude: 1, Longitude: 1,
				})
				if err != ErrCapReached {
					t.Errorf("expected ErrCapReached, got %v", err)
				}
				expected := "Households can save up to 10 locations of interest. Remove one to add another."
				if ErrCapReached.Error() != expected {
					t.Errorf("ErrCapReached message must match PRD §4.1 verbatim.\n got: %q\nwant: %q", ErrCapReached.Error(), expected)
				}
			},
		},
		{
			name: "cap is per household",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				p := newTestProcessor(t, db, nil)

				tenantID := uuid.New()
				householdA := uuid.New()
				householdB := uuid.New()
				for i := 0; i < MaxLocations; i++ {
					if _, err := p.Create(tenantID, householdA, CreateInput{
						PlaceName: "A", Latitude: 1, Longitude: 1,
					}); err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
				}
				if _, err := p.Create(tenantID, householdB, CreateInput{
					PlaceName: "B", Latitude: 2, Longitude: 2,
				}); err != nil {
					t.Errorf("expected household B unaffected by household A cap, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "missing id returns ErrNotFound",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				p := newTestProcessor(t, db, nil)

				if _, err := p.Get(uuid.New(), uuid.New()); err != ErrNotFound {
					t.Errorf("expected ErrNotFound, got %v", err)
				}
			},
		},
		{
			name: "cross-household returns ErrNotFound but owner can read",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				p := newTestProcessor(t, db, nil)

				tenantID := uuid.New()
				ownerHousehold := uuid.New()
				otherHousehold := uuid.New()

				m, err := p.Create(tenantID, ownerHousehold, CreateInput{
					PlaceName: "Owned", Latitude: 10, Longitude: 10,
				})
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}

				if _, err := p.Get(otherHousehold, m.Id()); err != ErrNotFound {
					t.Errorf("expected ErrNotFound for cross-household read, got %v", err)
				}
				got, err := p.Get(ownerHousehold, m.Id())
				if err != nil {
					t.Fatalf("owner read failed: %v", err)
				}
				if got.Id() != m.Id() {
					t.Errorf("owner got wrong id")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "only returns rows for the requested household",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				p := newTestProcessor(t, db, nil)

				tenantID := uuid.New()
				householdA := uuid.New()
				householdB := uuid.New()
				for i := 0; i < 3; i++ {
					if _, err := p.Create(tenantID, householdA, CreateInput{
						PlaceName: "A", Latitude: float64(i), Longitude: float64(i),
					}); err != nil {
						t.Fatalf("setup A failed: %v", err)
					}
				}
				if _, err := p.Create(tenantID, householdB, CreateInput{
					PlaceName: "B", Latitude: 1, Longitude: 1,
				}); err != nil {
					t.Fatalf("setup B failed: %v", err)
				}

				rows, err := p.List(householdA)
				if err != nil {
					t.Fatalf("list failed: %v", err)
				}
				if len(rows) != 3 {
					t.Errorf("expected 3 rows for household A, got %d", len(rows))
				}
				for _, r := range rows {
					if r.HouseholdID() != householdA {
						t.Errorf("list returned row from wrong household: %v", r.HouseholdID())
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestUpdateLabel(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "rejects label too long",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				p := newTestProcessor(t, db, nil)

				m, err := p.Create(uuid.New(), uuid.New(), CreateInput{
					PlaceName: "X", Latitude: 0, Longitude: 0,
				})
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				long := strings.Repeat("x", 65)
				if _, err := p.UpdateLabel(m.HouseholdID(), m.Id(), &long); err != ErrLabelTooLong {
					t.Errorf("expected ErrLabelTooLong, got %v", err)
				}
			},
		},
		{
			name: "empty string clears label",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				p := newTestProcessor(t, db, nil)

				original := "Original"
				m, err := p.Create(uuid.New(), uuid.New(), CreateInput{
					Label: &original, PlaceName: "X", Latitude: 0, Longitude: 0,
				})
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}

				empty := ""
				updated, err := p.UpdateLabel(m.HouseholdID(), m.Id(), &empty)
				if err != nil {
					t.Fatalf("update failed: %v", err)
				}
				if updated.Label() != nil {
					t.Errorf("expected cleared label, got %v", *updated.Label())
				}
			},
		},
		{
			name: "cross-household returns ErrNotFound",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				p := newTestProcessor(t, db, nil)

				owner := uuid.New()
				other := uuid.New()
				m, err := p.Create(uuid.New(), owner, CreateInput{
					PlaceName: "X", Latitude: 0, Longitude: 0,
				})
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				label := "Sneaky"
				if _, err := p.UpdateLabel(other, m.Id(), &label); err != ErrNotFound {
					t.Errorf("expected ErrNotFound, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "cross-household returns ErrNotFound and leaves row intact",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				p := newTestProcessor(t, db, nil)

				owner := uuid.New()
				other := uuid.New()
				m, err := p.Create(uuid.New(), owner, CreateInput{
					PlaceName: "X", Latitude: 0, Longitude: 0,
				})
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				if err := p.Delete(other, m.Id()); err != ErrNotFound {
					t.Errorf("expected ErrNotFound for cross-household delete, got %v", err)
				}
				if _, err := p.Get(owner, m.Id()); err != nil {
					t.Errorf("expected owner row to still exist, got %v", err)
				}
			},
		},
		{
			name: "happy path removes row",
			run: func(t *testing.T) {
				db := setupTestDB(t)
				p := newTestProcessor(t, db, nil)

				householdID := uuid.New()
				m, err := p.Create(uuid.New(), householdID, CreateInput{
					PlaceName: "X", Latitude: 0, Longitude: 0,
				})
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				if err := p.Delete(householdID, m.Id()); err != nil {
					t.Fatalf("delete failed: %v", err)
				}
				if _, err := p.Get(householdID, m.Id()); err != ErrNotFound {
					t.Errorf("expected ErrNotFound after delete, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
