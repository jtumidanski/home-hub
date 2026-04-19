package weekview

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/exercise"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/performance"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/region"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/theme"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
	"github.com/jtumidanski/home-hub/shared/go/database"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupNearestHandler(t *testing.T) (*mux.Router, *gorm.DB) {
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

	router := mux.NewRouter()
	si := server.GetServerInformation()
	InitializeRoutes(db)(l, si, router)
	return router, db
}

func seedPopulated(t *testing.T, db *gorm.DB, tenantID, userID uuid.UUID, start time.Time) {
	t.Helper()
	now := time.Now().UTC()
	wkID := uuid.New()
	require.NoError(t, db.Create(&week.Entity{
		Id: wkID, TenantId: tenantID, UserId: userID,
		WeekStartDate: start, RestDayFlags: json.RawMessage("[]"),
		CreatedAt: now, UpdatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&planneditem.Entity{
		Id: uuid.New(), TenantId: tenantID, UserId: userID,
		WeekId: wkID, ExerciseId: uuid.New(),
		DayOfWeek: 0, Position: 0,
		CreatedAt: now, UpdatedAt: now,
	}).Error)
}

func requestNearest(t *testing.T, router *mux.Router, tenantID, userID uuid.UUID, ref, direction string) *httptest.ResponseRecorder {
	t.Helper()
	url := "/workouts/weeks/nearest"
	params := ""
	if ref != "" || direction != "" {
		params = "?"
		if ref != "" {
			params += "reference=" + ref
		}
		if direction != "" {
			if ref != "" {
				params += "&"
			}
			params += "direction=" + direction
		}
	}
	req := httptest.NewRequest(http.MethodGet, url+params, nil)
	ctx := tenantctx.WithContext(req.Context(), tenantctx.New(tenantID, uuid.Nil, userID))
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestNearest_Success_Prev(t *testing.T) {
	router, db := setupNearestHandler(t)
	tenantID, userID := uuid.New(), uuid.New()
	seedPopulated(t, db, tenantID, userID, time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC))
	seedPopulated(t, db, tenantID, userID, time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC))

	w := requestNearest(t, router, tenantID, userID, "2026-04-13", "prev")
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	assert.Equal(t, "workoutWeekPointer", data["type"])
	assert.Equal(t, "2026-03-30", data["id"])
	attrs := data["attributes"].(map[string]any)
	assert.Equal(t, "2026-03-30", attrs["weekStartDate"])
}

func TestNearest_Success_Next(t *testing.T) {
	router, db := setupNearestHandler(t)
	tenantID, userID := uuid.New(), uuid.New()
	seedPopulated(t, db, tenantID, userID, time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC))

	w := requestNearest(t, router, tenantID, userID, "2026-04-06", "next")
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	assert.Equal(t, "2026-04-13", data["id"])
}

func TestNearest_NormalizesReferenceToMonday(t *testing.T) {
	router, db := setupNearestHandler(t)
	tenantID, userID := uuid.New(), uuid.New()
	seedPopulated(t, db, tenantID, userID, time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC))

	// Wednesday (2026-04-08) normalizes to Monday 2026-04-06.
	w := requestNearest(t, router, tenantID, userID, "2026-04-08", "prev")
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	data := body["data"].(map[string]any)
	assert.Equal(t, "2026-03-30", data["id"])
}

func TestNearest_404WhenNoMatch(t *testing.T) {
	router, _ := setupNearestHandler(t)
	tenantID, userID := uuid.New(), uuid.New()
	w := requestNearest(t, router, tenantID, userID, "2026-04-06", "prev")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestNearest_400MissingReference(t *testing.T) {
	router, _ := setupNearestHandler(t)
	w := requestNearest(t, router, uuid.New(), uuid.New(), "", "prev")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNearest_400InvalidReference(t *testing.T) {
	router, _ := setupNearestHandler(t)
	w := requestNearest(t, router, uuid.New(), uuid.New(), "not-a-date", "prev")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNearest_400BadDirection(t *testing.T) {
	router, _ := setupNearestHandler(t)
	w := requestNearest(t, router, uuid.New(), uuid.New(), "2026-04-06", "sideways")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNearest_CrossUserIsolation(t *testing.T) {
	router, db := setupNearestHandler(t)
	tenantID := uuid.New()
	userA, userB := uuid.New(), uuid.New()
	seedPopulated(t, db, tenantID, userA, time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC))

	// userB asks "next" from 2026-04-06 — userA's populated week must not leak.
	w := requestNearest(t, router, tenantID, userB, "2026-04-06", "next")
	assert.Equal(t, http.StatusNotFound, w.Code, "userB must not see userA's populated week")
}
