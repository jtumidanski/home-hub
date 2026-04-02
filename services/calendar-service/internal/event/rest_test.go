package event

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func testModel(userID uuid.UUID, visibility string) Model {
	return Model{
		id:              uuid.New(),
		tenantID:        uuid.New(),
		householdID:     uuid.New(),
		connectionID:    uuid.New(),
		sourceID:        uuid.New(),
		userID:          userID,
		externalID:      "ext-123",
		title:           "Secret Meeting",
		description:     "Very confidential details",
		startTime:       time.Now(),
		endTime:         time.Now().Add(time.Hour),
		location:        "Room 101",
		visibility:      visibility,
		userDisplayName: "Jane Doe",
		userColor:       "#4285F4",
	}
}

func TestTransformWithPrivacy(t *testing.T) {
	ownerID := uuid.New()
	otherID := uuid.New()

	tests := []struct {
		name        string
		ownerID     uuid.UUID
		requesterID uuid.UUID
		visibility  string
		validate    func(t *testing.T, r RestModel)
	}{
		{
			"owner sees full details on private event",
			ownerID,
			ownerID,
			"private",
			func(t *testing.T, r RestModel) {
				if r.Title != "Secret Meeting" {
					t.Fatalf("owner should see full title, got %q", r.Title)
				}
				if r.Description == nil || *r.Description != "Very confidential details" {
					t.Fatal("owner should see full description")
				}
				if r.Location == nil || *r.Location != "Room 101" {
					t.Fatal("owner should see full location")
				}
				if !r.IsOwner {
					t.Fatal("isOwner should be true for owner")
				}
			},
		},
		{
			"non-owner sees busy on private event",
			ownerID,
			otherID,
			"private",
			func(t *testing.T, r RestModel) {
				if r.Title != "Busy" {
					t.Fatalf("non-owner should see 'Busy', got %q", r.Title)
				}
				if r.Description != nil {
					t.Fatal("non-owner should not see description for private event")
				}
				if r.Location != nil {
					t.Fatal("non-owner should not see location for private event")
				}
				if r.IsOwner {
					t.Fatal("isOwner should be false for non-owner")
				}
			},
		},
		{
			"non-owner sees busy on confidential event",
			ownerID,
			otherID,
			"confidential",
			func(t *testing.T, r RestModel) {
				if r.Title != "Busy" {
					t.Fatalf("non-owner should see 'Busy' for confidential, got %q", r.Title)
				}
			},
		},
		{
			"public event visible to all",
			ownerID,
			otherID,
			"default",
			func(t *testing.T, r RestModel) {
				if r.Title != "Secret Meeting" {
					t.Fatalf("non-owner should see full title for default visibility, got %q", r.Title)
				}
			},
		},
		{
			"display name and color always visible",
			ownerID,
			otherID,
			"private",
			func(t *testing.T, r RestModel) {
				if r.UserDisplayName != "Jane Doe" {
					t.Fatalf("userDisplayName should always be visible, got %q", r.UserDisplayName)
				}
				if r.UserColor != "#4285F4" {
					t.Fatalf("userColor should always be visible, got %q", r.UserColor)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := testModel(tc.ownerID, tc.visibility)
			rest, err := TransformWithPrivacy(m, tc.requesterID)
			if err != nil {
				t.Fatal(err)
			}
			tc.validate(t, rest)
		})
	}
}
