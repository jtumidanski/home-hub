import type { Layout } from "@/lib/dashboard/schema";

function uuid(): string {
  return (crypto as Crypto).randomUUID();
}

/**
 * Seeds the "Kiosk" dashboard — a 4-column kiosk-style composition.
 * Columns each total h:12 grid units; per-widget heights tuned per design §6.2.
 */
export function kioskSeedLayout(): Layout {
  return {
    version: 1,
    widgets: [
      // Column 1 (x:0, w:3, h-total:12)
      { id: uuid(), type: "weather",          x: 0, y: 0,  w: 3, h: 3,  config: { units: "imperial", location: null } },
      { id: uuid(), type: "meal-plan-today",  x: 0, y: 3,  w: 3, h: 5,  config: { horizonDays: 3, view: "today-detail" } },
      { id: uuid(), type: "tasks-today",      x: 0, y: 8,  w: 3, h: 4,  config: { includeCompleted: true } },
      // Column 2 (x:3, w:3, h-total:12)
      { id: uuid(), type: "calendar-today",   x: 3, y: 0,  w: 3, h: 12, config: { horizonDays: 1, includeAllDay: true } },
      // Column 3 (x:6, w:3, h-total:12)
      { id: uuid(), type: "weather-tomorrow", x: 6, y: 0,  w: 3, h: 3,  config: { units: null } },
      { id: uuid(), type: "calendar-tomorrow",x: 6, y: 3,  w: 3, h: 5,  config: { includeAllDay: true, limit: 5 } },
      { id: uuid(), type: "tasks-tomorrow",   x: 6, y: 8,  w: 3, h: 4,  config: { limit: 5 } },
      // Column 4 (x:9, w:3, h-total:12)
      { id: uuid(), type: "reminders-today",  x: 9, y: 0,  w: 3, h: 12, config: { limit: 10 } },
    ],
  };
}
