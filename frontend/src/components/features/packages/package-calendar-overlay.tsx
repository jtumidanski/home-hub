import type { CalendarEvent } from "@/types/models/calendar";
import type { Package } from "@/types/models/package";
import { CARRIER_LABELS } from "@/types/models/package";

/**
 * Converts packages with ETAs into CalendarEvent objects so they can be
 * rendered alongside real calendar events in the CalendarGrid.
 *
 * Package events are distinguished by:
 * - A "package" userColor (teal)
 * - userDisplayName = carrier label
 */
export function packagesToCalendarEvents(packages: Package[]): CalendarEvent[] {
  return packages
    .filter((p) => p.attributes.estimatedDelivery)
    .map((p): CalendarEvent => {
      const eta = p.attributes.estimatedDelivery!;
      const label = p.attributes.label ?? "Package";
      const carrier = CARRIER_LABELS[p.attributes.carrier] ?? p.attributes.carrier;

      return {
        id: `pkg-${p.id}`,
        type: "calendar-events",
        attributes: {
          title: `${carrier}: ${label}`,
          description: null,
          startTime: `${eta}T00:00:00Z`,
          endTime: `${eta}T23:59:59Z`,
          allDay: true,
          location: null,
          visibility: "default",
          userDisplayName: carrier,
          userColor: "#0d9488", // teal-600 for package events
          isOwner: p.attributes.isOwner,
        },
      };
    });
}
