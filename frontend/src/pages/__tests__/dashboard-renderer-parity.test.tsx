import { describe, it, expect } from "vitest";
import { seedLayout } from "@/lib/dashboard/seed-layout";
import { widgetRegistry, findWidget } from "@/lib/dashboard/widget-registry";
import { WIDGET_TYPES } from "@/lib/dashboard/widget-types";

describe("dashboard parity — seed layout against widget registry", () => {
  it("seedLayout yields one widget per registry type", () => {
    const layout = seedLayout();
    const seededTypes = layout.widgets.map((w) => w.type).sort();
    const registryTypes = widgetRegistry.map((d) => d.type).sort();
    expect(seededTypes).toEqual(registryTypes);
  });

  it("every seeded widget resolves via findWidget", () => {
    for (const w of seedLayout().widgets) {
      expect(findWidget(w.type), `missing registry entry for ${w.type}`).toBeDefined();
    }
  });

  it("widget-type allowlist matches seed layout exactly", () => {
    const seeded = new Set(seedLayout().widgets.map((w) => w.type));
    for (const t of WIDGET_TYPES) {
      expect(seeded.has(t), `allowlist contains ${t} but seedLayout does not`).toBe(true);
    }
    expect(seeded.size).toBe(WIDGET_TYPES.length);
  });
});
