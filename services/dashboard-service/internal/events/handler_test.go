package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/dashboard-service/internal/dashboard"
	sharedDB "github.com/jtumidanski/home-hub/shared/go/database"
	sharedevents "github.com/jtumidanski/home-hub/shared/go/events"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	l, _ := test.NewNullLogger()
	sharedDB.RegisterTenantCallbacks(l, db)
	require.NoError(t, db.AutoMigrate(&dashboard.Entity{}))
	return db
}

func seedUserDashboard(t *testing.T, db *gorm.DB, tid, hid, uid uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	now := time.Now().UTC()
	e := dashboard.Entity{
		Id:            id,
		TenantId:      tid,
		HouseholdId:   hid,
		UserId:        &uid,
		Name:          "scratch",
		SortOrder:     0,
		Layout:        datatypes.JSON([]byte(`{"version":1,"widgets":[]}`)),
		SchemaVersion: 1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	require.NoError(t, db.Create(&e).Error)
	return id
}

func newHandler(t *testing.T, db *gorm.DB) *Handler {
	t.Helper()
	l, _ := test.NewNullLogger()
	return NewHandler(l, db)
}

func TestHandleUserDeletedCascades(t *testing.T) {
	db := setupTestDB(t)
	tid, uid, other := uuid.New(), uuid.New(), uuid.New()
	hid := uuid.New()
	seedUserDashboard(t, db, tid, hid, uid)
	seedUserDashboard(t, db, tid, hid, other)

	h := newHandler(t, db)

	evt := sharedevents.UserDeletedEvent{TenantID: tid, UserID: uid, DeletedAt: time.Now().UTC()}
	env, err := sharedevents.NewEnvelope(sharedevents.TypeUserDeleted, evt)
	require.NoError(t, err)
	raw, err := json.Marshal(env)
	require.NoError(t, err)

	require.NoError(t, h.Dispatch(context.Background(), kafka.Message{Value: raw}))

	var nUser, nOther int64
	require.NoError(t, db.Model(&dashboard.Entity{}).Where("user_id = ?", uid).Count(&nUser).Error)
	require.NoError(t, db.Model(&dashboard.Entity{}).Where("user_id = ?", other).Count(&nOther).Error)
	require.Equal(t, int64(0), nUser, "expected cascade to delete the target user's rows")
	require.Equal(t, int64(1), nOther, "unrelated user's rows must survive")

	// Re-dispatch is idempotent — no error, still zero rows for uid.
	require.NoError(t, h.Dispatch(context.Background(), kafka.Message{Value: raw}))
	require.NoError(t, db.Model(&dashboard.Entity{}).Where("user_id = ?", uid).Count(&nUser).Error)
	require.Equal(t, int64(0), nUser)
}

func TestHandleMalformedEnvelopeIsSwallowed(t *testing.T) {
	db := setupTestDB(t)
	h := newHandler(t, db)

	require.NoError(t, h.Dispatch(context.Background(), kafka.Message{Value: []byte("not json")}))
}

func TestHandleUnknownEventTypeIsNoOp(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	seedUserDashboard(t, db, tid, hid, uid)

	h := newHandler(t, db)

	env := sharedevents.Envelope{Type: "SOMETHING_UNRELATED", Version: 1, Payload: json.RawMessage(`{}`)}
	raw, err := json.Marshal(env)
	require.NoError(t, err)
	require.NoError(t, h.Dispatch(context.Background(), kafka.Message{Value: raw}))

	var n int64
	require.NoError(t, db.Model(&dashboard.Entity{}).Where("user_id = ?", uid).Count(&n).Error)
	require.Equal(t, int64(1), n, "unknown event type must not touch rows")
}

func TestHandleMalformedPayloadIsSwallowed(t *testing.T) {
	db := setupTestDB(t)
	tid, hid, uid := uuid.New(), uuid.New(), uuid.New()
	seedUserDashboard(t, db, tid, hid, uid)

	h := newHandler(t, db)

	// Valid envelope, unparseable payload.
	env := sharedevents.Envelope{Type: sharedevents.TypeUserDeleted, Version: 1, Payload: json.RawMessage(`"not-an-object"`)}
	raw, err := json.Marshal(env)
	require.NoError(t, err)
	require.NoError(t, h.Dispatch(context.Background(), kafka.Message{Value: raw}))

	var n int64
	require.NoError(t, db.Model(&dashboard.Entity{}).Where("user_id = ?", uid).Count(&n).Error)
	require.Equal(t, int64(1), n)
}
