import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { PackageCard } from "../package-card";
import type { Package } from "@/types/models/package";

const mockMutate = vi.fn();

vi.mock("@/lib/hooks/api/use-packages", () => ({
  useRefreshPackage: () => ({ mutate: mockMutate, isPending: false }),
  useArchivePackage: () => ({ mutate: mockMutate, isPending: false }),
  useUnarchivePackage: () => ({ mutate: mockMutate, isPending: false }),
  useDeletePackage: () => ({ mutate: mockMutate, isPending: false }),
  useUpdatePackage: () => ({ mutate: mockMutate, isPending: false }),
}));

function makePkg(overrides: Partial<Package["attributes"]> = {}): Package {
  return {
    id: "pkg-1",
    type: "packages",
    attributes: {
      trackingNumber: "1Z999AA10123456784",
      carrier: "ups",
      label: "New Keyboard",
      notes: "Leave at back door",
      status: "in_transit",
      private: false,
      estimatedDelivery: "2026-04-01",
      actualDelivery: null,
      lastPolledAt: null,
      archivedAt: null,
      isOwner: true,
      trackingEvents: [],
      createdAt: "2026-03-20T00:00:00Z",
      updatedAt: "2026-03-25T00:00:00Z",
      ...overrides,
    },
  };
}

describe("PackageCard", () => {
  it("renders label, tracking number, status, and ETA", () => {
    render(<PackageCard pkg={makePkg()} />);

    expect(screen.getByText("New Keyboard")).toBeInTheDocument();
    expect(screen.getByText("1Z999AA1...6784")).toBeInTheDocument();
    expect(screen.getByText("In Transit")).toBeInTheDocument();
    expect(screen.getByText("ETA: 2026-04-01")).toBeInTheDocument();
  });

  it("renders notes when not redacted", () => {
    render(<PackageCard pkg={makePkg()} />);
    expect(screen.getByText("Leave at back door")).toBeInTheDocument();
  });

  it("shows owner actions for package owner", () => {
    render(<PackageCard pkg={makePkg()} />);

    expect(screen.getByTitle("Refresh tracking")).toBeInTheDocument();
    expect(screen.getByTitle("Edit")).toBeInTheDocument();
    expect(screen.getByTitle("Make private")).toBeInTheDocument();
    expect(screen.getByTitle("Archive")).toBeInTheDocument();
    expect(screen.getByTitle("Delete")).toBeInTheDocument();
  });

  it("hides owner actions for non-owner", () => {
    render(<PackageCard pkg={makePkg({ isOwner: false })} />);

    expect(screen.queryByTitle("Refresh tracking")).not.toBeInTheDocument();
    expect(screen.queryByTitle("Edit")).not.toBeInTheDocument();
    expect(screen.queryByTitle("Delete")).not.toBeInTheDocument();
  });

  it("redacts tracking number and notes for private packages viewed by non-owner", () => {
    render(<PackageCard pkg={makePkg({ private: true, isOwner: false })} />);

    expect(screen.queryByText("1Z999AA1...6784")).not.toBeInTheDocument();
    expect(screen.queryByText("Leave at back door")).not.toBeInTheDocument();
  });

  it("shows tracking number and notes for private packages viewed by owner", () => {
    render(<PackageCard pkg={makePkg({ private: true, isOwner: true })} />);

    expect(screen.getByText("1Z999AA1...6784")).toBeInTheDocument();
    expect(screen.getByText("Leave at back door")).toBeInTheDocument();
  });

  it("renders Unarchive button for archived packages", () => {
    render(<PackageCard pkg={makePkg({ status: "archived" })} />);
    expect(screen.getByTitle("Unarchive")).toBeInTheDocument();
    expect(screen.queryByTitle("Archive")).not.toBeInTheDocument();
  });

  it("does not render expand button for redacted packages", () => {
    render(<PackageCard pkg={makePkg({ private: true, isOwner: false })} />);
    expect(screen.queryByTitle("Expand")).not.toBeInTheDocument();
  });

  it("displays 'Package' as label when no label provided", () => {
    render(<PackageCard pkg={makePkg({ label: null })} />);
    expect(screen.getByText("Package")).toBeInTheDocument();
  });
});
