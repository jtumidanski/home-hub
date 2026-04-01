package list

import (
	"context"
	"testing"

	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/services/shopping-service/internal/item"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
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

	err = db.AutoMigrate(&Entity{}, &item.Entity{})
	require.NoError(t, err)
	return db
}

func testContext(tenantID uuid.UUID) context.Context {
	t := tenantctx.New(tenantID, uuid.New(), uuid.New())
	return tenantctx.WithContext(context.Background(), t)
}

func newTestProcessor(t *testing.T, db *gorm.DB, ctx context.Context) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, ctx, db)
}

func createTestList(t *testing.T, p *Processor, tenantID, householdID, userID uuid.UUID, name string) Model {
	t.Helper()
	m, err := p.Create(tenantID, householdID, userID, name)
	require.NoError(t, err)
	return m
}

func TestProcessor_Create(t *testing.T) {
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name     string
		listName string
		wantErr  error
		validate func(t *testing.T, m Model)
	}{
		{
			name:     "creates active list",
			listName: "Weekly Groceries",
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, "Weekly Groceries", m.Name())
				assert.Equal(t, "active", m.Status())
				assert.NotEqual(t, uuid.Nil, m.Id())
				assert.Equal(t, tenantID, m.TenantID())
				assert.Equal(t, householdID, m.HouseholdID())
				assert.Equal(t, userID, m.CreatedBy())
			},
		},
		{
			name:     "rejects empty name",
			listName: "",
			wantErr:  ErrNameRequired,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			ctx := testContext(tenantID)
			p := newTestProcessor(t, db, ctx)

			m, err := p.Create(tenantID, householdID, userID, tc.listName)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			tc.validate(t, m)
		})
	}
}

func TestProcessor_Get(t *testing.T) {
	tenantID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID)
	p := newTestProcessor(t, db, ctx)

	created := createTestList(t, p, tenantID, uuid.New(), uuid.New(), "Groceries")

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{name: "existing list", id: created.Id()},
		{name: "non-existent list", id: uuid.New(), wantErr: ErrNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := p.Get(tc.id)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "Groceries", m.Name())
		})
	}
}

func TestProcessor_List(t *testing.T) {
	tenantID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID)
	p := newTestProcessor(t, db, ctx)

	createTestList(t, p, tenantID, uuid.New(), uuid.New(), "Active List")
	archived := createTestList(t, p, tenantID, uuid.New(), uuid.New(), "Archived List")
	p.Archive(archived.Id())

	tests := []struct {
		name      string
		status    string
		wantCount int
	}{
		{name: "active lists", status: "active", wantCount: 1},
		{name: "archived lists", status: "archived", wantCount: 1},
		{name: "default returns active", status: "", wantCount: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			models, err := p.List(tc.status)
			require.NoError(t, err)
			assert.Len(t, models, tc.wantCount)
		})
	}
}

func TestProcessor_Update(t *testing.T) {
	tenantID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID)
	p := newTestProcessor(t, db, ctx)

	created := createTestList(t, p, tenantID, uuid.New(), uuid.New(), "Original")
	archived := createTestList(t, p, tenantID, uuid.New(), uuid.New(), "To Archive")
	p.Archive(archived.Id())

	tests := []struct {
		name    string
		id      uuid.UUID
		newName string
		wantErr error
	}{
		{name: "updates name", id: created.Id(), newName: "Updated"},
		{name: "not found", id: uuid.New(), newName: "X", wantErr: ErrNotFound},
		{name: "archived list", id: archived.Id(), newName: "X", wantErr: ErrArchived},
		{name: "empty name", id: created.Id(), newName: "", wantErr: ErrNameRequired},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := p.Update(tc.id, tc.newName)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.newName, m.Name())
		})
	}
}

func TestProcessor_Delete(t *testing.T) {
	tenantID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID)
	p := newTestProcessor(t, db, ctx)

	created := createTestList(t, p, tenantID, uuid.New(), uuid.New(), "To Delete")

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{name: "deletes existing list", id: created.Id()},
		{name: "not found", id: uuid.New(), wantErr: ErrNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := p.Delete(tc.id)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestProcessor_Archive(t *testing.T) {
	tenantID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID)
	p := newTestProcessor(t, db, ctx)

	created := createTestList(t, p, tenantID, uuid.New(), uuid.New(), "To Archive")

	t.Run("archives active list", func(t *testing.T) {
		m, err := p.Archive(created.Id())
		require.NoError(t, err)
		assert.Equal(t, "archived", m.Status())
		assert.NotNil(t, m.ArchivedAt())
	})

	t.Run("already archived", func(t *testing.T) {
		_, err := p.Archive(created.Id())
		assert.ErrorIs(t, err, ErrAlreadyArchived)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := p.Archive(uuid.New())
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestProcessor_Unarchive(t *testing.T) {
	tenantID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID)
	p := newTestProcessor(t, db, ctx)

	created := createTestList(t, p, tenantID, uuid.New(), uuid.New(), "To Unarchive")
	p.Archive(created.Id())

	t.Run("unarchives archived list", func(t *testing.T) {
		m, err := p.Unarchive(created.Id())
		require.NoError(t, err)
		assert.Equal(t, "active", m.Status())
		assert.Nil(t, m.ArchivedAt())
	})

	t.Run("not archived", func(t *testing.T) {
		_, err := p.Unarchive(created.Id())
		assert.ErrorIs(t, err, ErrNotArchived)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := p.Unarchive(uuid.New())
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestProcessor_GetWithItems(t *testing.T) {
	tenantID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID)
	p := newTestProcessor(t, db, ctx)

	created := createTestList(t, p, tenantID, uuid.New(), uuid.New(), "With Items")
	p.AddItem(created.Id(), item.AddInput{Name: "Milk"})
	p.AddItem(created.Id(), item.AddInput{Name: "Bread"})

	t.Run("returns list with items", func(t *testing.T) {
		m, items, err := p.GetWithItems(created.Id())
		require.NoError(t, err)
		assert.Equal(t, "With Items", m.Name())
		assert.Len(t, items, 2)
	})

	t.Run("not found", func(t *testing.T) {
		_, _, err := p.GetWithItems(uuid.New())
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestProcessor_AddItem(t *testing.T) {
	tenantID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID)
	p := newTestProcessor(t, db, ctx)

	active := createTestList(t, p, tenantID, uuid.New(), uuid.New(), "Active")
	archived := createTestList(t, p, tenantID, uuid.New(), uuid.New(), "Archived")
	p.Archive(archived.Id())

	tests := []struct {
		name    string
		listID  uuid.UUID
		input   item.AddInput
		wantErr error
	}{
		{
			name:   "adds item to active list",
			listID: active.Id(),
			input:  item.AddInput{Name: "Milk"},
		},
		{
			name:    "rejects archived list",
			listID:  archived.Id(),
			input:   item.AddInput{Name: "Milk"},
			wantErr: ErrArchived,
		},
		{
			name:    "list not found",
			listID:  uuid.New(),
			input:   item.AddInput{Name: "Milk"},
			wantErr: ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := p.AddItem(tc.listID, tc.input)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "Milk", m.Name())
		})
	}
}

func TestProcessor_RemoveItem(t *testing.T) {
	tenantID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID)
	p := newTestProcessor(t, db, ctx)

	active := createTestList(t, p, tenantID, uuid.New(), uuid.New(), "Active")
	added, _ := p.AddItem(active.Id(), item.AddInput{Name: "To Remove"})

	tests := []struct {
		name    string
		listID  uuid.UUID
		itemID  uuid.UUID
		wantErr error
	}{
		{name: "removes item", listID: active.Id(), itemID: added.Id()},
		{name: "list not found", listID: uuid.New(), itemID: added.Id(), wantErr: ErrNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := p.RemoveItem(tc.listID, tc.itemID)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestProcessor_UncheckAllItems(t *testing.T) {
	tenantID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID)
	p := newTestProcessor(t, db, ctx)

	active := createTestList(t, p, tenantID, uuid.New(), uuid.New(), "Active")
	i1, _ := p.AddItem(active.Id(), item.AddInput{Name: "Item 1"})
	i2, _ := p.AddItem(active.Id(), item.AddInput{Name: "Item 2"})
	p.CheckItem(active.Id(), i1.Id(), true)
	p.CheckItem(active.Id(), i2.Id(), true)

	m, items, err := p.UncheckAllItems(active.Id())
	require.NoError(t, err)
	assert.Equal(t, "Active", m.Name())
	assert.Len(t, items, 2)
	for _, i := range items {
		assert.False(t, i.Checked(), "expected all items unchecked")
	}
}

func TestProcessor_ImportItems(t *testing.T) {
	tenantID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID)
	p := newTestProcessor(t, db, ctx)

	active := createTestList(t, p, tenantID, uuid.New(), uuid.New(), "Import Target")

	inputs := []item.AddInput{
		{Name: "Flour"},
		{Name: "Sugar"},
		{Name: "Butter"},
	}

	m, items, err := p.ImportItems(active.Id(), inputs)
	require.NoError(t, err)
	assert.Equal(t, "Import Target", m.Name())
	assert.Len(t, items, 3)
}
