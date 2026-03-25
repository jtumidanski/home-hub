import { describe, it, expect } from "vitest";
import { createTaskSchema } from "../task.schema";

describe("createTaskSchema", () => {
  it("accepts valid data with only title", () => {
    const result = createTaskSchema.safeParse({ title: "Buy milk" });
    expect(result.success).toBe(true);
  });

  it("accepts valid data with all fields", () => {
    const result = createTaskSchema.safeParse({
      title: "Buy milk",
      notes: "From the store",
      dueOn: "2026-04-01",
    });
    expect(result.success).toBe(true);
  });

  it("rejects empty title", () => {
    const result = createTaskSchema.safeParse({ title: "" });
    expect(result.success).toBe(false);
  });

  it("rejects title exceeding max length", () => {
    const result = createTaskSchema.safeParse({ title: "a".repeat(201) });
    expect(result.success).toBe(false);
  });

  it("rejects notes exceeding max length", () => {
    const result = createTaskSchema.safeParse({
      title: "Valid",
      notes: "a".repeat(1001),
    });
    expect(result.success).toBe(false);
  });

  it("allows optional notes and dueOn", () => {
    const result = createTaskSchema.safeParse({ title: "Task" });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.notes).toBeUndefined();
      expect(result.data.dueOn).toBeUndefined();
    }
  });
});
