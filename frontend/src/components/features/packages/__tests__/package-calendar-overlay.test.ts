import { describe, it, expect } from "vitest";
import { packagesToCalendarEvents } from "../package-calendar-overlay";
import type { Package } from "@/types/models/package";

function makePkg(overrides: Partial<Package["attributes"]> = {}): Package {
  return {
    id: "pkg-1",
    type: "packages",
    attributes: {
      trackingNumber: "1Z999AA10123456784",
      carrier: "ups",
      label: "New Keyboard",
      notes: null,
      status: "in_transit",
      private: false,
      estimatedDelivery: "2026-04-01",
      actualDelivery: null,
      lastPolledAt: null,
      archivedAt: null,
      isOwner: true,
      trackingEvents: [],
      createdAt: "2026-03-20T00:00:00Z",
      updatedAt: "2026-03-25T00:00:00Z",
      ...overrides,
    },
  };
}

describe("packagesToCalendarEvents", () => {
  it("converts package with ETA to all-day calendar event", () => {
    const events = packagesToCalendarEvents([makePkg()]);

    expect(events).toHaveLength(1);
    expect(events[0].id).toBe("pkg-pkg-1");
    expect(events[0].attributes.title).toBe("UPS: New Keyboard");
    expect(events[0].attributes.allDay).toBe(true);
    expect(events[0].attributes.startTime).toBe("2026-04-01T00:00:00Z");
    expect(events[0].attributes.endTime).toBe("2026-04-01T23:59:59Z");
    expect(events[0].attributes.userColor).toBe("#0d9488");
  });

  it("filters out packages without ETA", () => {
    const events = packagesToCalendarEvents([
      makePkg({ estimatedDelivery: "2026-04-01" }),
      makePkg({ estimatedDelivery: null }),
    ]);

    expect(events).toHaveLength(1);
  });

  it("uses 'Package' as label when none provided", () => {
    const events = packagesToCalendarEvents([makePkg({ label: null })]);
    expect(events[0].attributes.title).toBe("UPS: Package");
  });

  it("uses carrier label as userDisplayName", () => {
    const events = packagesToCalendarEvents([makePkg({ carrier: "fedex" })]);
    expect(events[0].attributes.userDisplayName).toBe("FedEx");
  });

  it("returns empty array for empty input", () => {
    expect(packagesToCalendarEvents([])).toEqual([]);
  });
});
