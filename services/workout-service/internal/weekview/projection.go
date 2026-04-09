package weekview

import (
	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/exercise"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/performance"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
	"gorm.io/gorm"
)

// AssembleItems builds the []ItemRest projection for a week. It performs three
// reads: every planned item for the week, every referenced exercise (including
// soft-deleted ones so historical names still resolve), and every performance
// row associated with the items. The reads live in the weekview package
// because the embedded shape spans multiple domain packages.
func AssembleItems(db *gorm.DB, weekID uuid.UUID) ([]ItemRest, error) {
	rows, err := planneditem.GetByWeek(weekID)(db)()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []ItemRest{}, nil
	}

	exerciseIDs := make([]uuid.UUID, 0, len(rows))
	itemIDs := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		exerciseIDs = append(exerciseIDs, r.ExerciseId)
		itemIDs = append(itemIDs, r.Id)
	}

	exMap, err := loadExerciseMap(db, exerciseIDs)
	if err != nil {
		return nil, err
	}

	perfMap, setMap, err := performance.LoadByPlannedItems(db, itemIDs)
	if err != nil {
		return nil, err
	}

	out := make([]ItemRest, 0, len(rows))
	for _, r := range rows {
		ex, ok := exMap[r.ExerciseId]
		if !ok {
			// FK guarantees this never happens, but stay defensive.
			continue
		}
		item := ItemRest{
			ID:              r.Id,
			DayOfWeek:       r.DayOfWeek,
			Position:        r.Position,
			ExerciseID:      ex.Id,
			ExerciseName:    ex.Name,
			ExerciseDeleted: ex.DeletedAt != nil,
			Kind:            ex.Kind,
			WeightType:      ex.WeightType,
			Planned: PlannedRest{
				Sets:            r.PlannedSets,
				Reps:            r.PlannedReps,
				Weight:          r.PlannedWeight,
				WeightUnit:      r.PlannedWeightUnit,
				DurationSeconds: r.PlannedDurationSeconds,
				Distance:        r.PlannedDistance,
				DistanceUnit:    r.PlannedDistanceUnit,
			},
			Notes: r.Notes,
		}
		if perf, ok := perfMap[r.Id]; ok {
			item.Performance = buildPerformanceRest(perf, setMap[perf.Id])
		}
		out = append(out, item)
	}
	return out, nil
}

// BuildDocument wraps a week model + its assembled items in the JSON:API
// envelope. The caller is responsible for marshaling.
func BuildDocument(m week.Model, items []ItemRest) Document {
	if items == nil {
		items = []ItemRest{}
	}
	return Document{
		Data: data{
			Type: "weeks",
			ID:   m.Id().String(),
			Attributes: attributes{
				WeekStartDate: m.WeekStartDate().Format("2006-01-02"),
				RestDayFlags:  m.RestDayFlags(),
				Items:         items,
			},
		},
	}
}

func loadExerciseMap(db *gorm.DB, ids []uuid.UUID) (map[uuid.UUID]exercise.Entity, error) {
	out := make(map[uuid.UUID]exercise.Entity, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	var rows []exercise.Entity
	// Include soft-deleted: planned items continue to display the original
	// exercise name even after the catalog row is removed.
	if err := db.Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.Id] = r
	}
	return out, nil
}

func buildPerformanceRest(p performance.Entity, sets []performance.SetEntity) *PerformanceRest {
	pr := &PerformanceRest{
		Status:     p.Status,
		Mode:       p.Mode,
		WeightUnit: p.WeightUnit,
		Notes:      p.Notes,
	}
	t := p.UpdatedAt
	pr.UpdatedAt = &t

	if p.Mode == performance.ModePerSet {
		out := make([]PerformanceSetRest, 0, len(sets))
		for _, s := range sets {
			out = append(out, PerformanceSetRest{
				SetNumber: s.SetNumber,
				Reps:      s.Reps,
				Weight:    s.Weight,
			})
		}
		pr.Sets = out
		return pr
	}
	pr.Actuals = &ActualsRest{
		Sets:            p.ActualSets,
		Reps:            p.ActualReps,
		Weight:          p.ActualWeight,
		DurationSeconds: p.ActualDurationSeconds,
		Distance:        p.ActualDistance,
		DistanceUnit:    p.ActualDistanceUnit,
	}
	return pr
}
