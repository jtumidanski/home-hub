export const WIDGET_TYPES = [
  "weather",
  "tasks-summary",
  "reminders-summary",
  "overdue-summary",
  "meal-plan-today",
  "calendar-today",
  "packages-summary",
  "habits-today",
  "workout-today",
] as const;

export type WidgetType = (typeof WIDGET_TYPES)[number];

export function isKnownWidgetType(t: string): t is WidgetType {
  return (WIDGET_TYPES as readonly string[]).includes(t);
}

export const LAYOUT_SCHEMA_VERSION = 1;
export const GRID_COLUMNS = 12;
export const MAX_WIDGETS = 40;
