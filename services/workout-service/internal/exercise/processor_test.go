package exercise

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/region"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/theme"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB stands up an in-memory SQLite for the exercise processor.
// We bypass the per-package Migration helpers because they emit Postgres-only
// DDL (partial unique indexes, jsonb @> filters). The behavioural tests below
// don't exercise the list filter, so SQLite is sufficient.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	require.NoError(t, db.AutoMigrate(&Entity{}, &theme.Entity{}, &region.Entity{}))
	return db
}

func newProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

// seedTheme inserts a theme owned by the supplied user. Tests use it as the
// `themeId` argument to Create/Update.
func seedTheme(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Create(&theme.Entity{
		Id:        id,
		TenantId:  tenantID,
		UserId:    userID,
		Name:      "Strength " + id.String()[:8],
		SortOrder: 0,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)
	return id
}

// seedRegion inserts a region owned by the supplied user.
func seedRegion(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Create(&region.Entity{
		Id:        id,
		TenantId:  tenantID,
		UserId:    userID,
		Name:      "Chest " + id.String()[:8],
		SortOrder: 0,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}).Error)
	return id
}

func TestProcessor_Create_HappyPathStrength(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()
	themeID := seedTheme(t, db, tenantID, userID)
	regionID := seedRegion(t, db, tenantID, userID)

	m, err := p.Create(tenantID, userID, CreateInput{
		Name:     "Bench Press",
		Kind:     KindStrength,
		ThemeID:  themeID,
		RegionID: regionID,
		Defaults: Defaults{Sets: ptrInt(3), Reps: ptrInt(8), Weight: ptrFloat(135), WeightUnit: ptrString("lb")},
	})
	require.NoError(t, err)
	assert.Equal(t, "Bench Press", m.Name())
	assert.Equal(t, KindStrength, m.Kind())
}

func TestProcessor_Create_DuplicateNameAllowedAcrossUsers(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID := uuid.New()
	userA, userB := uuid.New(), uuid.New()
	themeA := seedTheme(t, db, tenantID, userA)
	regionA := seedRegion(t, db, tenantID, userA)
	themeB := seedTheme(t, db, tenantID, userB)
	regionB := seedRegion(t, db, tenantID, userB)

	in := func(tID, rID uuid.UUID) CreateInput {
		return CreateInput{Name: "Deadlift", Kind: KindStrength, ThemeID: tID, RegionID: rID,
			Defaults: Defaults{Sets: ptrInt(5), Reps: ptrInt(5), Weight: ptrFloat(225), WeightUnit: ptrString("lb")}}
	}
	_, err := p.Create(tenantID, userA, in(themeA, regionA))
	require.NoError(t, err)
	_, err = p.Create(tenantID, userB, in(themeB, regionB))
	assert.NoError(t, err, "duplicate names are scoped per user, not per tenant")
}

func TestProcessor_Create_Rejections(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID) CreateInput
		wantErr error
	}{
		{
			name: "duplicate name same user",
			setup: func(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID) CreateInput {
				themeID := seedTheme(t, db, tenantID, userID)
				regionID := seedRegion(t, db, tenantID, userID)
				p := newProcessor(t, db)
				in := CreateInput{Name: "Squat", Kind: KindStrength, ThemeID: themeID, RegionID: regionID, Defaults: Defaults{Sets: ptrInt(5), Reps: ptrInt(5), Weight: ptrFloat(225), WeightUnit: ptrString("lb")}}
				_, err := p.Create(tenantID, userID, in)
				require.NoError(t, err)
				return in
			},
			wantErr: ErrDuplicateName,
		},
		{
			name: "theme from other user",
			setup: func(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID) CreateInput {
				otherUser := uuid.New()
				otherTheme := seedTheme(t, db, tenantID, otherUser)
				regionID := seedRegion(t, db, tenantID, userID)
				return CreateInput{Name: "OHP", Kind: KindStrength, ThemeID: otherTheme, RegionID: regionID, Defaults: Defaults{Sets: ptrInt(3), Reps: ptrInt(8), Weight: ptrFloat(95), WeightUnit: ptrString("lb")}}
			},
			wantErr: ErrThemeNotFound,
		},
		{
			name: "missing region",
			setup: func(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID) CreateInput {
				themeID := seedTheme(t, db, tenantID, userID)
				return CreateInput{Name: "Row", Kind: KindStrength, ThemeID: themeID, RegionID: uuid.New(), Defaults: Defaults{Sets: ptrInt(3), Reps: ptrInt(8), Weight: ptrFloat(95), WeightUnit: ptrString("lb")}}
			},
			wantErr: ErrRegionNotFound,
		},
		{
			name: "secondary region from other user",
			setup: func(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID) CreateInput {
				otherUser := uuid.New()
				themeID := seedTheme(t, db, tenantID, userID)
				regionID := seedRegion(t, db, tenantID, userID)
				otherRegion := seedRegion(t, db, tenantID, otherUser)
				return CreateInput{Name: "Pull", Kind: KindStrength, ThemeID: themeID, RegionID: regionID, SecondaryRegionIDs: []uuid.UUID{otherRegion}, Defaults: Defaults{Sets: ptrInt(3), Reps: ptrInt(8), Weight: ptrFloat(95), WeightUnit: ptrString("lb")}}
			},
			wantErr: ErrSecondaryNotFound,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			tenantID, userID := uuid.New(), uuid.New()
			in := tc.setup(t, db, tenantID, userID)
			p := newProcessor(t, db)
			_, err := p.Create(tenantID, userID, in)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestProcessor_Update_OwnershipReValidatedOnThemeChange(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID := uuid.New()
	userA, userB := uuid.New(), uuid.New()
	themeA := seedTheme(t, db, tenantID, userA)
	regionA := seedRegion(t, db, tenantID, userA)
	otherTheme := seedTheme(t, db, tenantID, userB)

	m, err := p.Create(tenantID, userA, CreateInput{
		Name:     "Press",
		Kind:     KindStrength,
		ThemeID:  themeA,
		RegionID: regionA,
		Defaults: Defaults{Sets: ptrInt(3), Reps: ptrInt(5), Weight: ptrFloat(155), WeightUnit: ptrString("lb")},
	})
	require.NoError(t, err)

	// Re-pointing the theme to one belonging to another user must be rejected.
	_, err = p.Update(m.Id(), UpdateInput{ThemeID: &otherTheme})
	assert.ErrorIs(t, err, ErrThemeNotFound, "Update must re-validate theme ownership")
}

func TestProcessor_Delete_SoftDeletesAndHidesFromGet(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()
	themeID := seedTheme(t, db, tenantID, userID)
	regionID := seedRegion(t, db, tenantID, userID)

	m, err := p.Create(tenantID, userID, CreateInput{
		Name:     "Curl",
		Kind:     KindStrength,
		ThemeID:  themeID,
		RegionID: regionID,
		Defaults: Defaults{Sets: ptrInt(3), Reps: ptrInt(12), Weight: ptrFloat(30), WeightUnit: ptrString("lb")},
	})
	require.NoError(t, err)

	require.NoError(t, p.Delete(m.Id()))
	_, err = p.Get(m.Id())
	assert.ErrorIs(t, err, ErrNotFound, "Get must hide soft-deleted exercises")

	// Confirm the row still exists in the table — soft delete, not hard delete.
	var ent Entity
	require.NoError(t, db.Where("id = ?", m.Id()).First(&ent).Error)
	require.NotNil(t, ent.DeletedAt)
}

// _ keeps json import alive when only entity inserts use it via tag-side serialization.
var _ = json.RawMessage(nil)
