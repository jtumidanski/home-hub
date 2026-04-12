package oauthstate

import (
	"context"
	"testing"
	"time"

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

func newProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

func TestCreate_PersistsState(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()

	m, err := p.Create(tenantID, householdID, userID, "https://example.com/callback", false)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if m.TenantID() != tenantID {
		t.Errorf("expected tenantID %v, got %v", tenantID, m.TenantID())
	}
	if m.HouseholdID() != householdID {
		t.Errorf("expected householdID %v, got %v", householdID, m.HouseholdID())
	}
	if m.UserID() != userID {
		t.Errorf("expected userID %v, got %v", userID, m.UserID())
	}
	if m.RedirectURI() != "https://example.com/callback" {
		t.Errorf("expected redirect URI, got %q", m.RedirectURI())
	}
	if m.IsExpired() {
		t.Error("newly created state should not be expired")
	}

	var count int64
	db.WithContext(database.WithoutTenantFilter(context.Background())).Model(&Entity{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestCreate_SetsReauthorizeFlag(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	m, err := p.Create(uuid.New(), uuid.New(), uuid.New(), "https://example.com/callback", true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !m.Reauthorize() {
		t.Error("expected Reauthorize() to be true")
	}
}

func TestValidateAndConsume_ValidState(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	created, err := p.Create(uuid.New(), uuid.New(), uuid.New(), "https://example.com/callback", false)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	consumed, err := p.ValidateAndConsume(created.Id())
	if err != nil {
		t.Fatalf("ValidateAndConsume: %v", err)
	}
	if consumed.Id() != created.Id() {
		t.Errorf("expected id %v, got %v", created.Id(), consumed.Id())
	}

	var count int64
	db.WithContext(database.WithoutTenantFilter(context.Background())).Model(&Entity{}).Count(&count)
	if count != 0 {
		t.Errorf("expected state to be consumed (deleted), got %d rows", count)
	}
}

func TestValidateAndConsume_ExpiredState(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	now := time.Now().UTC()
	e := Entity{
		Id:          uuid.New(),
		TenantId:    uuid.New(),
		HouseholdId: uuid.New(),
		UserId:      uuid.New(),
		RedirectUri: "https://example.com/callback",
		ExpiresAt:   now.Add(-1 * time.Minute),
		CreatedAt:   now.Add(-11 * time.Minute),
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("failed to seed expired state: %v", err)
	}

	_, err := p.ValidateAndConsume(e.Id)
	if err != ErrStateExpired {
		t.Fatalf("expected ErrStateExpired, got %v", err)
	}

	var count int64
	db.WithContext(database.WithoutTenantFilter(context.Background())).Model(&Entity{}).Count(&count)
	if count != 0 {
		t.Errorf("expected expired state to be deleted, got %d rows", count)
	}
}

func TestValidateAndConsume_NotFound(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	_, err := p.ValidateAndConsume(uuid.New())
	if err != ErrStateNotFound {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

func TestCleanupExpired_RemovesOnlyExpired(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)

	now := time.Now().UTC()

	expired := Entity{
		Id:          uuid.New(),
		TenantId:    uuid.New(),
		HouseholdId: uuid.New(),
		UserId:      uuid.New(),
		RedirectUri: "https://example.com/callback",
		ExpiresAt:   now.Add(-5 * time.Minute),
		CreatedAt:   now.Add(-15 * time.Minute),
	}
	valid := Entity{
		Id:          uuid.New(),
		TenantId:    uuid.New(),
		HouseholdId: uuid.New(),
		UserId:      uuid.New(),
		RedirectUri: "https://example.com/callback",
		ExpiresAt:   now.Add(5 * time.Minute),
		CreatedAt:   now,
	}
	if err := db.Create(&expired).Error; err != nil {
		t.Fatalf("failed to seed expired: %v", err)
	}
	if err := db.Create(&valid).Error; err != nil {
		t.Fatalf("failed to seed valid: %v", err)
	}

	p.CleanupExpired()

	var count int64
	db.WithContext(database.WithoutTenantFilter(context.Background())).Model(&Entity{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 remaining state, got %d", count)
	}

	var remaining Entity
	db.WithContext(database.WithoutTenantFilter(context.Background())).First(&remaining)
	if remaining.Id != valid.Id {
		t.Errorf("expected valid state to remain, got id %v", remaining.Id)
	}
}
