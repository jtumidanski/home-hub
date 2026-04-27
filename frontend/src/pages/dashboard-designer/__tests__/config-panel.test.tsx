import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { ConfigPanel } from "@/pages/dashboard-designer/config-panel";
import type { WidgetInstance } from "@/lib/dashboard/schema";

vi.mock("@/components/features/dashboard-widgets/tasks-summary", () => ({
  TasksSummaryWidget: () => null,
}));

function tasksWidget(): WidgetInstance {
  return {
    id: "w-1",
    type: "tasks-summary",
    x: 0,
    y: 0,
    w: 4,
    h: 2,
    config: { status: "pending" },
  };
}

describe("ConfigPanel", () => {
  it("renders status radio and an optional title input for tasks-summary", () => {
    render(<ConfigPanel widget={tasksWidget()} dispatch={vi.fn()} />);
    expect(screen.getByTestId("zod-field-status")).toBeInTheDocument();
    expect(screen.getByTestId("zod-field-title")).toBeInTheDocument();
    expect(screen.getByTestId("config-apply")).toBeInTheDocument();
    expect(screen.getByTestId("config-cancel")).toBeInTheDocument();
    expect(screen.getByTestId("config-reset")).toBeInTheDocument();
  });

  it("Apply dispatches update-config with the edited values and closes", async () => {
    const dispatch = vi.fn();
    render(<ConfigPanel widget={tasksWidget()} dispatch={dispatch} />);

    // Switch status from pending -> overdue by clicking the radio.
    const overdue = screen.getByLabelText("Overdue") as HTMLInputElement;
    fireEvent.click(overdue);

    // Type a title.
    const titleInput = screen.getByTestId("zod-field-title") as HTMLInputElement;
    fireEvent.change(titleInput, { target: { value: "My Tasks" } });

    fireEvent.click(screen.getByTestId("config-apply"));
    await waitFor(() => expect(dispatch).toHaveBeenCalled());
    const updateCall = dispatch.mock.calls.find(
      ([a]) => (a as { type: string }).type === "update-config",
    );
    expect(updateCall).toBeTruthy();
    expect(updateCall![0]).toMatchObject({
      type: "update-config",
      id: "w-1",
      config: { status: "overdue", title: "My Tasks" },
    });
    // A following `select: null` should close the panel.
    expect(
      dispatch.mock.calls.some(
        ([a]) => (a as { type: string; id: string | null }).type === "select" && (a as { id: string | null }).id === null,
      ),
    ).toBe(true);
  });

  it("Cancel closes without dispatching update-config", () => {
    const dispatch = vi.fn();
    render(<ConfigPanel widget={tasksWidget()} dispatch={dispatch} />);
    fireEvent.click(screen.getByTestId("config-cancel"));
    const hadUpdate = dispatch.mock.calls.some(
      ([a]) => (a as { type: string }).type === "update-config",
    );
    expect(hadUpdate).toBe(false);
    expect(dispatch).toHaveBeenCalledWith({ type: "select", id: null });
  });

  it("Reset to defaults clears the form values", () => {
    const dispatch = vi.fn();
    render(
      <ConfigPanel
        widget={{ ...tasksWidget(), config: { status: "overdue", title: "Keep" } }}
        dispatch={dispatch}
      />,
    );
    const titleInput = screen.getByTestId("zod-field-title") as HTMLInputElement;
    expect(titleInput.value).toBe("Keep");
    fireEvent.click(screen.getByTestId("config-reset"));
    // tasks-summary defaultConfig is { status: "pending" } — title is absent.
    expect(titleInput.value).toBe("");
  });
});
