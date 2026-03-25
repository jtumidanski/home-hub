import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";

const mockUseProviders = vi.fn();

vi.mock("@/lib/hooks/api/use-auth", () => ({
  useProviders: () => mockUseProviders(),
}));

vi.mock("@/services/api/auth", () => ({
  authService: {
    getLoginUrl: (provider: string) => `/api/v1/auth/login/${provider}?redirect=%2Fapp`,
  },
}));

import { LoginPage } from "../LoginPage";

describe("LoginPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseProviders.mockReturnValue({ data: null, isLoading: false, isError: false });
  });

  it("renders loading skeleton when isLoading is true", () => {
    mockUseProviders.mockReturnValue({ data: null, isLoading: true, isError: false });
    const { container } = render(<LoginPage />);
    expect(screen.getByText("Home Hub")).toBeInTheDocument();
    expect(container.querySelector(".animate-pulse")).toBeTruthy();
  });

  it("renders error state when isError is true", () => {
    mockUseProviders.mockReturnValue({ data: null, isLoading: false, isError: true });
    render(<LoginPage />);
    expect(screen.getByText(/failed to load login providers/i)).toBeInTheDocument();
  });

  it("renders provider login buttons", () => {
    mockUseProviders.mockReturnValue({
      data: {
        data: [
          { id: "google", type: "auth-providers", attributes: { displayName: "Google" } },
          { id: "github", type: "auth-providers", attributes: { displayName: "GitHub" } },
        ],
      },
      isLoading: false,
      isError: false,
    });
    render(<LoginPage />);
    expect(screen.getByText("Sign in with Google")).toBeInTheDocument();
    expect(screen.getByText("Sign in with GitHub")).toBeInTheDocument();
  });

  it("renders empty providers message when no providers exist", () => {
    mockUseProviders.mockReturnValue({ data: { data: [] }, isLoading: false, isError: false });
    render(<LoginPage />);
    expect(screen.getByText(/no login providers configured/i)).toBeInTheDocument();
  });

  it("renders correct login URLs", () => {
    mockUseProviders.mockReturnValue({
      data: {
        data: [
          { id: "google", type: "auth-providers", attributes: { displayName: "Google" } },
        ],
      },
      isLoading: false,
      isError: false,
    });
    render(<LoginPage />);
    const link = screen.getByText("Sign in with Google").closest("a");
    expect(link).toHaveAttribute("href", "/api/v1/auth/login/google?redirect=%2Fapp");
  });
});
