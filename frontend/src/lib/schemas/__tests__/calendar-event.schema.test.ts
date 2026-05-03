import { describe, it, expect } from "vitest";
import { eventFormSchema, createEventDefaults } from "@/lib/schemas/calendar-event.schema";

function baseValid() {
  return {
    title: "x",
    allDay: false,
    startDate: "2026-05-06",
    startTime: "09:00",
    endDate: "2026-05-06",
    endTime: "10:00",
    recurrence: "RRULE:FREQ=WEEKLY",
    location: "",
    description: "",
    calendarId: "cal-1",
    connectionId: "conn-1",
    endsMode: "on" as const,
    endsOnDate: "2027-05-06",
    endsAfterCount: 10,
    endsNeverConfirmed: false,
    endsOnDateUserEdited: false,
  };
}

describe("eventFormSchema — Ends fields", () => {
  it("createEventDefaults seeds the new fields", () => {
    const d = createEventDefaults();
    expect(d.endsMode).toBe("on");
    expect(d.endsOnDate).toBe("");
    expect(d.endsAfterCount).toBe(10);
    expect(d.endsNeverConfirmed).toBe(false);
    expect(d.endsOnDateUserEdited).toBe(false);
  });

  it("accepts a valid bounded recurring event", () => {
    expect(eventFormSchema.safeParse(baseValid()).success).toBe(true);
  });

  it("ignores Ends fields when recurrence is empty", () => {
    const r = eventFormSchema.safeParse({ ...baseValid(), recurrence: "", endsOnDate: "" });
    expect(r.success).toBe(true);
  });

  it("rejects mode=on with an end date before the start date", () => {
    const r = eventFormSchema.safeParse({ ...baseValid(), endsOnDate: "2026-04-01" });
    expect(r.success).toBe(false);
  });

  it("rejects mode=on with an end date more than 5 years out", () => {
    const r = eventFormSchema.safeParse({ ...baseValid(), endsOnDate: "2031-06-01" });
    expect(r.success).toBe(false);
  });

  it("rejects mode=after with count = 0", () => {
    const r = eventFormSchema.safeParse({ ...baseValid(), endsMode: "after", endsAfterCount: 0 });
    expect(r.success).toBe(false);
  });

  it("rejects mode=after with count = 731", () => {
    const r = eventFormSchema.safeParse({ ...baseValid(), endsMode: "after", endsAfterCount: 731 });
    expect(r.success).toBe(false);
  });

  it("rejects mode=never without confirmation", () => {
    const r = eventFormSchema.safeParse({
      ...baseValid(),
      endsMode: "never",
      endsNeverConfirmed: false,
    });
    expect(r.success).toBe(false);
  });

  it("accepts mode=never when confirmed", () => {
    const r = eventFormSchema.safeParse({
      ...baseValid(),
      endsMode: "never",
      endsNeverConfirmed: true,
    });
    expect(r.success).toBe(true);
  });
});
