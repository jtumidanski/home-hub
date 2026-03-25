import { describe, it, expect } from "vitest";
import { createReminderSchema } from "../reminder.schema";

describe("createReminderSchema", () => {
  it("accepts valid data", () => {
    const result = createReminderSchema.safeParse({
      title: "Take medicine",
      scheduledFor: "2026-04-01T09:00:00Z",
    });
    expect(result.success).toBe(true);
  });

  it("accepts valid data with notes", () => {
    const result = createReminderSchema.safeParse({
      title: "Take medicine",
      notes: "After breakfast",
      scheduledFor: "2026-04-01T09:00:00Z",
    });
    expect(result.success).toBe(true);
  });

  it("rejects empty title", () => {
    const result = createReminderSchema.safeParse({
      title: "",
      scheduledFor: "2026-04-01T09:00:00Z",
    });
    expect(result.success).toBe(false);
  });

  it("rejects title exceeding max length", () => {
    const result = createReminderSchema.safeParse({
      title: "a".repeat(201),
      scheduledFor: "2026-04-01T09:00:00Z",
    });
    expect(result.success).toBe(false);
  });

  it("rejects notes exceeding max length", () => {
    const result = createReminderSchema.safeParse({
      title: "Valid",
      notes: "a".repeat(1001),
      scheduledFor: "2026-04-01T09:00:00Z",
    });
    expect(result.success).toBe(false);
  });

  it("rejects empty scheduledFor", () => {
    const result = createReminderSchema.safeParse({
      title: "Reminder",
      scheduledFor: "",
    });
    expect(result.success).toBe(false);
  });

  it("rejects missing required fields", () => {
    const result = createReminderSchema.safeParse({});
    expect(result.success).toBe(false);
  });
});
