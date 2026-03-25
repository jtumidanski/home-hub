import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";

const mockUseAuth = vi.fn();
const mockCreateTenantMutateAsync = vi.fn();
const mockCreateHouseholdMutateAsync = vi.fn();
const mockNavigate = vi.fn();

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => mockUseAuth(),
}));

vi.mock("@/lib/hooks/api/use-context", () => ({
  useCreateTenant: () => ({ mutateAsync: mockCreateTenantMutateAsync }),
  useOnboardingCreateHousehold: () => ({ mutateAsync: mockCreateHouseholdMutateAsync }),
}));

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock("@/lib/api/errors", () => ({
  createErrorFromUnknown: (_err: unknown, fallback: string) => ({ message: fallback, type: "unknown" }),
}));

import { OnboardingPage } from "../OnboardingPage";

function renderPage() {
  return render(
    <MemoryRouter>
      <OnboardingPage />
    </MemoryRouter>,
  );
}

describe("OnboardingPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAuth.mockReturnValue({
      user: { attributes: { displayName: "Test User", email: "test@example.com" } },
    });
  });

  it("renders tenant creation step initially", () => {
    renderPage();
    expect(screen.getByText("Welcome to Home Hub")).toBeInTheDocument();
    expect(screen.getByText("Let's set up your account")).toBeInTheDocument();
    expect(screen.getByText("Account Name")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("e.g., The Smith Family")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /continue/i })).toBeInTheDocument();
  });

  it("pre-fills tenant name from user display name", () => {
    renderPage();
    const input = screen.getByPlaceholderText("e.g., The Smith Family");
    expect(input).toHaveValue("Test User's Home");
  });

  it("shows validation error when submitting empty tenant name", async () => {
    const user = userEvent.setup();
    mockUseAuth.mockReturnValue({ user: { attributes: { displayName: "", email: "" } } });
    renderPage();

    const input = screen.getByPlaceholderText("e.g., The Smith Family");
    await user.clear(input);
    await user.click(screen.getByRole("button", { name: /continue/i }));
    await waitFor(() => {
      expect(screen.getByText("Name is required")).toBeInTheDocument();
    });
  });

  it("advances to household step after successful tenant creation", async () => {
    const user = userEvent.setup();
    mockCreateTenantMutateAsync.mockResolvedValue({
      data: { id: "t-1", type: "tenants", attributes: { name: "Test User's Home" } },
    });

    renderPage();
    await user.click(screen.getByRole("button", { name: /continue/i }));

    await waitFor(() => {
      expect(screen.getByText("Now create your first household")).toBeInTheDocument();
    });
    expect(screen.getByText("Household Name")).toBeInTheDocument();
  });

  it("shows error toast when tenant creation fails", async () => {
    const user = userEvent.setup();
    const { toast } = await import("sonner");
    mockCreateTenantMutateAsync.mockRejectedValue(new Error("fail"));

    renderPage();
    await user.click(screen.getByRole("button", { name: /continue/i }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith("Failed to create account");
    });
  });

  it("navigates to /app after household creation", async () => {
    const user = userEvent.setup();
    mockCreateTenantMutateAsync.mockResolvedValue({
      data: { id: "t-1", type: "tenants", attributes: { name: "Home" } },
    });
    mockCreateHouseholdMutateAsync.mockResolvedValue({
      data: { id: "h-1", type: "households", attributes: { name: "Main Home" } },
    });

    renderPage();
    await user.click(screen.getByRole("button", { name: /continue/i }));
    await waitFor(() => {
      expect(screen.getByText("Household Name")).toBeInTheDocument();
    });
    await user.click(screen.getByRole("button", { name: /get started/i }));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith("/app");
    });
  });
});
