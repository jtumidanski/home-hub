package performance

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/exercise"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Table-driven coverage of the §4.4.1 status state machine. The processor
// implements the transitions in two helpers (`applyExplicitStatus` for
// explicit requests, `deriveStatusFromActuals` for the auto path); we test
// both directly so the rules don't have to round-trip through a real DB.
func TestApplyExplicitStatus(t *testing.T) {
	cases := []struct {
		name       string
		prev       string
		requested  string
		hasActuals bool
		want       string
	}{
		{"pending->done", StatusPending, StatusDone, false, StatusDone},
		{"pending->skipped", StatusPending, StatusSkipped, false, StatusSkipped},
		{"partial->done", StatusPartial, StatusDone, true, StatusDone},
		{"partial->skipped clears actuals", StatusPartial, StatusSkipped, true, StatusSkipped},
		{"done->skipped", StatusDone, StatusSkipped, true, StatusSkipped},
		{"done->pending with actuals -> partial", StatusDone, StatusPending, true, StatusPartial},
		{"done->pending without actuals -> pending", StatusDone, StatusPending, false, StatusPending},
		{"skipped->pending unskip", StatusSkipped, StatusPending, false, StatusPending},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := applyExplicitStatus(tc.prev, tc.requested, tc.hasActuals)
			if got != tc.want {
				t.Fatalf("applyExplicitStatus(%s,%s,%v) = %s, want %s", tc.prev, tc.requested, tc.hasActuals, got, tc.want)
			}
		})
	}
}

func TestDeriveStatusFromActuals(t *testing.T) {
	cases := []struct {
		prev string
		want string
	}{
		{StatusPending, StatusPartial}, // pending → partial
		{StatusSkipped, StatusPartial}, // skip is cleared when actuals arrive
		{StatusPartial, StatusPartial}, // stays
		{StatusDone, StatusDone},       // user is correcting, not retracting
	}
	for _, tc := range cases {
		got := deriveStatusFromActuals(tc.prev)
		if got != tc.want {
			t.Fatalf("deriveStatusFromActuals(%s) = %s, want %s", tc.prev, got, tc.want)
		}
	}
}

func TestHasActuals(t *testing.T) {
	v := 3
	in := PatchInput{ActualSets: &v}
	if !in.hasActuals() {
		t.Fatal("expected hasActuals true with sets set")
	}
	if (PatchInput{}).hasActuals() {
		t.Fatal("expected hasActuals false on empty input")
	}
}

// --- DB-backed reject-path tests -----------------------------------------
//
// These exercise the §5.1 / §5.2 guardrails through the real processor and
// SQLite. The harness side-steps the per-package Migration helpers because
// they emit Postgres-only DDL.

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	require.NoError(t, db.AutoMigrate(
		&Entity{},
		&SetEntity{},
		&planneditem.Entity{},
		&exercise.Entity{},
	))
	return db
}

func newProcessor(t *testing.T, db *gorm.DB) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, context.Background(), db)
}

// seedItem inserts an exercise + planned item pair so the performance
// processor has a parent to attach to. The exercise kind defaults to strength
// (the only kind allowed in per-set mode); cardio/isometric tests pass an
// override.
func seedItem(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID, kind string) (exerciseID, plannedItemID uuid.UUID) {
	t.Helper()
	exerciseID = uuid.New()
	require.NoError(t, db.Create(&exercise.Entity{
		Id:                 exerciseID,
		TenantId:           tenantID,
		UserId:             userID,
		Name:               "Bench " + exerciseID.String()[:8],
		Kind:               kind,
		WeightType:         exercise.WeightTypeFree,
		ThemeId:            uuid.New(),
		RegionId:           uuid.New(),
		SecondaryRegionIds: json.RawMessage("[]"),
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}).Error)
	plannedItemID = uuid.New()
	require.NoError(t, db.Create(&planneditem.Entity{
		Id:         plannedItemID,
		TenantId:   tenantID,
		UserId:     userID,
		WeekId:     uuid.New(),
		ExerciseId: exerciseID,
		DayOfWeek:  0,
		Position:   0,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}).Error)
	return exerciseID, plannedItemID
}

func TestProcessor_ReplaceSets_RejectsNonStrength(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()
	_, itemID := seedItem(t, db, tenantID, userID, exercise.KindCardio)

	_, _, err := p.ReplaceSets(tenantID, userID, itemID, "lb", []SetInput{{Reps: 5, Weight: 100}})
	assert.ErrorIs(t, err, ErrPerSetNotAllowed, "cardio exercises must reject per-set logging with §5.2 422")
}

func TestProcessor_Patch_RejectsSummaryWritesInPerSetMode(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()
	_, itemID := seedItem(t, db, tenantID, userID, exercise.KindStrength)

	// Switch the performance into per-set mode by writing a single set row.
	_, _, err := p.ReplaceSets(tenantID, userID, itemID, "lb", []SetInput{{Reps: 5, Weight: 100}})
	require.NoError(t, err)

	// A subsequent summary actuals write must be rejected with the §5.1 409.
	sets := 3
	_, _, err = p.Patch(tenantID, userID, itemID, PatchInput{ActualSets: &sets})
	assert.ErrorIs(t, err, ErrSummaryWhilePerSet)
}

func TestProcessor_Patch_RejectsWeightUnitChangeWhilePerSetRowsExist(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()
	_, itemID := seedItem(t, db, tenantID, userID, exercise.KindStrength)

	_, _, err := p.ReplaceSets(tenantID, userID, itemID, "lb", []SetInput{{Reps: 5, Weight: 100}})
	require.NoError(t, err)

	// Per-set rows exist; weightUnit must be locked until the user collapses.
	kg := "kg"
	_, _, err = p.Patch(tenantID, userID, itemID, PatchInput{WeightUnit: &kg})
	assert.ErrorIs(t, err, ErrUnitChangeWithSets)
}

func TestProcessor_Patch_RejectsCrossUserPlannedItem(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID := uuid.New()
	userA, userB := uuid.New(), uuid.New()
	_, itemID := seedItem(t, db, tenantID, userA, exercise.KindStrength)

	sets := 3
	_, _, err := p.Patch(tenantID, userB, itemID, PatchInput{ActualSets: &sets})
	assert.ErrorIs(t, err, ErrPlannedItemNotFound, "cross-user access must surface as 404")
}

// seedStrengthPlanned seeds a strength item with explicit planned values so
// the backfill tests can assert the auto-copy.
func seedStrengthPlanned(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID, sets, reps int, weight float64, unit string) uuid.UUID {
	t.Helper()
	exerciseID := uuid.New()
	require.NoError(t, db.Create(&exercise.Entity{
		Id:                 exerciseID,
		TenantId:           tenantID,
		UserId:             userID,
		Name:               "Bench " + exerciseID.String()[:8],
		Kind:               exercise.KindStrength,
		WeightType:         exercise.WeightTypeFree,
		ThemeId:            uuid.New(),
		RegionId:           uuid.New(),
		SecondaryRegionIds: json.RawMessage("[]"),
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}).Error)
	itemID := uuid.New()
	u := unit
	require.NoError(t, db.Create(&planneditem.Entity{
		Id:                itemID,
		TenantId:          tenantID,
		UserId:            userID,
		WeekId:            uuid.New(),
		ExerciseId:        exerciseID,
		DayOfWeek:         0,
		Position:          0,
		PlannedSets:       &sets,
		PlannedReps:       &reps,
		PlannedWeight:     &weight,
		PlannedWeightUnit: &u,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}).Error)
	return itemID
}

// Bare "done" with no actuals must backfill planned → actual so downstream
// consumers (Review page, dashboards) see authoritative numbers instead of a
// status-only row. This is the bug fix for the "?x?" display on Review.
func TestProcessor_Patch_BareDone_BackfillsActualsFromPlanned(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()
	itemID := seedStrengthPlanned(t, db, tenantID, userID, 3, 8, 135, "lb")

	done := StatusDone
	m, _, err := p.Patch(tenantID, userID, itemID, PatchInput{Status: &done})
	require.NoError(t, err)

	assert.Equal(t, StatusDone, m.Status())
	require.NotNil(t, m.ActualSets())
	assert.Equal(t, 3, *m.ActualSets())
	require.NotNil(t, m.ActualReps())
	assert.Equal(t, 8, *m.ActualReps())
	require.NotNil(t, m.ActualWeight())
	assert.Equal(t, 135.0, *m.ActualWeight())
	require.NotNil(t, m.WeightUnit())
	assert.Equal(t, "lb", *m.WeightUnit())
}

// When the client supplies some actuals alongside `done`, the user-supplied
// values win; only the remaining nil fields are filled from planned.
func TestProcessor_Patch_DoneWithPartialActuals_PreservesUserValues(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()
	itemID := seedStrengthPlanned(t, db, tenantID, userID, 3, 8, 135, "lb")

	done := StatusDone
	reps := 10 // user overrides reps; sets/weight fall back to planned
	m, _, err := p.Patch(tenantID, userID, itemID, PatchInput{Status: &done, ActualReps: &reps})
	require.NoError(t, err)

	require.NotNil(t, m.ActualReps())
	assert.Equal(t, 10, *m.ActualReps(), "user-supplied reps must not be overwritten by planned")
	require.NotNil(t, m.ActualSets())
	assert.Equal(t, 3, *m.ActualSets())
	require.NotNil(t, m.ActualWeight())
	assert.Equal(t, 135.0, *m.ActualWeight())
}

// The backfill only fires in summary mode. A per-set performance with a bare
// `done` patch keeps its set rows as the source of truth — no summary actuals
// get written.
func TestProcessor_Patch_BareDone_NoBackfillInPerSetMode(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()
	itemID := seedStrengthPlanned(t, db, tenantID, userID, 3, 8, 135, "lb")

	_, _, err := p.ReplaceSets(tenantID, userID, itemID, "lb", []SetInput{{Reps: 5, Weight: 100}})
	require.NoError(t, err)

	done := StatusDone
	m, _, err := p.Patch(tenantID, userID, itemID, PatchInput{Status: &done})
	require.NoError(t, err)

	assert.Equal(t, ModePerSet, m.Mode())
	assert.Nil(t, m.ActualSets(), "per-set mode must not gain summary actuals from backfill")
	assert.Nil(t, m.ActualReps())
	assert.Nil(t, m.ActualWeight())
}

func TestProcessor_CollapseSets_RoundTripsPerSetThroughCollapse(t *testing.T) {
	db := setupTestDB(t)
	p := newProcessor(t, db)
	tenantID, userID := uuid.New(), uuid.New()
	_, itemID := seedItem(t, db, tenantID, userID, exercise.KindStrength)

	// Three sets at varying reps/weight; collapse rule = count, max-reps, max-weight.
	_, _, err := p.ReplaceSets(tenantID, userID, itemID, "lb", []SetInput{
		{Reps: 8, Weight: 100},
		{Reps: 6, Weight: 110},
		{Reps: 4, Weight: 120},
	})
	require.NoError(t, err)

	m, err := p.CollapseSets(userID, itemID)
	require.NoError(t, err)
	assert.Equal(t, ModeSummary, m.Mode())
	require.NotNil(t, m.ActualSets())
	assert.Equal(t, 3, *m.ActualSets(), "collapse must use the per-set count for sets")
	require.NotNil(t, m.ActualReps())
	assert.Equal(t, 8, *m.ActualReps(), "collapse must use max reps")
	require.NotNil(t, m.ActualWeight())
	assert.Equal(t, 120.0, *m.ActualWeight(), "collapse must use max weight")
}
