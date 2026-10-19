import { describe, it, expect } from "vitest";
import { parseColor, contrastRatio, readableForeground } from "../color-contrast";

describe("parseColor", () => {
  it("parses short and long hex", () => {
    expect(parseColor("#fff")).toEqual({ r: 1, g: 1, b: 1 });
    expect(parseColor("#000000")).toEqual({ r: 0, g: 0, b: 0 });
  });
  it("parses oklch white and black", () => {
    const w = parseColor("oklch(1 0 0)")!;
    expect(w.r).toBeCloseTo(1, 3);
    expect(w.g).toBeCloseTo(1, 3);
    expect(w.b).toBeCloseTo(1, 3);
    expect(parseColor("oklch(0% 0 0)")).toEqual({ r: 0, g: 0, b: 0 });
  });
  it("ignores alpha", () => {
    expect(parseColor("oklch(1 0 0 / 10%)")).not.toBeNull();
  });
  it("returns null for garbage", () => {
    for (const s of ["", "teal", "#ggg", "rgb(1,2,3)"]) expect(parseColor(s)).toBeNull();
  });
});

describe("contrastRatio", () => {
  it("is 21 for black on white", () => {
    expect(contrastRatio("#000", "#fff")).toBeCloseTo(21, 1);
  });
  it("is symmetric", () => {
    expect(contrastRatio("#777", "#fff")).toBeCloseTo(contrastRatio("#fff", "#777"), 5);
  });
  it("matches a known pair", () => {
    expect(contrastRatio("#767676", "#ffffff")).toBeCloseTo(4.54, 1);
  });
  it("throws on unparseable input", () => {
    expect(() => contrastRatio("nope", "#fff")).toThrow("unparseable colour: nope");
  });
});

describe("readableForeground", () => {
  it("picks ink on light and paper on dark", () => {
    expect(readableForeground("#ffff00")).toBe("ink");
    expect(readableForeground("#000000")).toBe("paper");
  });
  it("falls back to ink when unparseable", () => {
    expect(readableForeground("")).toBe("ink");
  });
});
