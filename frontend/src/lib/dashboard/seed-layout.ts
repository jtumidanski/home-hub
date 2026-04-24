import type { Layout } from "@/lib/dashboard/schema";

function uuid(): string {
  return (crypto as Crypto).randomUUID();
}

/**
 * Produces the first-run dashboard layout, replicating the widget set and
 * arrangement of the legacy DashboardPage. Each invocation returns fresh
 * UUIDs so the same seed can be used repeatedly without collision.
 */
export function seedLayout(): Layout {
  return {
    version: 1,
    widgets: [
      { id: uuid(), type: "weather",           x: 0, y: 0, w: 12, h: 3, config: { units: "imperial", location: null } },
      { id: uuid(), type: "tasks-summary",     x: 0, y: 3, w: 4,  h: 2, config: { status: "pending", title: "Pending Tasks" } },
      { id: uuid(), type: "reminders-summary", x: 4, y: 3, w: 4,  h: 2, config: { filter: "active", title: "Active Reminders" } },
      { id: uuid(), type: "overdue-summary",   x: 8, y: 3, w: 4,  h: 2, config: { title: "Overdue" } },
      { id: uuid(), type: "meal-plan-today",   x: 0, y: 5, w: 4,  h: 3, config: { horizonDays: 1 } },
      { id: uuid(), type: "habits-today",      x: 4, y: 5, w: 4,  h: 3, config: {} },
      { id: uuid(), type: "packages-summary",  x: 8, y: 5, w: 4,  h: 3, config: {} },
      { id: uuid(), type: "calendar-today",    x: 0, y: 8, w: 6,  h: 3, config: { horizonDays: 1, includeAllDay: true } },
      { id: uuid(), type: "workout-today",     x: 6, y: 8, w: 6,  h: 3, config: {} },
    ],
  };
}
