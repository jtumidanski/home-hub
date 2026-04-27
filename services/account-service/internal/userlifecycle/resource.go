// Package userlifecycle exposes a service-to-service endpoint that
// account-service callers can invoke when a user is hard-deleted. The
// endpoint removes the user's local per-user preferences and emits a
// UserDeletedEvent on the shared Kafka bus so other services can cascade.
package userlifecycle

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/services/account-service/internal/householdpreference"
	sharedevents "github.com/jtumidanski/home-hub/shared/go/events"
	"github.com/jtumidanski/home-hub/shared/go/server"
	sharedtenant "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Producer is the subset of kafka/producer.Producer this package depends on.
// Tests supply a stub; production wiring passes *producer.Producer.
type Producer interface {
	Produce(ctx context.Context, topic string, key, value []byte, headers map[string]string) error
}

type Config struct {
	Topic         string
	InternalToken string
}

// InitializeRoutes mounts the internal user-deleted endpoint on the supplied
// router. The endpoint deliberately sits outside the /api/v1 JWT subrouter:
// the caller is another service, authenticated via the shared internal token.
//
// Because the caller is a service (not a user), there is no tenant in the
// incoming JWT. The tenant must be supplied explicitly via X-Tenant-ID so the
// handler can scope the row deletion. Missing or malformed header → 400.
func InitializeRoutes(db *gorm.DB, prod Producer, cfg Config) func(l logrus.FieldLogger, router *mux.Router) {
	return func(l logrus.FieldLogger, router *mux.Router) {
		router.HandleFunc("/internal/users/{id}/deleted", deletedHandler(l, db, prod, cfg)).Methods(http.MethodPost)
	}
}

func deletedHandler(l logrus.FieldLogger, db *gorm.DB, prod Producer, cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.InternalToken == "" || r.Header.Get("X-Internal-Token") != cfg.InternalToken {
			server.WriteError(w, http.StatusUnauthorized, "Unauthorized", "")
			return
		}
		userIDStr := mux.Vars(r)["id"]
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			server.WriteError(w, http.StatusBadRequest, "Bad Request", "invalid user id")
			return
		}
		tenantIDStr := r.Header.Get("X-Tenant-ID")
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			server.WriteError(w, http.StatusBadRequest, "Bad Request", "X-Tenant-ID header required")
			return
		}

		// Inject the tenant into the request context so GORM's tenant-callback
		// filters scope the DELETE, matching every other handler in this service.
		t := sharedtenant.New(tenantID, uuid.Nil, uuid.Nil)
		ctx := sharedtenant.WithContext(r.Context(), t)

		if err := db.WithContext(ctx).
			Where("tenant_id = ? AND user_id = ?", tenantID, userID).
			Delete(&householdpreference.Entity{}).Error; err != nil {
			l.WithError(err).Error("failed to delete household_preferences")
			server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
			return
		}

		evt := sharedevents.UserDeletedEvent{TenantID: tenantID, UserID: userID, DeletedAt: time.Now().UTC()}
		env, err := sharedevents.NewEnvelope(sharedevents.TypeUserDeleted, evt)
		if err != nil {
			l.WithError(err).Warn("envelope build")
		} else if prod != nil {
			payload, _ := json.Marshal(env)
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, uint64(userID.ID()))
			if err := prod.Produce(ctx, cfg.Topic, key, payload, nil); err != nil {
				l.WithError(err).Warn("produce UserDeletedEvent failed; event dropped")
			}
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
