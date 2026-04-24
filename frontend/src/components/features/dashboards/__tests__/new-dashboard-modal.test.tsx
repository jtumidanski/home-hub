import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { Dashboard } from "@/types/models/dashboard";

const mockNavigate = vi.fn();
const mockCreateMutateAsync = vi.fn();
const mockCopyToMineMutateAsync = vi.fn();
const mockUseDashboards = vi.fn();

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual<typeof import("react-router-dom")>("react-router-dom");
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("@/lib/hooks/api/use-dashboards", () => ({
  useDashboards: () => mockUseDashboards(),
  useCreateDashboard: () => ({ mutateAsync: mockCreateMutateAsync, isPending: false }),
  useCopyDashboardToMine: () => ({ mutateAsync: mockCopyToMineMutateAsync, isPending: false }),
}));

import { NewDashboardModal } from "../new-dashboard-modal";

function dash(overrides: Partial<Dashboard["attributes"]> & { id: string }): Dashboard {
  const { id, ...attrs } = overrides;
  return {
    id,
    type: "dashboards",
    attributes: {
      name: id,
      scope: "household",
      sortOrder: 0,
      layout: { version: 1, widgets: [{ id: "w-1", type: "clock", x: 0, y: 0, w: 3, h: 2, config: {} }] },
      schemaVersion: 1,
      createdAt: "2025-01-01T00:00:00Z",
      updatedAt: "2025-01-01T00:00:00Z",
      ...attrs,
    },
  };
}

function renderModal() {
  return render(
    <MemoryRouter>
      <NewDashboardModal open={true} onOpenChange={() => {}} />
    </MemoryRouter>,
  );
}

describe("NewDashboardModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseDashboards.mockReturnValue({ data: { data: [] } });
  });

  it("renders all form fields", () => {
    renderModal();
    expect(screen.getByLabelText(/name/i)).toBeInTheDocument();
    expect(screen.getByText("Household")).toBeInTheDocument();
    expect(screen.getByText("My Dashboards")).toBeInTheDocument();
    expect(screen.getByText(/copy from/i)).toBeInTheDocument();
  });

  it("rejects empty name", async () => {
    const user = userEvent.setup();
    renderModal();
    await user.click(screen.getByRole("button", { name: /create/i }));
    expect(await screen.findByText(/name is required/i)).toBeInTheDocument();
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
  });

  it("creates a blank household dashboard when no copy source", async () => {
    mockCreateMutateAsync.mockResolvedValue({
      data: dash({ id: "new-1", scope: "household", name: "Fresh" }),
    });
    const user = userEvent.setup();
    renderModal();
    await user.type(screen.getByLabelText(/name/i), "Fresh");
    await user.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() => {
      expect(mockCreateMutateAsync).toHaveBeenCalledWith({
        name: "Fresh",
        scope: "household",
        layout: { version: 1, widgets: [] },
      });
    });
    expect(mockNavigate).toHaveBeenCalledWith("/app/dashboards/new-1/edit");
  });

  it("calls copyDashboardToMine when copying household source as user", async () => {
    mockUseDashboards.mockReturnValue({
      data: { data: [dash({ id: "src-hh", name: "Source HH", scope: "household" })] },
    });
    mockCopyToMineMutateAsync.mockResolvedValue({
      data: dash({ id: "copy-1", scope: "user", name: "Copy" }),
    });

    const user = userEvent.setup();
    renderModal();

    await user.type(screen.getByLabelText(/name/i), "My Copy");
    // Switch scope to "user"
    await user.click(screen.getByText(/my dashboards/i));
    // Select the source
    await user.click(screen.getByRole("combobox"));
    await user.click(await screen.findByRole("option", { name: /source hh/i }));

    await user.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() => {
      expect(mockCopyToMineMutateAsync).toHaveBeenCalledWith("src-hh");
    });
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith("/app/dashboards/copy-1/edit");
  });
});
