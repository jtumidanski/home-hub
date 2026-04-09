package exercise

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity is the GORM mapping for `workout.exercises`.
//
// `secondary_region_ids` is stored as a `jsonb` array of UUID strings rather
// than a postgres native uuid[]. The `@>` containment operator on jsonb gives
// a clean way to filter "exercises whose secondary regions include X" from
// the list endpoint, and matches the storage decision documented in
// data-model.md.
type Entity struct {
	Id                     uuid.UUID       `gorm:"type:uuid;primaryKey"`
	TenantId               uuid.UUID       `gorm:"type:uuid;not null;index:idx_workout_exercise_tenant_user_deleted,priority:1;index:idx_workout_exercise_tenant_user_theme,priority:1;index:idx_workout_exercise_tenant_user_region,priority:1"`
	UserId                 uuid.UUID       `gorm:"type:uuid;not null;index:idx_workout_exercise_tenant_user_deleted,priority:2;index:idx_workout_exercise_tenant_user_theme,priority:2;index:idx_workout_exercise_tenant_user_region,priority:2"`
	Name                   string          `gorm:"type:varchar(100);not null"`
	Kind                   string          `gorm:"type:varchar(16);not null"`
	WeightType             string          `gorm:"type:varchar(16);not null;default:'free'"`
	ThemeId                uuid.UUID       `gorm:"type:uuid;not null;index:idx_workout_exercise_tenant_user_theme,priority:3"`
	RegionId               uuid.UUID       `gorm:"type:uuid;not null;index:idx_workout_exercise_tenant_user_region,priority:3"`
	SecondaryRegionIds     json.RawMessage `gorm:"type:jsonb;not null;default:'[]'"`
	DefaultSets            *int            `gorm:""`
	DefaultReps            *int            `gorm:""`
	DefaultWeight          *float64        `gorm:"type:numeric(7,2)"`
	DefaultWeightUnit      *string         `gorm:"type:varchar(4)"`
	DefaultDurationSeconds *int            `gorm:""`
	DefaultDistance        *float64        `gorm:"type:numeric(8,3)"`
	DefaultDistanceUnit    *string         `gorm:"type:varchar(4)"`
	Notes                  *string         `gorm:"type:varchar(500)"`
	CreatedAt              time.Time       `gorm:"not null"`
	UpdatedAt              time.Time       `gorm:"not null"`
	DeletedAt              *time.Time      `gorm:"index:idx_workout_exercise_tenant_user_deleted,priority:3"`
}

func (Entity) TableName() string { return "exercises" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	return db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_workout_exercises_tenant_user_name_active
		ON exercises (tenant_id, user_id, name) WHERE deleted_at IS NULL`).Error
}

func (m Model) ToEntity() Entity {
	secondary, _ := json.Marshal(uuidsToStrings(m.secondaryRegionIDs))
	return Entity{
		Id:                     m.id,
		TenantId:               m.tenantID,
		UserId:                 m.userID,
		Name:                   m.name,
		Kind:                   m.kind,
		WeightType:             m.weightType,
		ThemeId:                m.themeID,
		RegionId:               m.regionID,
		SecondaryRegionIds:     secondary,
		DefaultSets:            m.defaultSets,
		DefaultReps:            m.defaultReps,
		DefaultWeight:          m.defaultWeight,
		DefaultWeightUnit:      m.defaultWeightUnit,
		DefaultDurationSeconds: m.defaultDurationSeconds,
		DefaultDistance:        m.defaultDistance,
		DefaultDistanceUnit:    m.defaultDistanceUnit,
		Notes:                  m.notes,
		CreatedAt:              m.createdAt,
		UpdatedAt:              m.updatedAt,
		DeletedAt:              m.deletedAt,
	}
}

func Make(e Entity) (Model, error) {
	var secondary []uuid.UUID
	if len(e.SecondaryRegionIds) > 0 && string(e.SecondaryRegionIds) != "null" {
		var raw []string
		if err := json.Unmarshal(e.SecondaryRegionIds, &raw); err != nil {
			return Model{}, err
		}
		secondary = make([]uuid.UUID, 0, len(raw))
		for _, s := range raw {
			id, err := uuid.Parse(s)
			if err != nil {
				return Model{}, err
			}
			secondary = append(secondary, id)
		}
	}

	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetUserID(e.UserId).
		SetName(e.Name).
		SetKind(e.Kind).
		SetWeightType(e.WeightType).
		SetThemeID(e.ThemeId).
		SetRegionID(e.RegionId).
		SetSecondaryRegionIDs(secondary).
		SetDefaultSets(e.DefaultSets).
		SetDefaultReps(e.DefaultReps).
		SetDefaultWeight(e.DefaultWeight).
		SetDefaultWeightUnit(e.DefaultWeightUnit).
		SetDefaultDurationSeconds(e.DefaultDurationSeconds).
		SetDefaultDistance(e.DefaultDistance).
		SetDefaultDistanceUnit(e.DefaultDistanceUnit).
		SetNotes(e.Notes).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		SetDeletedAt(e.DeletedAt).
		Build()
}

func uuidsToStrings(ids []uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}
