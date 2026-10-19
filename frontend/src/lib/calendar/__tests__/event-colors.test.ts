import { describe, it, expect } from "vitest";
import { eventChipColors, PACKAGE_EVENT_COLOR } from "../event-colors";

const base = { title: "Dentist", isOwner: true, userColor: "#1e3a8a" };

describe("eventChipColors", () => {
  it("uses paper text on a dark user colour and ink text on a light one", () => {
    expect(eventChipColors(base).color).toBe("var(--on-color-paper)");
    expect(eventChipColors({ ...base, userColor: "#fde68a" })).toEqual({
      backgroundColor: "#fde68a",
      color: "var(--on-color-ink)",
    });
  });

  it("uses the busy token pair for another user's Busy placeholder", () => {
    expect(eventChipColors({ ...base, title: "Busy", isOwner: false })).toEqual({
      backgroundColor: "var(--calendar-busy)",
      color: "var(--calendar-busy-foreground)",
    });
  });

  it("uses the user colour for the owner's own Busy event", () => {
    expect(eventChipColors({ ...base, title: "Busy", isOwner: true }).backgroundColor).toBe("#1e3a8a");
  });

  it("uses the package token pair for the package sentinel", () => {
    expect(eventChipColors({ ...base, userColor: PACKAGE_EVENT_COLOR })).toEqual({
      backgroundColor: "var(--calendar-package)",
      color: "var(--calendar-package-foreground)",
    });
  });

  it("lets busy win over the package sentinel", () => {
    expect(
      eventChipColors({ title: "Busy", isOwner: false, userColor: PACKAGE_EVENT_COLOR }),
    ).toEqual({
      backgroundColor: "var(--calendar-busy)",
      color: "var(--calendar-busy-foreground)",
    });
  });

  it("picks ink on yellow and paper on black", () => {
    expect(eventChipColors({ ...base, userColor: "#ffff00" }).color).toBe("var(--on-color-ink)");
    expect(eventChipColors({ ...base, userColor: "#000000" }).color).toBe("var(--on-color-paper)");
  });

  it.each(["", "teal"])("does not throw on unparseable colour %j", (userColor) => {
    expect(eventChipColors({ ...base, userColor })).toEqual({
      backgroundColor: userColor,
      color: "var(--on-color-ink)",
    });
  });
});
