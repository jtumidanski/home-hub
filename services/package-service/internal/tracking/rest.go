package tracking

import (
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/package-service/internal/carrier"
	"github.com/jtumidanski/home-hub/services/package-service/internal/trackingevent"
)

// RestModel is the JSON:API representation for a package in list responses.
type RestModel struct {
	Id                uuid.UUID  `json:"-"`
	TrackingNumber    *string    `json:"trackingNumber"`
	Carrier           string     `json:"carrier"`
	Label             *string    `json:"label"`
	Notes             *string    `json:"notes"`
	Status            *string    `json:"status"`
	Private           bool       `json:"private"`
	EstimatedDelivery *string    `json:"estimatedDelivery"`
	ActualDelivery    *time.Time `json:"actualDelivery"`
	LastPolledAt      *time.Time `json:"lastPolledAt"`
	ArchivedAt        *time.Time `json:"archivedAt"`
	IsOwner           bool       `json:"isOwner"`
	UserID            uuid.UUID  `json:"userId"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

func (r RestModel) GetName() string       { return "packages" }
func (r RestModel) GetID() string          { return r.Id.String() }
func (r *RestModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

// RestSummaryModel is the JSON:API representation for the package summary.
type RestSummaryModel struct {
	ArrivingTodayCount int64 `json:"arrivingTodayCount"`
	InTransitCount     int64 `json:"inTransitCount"`
	ExceptionCount     int64 `json:"exceptionCount"`
}

func (r RestSummaryModel) GetName() string      { return "packageSummaries" }
func (r RestSummaryModel) GetID() string         { return "summary" }
func (r *RestSummaryModel) SetID(_ string) error { return nil }

// RestTrackingEventModel is the JSON:API representation for a tracking event.
type RestTrackingEventModel struct {
	Id          uuid.UUID `json:"-"`
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Location    *string   `json:"location"`
	RawStatus   *string   `json:"rawStatus"`
}

func (r RestTrackingEventModel) GetName() string       { return "trackingEvents" }
func (r RestTrackingEventModel) GetID() string          { return r.Id.String() }
func (r *RestTrackingEventModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

// RestDetailModel is the JSON:API representation for a single package with tracking events.
type RestDetailModel struct {
	Id                uuid.UUID               `json:"-"`
	TrackingNumber    *string                 `json:"trackingNumber"`
	Carrier           string                  `json:"carrier"`
	Label             *string                 `json:"label"`
	Notes             *string                 `json:"notes"`
	Status            *string                 `json:"status"`
	Private           bool                    `json:"private"`
	EstimatedDelivery *string                 `json:"estimatedDelivery"`
	ActualDelivery    *time.Time              `json:"actualDelivery"`
	LastPolledAt      *time.Time              `json:"lastPolledAt"`
	ArchivedAt        *time.Time              `json:"archivedAt"`
	IsOwner           bool                    `json:"isOwner"`
	UserID            uuid.UUID               `json:"userId"`
	TrackingEvents    []RestTrackingEventModel `json:"trackingEvents"`
	CreatedAt         time.Time               `json:"createdAt"`
	UpdatedAt         time.Time               `json:"updatedAt"`
}

func (r RestDetailModel) GetName() string       { return "packages" }
func (r RestDetailModel) GetID() string          { return r.Id.String() }
func (r *RestDetailModel) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func TransformDetail(m Model, events []RestTrackingEventModel, requesterUserID uuid.UUID) (RestDetailModel, error) {
	rm, err := TransformWithPrivacy(m, requesterUserID)
	if err != nil {
		return RestDetailModel{}, err
	}
	evts := events
	if evts == nil {
		evts = []RestTrackingEventModel{}
	}
	return RestDetailModel{
		Id:                rm.Id,
		TrackingNumber:    rm.TrackingNumber,
		Carrier:           rm.Carrier,
		Label:             rm.Label,
		Notes:             rm.Notes,
		Status:            rm.Status,
		Private:           rm.Private,
		EstimatedDelivery: rm.EstimatedDelivery,
		ActualDelivery:    rm.ActualDelivery,
		LastPolledAt:      rm.LastPolledAt,
		ArchivedAt:        rm.ArchivedAt,
		IsOwner:           rm.IsOwner,
		UserID:            rm.UserID,
		TrackingEvents:    evts,
		CreatedAt:         rm.CreatedAt,
		UpdatedAt:         rm.UpdatedAt,
	}, nil
}

func TransformTrackingEvent(m trackingevent.Model) RestTrackingEventModel {
	return RestTrackingEventModel{
		Id:          m.Id(),
		Timestamp:   m.Timestamp(),
		Status:      m.Status(),
		Description: m.Description(),
		Location:    m.Location(),
		RawStatus:   m.RawStatus(),
	}
}

func TransformTrackingEventSlice(models []trackingevent.Model) []RestTrackingEventModel {
	result := make([]RestTrackingEventModel, len(models))
	for i, m := range models {
		result[i] = TransformTrackingEvent(m)
	}
	return result
}

// CreateRequest is the JSON:API request body for creating a package.
type CreateRequest struct {
	Id             uuid.UUID `json:"-"`
	TrackingNumber string    `json:"trackingNumber"`
	Carrier        string    `json:"carrier"`
	Label          string    `json:"label,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	Private        bool      `json:"private"`
}

func (r CreateRequest) GetName() string       { return "packages" }
func (r CreateRequest) GetID() string          { return r.Id.String() }
func (r *CreateRequest) SetID(id string) error {
	if id == "" {
		return nil
	}
	var err error
	r.Id, err = uuid.Parse(id)
	return err
}

// UpdateRequest is the JSON:API request body for updating a package.
type UpdateRequest struct {
	Id      uuid.UUID `json:"-"`
	Label   *string   `json:"label,omitempty"`
	Notes   *string   `json:"notes,omitempty"`
	Carrier *string   `json:"carrier,omitempty"`
	Private *bool     `json:"private,omitempty"`
}

func (r UpdateRequest) GetName() string       { return "packages" }
func (r UpdateRequest) GetID() string          { return r.Id.String() }
func (r *UpdateRequest) SetID(id string) error { var err error; r.Id, err = uuid.Parse(id); return err }

func Transform(m Model) (RestModel, error) {
	tn := m.TrackingNumber()
	s := m.Status()

	rm := RestModel{
		Id:             m.Id(),
		TrackingNumber: &tn,
		Carrier:        m.Carrier(),
		Label:          m.Label(),
		Notes:          m.Notes(),
		Status:         &s,
		Private:        m.Private(),
		ActualDelivery: m.ActualDelivery(),
		LastPolledAt:   m.LastPolledAt(),
		ArchivedAt:     m.ArchivedAt(),
		UserID:         m.UserID(),
		CreatedAt:      m.CreatedAt(),
		UpdatedAt:      m.UpdatedAt(),
	}

	if m.EstimatedDelivery() != nil {
		ed := m.EstimatedDelivery().Format("2006-01-02")
		rm.EstimatedDelivery = &ed
	}

	return rm, nil
}

func TransformSlice(models []Model) ([]RestModel, error) {
	result := make([]RestModel, len(models))
	for i, m := range models {
		rm, err := Transform(m)
		if err != nil {
			return nil, err
		}
		result[i] = rm
	}
	return result, nil
}

func TransformWithPrivacy(m Model, requesterUserID uuid.UUID) (RestModel, error) {
	isOwner := m.UserID() == requesterUserID

	rm := RestModel{
		Id:        m.Id(),
		Carrier:   m.Carrier(),
		Private:   m.Private(),
		IsOwner:   isOwner,
		UserID:    m.UserID(),
		CreatedAt: m.CreatedAt(),
		UpdatedAt: m.UpdatedAt(),
	}

	if m.EstimatedDelivery() != nil {
		s := m.EstimatedDelivery().Format("2006-01-02")
		rm.EstimatedDelivery = &s
	}
	rm.ActualDelivery = m.ActualDelivery()
	rm.ArchivedAt = m.ArchivedAt()

	if m.IsPrivate() && !isOwner {
		placeholder := "Package"
		rm.Label = &placeholder
		rm.TrackingNumber = nil
		rm.Notes = nil
		rm.Status = nil
		rm.LastPolledAt = nil
	} else {
		tn := m.TrackingNumber()
		rm.TrackingNumber = &tn
		rm.Label = m.Label()
		rm.Notes = m.Notes()
		s := m.Status()
		rm.Status = &s
		rm.LastPolledAt = m.LastPolledAt()
	}

	return rm, nil
}

func TransformDetection(result carrier.DetectionResult) carrier.RestDetectionModel {
	return carrier.RestDetectionModel{
		TrackingNumber:  result.TrackingNumber,
		DetectedCarrier: result.DetectedCarrier,
		Confidence:      result.Confidence,
	}
}

func TransformSliceWithPrivacy(models []Model, requesterUserID uuid.UUID) ([]RestModel, error) {
	result := make([]RestModel, len(models))
	for i, m := range models {
		rm, err := TransformWithPrivacy(m, requesterUserID)
		if err != nil {
			return nil, err
		}
		result[i] = rm
	}
	return result, nil
}
