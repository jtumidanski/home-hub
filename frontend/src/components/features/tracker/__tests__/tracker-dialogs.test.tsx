import { render, screen, fireEvent } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { COLOR_PALETTE, type Tracker } from "@/types/models/tracker";
import { CreateTrackerDialog } from "../create-tracker-dialog";
import { EditTrackerDialog } from "../edit-tracker-dialog";

vi.mock("@/lib/hooks/api/use-trackers", () => ({
  useCreateTracker: () => ({ mutate: vi.fn(), isPending: false }),
  useUpdateTracker: () => ({ mutate: vi.fn(), isPending: false }),
}));

const tracker = {
  id: "t1",
  type: "trackers",
  attributes: {
    name: "Run",
    scale_type: "sentiment",
    scale_config: null,
    schedule: [],
    color: "blue",
    sort_order: 0,
    created_at: "",
    updated_at: "",
  },
} as unknown as Tracker;

describe("colour pickers", () => {
  it("create dialog labels every swatch by colour name and marks one selected", () => {
    render(<CreateTrackerDialog open onClose={() => {}} />);
    for (const c of COLOR_PALETTE) {
      expect(screen.getByRole("button", { name: c })).toBeInTheDocument();
    }
    expect(document.querySelectorAll('button[aria-pressed="true"][aria-label]').length).toBeGreaterThan(0);
    const teal = screen.getByRole("button", { name: "teal" });
    fireEvent.click(teal);
    expect(teal).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "red" })).toHaveAttribute("aria-pressed", "false");
  });

  it("edit dialog marks the tracker's colour as pressed", () => {
    render(<EditTrackerDialog open onClose={() => {}} tracker={tracker} />);
    for (const c of COLOR_PALETTE) {
      expect(screen.getByRole("button", { name: c })).toBeInTheDocument();
    }
    expect(screen.getByRole("button", { name: "blue" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "red" })).toHaveAttribute("aria-pressed", "false");
    fireEvent.click(screen.getByRole("button", { name: "pink" }));
    expect(screen.getByRole("button", { name: "pink" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "blue" })).toHaveAttribute("aria-pressed", "false");
  });
});
