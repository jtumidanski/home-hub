import { describe, it, expect } from "vitest";
import { seedLayout } from "@/lib/dashboard/seed-layout";
import { kioskSeedLayout } from "@/lib/dashboard/kiosk-seed-layout";
import { widgetRegistry, findWidget } from "@/lib/dashboard/widget-registry";
import { WIDGET_TYPES } from "@/lib/dashboard/widget-types";

describe("dashboard parity — seed layouts against widget registry", () => {
  it("union of seedLayout + kioskSeedLayout covers every registry type", () => {
    const seeded = new Set([
      ...seedLayout().widgets.map((w) => w.type),
      ...kioskSeedLayout().widgets.map((w) => w.type),
    ]);
    for (const def of widgetRegistry) {
      expect(seeded.has(def.type), `${def.type} appears in registry but not in any seed layout`).toBe(true);
    }
  });

  it("every seeded widget resolves via findWidget", () => {
    for (const w of [...seedLayout().widgets, ...kioskSeedLayout().widgets]) {
      expect(findWidget(w.type), `missing registry entry for ${w.type}`).toBeDefined();
    }
  });

  it("WIDGET_TYPES is fully covered by the union of seed layouts", () => {
    const seeded = new Set([
      ...seedLayout().widgets.map((w) => w.type),
      ...kioskSeedLayout().widgets.map((w) => w.type),
    ]);
    for (const t of WIDGET_TYPES) {
      expect(seeded.has(t), `allowlist contains ${t} but neither seed layout includes it`).toBe(true);
    }
  });
});
