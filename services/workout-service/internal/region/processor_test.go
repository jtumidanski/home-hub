package region

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB stands up an in-memory SQLite for the region processor. The
// per-package Migration helper emits Postgres-only DDL (partial unique
// indexes), so we use AutoMigrate directly. Tenant callbacks are registered
// because every read flows through the shared tenant scope.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	require.NoError(t, db.AutoMigrate(&Entity{}))
	return db
}

func newProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

func TestProcessor_List_SeedsDefaults(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()

	models, err := p.List(tenantID, userID)
	require.NoError(t, err)
	assert.Len(t, models, len(DefaultRegions), "fresh user should be seeded with the default region list")

	// Idempotent — second call must not duplicate the seed.
	models, err = p.List(tenantID, userID)
	require.NoError(t, err)
	assert.Len(t, models, len(DefaultRegions))
}

func TestProcessor_Create_Validation(t *testing.T) {
	tenantID, userID := uuid.New(), uuid.New()
	cases := []struct {
		name      string
		regionName string
		sortOrder int
		wantErr   error
	}{
		{"empty name", "", 0, ErrNameRequired},
		{"too long", string(make([]byte, 51)), 0, ErrNameTooLong},
		{"negative sort", "Calves", -1, ErrInvalidSortOrder},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			p := newProcessor(t, db)
			_, err := p.Create(tenantID, userID, tc.regionName, tc.sortOrder)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestProcessor_Create_RejectsDuplicateForSameUser(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()

	_, err := p.Create(tenantID, userID, "Calves", 0)
	require.NoError(t, err)
	_, err = p.Create(tenantID, userID, "Calves", 0)
	assert.ErrorIs(t, err, ErrDuplicateName)
}

func TestProcessor_Update_RenameAndReorder(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()

	created, err := p.Create(tenantID, userID, "Forearms", 5)
	require.NoError(t, err)

	newName := "Lower Arms"
	newOrder := 9
	updated, err := p.Update(created.Id(), &newName, &newOrder)
	require.NoError(t, err)
	assert.Equal(t, "Lower Arms", updated.Name())
	assert.Equal(t, 9, updated.SortOrder())
}

func TestProcessor_Delete_SoftDeletes(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()

	created, err := p.Create(tenantID, userID, "Traps", 0)
	require.NoError(t, err)

	require.NoError(t, p.Delete(created.Id()))
	_, err = p.Get(created.Id())
	assert.ErrorIs(t, err, ErrNotFound, "Get must hide soft-deleted regions")

	// The row still exists in the table — soft delete, not hard delete.
	var ent Entity
	require.NoError(t, db.Where("id = ?", created.Id()).First(&ent).Error)
	require.NotNil(t, ent.DeletedAt)
}
