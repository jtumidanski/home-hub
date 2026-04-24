import { describe, it, expect } from "vitest";
import fixture from "@/lib/dashboard/fixtures/widget-types.json";
import { WIDGET_TYPES } from "@/lib/dashboard/widget-types";

describe("widget-types parity with Go allowlist", () => {
  it("matches the shared fixture exactly", () => {
    const ts = [...WIDGET_TYPES].sort();
    const go = [...(fixture as string[])].sort();
    expect(ts).toEqual(go);
  });
});
