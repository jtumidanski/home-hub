package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserDeletedEventRoundtrip(t *testing.T) {
	in := UserDeletedEvent{
		TenantID:  uuid.New(),
		UserID:    uuid.New(),
		DeletedAt: time.Now().UTC().Truncate(time.Second),
	}
	env, err := NewEnvelope(TypeUserDeleted, in)
	if err != nil {
		t.Fatal(err)
	}
	if env.Type != TypeUserDeleted {
		t.Fatalf("type: got %s", env.Type)
	}
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	var got Envelope
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	var out UserDeletedEvent
	if err := json.Unmarshal(got.Payload, &out); err != nil {
		t.Fatal(err)
	}
	if out.UserID != in.UserID || out.TenantID != in.TenantID {
		t.Fatalf("ids drifted")
	}
}
