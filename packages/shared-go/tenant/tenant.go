package tenant

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type Model struct {
	id uuid.UUID
}

func (m *Model) Id() uuid.UUID {
	return m.id
}

func (m *Model) MarshalJSON() ([]byte, error) {
	type alias Model
	return json.Marshal(&struct {
		Id           uuid.UUID `json:"id"`
		Region       string    `json:"region"`
		MajorVersion uint16    `json:"majorVersion"`
		MinorVersion uint16    `json:"minorVersion"`
	}{
		Id: m.id,
	})
}

func (m *Model) UnmarshalJSON(data []byte) error {
	t := &struct {
		Id           uuid.UUID `json:"id"`
		Region       string    `json:"region"`
		MajorVersion uint16    `json:"majorVersion"`
		MinorVersion uint16    `json:"minorVersion"`
	}{}

	if err := json.Unmarshal(data, t); err != nil {
		return err
	}

	m.id = t.Id
	return nil
}

func (m *Model) Is(tenant Model) bool {
	if tenant.Id() != m.Id() {
		return false
	}
	return true
}

func (m *Model) String() string {
	return fmt.Sprintf("Id [%s]", m.Id().String())
}
