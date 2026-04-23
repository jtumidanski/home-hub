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

func TestProcessorListScopesVisibility(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid := uuid.New(), uuid.New()
	callerUID := uuid.New()
	otherUID := uuid.New()
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)

	_, err := p.Create(tid, hid, callerUID, CreateAttrs{Name: "Shared", Scope: "household", Layout: layoutJSON})
	require.NoError(t, err)
	_, err = p.Create(tid, hid, callerUID, CreateAttrs{Name: "Mine", Scope: "user", Layout: layoutJSON})
	require.NoError(t, err)
	_, err = p.Create(tid, hid, otherUID, CreateAttrs{Name: "Theirs", Scope: "user", Layout: layoutJSON})
	require.NoError(t, err)

	list, err := p.List(tid, hid, callerUID)
	require.NoError(t, err)
	require.Len(t, list, 2)

	names := map[string]bool{}
	for _, m := range list {
		names[m.Name()] = true
	}
	require.True(t, names["Shared"])
	require.True(t, names["Mine"])
	require.False(t, names["Theirs"])
}

func TestProcessorGetBlocksInvisibleRow(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid := uuid.New(), uuid.New()
	callerUID := uuid.New()
	otherUID := uuid.New()

	m, err := p.Create(tid, hid, otherUID, CreateAttrs{
		Name:   "Theirs",
		Scope:  "user",
		Layout: json.RawMessage(`{"version":1,"widgets":[]}`),
	})
	require.NoError(t, err)

	_, err = p.GetByID(m.Id(), tid, hid, callerUID)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestProcessorUpdateHouseholdAllowsAnyMember(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid := uuid.New(), uuid.New()
	creator, other := uuid.New(), uuid.New()
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)

	m, err := p.Create(tid, hid, creator, CreateAttrs{Name: "Home", Scope: "household", Layout: layoutJSON})
	require.NoError(t, err)

	newName := "Household Home"
	updated, err := p.Update(m.Id(), tid, hid, other, UpdateAttrs{Name: &newName})
	require.NoError(t, err)
	require.Equal(t, "Household Home", updated.Name())
}

func TestProcessorUpdateUserScopedRejectsNonOwner(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid := uuid.New(), uuid.New()
	owner, other := uuid.New(), uuid.New()
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)

	m, err := p.Create(tid, hid, owner, CreateAttrs{Name: "Mine", Scope: "user", Layout: layoutJSON})
	require.NoError(t, err)

	newName := "Hacked"
	_, err = p.Update(m.Id(), tid, hid, other, UpdateAttrs{Name: &newName})
	require.ErrorIs(t, err, ErrForbidden)
}

func TestProcessorDeleteUserScopedRejectsNonOwner(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid := uuid.New(), uuid.New()
	owner, other := uuid.New(), uuid.New()
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)

	mine, err := p.Create(tid, hid, owner, CreateAttrs{Name: "Mine", Scope: "user", Layout: layoutJSON})
	require.NoError(t, err)

	err = p.Delete(mine.Id(), tid, hid, other)
	require.ErrorIs(t, err, ErrForbidden)
}

func TestProcessorReorderSingleScope(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)

	a, err := p.Create(tid, hid, uid, CreateAttrs{Name: "A", Scope: "household", Layout: layoutJSON})
	require.NoError(t, err)
	b, err := p.Create(tid, hid, uid, CreateAttrs{Name: "B", Scope: "household", Layout: layoutJSON})
	require.NoError(t, err)

	_, err = p.Reorder(tid, hid, uid, []ReorderPair{
		{ID: a.Id(), SortOrder: 10},
		{ID: b.Id(), SortOrder: 5},
	})
	require.NoError(t, err)

	got, err := p.GetByID(a.Id(), tid, hid, uid)
	require.NoError(t, err)
	require.Equal(t, 10, got.SortOrder())
	got, err = p.GetByID(b.Id(), tid, hid, uid)
	require.NoError(t, err)
	require.Equal(t, 5, got.SortOrder())
}

func TestProcessorReorderMixedScopeFails(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)

	hh, err := p.Create(tid, hid, uid, CreateAttrs{Name: "HH", Scope: "household", Layout: layoutJSON})
	require.NoError(t, err)
	mine, err := p.Create(tid, hid, uid, CreateAttrs{Name: "Mine", Scope: "user", Layout: layoutJSON})
	require.NoError(t, err)

	_, err = p.Reorder(tid, hid, uid, []ReorderPair{
		{ID: hh.Id(), SortOrder: 0},
		{ID: mine.Id(), SortOrder: 1},
	})
	require.ErrorIs(t, err, ErrMixedScope)
}

func TestProcessorReorderUnknownIDFails(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)

	hh, err := p.Create(tid, hid, uid, CreateAttrs{Name: "HH", Scope: "household", Layout: layoutJSON})
	require.NoError(t, err)

	_, err = p.Reorder(tid, hid, uid, []ReorderPair{
		{ID: hh.Id(), SortOrder: 0},
		{ID: uuid.New(), SortOrder: 1},
	})
	require.ErrorIs(t, err, ErrNotFound)
}

func TestProcessorReorderInvisibleIDFails(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid := uuid.New(), uuid.New()
	caller, other := uuid.New(), uuid.New()
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)

	theirs, err := p.Create(tid, hid, other, CreateAttrs{Name: "Theirs", Scope: "user", Layout: layoutJSON})
	require.NoError(t, err)

	_, err = p.Reorder(tid, hid, caller, []ReorderPair{{ID: theirs.Id(), SortOrder: 0}})
	require.ErrorIs(t, err, ErrNotFound)
}

func TestProcessorDeleteHouseholdAllowsAnyMember(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	tid, hid := uuid.New(), uuid.New()
	creator, other := uuid.New(), uuid.New()
	layoutJSON := json.RawMessage(`{"version":1,"widgets":[]}`)

	m, err := p.Create(tid, hid, creator, CreateAttrs{Name: "Home", Scope: "household", Layout: layoutJSON})
	require.NoError(t, err)

	err = p.Delete(m.Id(), tid, hid, other)
	require.NoError(t, err)

	_, err = p.GetByID(m.Id(), tid, hid, creator)
	require.ErrorIs(t, err, ErrNotFound)
}
