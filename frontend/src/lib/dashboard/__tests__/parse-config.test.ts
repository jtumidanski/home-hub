import { describe, it, expect } from "vitest";
import { z } from "zod";
import { parseConfig } from "@/lib/dashboard/parse-config";

const def = {
  type: "tasks-summary" as const,
  defaultConfig: { status: "pending" as const, title: "" },
  configSchema: z.object({
    status: z.enum(["pending", "overdue", "completed"]),
    title: z.string().max(80).optional(),
  }),
} as const;

describe("parseConfig", () => {
  it("returns parsed config when valid", () => {
    const r = parseConfig(def, { status: "overdue" });
    expect(r.config.status).toBe("overdue");
    expect(r.lossy).toBe(false);
  });

  it("fills missing fields from defaultConfig", () => {
    const r = parseConfig(def, { title: "hi" });
    expect(r.config.status).toBe("pending");
    expect(r.lossy).toBe(false);
  });

  it("falls back to defaults on invalid input", () => {
    const r = parseConfig(def, { status: "not-real" });
    expect(r.config).toEqual(def.defaultConfig);
    expect(r.lossy).toBe(true);
  });

  it("treats non-object raw as empty", () => {
    const r = parseConfig(def, 42);
    expect(r.config).toEqual(def.defaultConfig);
  });
});
