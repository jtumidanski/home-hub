import { describe, it, expect } from "vitest";
import { weatherTomorrowWidget } from "@/lib/dashboard/widgets/weather-tomorrow";

describe("weather-tomorrow widget definition", () => {
  it("declares metadata", () => {
    expect(weatherTomorrowWidget.type).toBe("weather-tomorrow");
    expect(weatherTomorrowWidget.dataScope).toBe("household");
    expect(weatherTomorrowWidget.defaultSize).toEqual({ w: 3, h: 2 });
    expect(weatherTomorrowWidget.minSize).toEqual({ w: 2, h: 2 });
    expect(weatherTomorrowWidget.maxSize).toEqual({ w: 6, h: 3 });
  });

  it("schema accepts null/imperial/metric for units", () => {
    expect(weatherTomorrowWidget.configSchema.safeParse({ units: "imperial" }).success).toBe(true);
    expect(weatherTomorrowWidget.configSchema.safeParse({ units: "metric" }).success).toBe(true);
    expect(weatherTomorrowWidget.configSchema.safeParse({ units: null }).success).toBe(true);
    expect(weatherTomorrowWidget.configSchema.safeParse({ units: "kelvin" }).success).toBe(false);
  });

  it("default config has units=null", () => {
    expect(weatherTomorrowWidget.defaultConfig).toEqual({ units: null });
  });
});
