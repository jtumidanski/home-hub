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

func TestPrivacyMaskingOwnerSeesFullDetails(t *testing.T) {
	ownerID := uuid.New()
	m := testModel(ownerID, "private")

	rest, err := TransformWithPrivacy(m, ownerID)
	if err != nil {
		t.Fatal(err)
	}

	if rest.Title != "Secret Meeting" {
		t.Fatalf("owner should see full title, got %q", rest.Title)
	}
	if rest.Description == nil || *rest.Description != "Very confidential details" {
		t.Fatal("owner should see full description")
	}
	if rest.Location == nil || *rest.Location != "Room 101" {
		t.Fatal("owner should see full location")
	}
	if !rest.IsOwner {
		t.Fatal("isOwner should be true for owner")
	}
}

func TestPrivacyMaskingNonOwnerSeesBusy(t *testing.T) {
	ownerID := uuid.New()
	requesterID := uuid.New()
	m := testModel(ownerID, "private")

	rest, err := TransformWithPrivacy(m, requesterID)
	if err != nil {
		t.Fatal(err)
	}

	if rest.Title != "Busy" {
		t.Fatalf("non-owner should see 'Busy', got %q", rest.Title)
	}
	if rest.Description != nil {
		t.Fatal("non-owner should not see description for private event")
	}
	if rest.Location != nil {
		t.Fatal("non-owner should not see location for private event")
	}
	if rest.IsOwner {
		t.Fatal("isOwner should be false for non-owner")
	}
}

func TestPrivacyMaskingConfidentialSameAsBusy(t *testing.T) {
	ownerID := uuid.New()
	requesterID := uuid.New()
	m := testModel(ownerID, "confidential")

	rest, err := TransformWithPrivacy(m, requesterID)
	if err != nil {
		t.Fatal(err)
	}

	if rest.Title != "Busy" {
		t.Fatalf("non-owner should see 'Busy' for confidential, got %q", rest.Title)
	}
}

func TestPublicEventVisibleToAll(t *testing.T) {
	ownerID := uuid.New()
	requesterID := uuid.New()
	m := testModel(ownerID, "default")

	rest, err := TransformWithPrivacy(m, requesterID)
	if err != nil {
		t.Fatal(err)
	}

	if rest.Title != "Secret Meeting" {
		t.Fatalf("non-owner should see full title for default visibility, got %q", rest.Title)
	}
}

func TestUserDisplayNameAndColorAlwaysVisible(t *testing.T) {
	ownerID := uuid.New()
	requesterID := uuid.New()
	m := testModel(ownerID, "private")

	rest, err := TransformWithPrivacy(m, requesterID)
	if err != nil {
		t.Fatal(err)
	}

	if rest.UserDisplayName != "Jane Doe" {
		t.Fatalf("userDisplayName should always be visible, got %q", rest.UserDisplayName)
	}
	if rest.UserColor != "#4285F4" {
		t.Fatalf("userColor should always be visible, got %q", rest.UserColor)
	}
}
