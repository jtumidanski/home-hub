// Package events defines versioned envelopes for domain events crossing
// service boundaries on the shared Kafka bus. Each event type has a stable
// string tag carried on the envelope so consumers can ignore types they do
// not handle.
package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	TypeUserDeleted EventType = "USER_DELETED"
)

// Envelope is the wire format for every cross-service event.
type Envelope struct {
	Type    EventType       `json:"type"`
	Version int             `json:"version"`
	Payload json.RawMessage `json:"payload"`
}

func NewEnvelope(t EventType, payload any) (Envelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{Type: t, Version: 1, Payload: raw}, nil
}

// UserDeletedEvent is emitted by account-service when a user is hard-deleted
// from a tenant. Consumers remove any user-scoped rows they own.
type UserDeletedEvent struct {
	TenantID  uuid.UUID `json:"tenantId"`
	UserID    uuid.UUID `json:"userId"`
	DeletedAt time.Time `json:"deletedAt"`
}
