import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CreatePackageDialog } from "../create-package-dialog";

const mockCreateMutate = vi.fn();
const mockDetectMutate = vi.fn();

vi.mock("@/lib/hooks/api/use-packages", () => ({
  useCreatePackage: () => ({
    mutate: mockCreateMutate,
    isPending: false,
  }),
  useDetectCarrier: () => ({
    mutate: mockDetectMutate,
    data: null,
  }),
}));

describe("CreatePackageDialog", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("does not render when closed", () => {
    render(<CreatePackageDialog open={false} onClose={vi.fn()} />);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("renders form fields when open", () => {
    render(<CreatePackageDialog open={true} onClose={vi.fn()} />);

    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Add Package" })).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Enter tracking number")).toBeInTheDocument();
    expect(screen.getByLabelText("Label (optional)")).toBeInTheDocument();
    expect(screen.getByLabelText("Notes (optional)")).toBeInTheDocument();
    expect(screen.getByLabelText(/Private/)).toBeInTheDocument();
  });

  it("shows validation error when submitting with empty tracking number", async () => {
    const user = userEvent.setup();
    render(<CreatePackageDialog open={true} onClose={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /add package/i }));

    await waitFor(() => {
      expect(screen.getByText("Tracking number is required")).toBeInTheDocument();
    });
    expect(mockCreateMutate).not.toHaveBeenCalled();
  });

  it("calls createMutation with correct data on valid submission", async () => {
    const user = userEvent.setup();
    render(<CreatePackageDialog open={true} onClose={vi.fn()} />);

    await user.clear(screen.getByPlaceholderText("Enter tracking number"));
    await user.type(screen.getByPlaceholderText("Enter tracking number"), "1Z999AA10123456784");
    await user.type(screen.getByLabelText("Label (optional)"), "My Package");
    await user.click(screen.getByRole("button", { name: /add package/i }));

    await waitFor(() => {
      expect(mockCreateMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          trackingNumber: "1Z999AA10123456784",
          carrier: "usps",
          label: "My Package",
        }),
        expect.any(Object),
      );
    });
  });

  it("triggers carrier detection on tracking number blur when length >= 8", async () => {
    const user = userEvent.setup();
    render(<CreatePackageDialog open={true} onClose={vi.fn()} />);

    const input = screen.getByPlaceholderText("Enter tracking number");
    await user.type(input, "1Z999AA10123456784");
    await user.tab();

    expect(mockDetectMutate).toHaveBeenCalledWith(
      "1Z999AA10123456784",
      expect.any(Object),
    );
  });

  it("does not trigger carrier detection when tracking number is too short", async () => {
    const user = userEvent.setup();
    render(<CreatePackageDialog open={true} onClose={vi.fn()} />);

    const input = screen.getByPlaceholderText("Enter tracking number");
    await user.type(input, "SHORT");
    await user.tab();

    expect(mockDetectMutate).not.toHaveBeenCalled();
  });
});
