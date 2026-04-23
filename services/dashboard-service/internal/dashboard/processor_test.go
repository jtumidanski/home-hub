package dashboard

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)

	err = db.AutoMigrate(&Entity{})
	require.NoError(t, err)
	return db
}

func newTestProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

func TestProcessorCreateHousehold(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()

	m, err := p.Create(tid, hid, uid, CreateAttrs{
		Name:   "Home",
		Scope:  "household",
		Layout: json.RawMessage(`{"version":1,"widgets":[]}`),
	})
	require.NoError(t, err)
	require.Equal(t, "Home", m.Name())
	require.Nil(t, m.UserID())
	require.Equal(t, tid, m.TenantID())
	require.Equal(t, hid, m.HouseholdID())
	require.Equal(t, 0, m.SortOrder())
}
