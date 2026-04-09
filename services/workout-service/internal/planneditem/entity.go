package planneditem

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	Id                     uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId               uuid.UUID `gorm:"type:uuid;not null;index:idx_workout_planned_tenant_user_exercise,priority:1"`
	UserId                 uuid.UUID `gorm:"type:uuid;not null;index:idx_workout_planned_tenant_user_exercise,priority:2"`
	WeekId                 uuid.UUID `gorm:"type:uuid;not null;index:idx_workout_planned_week_day_position,priority:1"`
	ExerciseId             uuid.UUID `gorm:"type:uuid;not null;index:idx_workout_planned_tenant_user_exercise,priority:3"`
	DayOfWeek              int       `gorm:"not null;index:idx_workout_planned_week_day_position,priority:2"`
	Position               int       `gorm:"not null;index:idx_workout_planned_week_day_position,priority:3"`
	PlannedSets            *int      `gorm:""`
	PlannedReps            *int      `gorm:""`
	PlannedWeight          *float64  `gorm:"type:numeric(7,2)"`
	PlannedWeightUnit      *string   `gorm:"type:varchar(4)"`
	PlannedDurationSeconds *int      `gorm:""`
	PlannedDistance        *float64  `gorm:"type:numeric(8,3)"`
	PlannedDistanceUnit    *string   `gorm:"type:varchar(4)"`
	Notes                  *string   `gorm:"type:varchar(500)"`
	CreatedAt              time.Time `gorm:"not null"`
	UpdatedAt              time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "planned_items" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}); err != nil {
		return err
	}
	// FK constraints with explicit cascade/restrict semantics. AutoMigrate's
	// inferred FKs default to NO ACTION, which would block week deletion. We
	// also enforce ON DELETE RESTRICT for exercise_id so a hard exercise
	// delete cannot orphan history; only soft deletes are exposed in the API.
	stmts := []string{
		`ALTER TABLE planned_items DROP CONSTRAINT IF EXISTS planned_items_week_id_fkey`,
		`ALTER TABLE planned_items ADD CONSTRAINT planned_items_week_id_fkey
			FOREIGN KEY (week_id) REFERENCES weeks(id) ON DELETE CASCADE`,
		`ALTER TABLE planned_items DROP CONSTRAINT IF EXISTS planned_items_exercise_id_fkey`,
		`ALTER TABLE planned_items ADD CONSTRAINT planned_items_exercise_id_fkey
			FOREIGN KEY (exercise_id) REFERENCES exercises(id) ON DELETE RESTRICT`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}
	return nil
}

func (m Model) ToEntity() Entity {
	return Entity{
		Id:                     m.id,
		TenantId:               m.tenantID,
		UserId:                 m.userID,
		WeekId:                 m.weekID,
		ExerciseId:             m.exerciseID,
		DayOfWeek:              m.dayOfWeek,
		Position:               m.position,
		PlannedSets:            m.plannedSets,
		PlannedReps:            m.plannedReps,
		PlannedWeight:          m.plannedWeight,
		PlannedWeightUnit:      m.plannedWeightUnit,
		PlannedDurationSeconds: m.plannedDurationSeconds,
		PlannedDistance:        m.plannedDistance,
		PlannedDistanceUnit:    m.plannedDistanceUnit,
		Notes:                  m.notes,
		CreatedAt:              m.createdAt,
		UpdatedAt:              m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetUserID(e.UserId).
		SetWeekID(e.WeekId).
		SetExerciseID(e.ExerciseId).
		SetDayOfWeek(e.DayOfWeek).
		SetPosition(e.Position).
		SetPlannedSets(e.PlannedSets).
		SetPlannedReps(e.PlannedReps).
		SetPlannedWeight(e.PlannedWeight).
		SetPlannedWeightUnit(e.PlannedWeightUnit).
		SetPlannedDurationSeconds(e.PlannedDurationSeconds).
		SetPlannedDistance(e.PlannedDistance).
		SetPlannedDistanceUnit(e.PlannedDistanceUnit).
		SetNotes(e.Notes).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}
