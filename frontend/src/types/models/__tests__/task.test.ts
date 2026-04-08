import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import type { Task } from "../task";
import { isTaskOverdue } from "../task";

function makeTask(overrides: Partial<Task["attributes"]> = {}): Task {
  return {
    id: "t1",
    type: "tasks",
    attributes: {
      title: "test",
      status: "pending",
      rolloverEnabled: false,
      createdAt: "2026-01-01T00:00:00Z",
      updatedAt: "2026-01-01T00:00:00Z",
      ...overrides,
    },
  };
}

function ymd(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

describe("isTaskOverdue", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    // Pick a fixed local "now" mid-day to avoid DST surprises
    vi.setSystemTime(new Date(2026, 3, 8, 12, 0, 0));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns false for a task due today", () => {
    const today = ymd(new Date());
    expect(isTaskOverdue(makeTask({ dueOn: today }))).toBe(false);
  });

  it("returns true for a task due yesterday", () => {
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    expect(isTaskOverdue(makeTask({ dueOn: ymd(yesterday) }))).toBe(true);
  });

  it("returns false for a task due tomorrow", () => {
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    expect(isTaskOverdue(makeTask({ dueOn: ymd(tomorrow) }))).toBe(false);
  });

  it("returns false for a completed task with a past due date", () => {
    expect(
      isTaskOverdue(makeTask({ status: "completed", dueOn: "2020-01-01" })),
    ).toBe(false);
  });

  it("returns false when dueOn is missing", () => {
    expect(isTaskOverdue(makeTask())).toBe(false);
  });
});
