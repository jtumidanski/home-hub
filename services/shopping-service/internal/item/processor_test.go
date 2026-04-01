package item

import (
	"context"
	"testing"

	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
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

	err = db.AutoMigrate(&Entity{})
	require.NoError(t, err)
	return db
}

func newTestProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

func TestProcessor_Add(t *testing.T) {
	tests := []struct {
		name     string
		input    AddInput
		wantErr  error
		validate func(t *testing.T, m Model)
	}{
		{
			name:  "creates item with auto position",
			input: AddInput{ListID: uuid.New(), Name: "Milk"},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, "Milk", m.Name())
				assert.Equal(t, 1, m.Position())
				assert.False(t, m.Checked())
				assert.NotEqual(t, uuid.Nil, m.Id())
			},
		},
		{
			name: "creates item with explicit position",
			input: func() AddInput {
				pos := 5
				return AddInput{ListID: uuid.New(), Name: "Bread", Position: &pos}
			}(),
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, "Bread", m.Name())
				assert.Equal(t, 5, m.Position())
			},
		},
		{
			name: "creates item with optional fields",
			input: func() AddInput {
				qty := "2 lb"
				catID := uuid.New()
				catName := "Produce"
				sortOrder := 1
				return AddInput{
					ListID: uuid.New(), Name: "Chicken",
					Quantity: &qty, CategoryID: &catID,
					CategoryName: &catName, CategorySortOrder: &sortOrder,
				}
			}(),
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, "Chicken", m.Name())
				assert.NotNil(t, m.Quantity())
				assert.NotNil(t, m.CategoryID())
				assert.NotNil(t, m.CategoryName())
				assert.NotNil(t, m.CategorySortOrder())
			},
		},
		{
			name:    "rejects empty name",
			input:   AddInput{ListID: uuid.New(), Name: ""},
			wantErr: ErrNameRequired,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			p := newTestProcessor(t, db)

			m, err := p.Add(tc.input)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			tc.validate(t, m)
		})
	}
}

func TestProcessor_Add_AutoIncrementsPosition(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	listID := uuid.New()

	m1, err := p.Add(AddInput{ListID: listID, Name: "First"})
	require.NoError(t, err)
	assert.Equal(t, 1, m1.Position())

	m2, err := p.Add(AddInput{ListID: listID, Name: "Second"})
	require.NoError(t, err)
	assert.Equal(t, 2, m2.Position())
}

func TestProcessor_GetByListID(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	listID := uuid.New()

	tests := []struct {
		name      string
		setup     func()
		wantCount int
	}{
		{
			name:      "returns empty for no items",
			setup:     func() {},
			wantCount: 0,
		},
		{
			name: "returns items for list",
			setup: func() {
				p.Add(AddInput{ListID: listID, Name: "Item A"})
				p.Add(AddInput{ListID: listID, Name: "Item B"})
			},
			wantCount: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			models, err := p.GetByListID(listID)
			require.NoError(t, err)
			assert.Len(t, models, tc.wantCount)
		})
	}
}

func TestProcessor_Update(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	created, err := p.Add(AddInput{ListID: uuid.New(), Name: "Original"})
	require.NoError(t, err)

	newName := "Updated"
	tests := []struct {
		name     string
		id       uuid.UUID
		input    UpdateInput
		wantErr  error
		validate func(t *testing.T, m Model)
	}{
		{
			name:  "updates name",
			id:    created.Id(),
			input: UpdateInput{Name: &newName},
			validate: func(t *testing.T, m Model) {
				assert.Equal(t, "Updated", m.Name())
			},
		},
		{
			name:    "not found",
			id:      uuid.New(),
			input:   UpdateInput{Name: &newName},
			wantErr: ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := p.Update(tc.id, tc.input)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			tc.validate(t, m)
		})
	}
}

func TestProcessor_Delete(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	created, err := p.Add(AddInput{ListID: uuid.New(), Name: "To Delete"})
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{name: "deletes existing item", id: created.Id()},
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

func TestProcessor_Check(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)

	created, err := p.Add(AddInput{ListID: uuid.New(), Name: "Checkable"})
	require.NoError(t, err)
	assert.False(t, created.Checked())

	tests := []struct {
		name    string
		checked bool
		want    bool
	}{
		{name: "check item", checked: true, want: true},
		{name: "uncheck item", checked: false, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := p.Check(created.Id(), tc.checked)
			require.NoError(t, err)
			assert.Equal(t, tc.want, m.Checked())
		})
	}
}

func TestProcessor_UncheckAll(t *testing.T) {
	db := setupTestDB(t)
	p := newTestProcessor(t, db)
	listID := uuid.New()

	m1, _ := p.Add(AddInput{ListID: listID, Name: "Item 1"})
	m2, _ := p.Add(AddInput{ListID: listID, Name: "Item 2"})
	p.Check(m1.Id(), true)
	p.Check(m2.Id(), true)

	err := p.UncheckAll(listID)
	require.NoError(t, err)

	items, err := p.GetByListID(listID)
	require.NoError(t, err)
	for _, item := range items {
		assert.False(t, item.Checked(), "expected all items unchecked")
	}
}
