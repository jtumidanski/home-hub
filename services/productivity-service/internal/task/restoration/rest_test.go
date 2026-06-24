package restoration

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/stretchr/testify/require"
)

// TestCreateRequest_UnmarshalRelationship guards the regression where the
// frontend sends the task id as a JSON:API to-one relationship but the backend
// only read it from an attribute, leaving TaskId empty and causing a 400. The
// body mirrors what services/api/productivity.ts restoreTask sends.
func TestCreateRequest_UnmarshalRelationship(t *testing.T) {
	taskID := uuid.New()
	body := fmt.Sprintf(`{
		"data": {
			"type": "task-restorations",
			"relationships": {
				"task": { "data": { "type": "tasks", "id": %q } }
			}
		}
	}`, taskID.String())

	var input CreateRequest
	require.NoError(t, jsonapi.Unmarshal([]byte(body), &input))
	require.Equal(t, taskID.String(), input.TaskId)
}
