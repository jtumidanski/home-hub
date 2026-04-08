package month

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/entry"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/schedule"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/trackingitem"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrInvalidMonth    = errors.New("month must be in YYYY-MM format")
	ErrMonthIncomplete = errors.New("report is only available for completed months")
)

type Processor struct {
	l   logrus.FieldLogger
	ctx context.Context
	db  *gorm.DB
}

func NewProcessor(l logrus.FieldLogger, ctx context.Context, db *gorm.DB) *Processor {
	return &Processor{l: l, ctx: ctx, db: db}
}

// SummaryDetail bundles the month summary together with the active items, the
// entries that fall within the month, and the schedule snapshots that govern
// expected days. The REST layer needs all four to render its document, so the
// processor returns them as a single value rather than forcing the handler to
// re-query providers.
type SummaryDetail struct {
	Summary         MonthSummary
	Items           []trackingitem.Model
	Entries         []entry.Model
	SnapshotsByItem map[uuid.UUID][]schedule.Model
}

func (p *Processor) ComputeMonthSummaryDetail(userID uuid.UUID, monthStr string) (SummaryDetail, error) {
	summary, items, entries, snapshots, err := p.computeMonthSummary(userID, monthStr)
	if err != nil {
		return SummaryDetail{}, err
	}
	return SummaryDetail{
		Summary:         summary,
		Items:           items,
		Entries:         entries,
		SnapshotsByItem: snapshots,
	}, nil
}

// ComputeMonthSummary preserves the original 4-tuple return for tests and
// internal callers that don't need the snapshot map.
func (p *Processor) ComputeMonthSummary(userID uuid.UUID, monthStr string) (MonthSummary, []trackingitem.Model, []entry.Model, error) {
	summary, items, entries, _, err := p.computeMonthSummary(userID, monthStr)
	return summary, items, entries, err
}

func (p *Processor) computeMonthSummary(userID uuid.UUID, monthStr string) (MonthSummary, []trackingitem.Model, []entry.Model, map[uuid.UUID][]schedule.Model, error) {
	monthStart, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return MonthSummary{}, nil, nil, nil, ErrInvalidMonth
	}
	monthEnd := monthStart.AddDate(0, 1, -1)
	today := time.Now().UTC().Truncate(24 * time.Hour)

	items, err := trackingitem.GetAllByUserIncludeDeleted(userID)(p.db.WithContext(p.ctx))()
	if err != nil {
		return MonthSummary{}, nil, nil, nil, err
	}

	var activeItems []trackingitem.Model
	for _, e := range items {
		m, err := trackingitem.Make(e)
		if err != nil {
			continue
		}
		itemStart := m.CreatedAt().Truncate(24 * time.Hour)
		if itemStart.After(monthEnd) {
			continue
		}
		if m.DeletedAt() != nil {
			deletedAt := m.DeletedAt().Truncate(24 * time.Hour)
			if deletedAt.Before(monthStart) {
				continue
			}
		}
		activeItems = append(activeItems, m)
	}

	itemIDs := make([]uuid.UUID, len(activeItems))
	for i, m := range activeItems {
		itemIDs[i] = m.Id()
	}

	snapshotsByItem, err := schedule.NewProcessor(p.l, p.ctx, p.db).GetHistoriesByItems(itemIDs)
	if err != nil {
		return MonthSummary{}, nil, nil, nil, err
	}

	entries, err := entry.GetByUserAndMonth(userID, monthStart, monthEnd)(p.db.WithContext(p.ctx))()
	if err != nil {
		return MonthSummary{}, nil, nil, nil, err
	}

	var entryModels []entry.Model
	for _, e := range entries {
		m, err := entry.Make(e)
		if err != nil {
			continue
		}
		entryModels = append(entryModels, m)
	}

	entryMap := make(map[string]entry.Model)
	for _, e := range entryModels {
		key := e.TrackingItemID().String() + ":" + e.Date().Format("2006-01-02")
		entryMap[key] = e
	}

	totalExpected := 0
	totalFilled := 0
	totalSkipped := 0
	hasFutureScheduled := false

	for _, item := range activeItems {
		snapshots := snapshotsByItem[item.Id()]
		itemStart := item.CreatedAt().Truncate(24 * time.Hour)
		if itemStart.Before(monthStart) {
			itemStart = monthStart
		}
		itemEnd := monthEnd
		if item.DeletedAt() != nil {
			deletedAt := item.DeletedAt().Truncate(24 * time.Hour)
			if deletedAt.Before(itemEnd) {
				itemEnd = deletedAt
			}
		}

		for d := itemStart; !d.After(itemEnd); d = d.AddDate(0, 0, 1) {
			if isScheduledDay(d, snapshots) {
				totalExpected++
				key := item.Id().String() + ":" + d.Format("2006-01-02")
				if e, ok := entryMap[key]; ok {
					if e.Skipped() {
						totalSkipped++
					} else {
						totalFilled++
					}
				} else if d.After(today) {
					hasFutureScheduled = true
				}
			}
		}
	}

	remaining := totalExpected - totalFilled - totalSkipped
	complete := remaining == 0 && !hasFutureScheduled

	return MonthSummary{
		Month:    monthStr,
		Complete: complete,
		Completion: CompletionStats{
			Expected:  totalExpected,
			Filled:    totalFilled,
			Skipped:   totalSkipped,
			Remaining: remaining,
		},
	}, activeItems, entryModels, snapshotsByItem, nil
}

func (p *Processor) ComputeReport(userID uuid.UUID, monthStr string) (Report, error) {
	summary, activeItems, entryModels, snapshotsByItem, err := p.computeMonthSummary(userID, monthStr)
	if err != nil {
		return Report{}, err
	}
	if !summary.Complete {
		return Report{}, ErrMonthIncomplete
	}

	monthStart, _ := time.Parse("2006-01", monthStr)
	monthEnd := monthStart.AddDate(0, 1, -1)

	entriesByItem := make(map[uuid.UUID][]entry.Model)
	for _, e := range entryModels {
		entriesByItem[e.TrackingItemID()] = append(entriesByItem[e.TrackingItemID()], e)
	}

	var itemReports []ItemReport
	for _, item := range activeItems {
		snapshots := snapshotsByItem[item.Id()]
		itemEntries := entriesByItem[item.Id()]
		itemStart := item.CreatedAt().Truncate(24 * time.Hour)
		if itemStart.Before(monthStart) {
			itemStart = monthStart
		}
		itemEnd := monthEnd
		if item.DeletedAt() != nil {
			deletedAt := item.DeletedAt().Truncate(24 * time.Hour)
			if deletedAt.Before(itemEnd) {
				itemEnd = deletedAt
			}
		}

		expectedDays := 0
		for d := itemStart; !d.After(itemEnd); d = d.AddDate(0, 0, 1) {
			if isScheduledDay(d, snapshots) {
				expectedDays++
			}
		}

		filledDays := 0
		skippedDays := 0
		for _, e := range itemEntries {
			if e.Skipped() {
				skippedDays++
			} else {
				filledDays++
			}
		}

		var stats json.RawMessage
		switch item.ScaleType() {
		case "sentiment":
			stats = computeSentimentStats(expectedDays, filledDays, skippedDays, itemEntries)
		case "numeric":
			stats = computeNumericStats(expectedDays, filledDays, skippedDays, itemEntries)
		case "range":
			stats = computeRangeStats(expectedDays, filledDays, skippedDays, itemEntries)
		}

		itemReports = append(itemReports, ItemReport{
			TrackingItemId: item.Id(),
			Name:           item.Name(),
			ScaleType:      item.ScaleType(),
			Stats:          stats,
		})
	}

	completionRate := 0.0
	skipRate := 0.0
	if summary.Completion.Expected > 0 {
		completionRate = float64(summary.Completion.Filled) / float64(summary.Completion.Expected)
		skipRate = float64(summary.Completion.Skipped) / float64(summary.Completion.Expected)
	}

	return Report{
		Month: monthStr,
		Summary: ReportSummary{
			TotalItems:     len(activeItems),
			CompletionRate: math.Round(completionRate*100) / 100,
			SkipRate:       math.Round(skipRate*100) / 100,
			TotalExpected:  summary.Completion.Expected,
			TotalFilled:    summary.Completion.Filled,
			TotalSkipped:   summary.Completion.Skipped,
		},
		Items: itemReports,
	}, nil
}

func isScheduledDay(date time.Time, snapshots []schedule.Model) bool {
	var applicable *schedule.Model
	for i := range snapshots {
		effDate := snapshots[i].EffectiveDate().Truncate(24 * time.Hour)
		if !effDate.After(date) {
			s := snapshots[i]
			applicable = &s
		}
	}
	if applicable == nil {
		return false
	}
	sched := applicable.Schedule()
	if len(sched) == 0 {
		return true
	}
	dow := int(date.Weekday())
	for _, d := range sched {
		if d == dow {
			return true
		}
	}
	return false
}

func computeSentimentStats(expectedDays, filledDays, skippedDays int, entries []entry.Model) json.RawMessage {
	stats := SentimentStats{
		ExpectedDays: expectedDays,
		FilledDays:   filledDays,
		SkippedDays:  skippedDays,
	}
	for _, e := range entries {
		if e.Skipped() || len(e.Value()) == 0 {
			continue
		}
		var sv struct{ Rating string `json:"rating"` }
		if err := json.Unmarshal(e.Value(), &sv); err != nil {
			continue
		}
		switch sv.Rating {
		case "positive":
			stats.Positive++
		case "neutral":
			stats.Neutral++
		case "negative":
			stats.Negative++
		}
		stats.DailyValues = append(stats.DailyValues, DailyRating{
			Date:   e.Date().Format("2006-01-02"),
			Rating: sv.Rating,
		})
	}
	total := stats.Positive + stats.Neutral + stats.Negative
	if total > 0 {
		stats.PositiveRatio = math.Round(float64(stats.Positive)/float64(total)*100) / 100
	}
	b, _ := json.Marshal(stats)
	return b
}

func computeNumericStats(expectedDays, filledDays, skippedDays int, entries []entry.Model) json.RawMessage {
	stats := NumericStats{
		ExpectedDays: expectedDays,
		FilledDays:   filledDays,
		SkippedDays:  skippedDays,
	}
	var values []DailyCount
	for _, e := range entries {
		if e.Skipped() || len(e.Value()) == 0 {
			continue
		}
		var nv struct{ Count int `json:"count"` }
		if err := json.Unmarshal(e.Value(), &nv); err != nil {
			continue
		}
		dc := DailyCount{Date: e.Date().Format("2006-01-02"), Count: nv.Count}
		values = append(values, dc)
		stats.Total += nv.Count
		if nv.Count > 0 {
			stats.DaysWithEntriesAboveZero++
		}
	}
	stats.DailyValues = values
	if filledDays > 0 {
		stats.DailyAverage = math.Round(float64(stats.Total)/float64(filledDays)*100) / 100
		stats.DaysWithEntriesAboveZeroPct = math.Round(float64(stats.DaysWithEntriesAboveZero)/float64(filledDays)*100) / 100
	}
	for _, v := range values {
		if stats.Min == nil || v.Count < stats.Min.Count {
			min := v
			stats.Min = &min
		}
		if stats.Max == nil || v.Count > stats.Max.Count {
			max := v
			stats.Max = &max
		}
	}
	b, _ := json.Marshal(stats)
	return b
}

func computeRangeStats(expectedDays, filledDays, skippedDays int, entries []entry.Model) json.RawMessage {
	stats := RangeStats{
		ExpectedDays: expectedDays,
		FilledDays:   filledDays,
		SkippedDays:  skippedDays,
	}
	var values []DailyValue
	var nums []float64
	for _, e := range entries {
		if e.Skipped() || len(e.Value()) == 0 {
			continue
		}
		var rv struct{ Value int `json:"value"` }
		if err := json.Unmarshal(e.Value(), &rv); err != nil {
			continue
		}
		dv := DailyValue{Date: e.Date().Format("2006-01-02"), Value: rv.Value}
		values = append(values, dv)
		nums = append(nums, float64(rv.Value))
	}
	stats.DailyValues = values
	if len(nums) > 0 {
		sum := 0.0
		for _, n := range nums {
			sum += n
		}
		stats.Average = math.Round(sum/float64(len(nums))*10) / 10

		for _, v := range values {
			if stats.Min == nil || v.Value < stats.Min.Value {
				min := v
				stats.Min = &min
			}
			if stats.Max == nil || v.Value > stats.Max.Value {
				max := v
				stats.Max = &max
			}
		}

		if len(nums) > 1 {
			mean := sum / float64(len(nums))
			variance := 0.0
			for _, n := range nums {
				diff := n - mean
				variance += diff * diff
			}
			variance /= float64(len(nums))
			stats.StdDev = math.Round(math.Sqrt(variance)*10) / 10
		}
	}
	b, _ := json.Marshal(stats)
	return b
}
