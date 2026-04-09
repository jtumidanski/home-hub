// Package workout_service_integration verifies the multi-tenancy contract
// across every workout-service domain. The shared database callback
// automatically injects `WHERE tenant_id = ?` from request context, so the
// real test is "do the processors propagate context correctly so the filter
// fires." We assert that against an in-memory SQLite database with the
// callback registered.
package internal_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/exercise"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/performance"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/region"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/theme"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB stands up an in-memory SQLite with the auto-tenant-filter
// callback registered against every domain entity. We bypass the per-package
// Migration helpers because they emit Postgres-only DDL.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	require.NoError(t, db.AutoMigrate(
		&theme.Entity{},
		&region.Entity{},
		&exercise.Entity{},
		&week.Entity{},
		&planneditem.Entity{},
		&performance.Entity{},
		&performance.SetEntity{},
	))
	return db
}

// withTenant returns a context bound to the given tenant + user. Both the
// real HTTP layer and the integration test use this exact wiring so the
// callback behaviour observed here matches production.
func withTenant(tenantID, userID uuid.UUID) context.Context {
	return tenant.WithContext(context.Background(), tenant.New(tenantID, uuid.New(), userID))
}

// seedFullStack creates one of every workout entity owned by the supplied
// tenant + user. Returns the IDs so the assertion phase can attempt to read
// them as a different tenant.
type seedIDs struct {
	themeID, regionID, exerciseID, weekID, plannedItemID, performanceID uuid.UUID
}

func seedFullStack(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID) seedIDs {
	t.Helper()
	now := time.Now().UTC()
	ids := seedIDs{
		themeID:        uuid.New(),
		regionID:       uuid.New(),
		exerciseID:     uuid.New(),
		weekID:         uuid.New(),
		plannedItemID:  uuid.New(),
		performanceID: uuid.New(),
	}
	require.NoError(t, db.Create(&theme.Entity{
		Id: ids.themeID, TenantId: tenantID, UserId: userID,
		Name: "Strength " + ids.themeID.String()[:8], SortOrder: 0,
		CreatedAt: now, UpdatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&region.Entity{
		Id: ids.regionID, TenantId: tenantID, UserId: userID,
		Name: "Chest " + ids.regionID.String()[:8], SortOrder: 0,
		CreatedAt: now, UpdatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&exercise.Entity{
		Id: ids.exerciseID, TenantId: tenantID, UserId: userID,
		Name: "Bench " + ids.exerciseID.String()[:8],
		Kind: exercise.KindStrength, WeightType: exercise.WeightTypeFree,
		ThemeId: ids.themeID, RegionId: ids.regionID,
		SecondaryRegionIds: json.RawMessage("[]"),
		CreatedAt:          now, UpdatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&week.Entity{
		Id: ids.weekID, TenantId: tenantID, UserId: userID,
		WeekStartDate: time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
		RestDayFlags:  json.RawMessage("[]"),
		CreatedAt:     now, UpdatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&planneditem.Entity{
		Id: ids.plannedItemID, TenantId: tenantID, UserId: userID,
		WeekId: ids.weekID, ExerciseId: ids.exerciseID,
		DayOfWeek: 0, Position: 0,
		CreatedAt: now, UpdatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&performance.Entity{
		Id: ids.performanceID, TenantId: tenantID, UserId: userID,
		PlannedItemId: ids.plannedItemID,
		Status:        performance.StatusPending,
		Mode:          performance.ModeSummary,
		CreatedAt:     now, UpdatedAt: now,
	}).Error)
	return ids
}

func TestCrossTenantIsolation_EveryDomainBlocksOtherTenantReads(t *testing.T) {
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()

	tenantA, userA := uuid.New(), uuid.New()
	tenantB, userB := uuid.New(), uuid.New()
	idsA := seedFullStack(t, db, tenantA, userA)
	_ = seedFullStack(t, db, tenantB, userB) // ensure tenantB has its own rows

	// Build processors bound to tenantB's context. Each lookup of a tenantA
	// resource id must miss because the callback adds `WHERE tenant_id = ?
	// = tenantB`.
	ctxB := withTenant(tenantB, userB)

	t.Run("theme", func(t *testing.T) {
		p := theme.NewProcessor(l, ctxB, db)
		_, err := p.Get(idsA.themeID)
		assert.ErrorIs(t, err, theme.ErrNotFound, "tenantB must not see tenantA's theme")
	})

	t.Run("region", func(t *testing.T) {
		p := region.NewProcessor(l, ctxB, db)
		_, err := p.Get(idsA.regionID)
		assert.ErrorIs(t, err, region.ErrNotFound, "tenantB must not see tenantA's region")
	})

	t.Run("exercise", func(t *testing.T) {
		p := exercise.NewProcessor(l, ctxB, db)
		_, err := p.Get(idsA.exerciseID)
		assert.ErrorIs(t, err, exercise.ErrNotFound, "tenantB must not see tenantA's exercise")
	})

	t.Run("week", func(t *testing.T) {
		// Week.Get takes (userID, weekStart). The callback still scopes to
		// tenantB, so even passing the right userId from tenantA returns 404.
		p := week.NewProcessor(l, ctxB, db)
		_, err := p.Get(userA, time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC))
		assert.ErrorIs(t, err, week.ErrNotFound, "tenantB must not load tenantA's week")
	})

	t.Run("planneditem and performance", func(t *testing.T) {
		// Performance Patch first looks up the planned item; with tenantB's
		// callback firing, the lookup fails and the processor surfaces
		// ErrPlannedItemNotFound — proving that planned-item access is also
		// tenant-scoped.
		p := performance.NewProcessor(l, ctxB, db)
		sets := 3
		_, _, err := p.Patch(tenantB, userA, idsA.plannedItemID, performance.PatchInput{ActualSets: &sets})
		assert.ErrorIs(t, err, performance.ErrPlannedItemNotFound, "tenantB must not see tenantA's planned item")
	})
}

func TestCrossUserIsolation_ListEndpointsHonorUserScope(t *testing.T) {
	// Within a single tenant the explicit user_id filter on list endpoints
	// (theme.GetAllByUser, region.GetAllByUser, exercise.ListByUser) keeps
	// each user's catalog visible only to that user.
	db := setupTestDB(t)
	l, _ := test.NewNullLogger()

	tenantID := uuid.New()
	userA, userB := uuid.New(), uuid.New()
	idsA := seedFullStack(t, db, tenantID, userA)
	_ = seedFullStack(t, db, tenantID, userB)

	ctx := withTenant(tenantID, userA)

	t.Run("theme.List excludes other-user rows", func(t *testing.T) {
		// theme.List EnsureSeeds default themes for the asking user, so we
		// expect at least 2 (Muscle, Cardio) plus the one we explicitly
		// inserted. Crucially, every row's user_id must equal userA.
		p := theme.NewProcessor(l, ctx, db)
		out, err := p.List(tenantID, userA)
		require.NoError(t, err)
		require.NotEmpty(t, out)
		for _, m := range out {
			assert.Equal(t, userA, m.UserID(), "theme.List must only return rows owned by the calling user")
		}
		// Sanity check: the seeded tenantA-userA theme is present.
		found := false
		for _, m := range out {
			if m.Id() == idsA.themeID {
				found = true
				break
			}
		}
		assert.True(t, found, "userA should see the theme they own")
	})

	t.Run("exercise.List excludes other-user rows", func(t *testing.T) {
		p := exercise.NewProcessor(l, ctx, db)
		out, err := p.List(userA, nil, nil)
		require.NoError(t, err)
		for _, m := range out {
			assert.Equal(t, userA, m.UserID(), "exercise.List must only return rows owned by the calling user")
		}
	})
}
