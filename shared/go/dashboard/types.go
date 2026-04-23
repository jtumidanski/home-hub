// Package dashboard holds cross-service constants for the dashboard feature:
// the widget-type allowlist, the layout schema version, and the hard caps on
// payload size / widget count / per-widget config. The matching TypeScript
// allowlist lives at frontend/src/lib/dashboard/widget-types.ts; both sides
// are asserted against fixtures/widget-types.json.
package dashboard

const (
	LayoutSchemaVersion  = 1
	MaxWidgets           = 40
	MaxLayoutBytes       = 64 * 1024
	MaxWidgetConfigBytes = 4 * 1024
	MaxWidgetConfigDepth = 5
	GridColumns          = 12
)

var WidgetTypes = map[string]struct{}{
	"weather":           {},
	"tasks-summary":     {},
	"reminders-summary": {},
	"overdue-summary":   {},
	"meal-plan-today":   {},
	"calendar-today":    {},
	"packages-summary":  {},
	"habits-today":      {},
	"workout-today":     {},
}

func IsKnownWidgetType(t string) bool { _, ok := WidgetTypes[t]; return ok }
