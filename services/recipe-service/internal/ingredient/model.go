package ingredient

import (
	"time"

	"github.com/google/uuid"
)

type UnitFamily string

const (
	UnitFamilyCount  UnitFamily = "count"
	UnitFamilyWeight UnitFamily = "weight"
	UnitFamilyVolume UnitFamily = "volume"
)

func ValidUnitFamily(s string) bool {
	switch UnitFamily(s) {
	case UnitFamilyCount, UnitFamilyWeight, UnitFamilyVolume:
		return true
	case "":
		return true
	default:
		return false
	}
}

type Alias struct {
	id   uuid.UUID
	name string
}

func (a Alias) Id() uuid.UUID { return a.id }
func (a Alias) Name() string  { return a.name }

type Model struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	name        string
	displayName string
	unitFamily  string
	aliases     []Alias
	aliasCount  int
	usageCount  int
	createdAt   time.Time
	updatedAt   time.Time
}

func (m Model) Id() uuid.UUID      { return m.id }
func (m Model) TenantID() uuid.UUID { return m.tenantID }
func (m Model) Name() string        { return m.name }
func (m Model) DisplayName() string  { return m.displayName }
func (m Model) UnitFamily() string   { return m.unitFamily }
func (m Model) Aliases() []Alias     { return m.aliases }
func (m Model) AliasCount() int      { return m.aliasCount }
func (m Model) UsageCount() int      { return m.usageCount }
func (m Model) CreatedAt() time.Time { return m.createdAt }
func (m Model) UpdatedAt() time.Time { return m.updatedAt }

func (m Model) WithUsageCount(count int) Model {
	m.usageCount = count
	return m
}
