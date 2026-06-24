package dismissal

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/stretchr/testify/require"
)

// TestCreateRequest_UnmarshalRelationship guards the regression where the
// frontend sends the reminder id as a JSON:API to-one relationship but the
// backend only read it from an attribute, leaving ReminderId empty and causing
// a 400. The body mirrors what services/api/productivity.ts dismissReminder sends.
func TestCreateRequest_UnmarshalRelationship(t *testing.T) {
	reminderID := uuid.New()
	body := fmt.Sprintf(`{
		"data": {
			"type": "reminder-dismissals",
			"relationships": {
				"reminder": { "data": { "type": "reminders", "id": %q } }
			}
		}
	}`, reminderID.String())

	var input CreateRequest
	require.NoError(t, jsonapi.Unmarshal([]byte(body), &input))
	require.Equal(t, reminderID.String(), input.ReminderId)
}
