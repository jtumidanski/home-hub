import { describe, it, expect, vi } from "vitest";
import { reminderKeys } from "../use-reminders";

vi.mock("@/context/tenant-context", () => ({
  useTenant: () => ({
    tenantId: "tenant-1",
    householdId: "household-1",
    tenant: null,
    household: null,
    setActiveHousehold: vi.fn(),
  }),
}));

vi.mock("@/services/api/productivity", () => ({
  productivityService: {
    listReminders: vi.fn(),
    getReminderSummary: vi.fn(),
    createReminder: vi.fn(),
    updateReminder: vi.fn(),
    deleteReminder: vi.fn(),
    snoozeReminder: vi.fn(),
    dismissReminder: vi.fn(),
  },
}));

describe("reminderKeys", () => {
  it("generates all key with household id", () => {
    expect(reminderKeys.all("hh-1")).toEqual(["reminders", "hh-1"]);
  });

  it("generates all key with no-household fallback", () => {
    expect(reminderKeys.all(null)).toEqual(["reminders", "no-household"]);
  });

  it("generates lists key", () => {
    expect(reminderKeys.lists("hh-1")).toEqual(["reminders", "hh-1", "list"]);
  });

  it("generates details key", () => {
    expect(reminderKeys.details("hh-1")).toEqual(["reminders", "hh-1", "detail"]);
  });

  it("generates detail key with id", () => {
    expect(reminderKeys.detail("hh-1", "rem-42")).toEqual(["reminders", "hh-1", "detail", "rem-42"]);
  });

  it("generates summary key", () => {
    expect(reminderKeys.summary("hh-1")).toEqual(["reminders", "hh-1", "summary"]);
  });
});
