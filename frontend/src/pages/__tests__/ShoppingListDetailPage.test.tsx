import { describe, it, expect } from "vitest";
import { progressPercent } from "../ShoppingListDetailPage";

describe("progressPercent", () => {
  it("returns 0 when total is 0 (no divide-by-zero)", () => {
    expect(progressPercent(0, 0)).toBe(0);
  });
  it("returns 0 when nothing is checked", () => {
    expect(progressPercent(0, 4)).toBe(0);
  });
  it("returns 50 at the halfway point", () => {
    expect(progressPercent(2, 4)).toBe(50);
  });
  it("returns 100 when everything is checked", () => {
    expect(progressPercent(4, 4)).toBe(100);
  });
  it("clamps to 100 when checked exceeds total", () => {
    expect(progressPercent(5, 4)).toBe(100);
  });
});
