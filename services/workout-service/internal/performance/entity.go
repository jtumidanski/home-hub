package performance

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity is the GORM mapping for `workout.performances`. 1:1 with planned_items.
type Entity struct {
	Id                     uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId               uuid.UUID `gorm:"type:uuid;not null"`
	UserId                 uuid.UUID `gorm:"type:uuid;not null"`
	PlannedItemId          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_workout_performance_planned"`
	Status                 string    `gorm:"type:varchar(16);not null;default:'pending'"`
	Mode                   string    `gorm:"type:varchar(16);not null;default:'summary'"`
	WeightUnit             *string   `gorm:"type:varchar(4)"`
	ActualSets             *int      `gorm:""`
	ActualReps             *int      `gorm:""`
	ActualWeight           *float64  `gorm:"type:numeric(7,2)"`
	ActualDurationSeconds  *int      `gorm:""`
	ActualDistance         *float64  `gorm:"type:numeric(8,3)"`
	ActualDistanceUnit     *string   `gorm:"type:varchar(4)"`
	Notes                  *string   `gorm:"type:varchar(500)"`
	CreatedAt              time.Time `gorm:"not null"`
	UpdatedAt              time.Time `gorm:"not null"`
}

func (Entity) TableName() string { return "performances" }

// SetEntity is the GORM mapping for `workout.performance_sets`.
type SetEntity struct {
	Id            uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId      uuid.UUID `gorm:"type:uuid;not null"`
	UserId        uuid.UUID `gorm:"type:uuid;not null"`
	PerformanceId uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_workout_perf_set_perf_setno,priority:1"`
	SetNumber     int       `gorm:"not null;uniqueIndex:idx_workout_perf_set_perf_setno,priority:2"`
	Reps          int       `gorm:"not null"`
	Weight        float64   `gorm:"type:numeric(7,2);not null"`
	CreatedAt     time.Time `gorm:"not null"`
}

func (SetEntity) TableName() string { return "performance_sets" }

func Migration(db *gorm.DB) error {
	if err := db.AutoMigrate(&Entity{}, &SetEntity{}); err != nil {
		return err
	}
	stmts := []string{
		`ALTER TABLE performances DROP CONSTRAINT IF EXISTS performances_planned_item_id_fkey`,
		`ALTER TABLE performances ADD CONSTRAINT performances_planned_item_id_fkey
			FOREIGN KEY (planned_item_id) REFERENCES planned_items(id) ON DELETE CASCADE`,
		`ALTER TABLE performance_sets DROP CONSTRAINT IF EXISTS performance_sets_performance_id_fkey`,
		`ALTER TABLE performance_sets ADD CONSTRAINT performance_sets_performance_id_fkey
			FOREIGN KEY (performance_id) REFERENCES performances(id) ON DELETE CASCADE`,
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
		Id:                    m.id,
		TenantId:              m.tenantID,
		UserId:                m.userID,
		PlannedItemId:         m.plannedItemID,
		Status:                m.status,
		Mode:                  m.mode,
		WeightUnit:            m.weightUnit,
		ActualSets:            m.actualSets,
		ActualReps:            m.actualReps,
		ActualWeight:          m.actualWeight,
		ActualDurationSeconds: m.actualDurationSeconds,
		ActualDistance:        m.actualDistance,
		ActualDistanceUnit:    m.actualDistanceUnit,
		Notes:                 m.notes,
		CreatedAt:             m.createdAt,
		UpdatedAt:             m.updatedAt,
	}
}

func Make(e Entity) (Model, error) {
	return NewBuilder().
		SetId(e.Id).
		SetTenantID(e.TenantId).
		SetUserID(e.UserId).
		SetPlannedItemID(e.PlannedItemId).
		SetStatus(e.Status).
		SetMode(e.Mode).
		SetWeightUnit(e.WeightUnit).
		SetActualSets(e.ActualSets).
		SetActualReps(e.ActualReps).
		SetActualWeight(e.ActualWeight).
		SetActualDurationSeconds(e.ActualDurationSeconds).
		SetActualDistance(e.ActualDistance).
		SetActualDistanceUnit(e.ActualDistanceUnit).
		SetNotes(e.Notes).
		SetCreatedAt(e.CreatedAt).
		SetUpdatedAt(e.UpdatedAt).
		Build()
}

// MakeSet projects a row into a SetModel.
func MakeSet(e SetEntity) SetModel {
	return SetModel{
		id:            e.Id,
		tenantID:      e.TenantId,
		userID:        e.UserId,
		performanceID: e.PerformanceId,
		setNumber:     e.SetNumber,
		reps:          e.Reps,
		weight:        e.Weight,
		createdAt:     e.CreatedAt,
	}
}
