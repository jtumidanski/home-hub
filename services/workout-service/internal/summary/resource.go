// Package summary serves the GET /workouts/weeks/{weekStart}/summary endpoint
// — the per-week reporting projection that powers the weekly summary screen.
//
// The projection is computed on-demand from planned_items + performances +
// performance_sets + the joined exercise/theme/region rows. Nothing is
// persisted: this endpoint is a pure read.
package summary

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/exercise"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/performance"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/planneditem"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/region"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/theme"
	"github.com/jtumidanski/home-hub/services/workout-service/internal/week"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const weekDateLayout = "2006-01-02"

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/workouts/weeks/{weekStart}/summary", rh("GetWeekSummary", summaryHandler(db))).Methods(http.MethodGet)
	}
}

// --- response types -------------------------------------------------------

type quantity struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type cardioBlock struct {
	TotalDurationSeconds int      `json:"totalDurationSeconds"`
	TotalDistance        quantity `json:"totalDistance"`
}

type groupRest struct {
	ItemCount      int          `json:"itemCount"`
	StrengthVolume *quantity    `json:"strengthVolume"`
	Cardio         *cardioBlock `json:"cardio"`
}

type themeGroup struct {
	ThemeID   uuid.UUID `json:"themeId"`
	ThemeName string    `json:"themeName"`
	groupRest
}

type regionGroup struct {
	RegionID   uuid.UUID `json:"regionId"`
	RegionName string    `json:"regionName"`
	groupRest
}

type dayItem struct {
	ItemID        uuid.UUID `json:"itemId"`
	ExerciseName  string    `json:"exerciseName"`
	Status        string    `json:"status"`
	Planned       any       `json:"planned"`
	ActualSummary any       `json:"actualSummary"`
}

type dayBlock struct {
	DayOfWeek int       `json:"dayOfWeek"`
	IsRestDay bool      `json:"isRestDay"`
	Items     []dayItem `json:"items"`
}

type document struct {
	Data data `json:"data"`
}

type data struct {
	Type       string     `json:"type"`
	ID         string     `json:"id"`
	Attributes attributes `json:"attributes"`
}

type attributes struct {
	WeekStartDate       string        `json:"weekStartDate"`
	RestDayFlags        []int         `json:"restDayFlags"`
	TotalPlannedItems   int           `json:"totalPlannedItems"`
	TotalPerformedItems int           `json:"totalPerformedItems"`
	TotalSkippedItems   int           `json:"totalSkippedItems"`
	ByDay               []dayBlock    `json:"byDay"`
	ByTheme             []themeGroup  `json:"byTheme"`
	ByRegion            []regionGroup `json:"byRegion"`
}

// --- handler --------------------------------------------------------------

func summaryHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			ws, err := time.ParseInLocation(weekDateLayout, mux.Vars(r)["weekStart"], time.UTC)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid weekStart", "weekStart must be YYYY-MM-DD")
				return
			}

			weekProc := week.NewProcessor(d.Logger(), r.Context(), db)
			wkModel, err := weekProc.Get(t.UserId(), ws)
			if err != nil {
				if errors.Is(err, week.ErrNotFound) {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Week not found")
					return
				}
				d.Logger().WithError(err).Error("Failed to load week for summary")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			doc, err := buildSummary(weekProc.DB(), wkModel)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to build week summary")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(doc)
		}
	}
}

// --- projection -----------------------------------------------------------

func buildSummary(db *gorm.DB, wk week.Model) (document, error) {
	items, err := planneditem.GetByWeek(wk.Id())(db)()
	if err != nil {
		return document{}, err
	}

	exMap, themeMap, regionMap, err := loadCatalog(db, items)
	if err != nil {
		return document{}, err
	}
	itemIDs := make([]uuid.UUID, 0, len(items))
	for _, it := range items {
		itemIDs = append(itemIDs, it.Id)
	}
	perfMap, setMap, err := performance.LoadByPlannedItems(db, itemIDs)
	if err != nil {
		return document{}, err
	}

	// Tally weight + distance unit usage so we can pick the most-used unit
	// for the cross-group volume math (with the documented tie-breakers).
	weightUnit := selectWeightUnit(items, exMap, perfMap)
	distanceUnit := selectDistanceUnit(items, exMap, perfMap)

	totalsByTheme := make(map[uuid.UUID]*themeGroup)
	totalsByRegion := make(map[uuid.UUID]*regionGroup)
	byDay := make([][]dayItem, 7)
	totals := struct {
		planned   int
		performed int
		skipped   int
	}{}

	for _, it := range items {
		ex, ok := exMap[it.ExerciseId]
		if !ok {
			continue
		}
		perf, hasPerf := perfMap[it.Id]
		status := performance.StatusPending
		if hasPerf {
			status = perf.Status
		}

		totals.planned++
		switch status {
		case performance.StatusDone, performance.StatusPartial:
			totals.performed++
		case performance.StatusSkipped:
			totals.skipped++
		}

		// Per-day projection. `planned` and `actualSummary` are emitted as
		// kind-shaped maps so the frontend can render them without a per-kind
		// switch in the JSON.
		byDay[it.DayOfWeek] = append(byDay[it.DayOfWeek], dayItem{
			ItemID:        it.Id,
			ExerciseName:  ex.Name,
			Status:        status,
			Planned:       buildPlannedShape(it, ex.Kind),
			ActualSummary: buildActualSummary(perf, setMap[performanceID(perf, hasPerf)], ex.Kind),
		})

		// Theme tally.
		tg := ensureThemeGroup(totalsByTheme, themeMap, ex.ThemeId, weightUnit, distanceUnit)
		tg.ItemCount++
		applyVolume(tg, it, ex, perf, setMap, weightUnit, distanceUnit, hasPerf)

		// Region tally — primary region only, never secondary, per §7.
		rg := ensureRegionGroup(totalsByRegion, regionMap, ex.RegionId, weightUnit, distanceUnit)
		rg.ItemCount++
		applyVolumeToRegion(rg, it, ex, perf, setMap, weightUnit, distanceUnit, hasPerf)
	}

	// Materialize the day blocks (always 7 entries, ordered Mon..Sun).
	dayList := make([]dayBlock, 7)
	restSet := make(map[int]bool, len(wk.RestDayFlags()))
	for _, d := range wk.RestDayFlags() {
		restSet[d] = true
	}
	for d := 0; d < 7; d++ {
		dayList[d] = dayBlock{
			DayOfWeek: d,
			IsRestDay: restSet[d],
			Items:     byDay[d],
		}
		if dayList[d].Items == nil {
			dayList[d].Items = []dayItem{}
		}
	}

	themes := make([]themeGroup, 0, len(totalsByTheme))
	for _, g := range totalsByTheme {
		themes = append(themes, *g)
	}
	regions := make([]regionGroup, 0, len(totalsByRegion))
	for _, g := range totalsByRegion {
		regions = append(regions, *g)
	}

	doc := document{
		Data: data{
			Type: "week-summaries",
			ID:   wk.WeekStartDate().Format("2006-01-02"),
			Attributes: attributes{
				WeekStartDate:       wk.WeekStartDate().Format("2006-01-02"),
				RestDayFlags:        wk.RestDayFlags(),
				TotalPlannedItems:   totals.planned,
				TotalPerformedItems: totals.performed,
				TotalSkippedItems:   totals.skipped,
				ByDay:               dayList,
				ByTheme:             themes,
				ByRegion:            regions,
			},
		},
	}
	return doc, nil
}

// --- catalog & unit-selection helpers ------------------------------------

func loadCatalog(db *gorm.DB, items []planneditem.Entity) (map[uuid.UUID]exercise.Entity, map[uuid.UUID]theme.Entity, map[uuid.UUID]region.Entity, error) {
	exIDs := make([]uuid.UUID, 0, len(items))
	for _, it := range items {
		exIDs = append(exIDs, it.ExerciseId)
	}
	var exRows []exercise.Entity
	if len(exIDs) > 0 {
		// Include soft-deleted: historical totals must still resolve.
		if err := db.Where("id IN ?", exIDs).Find(&exRows).Error; err != nil {
			return nil, nil, nil, err
		}
	}
	exMap := make(map[uuid.UUID]exercise.Entity, len(exRows))
	themeIDs := make(map[uuid.UUID]struct{})
	regionIDs := make(map[uuid.UUID]struct{})
	for _, e := range exRows {
		exMap[e.Id] = e
		themeIDs[e.ThemeId] = struct{}{}
		regionIDs[e.RegionId] = struct{}{}
	}

	themeMap := make(map[uuid.UUID]theme.Entity)
	if len(themeIDs) > 0 {
		ids := keysOf(themeIDs)
		var rows []theme.Entity
		if err := db.Where("id IN ?", ids).Find(&rows).Error; err != nil {
			return nil, nil, nil, err
		}
		for _, r := range rows {
			themeMap[r.Id] = r
		}
	}
	regionMap := make(map[uuid.UUID]region.Entity)
	if len(regionIDs) > 0 {
		ids := keysOf(regionIDs)
		var rows []region.Entity
		if err := db.Where("id IN ?", ids).Find(&rows).Error; err != nil {
			return nil, nil, nil, err
		}
		for _, r := range rows {
			regionMap[r.Id] = r
		}
	}
	return exMap, themeMap, regionMap, nil
}

func keysOf(m map[uuid.UUID]struct{}) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// selectWeightUnit picks the user's most-used weight unit in the week. Ties
// break in favor of `lb` per §7. Strength items contribute their planned or
// actual unit; cardio items don't contribute.
func selectWeightUnit(items []planneditem.Entity, exMap map[uuid.UUID]exercise.Entity, perfMap map[uuid.UUID]performance.Entity) string {
	counts := map[string]int{"lb": 0, "kg": 0}
	for _, it := range items {
		ex, ok := exMap[it.ExerciseId]
		if !ok || ex.Kind != exercise.KindStrength {
			continue
		}
		if perf, ok := perfMap[it.Id]; ok && perf.WeightUnit != nil {
			counts[*perf.WeightUnit]++
			continue
		}
		if it.PlannedWeightUnit != nil {
			counts[*it.PlannedWeightUnit]++
		}
	}
	if counts["kg"] > counts["lb"] {
		return "kg"
	}
	return "lb"
}

// selectDistanceUnit picks the most-used cardio distance unit, tie-broken to
// `mi`. Only cardio items contribute.
func selectDistanceUnit(items []planneditem.Entity, exMap map[uuid.UUID]exercise.Entity, perfMap map[uuid.UUID]performance.Entity) string {
	counts := map[string]int{"mi": 0, "km": 0, "m": 0}
	for _, it := range items {
		ex, ok := exMap[it.ExerciseId]
		if !ok || ex.Kind != exercise.KindCardio {
			continue
		}
		if perf, ok := perfMap[it.Id]; ok && perf.ActualDistanceUnit != nil {
			counts[*perf.ActualDistanceUnit]++
			continue
		}
		if it.PlannedDistanceUnit != nil {
			counts[*it.PlannedDistanceUnit]++
		}
	}
	// Highest wins; tie → mi.
	best := "mi"
	bestCount := counts["mi"]
	for _, u := range []string{"km", "m"} {
		if counts[u] > bestCount {
			best = u
			bestCount = counts[u]
		}
	}
	return best
}

// --- volume math ----------------------------------------------------------

func ensureThemeGroup(m map[uuid.UUID]*themeGroup, themeMap map[uuid.UUID]theme.Entity, id uuid.UUID, wu, du string) *themeGroup {
	if g, ok := m[id]; ok {
		return g
	}
	name := ""
	if t, ok := themeMap[id]; ok {
		name = t.Name
	}
	g := &themeGroup{
		ThemeID:   id,
		ThemeName: name,
		groupRest: groupRest{},
	}
	m[id] = g
	return g
}

func ensureRegionGroup(m map[uuid.UUID]*regionGroup, regionMap map[uuid.UUID]region.Entity, id uuid.UUID, wu, du string) *regionGroup {
	if g, ok := m[id]; ok {
		return g
	}
	name := ""
	if r, ok := regionMap[id]; ok {
		name = r.Name
	}
	g := &regionGroup{
		RegionID:   id,
		RegionName: name,
		groupRest:  groupRest{},
	}
	m[id] = g
	return g
}

// applyVolume adds one item's contribution to a group's totals. Bodyweight
// strength and isometric items only contribute to itemCount; only `free`
// strength items contribute to strengthVolume.
func applyVolume(g *themeGroup, it planneditem.Entity, ex exercise.Entity, perf performance.Entity, setMap map[uuid.UUID][]performance.SetEntity, weightUnit, distanceUnit string, hasPerf bool) {
	addVolume(&g.groupRest, it, ex, perf, setMap, weightUnit, distanceUnit, hasPerf)
}

func applyVolumeToRegion(g *regionGroup, it planneditem.Entity, ex exercise.Entity, perf performance.Entity, setMap map[uuid.UUID][]performance.SetEntity, weightUnit, distanceUnit string, hasPerf bool) {
	addVolume(&g.groupRest, it, ex, perf, setMap, weightUnit, distanceUnit, hasPerf)
}

func addVolume(g *groupRest, it planneditem.Entity, ex exercise.Entity, perf performance.Entity, setMap map[uuid.UUID][]performance.SetEntity, weightUnit, distanceUnit string, hasPerf bool) {
	switch ex.Kind {
	case exercise.KindStrength:
		if ex.WeightType == exercise.WeightTypeBodyweight {
			return // count only
		}
		volume := strengthVolume(it, perf, setMap, hasPerf, weightUnit)
		if volume > 0 {
			if g.StrengthVolume == nil {
				g.StrengthVolume = &quantity{Unit: weightUnit}
			}
			g.StrengthVolume.Value += volume
		}
	case exercise.KindCardio:
		if g.Cardio == nil {
			g.Cardio = &cardioBlock{TotalDistance: quantity{Unit: distanceUnit}}
		}
		dur, dist, distUnit := cardioActuals(it, perf, hasPerf)
		g.Cardio.TotalDurationSeconds += dur
		g.Cardio.TotalDistance.Value += convertDistance(dist, distUnit, distanceUnit)
	}
}

// strengthVolume computes Σ sets × reps × weight for one item, after unit
// conversion. Per-set rows are summed individually; summary mode falls back
// to the planned values when no actuals were logged so the projection still
// reflects the user's intent for not-yet-completed weeks.
func strengthVolume(it planneditem.Entity, perf performance.Entity, setMap map[uuid.UUID][]performance.SetEntity, hasPerf bool, targetUnit string) float64 {
	if hasPerf && perf.Mode == performance.ModePerSet {
		var total float64
		unit := "lb"
		if perf.WeightUnit != nil {
			unit = *perf.WeightUnit
		}
		for _, s := range setMap[perf.Id] {
			total += float64(s.Reps) * convertWeight(s.Weight, unit, targetUnit)
		}
		return total
	}
	sets := derefInt(it.PlannedSets)
	reps := derefInt(it.PlannedReps)
	weight := derefFloat(it.PlannedWeight)
	unit := "lb"
	if it.PlannedWeightUnit != nil {
		unit = *it.PlannedWeightUnit
	}
	if hasPerf && perf.Mode == performance.ModeSummary {
		if perf.ActualSets != nil {
			sets = *perf.ActualSets
		}
		if perf.ActualReps != nil {
			reps = *perf.ActualReps
		}
		if perf.ActualWeight != nil {
			weight = *perf.ActualWeight
		}
		if perf.WeightUnit != nil {
			unit = *perf.WeightUnit
		}
	}
	if sets == 0 || reps == 0 || weight == 0 {
		return 0
	}
	return float64(sets) * float64(reps) * convertWeight(weight, unit, targetUnit)
}

func cardioActuals(it planneditem.Entity, perf performance.Entity, hasPerf bool) (int, float64, string) {
	dur := derefInt(it.PlannedDurationSeconds)
	dist := derefFloat(it.PlannedDistance)
	unit := "mi"
	if it.PlannedDistanceUnit != nil {
		unit = *it.PlannedDistanceUnit
	}
	if hasPerf {
		if perf.ActualDurationSeconds != nil {
			dur = *perf.ActualDurationSeconds
		}
		if perf.ActualDistance != nil {
			dist = *perf.ActualDistance
		}
		if perf.ActualDistanceUnit != nil {
			unit = *perf.ActualDistanceUnit
		}
	}
	return dur, dist, unit
}

// convertWeight returns `value` expressed in `to` units, given source unit `from`.
// Conversion factor: 1 kg = 2.20462 lb. Identical units short-circuit.
func convertWeight(value float64, from, to string) float64 {
	if from == to {
		return value
	}
	switch {
	case from == "kg" && to == "lb":
		return value * 2.20462
	case from == "lb" && to == "kg":
		return value / 2.20462
	}
	return value
}

// convertDistance returns `value` expressed in `to` units. Internal unit is
// meters; we route through it so the matrix stays small.
func convertDistance(value float64, from, to string) float64 {
	if from == to || value == 0 {
		return value
	}
	meters := value
	switch from {
	case "mi":
		meters = value * 1609.344
	case "km":
		meters = value * 1000
	}
	switch to {
	case "mi":
		return meters / 1609.344
	case "km":
		return meters / 1000
	}
	return meters
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
func derefFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// --- per-day projection helpers -------------------------------------------

// performanceID returns a stable id for a performance row when one exists,
// or uuid.Nil otherwise. Used so the setMap lookup is safe even when the
// item has no performance row yet.
func performanceID(p performance.Entity, exists bool) uuid.UUID {
	if !exists {
		return uuid.Nil
	}
	return p.Id
}

func buildPlannedShape(it planneditem.Entity, kind string) any {
	switch kind {
	case exercise.KindStrength:
		return map[string]any{
			"sets":       it.PlannedSets,
			"reps":       it.PlannedReps,
			"weight":     it.PlannedWeight,
			"weightUnit": it.PlannedWeightUnit,
		}
	case exercise.KindIsometric:
		return map[string]any{
			"sets":            it.PlannedSets,
			"durationSeconds": it.PlannedDurationSeconds,
			"weight":          it.PlannedWeight,
			"weightUnit":      it.PlannedWeightUnit,
		}
	case exercise.KindCardio:
		return map[string]any{
			"durationSeconds": it.PlannedDurationSeconds,
			"distance":        it.PlannedDistance,
			"distanceUnit":    it.PlannedDistanceUnit,
		}
	}
	return map[string]any{}
}

func buildActualSummary(p performance.Entity, sets []performance.SetEntity, kind string) any {
	if p.Id == uuid.Nil {
		return nil
	}
	if p.Mode == performance.ModePerSet && len(sets) > 0 {
		count := len(sets)
		maxReps := 0
		var maxWeight float64
		for _, s := range sets {
			if s.Reps > maxReps {
				maxReps = s.Reps
			}
			if s.Weight > maxWeight {
				maxWeight = s.Weight
			}
		}
		return map[string]any{
			"sets":       count,
			"reps":       maxReps,
			"weight":     maxWeight,
			"weightUnit": p.WeightUnit,
		}
	}
	switch kind {
	case exercise.KindStrength:
		return map[string]any{
			"sets":       p.ActualSets,
			"reps":       p.ActualReps,
			"weight":     p.ActualWeight,
			"weightUnit": p.WeightUnit,
		}
	case exercise.KindIsometric:
		return map[string]any{
			"sets":            p.ActualSets,
			"durationSeconds": p.ActualDurationSeconds,
			"weight":          p.ActualWeight,
			"weightUnit":      p.WeightUnit,
		}
	case exercise.KindCardio:
		return map[string]any{
			"durationSeconds": p.ActualDurationSeconds,
			"distance":        p.ActualDistance,
			"distanceUnit":    p.ActualDistanceUnit,
		}
	}
	return nil
}

// _ exists to keep the logrus import alive when the package compiles in
// configurations where summary handlers don't directly emit logs.
var _ logrus.FieldLogger
