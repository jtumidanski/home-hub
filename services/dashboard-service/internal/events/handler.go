// Package events implements the Kafka consumer dispatch for dashboard-service.
//
// The handler decodes a shared events.Envelope and, for UserDeletedEvent,
// hard-deletes every user-scoped dashboard belonging to the (tenant, user)
// pair. The operation is naturally idempotent — a second delete with the same
// predicate matches zero rows.
//
// Malformed envelopes and unknown event types are logged and swallowed so the
// consumer can commit the offset and move on; returning an error would cause
// the Kafka manager to re-read the same offset forever.
package events

import (
	"context"
	"encoding/json"

	"github.com/jtumidanski/home-hub/services/dashboard-service/internal/dashboard"
	sharedevents "github.com/jtumidanski/home-hub/shared/go/events"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Handler struct {
	l  logrus.FieldLogger
	db *gorm.DB
}

func NewHandler(l logrus.FieldLogger, db *gorm.DB) *Handler {
	return &Handler{l: l, db: db}
}

func (h *Handler) Dispatch(ctx context.Context, msg kafka.Message) error {
	var env sharedevents.Envelope
	if err := json.Unmarshal(msg.Value, &env); err != nil {
		h.l.WithError(err).Warn("skipping malformed envelope")
		return nil
	}
	switch env.Type {
	case sharedevents.TypeUserDeleted:
		var evt sharedevents.UserDeletedEvent
		if err := json.Unmarshal(env.Payload, &evt); err != nil {
			h.l.WithError(err).Warn("skipping malformed UserDeletedEvent")
			return nil
		}
		res := h.db.WithContext(ctx).
			Where("tenant_id = ? AND user_id = ?", evt.TenantID, evt.UserID).
			Delete(&dashboard.Entity{})
		if res.Error != nil {
			return res.Error
		}
		h.l.WithField("tenant_id", evt.TenantID).WithField("user_id", evt.UserID).
			WithField("rows", res.RowsAffected).Info("user cascade")
	default:
		// ignore unknown types
	}
	return nil
}
