package recipeimage

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	database "github.com/jtumidanski/home-hub/shared/go/database"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
)

// newTestDB returns an in-memory SQLite DB pinned to one connection so the
// schema survives across goroutines (each new :memory: connection would be
// a fresh, empty database).
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	if err := Migration(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func tenantCtx(tenantID, householdID uuid.UUID) context.Context {
	return tenantctx.WithContext(context.Background(), tenantctx.New(tenantID, householdID, uuid.New()))
}
