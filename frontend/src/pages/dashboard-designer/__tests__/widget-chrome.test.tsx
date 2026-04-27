import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { WidgetChrome } from "@/pages/dashboard-designer/widget-chrome";
import type { WidgetInstance } from "@/lib/dashboard/schema";

vi.mock("@/components/features/dashboard-widgets/tasks-summary", () => ({
  TasksSummaryWidget: () => null,
}));

function makeWidget(overrides: Partial<WidgetInstance> = {}): WidgetInstance {
  return {
    id: "w-1",
    type: "tasks-summary",
    x: 0,
    y: 0,
    w: 4,
    h: 2,
    config: {},
    ...overrides,
  };
}

describe("WidgetChrome", () => {
  it("gear dispatches select", () => {
    const dispatch = vi.fn();
    const widget = makeWidget();
    render(
      <WidgetChrome widget={widget} dispatch={dispatch}>
        <div>body</div>
      </WidgetChrome>,
    );
    fireEvent.click(screen.getByTestId("widget-configure-w-1"));
    expect(dispatch).toHaveBeenCalledWith({ type: "select", id: "w-1" });
  });

  it("trash dispatches remove", () => {
    const dispatch = vi.fn();
    const widget = makeWidget();
    render(
      <WidgetChrome widget={widget} dispatch={dispatch}>
        <div>body</div>
      </WidgetChrome>,
    );
    fireEvent.click(screen.getByTestId("widget-remove-w-1"));
    expect(dispatch).toHaveBeenCalledWith({ type: "remove", id: "w-1" });
  });

  it("gear is disabled with tooltip when widget type is unknown", () => {
    const dispatch = vi.fn();
    const widget = makeWidget({ id: "w-x", type: "totally-made-up-type" });
    render(
      <WidgetChrome widget={widget} dispatch={dispatch}>
        <div>body</div>
      </WidgetChrome>,
    );
    const gear = screen.getByTestId("widget-configure-w-x") as HTMLButtonElement;
    expect(gear).toBeDisabled();
    expect(gear).toHaveAttribute("title", "No config for this widget type");
  });

  it("renders a drag handle with the class used by the grid", () => {
    const dispatch = vi.fn();
    const widget = makeWidget();
    render(
      <WidgetChrome widget={widget} dispatch={dispatch}>
        <div>body</div>
      </WidgetChrome>,
    );
    const handle = screen.getByTestId("widget-drag-w-1");
    expect(handle.className).toMatch(/widget-drag-handle/);
  });
});
