package wishlist

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	database "github.com/jtumidanski/home-hub/shared/go/database"
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
	// SQLite in-memory has very limited write concurrency. Pin the pool to a
	// single connection so concurrent goroutines serialize at the driver
	// level instead of failing with "database table is locked". The atomic
	// UPDATE itself is what we are testing — not the database backend's
	// concurrency model.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	l, _ := test.NewNullLogger()
	database.RegisterTenantCallbacks(l, db)
	require.NoError(t, db.AutoMigrate(&Entity{}))
	return db
}

func testContext(tenantID, householdID, userID uuid.UUID) context.Context {
	t := tenantctx.New(tenantID, householdID, userID)
	return tenantctx.WithContext(context.Background(), t)
}

func newTestProcessor(t *testing.T, db *gorm.DB, ctx context.Context) *Processor {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewProcessor(l, ctx, db)
}

func TestProcessor_Create_Defaults(t *testing.T) {
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID, householdID, userID)
	p := newTestProcessor(t, db, ctx)

	m, err := p.Create(tenantID, householdID, userID, CreateInput{Name: "Espresso machine"})
	require.NoError(t, err)
	assert.Equal(t, "Espresso machine", m.Name())
	assert.Equal(t, UrgencyWant, m.Urgency())
	assert.Equal(t, 0, m.VoteCount())
	assert.Equal(t, householdID, m.HouseholdID())
	assert.Equal(t, userID, m.CreatedBy())
	assert.Nil(t, m.PurchaseLocation())
}

func TestProcessor_Create_TrimAndValidate(t *testing.T) {
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID, householdID, userID)
	p := newTestProcessor(t, db, ctx)

	_, err := p.Create(tenantID, householdID, userID, CreateInput{Name: "   "})
	assert.ErrorIs(t, err, ErrNameRequired)

	bad := "nope"
	_, err = p.Create(tenantID, householdID, userID, CreateInput{Name: "X", Urgency: &bad})
	assert.ErrorIs(t, err, ErrInvalidUrgency)
}

func TestProcessor_List_SortedByVotes(t *testing.T) {
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID, householdID, userID)
	p := newTestProcessor(t, db, ctx)

	a, err := p.Create(tenantID, householdID, userID, CreateInput{Name: "A"})
	require.NoError(t, err)
	time.Sleep(2 * time.Millisecond)
	b, err := p.Create(tenantID, householdID, userID, CreateInput{Name: "B"})
	require.NoError(t, err)
	time.Sleep(2 * time.Millisecond)
	c, err := p.Create(tenantID, householdID, userID, CreateInput{Name: "C"})
	require.NoError(t, err)

	// Vote: B gets 2, A gets 1, C gets 0.
	_, err = p.Vote(householdID, b.Id())
	require.NoError(t, err)
	_, err = p.Vote(householdID, b.Id())
	require.NoError(t, err)
	_, err = p.Vote(householdID, a.Id())
	require.NoError(t, err)

	models, err := p.List(householdID)
	require.NoError(t, err)
	require.Len(t, models, 3)
	assert.Equal(t, b.Id(), models[0].Id(), "B has the most votes")
	assert.Equal(t, 2, models[0].VoteCount())
	assert.Equal(t, a.Id(), models[1].Id())
	assert.Equal(t, 1, models[1].VoteCount())
	assert.Equal(t, c.Id(), models[2].Id())
	assert.Equal(t, 0, models[2].VoteCount())
}

func TestProcessor_HouseholdIsolation(t *testing.T) {
	tenantID := uuid.New()
	hh1 := uuid.New()
	hh2 := uuid.New()
	userID := uuid.New()
	db := setupTestDB(t)

	p1 := newTestProcessor(t, db, testContext(tenantID, hh1, userID))
	p2 := newTestProcessor(t, db, testContext(tenantID, hh2, userID))

	created, err := p1.Create(tenantID, hh1, userID, CreateInput{Name: "HH1 only"})
	require.NoError(t, err)

	models, err := p2.List(hh2)
	require.NoError(t, err)
	assert.Empty(t, models, "household 2 must not see household 1's items")

	_, err = p2.Get(hh2, created.Id())
	assert.ErrorIs(t, err, ErrNotFound)

	_, err = p2.Vote(hh2, created.Id())
	assert.ErrorIs(t, err, ErrNotFound)

	err = p2.Delete(hh2, created.Id())
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestProcessor_Update_DropsVoteCount(t *testing.T) {
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID, householdID, userID)
	p := newTestProcessor(t, db, ctx)

	created, err := p.Create(tenantID, householdID, userID, CreateInput{Name: "Bike"})
	require.NoError(t, err)
	_, err = p.Vote(householdID, created.Id())
	require.NoError(t, err)
	_, err = p.Vote(householdID, created.Id())
	require.NoError(t, err)

	newName := "Mountain bike"
	urgent := UrgencyMustHave
	updated, err := p.Update(householdID, created.Id(), UpdateInput{
		Name:    &newName,
		Urgency: &urgent,
	})
	require.NoError(t, err)
	assert.Equal(t, "Mountain bike", updated.Name())
	assert.Equal(t, UrgencyMustHave, updated.Urgency())
	assert.Equal(t, 2, updated.VoteCount(), "vote_count must be untouched by Update")
}

func TestProcessor_Delete(t *testing.T) {
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID, householdID, userID)
	p := newTestProcessor(t, db, ctx)

	created, err := p.Create(tenantID, householdID, userID, CreateInput{Name: "Toy"})
	require.NoError(t, err)
	require.NoError(t, p.Delete(householdID, created.Id()))

	_, err = p.Get(householdID, created.Id())
	assert.ErrorIs(t, err, ErrNotFound)
	// Hard delete: re-deleting must report not found.
	assert.ErrorIs(t, p.Delete(householdID, created.Id()), ErrNotFound)
}

func TestProcessor_Vote_NotFound(t *testing.T) {
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID, householdID, userID)
	p := newTestProcessor(t, db, ctx)

	_, err := p.Vote(householdID, uuid.New())
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestProcessor_ConcurrentVotes(t *testing.T) {
	tenantID := uuid.New()
	householdID := uuid.New()
	userID := uuid.New()
	db := setupTestDB(t)
	ctx := testContext(tenantID, householdID, userID)
	p := newTestProcessor(t, db, ctx)

	created, err := p.Create(tenantID, householdID, userID, CreateInput{Name: "Hot item"})
	require.NoError(t, err)

	const N = 50
	var wg sync.WaitGroup
	errs := make(chan error, N)
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := p.Vote(householdID, created.Id())
			if err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("vote returned error: %v", err)
	}

	final, err := p.Get(householdID, created.Id())
	require.NoError(t, err)
	assert.Equal(t, N, final.VoteCount(), "all votes must be counted, no lost updates")
}
