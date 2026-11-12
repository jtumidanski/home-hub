import { describe, it, expect } from "vitest";
import { isNavItemActive, navGroups } from "../nav-config";

const allItems = navGroups.flatMap((g) => g.items);

describe("nav-config active matching for sub-routes", () => {
  it.each([
    ["/app/habits/setup", "/app/habits"],
    ["/app/shopping/grocery/archived", "/app/shopping/grocery"],
    ["/app/calendar/month", "/app/calendar"],
    ["/app/ingredients/bulk", "/app/ingredients"],
  ])("%s keeps %s active and the item is not exact", (pathname, to) => {
    expect(isNavItemActive(pathname, to)).toBe(true);
    const item = allItems.find((i) => i.to === to);
    expect(item).toBeDefined();
    expect(item?.end).toBeFalsy();
  });
});
