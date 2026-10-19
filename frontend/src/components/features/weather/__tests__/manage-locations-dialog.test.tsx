import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ManageLocationsDialog } from "../manage-locations-dialog";

const locations = [
  {
    id: "l1",
    type: "locations-of-interest",
    attributes: { label: "Cabin", placeName: "Bend, Oregon", latitude: 1, longitude: 2 },
  },
  {
    id: "l2",
    type: "locations-of-interest",
    attributes: { label: null, placeName: "Seattle, Washington", latitude: 3, longitude: 4 },
  },
];

vi.mock("@/lib/hooks/api/use-locations-of-interest", () => ({
  useLocationsOfInterest: () => ({ data: { data: locations }, isLoading: false }),
  useCreateLocationOfInterest: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateLocationOfInterest: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeleteLocationOfInterest: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock("@/lib/hooks/api/use-weather", () => ({
  useGeocodingSearch: () => ({ data: undefined }),
}));

describe("ManageLocationsDialog", () => {
  it("labels the rename and delete buttons with the location name", () => {
    render(<ManageLocationsDialog open onOpenChange={vi.fn()} />);

    expect(screen.getByRole("button", { name: "Rename Cabin" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Delete Cabin" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Rename Seattle, Washington" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Delete Seattle, Washington" })).toBeInTheDocument();
  });

  it("labels the save and cancel buttons while renaming", async () => {
    const user = userEvent.setup();
    render(<ManageLocationsDialog open onOpenChange={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Rename Cabin" }));

    expect(screen.getByRole("textbox", { name: "Location name" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save name" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Cancel rename" })).toBeInTheDocument();
  });
});
