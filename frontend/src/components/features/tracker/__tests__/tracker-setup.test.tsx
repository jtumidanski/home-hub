import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { Tracker } from "@/types/models/tracker";
import { TrackerSetup } from "../tracker-setup";

const trackers = [
  {
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
  },
] as unknown as Tracker[];

vi.mock("@/lib/hooks/api/use-trackers", () => ({
  useTrackers: () => ({ data: { data: trackers }, isLoading: false }),
  useDeleteTracker: () => ({ mutate: vi.fn(), isPending: false }),
  useCreateTracker: () => ({ mutate: vi.fn(), isPending: false }),
  useUpdateTracker: () => ({ mutate: vi.fn(), isPending: false }),
}));

describe("TrackerSetup", () => {
  it("names the icon-only edit and delete buttons after the habit", () => {
    render(<TrackerSetup />);
    expect(screen.getByRole("button", { name: "Edit Run" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Delete Run" })).toBeInTheDocument();
  });
});
