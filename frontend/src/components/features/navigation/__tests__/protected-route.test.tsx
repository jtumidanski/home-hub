import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

const mockUseAuth = vi.fn();

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => mockUseAuth(),
}));

import { ProtectedRoute } from "../protected-route";

function renderWithRouter(ui: React.ReactElement) {
  return render(<MemoryRouter>{ui}</MemoryRouter>);
}

describe("ProtectedRoute", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders loading skeleton when auth is loading", () => {
    mockUseAuth.mockReturnValue({ isAuthenticated: false, isLoading: true, needsOnboarding: false });
    renderWithRouter(
      <ProtectedRoute><div>Protected Content</div></ProtectedRoute>,
    );
    expect(screen.queryByText("Protected Content")).not.toBeInTheDocument();
  });

  it("redirects to /login when not authenticated", () => {
    mockUseAuth.mockReturnValue({ isAuthenticated: false, isLoading: false, needsOnboarding: false });
    renderWithRouter(
      <ProtectedRoute><div>Protected Content</div></ProtectedRoute>,
    );
    expect(screen.queryByText("Protected Content")).not.toBeInTheDocument();
  });

  it("redirects to /onboarding when needsOnboarding is true", () => {
    mockUseAuth.mockReturnValue({ isAuthenticated: true, isLoading: false, needsOnboarding: true });
    renderWithRouter(
      <ProtectedRoute><div>Protected Content</div></ProtectedRoute>,
    );
    expect(screen.queryByText("Protected Content")).not.toBeInTheDocument();
  });

  it("renders children when authenticated and no onboarding needed", () => {
    mockUseAuth.mockReturnValue({ isAuthenticated: true, isLoading: false, needsOnboarding: false });
    renderWithRouter(
      <ProtectedRoute><div>Protected Content</div></ProtectedRoute>,
    );
    expect(screen.getByText("Protected Content")).toBeInTheDocument();
  });
});
