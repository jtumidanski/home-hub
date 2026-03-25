import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const mockUseAuth = vi.fn();
const mockToggleTheme = vi.fn();
const mockUseThemeToggle = vi.fn();

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => mockUseAuth(),
}));

vi.mock("@/lib/hooks/use-theme-toggle", () => ({
  useThemeToggle: () => mockUseThemeToggle(),
}));

import { SettingsPage } from "../SettingsPage";

describe("SettingsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseThemeToggle.mockReturnValue({ theme: "light", toggleTheme: mockToggleTheme });
    mockUseAuth.mockReturnValue({
      user: {
        id: "u-1",
        type: "users",
        attributes: { displayName: "Jane Doe", email: "jane@example.com" },
      },
      appContext: {
        attributes: { resolvedRole: "owner" },
        relationships: {},
      },
      isLoading: false,
    });
  });

  it("renders loading skeleton when isLoading is true", () => {
    mockUseAuth.mockReturnValue({ user: null, appContext: null, isLoading: true });
    const { container } = render(<SettingsPage />);
    expect(screen.queryByText("Settings")).not.toBeInTheDocument();
    expect(container.querySelector(".animate-pulse")).toBeTruthy();
  });

  it("renders error state when user or appContext is missing", () => {
    mockUseAuth.mockReturnValue({ user: null, appContext: null, isLoading: false });
    render(<SettingsPage />);
    expect(screen.getByText(/failed to load settings/i)).toBeInTheDocument();
  });

  it("renders profile information", () => {
    render(<SettingsPage />);
    expect(screen.getByText("Settings")).toBeInTheDocument();
    expect(screen.getByText("Jane Doe")).toBeInTheDocument();
    expect(screen.getByText("jane@example.com")).toBeInTheDocument();
    expect(screen.getByText("owner")).toBeInTheDocument();
  });

  it("renders theme toggle button with correct label", () => {
    render(<SettingsPage />);
    expect(screen.getByRole("button", { name: /switch to dark mode/i })).toBeInTheDocument();
  });

  it("calls toggleTheme when button is clicked", async () => {
    const user = userEvent.setup();
    render(<SettingsPage />);
    await user.click(screen.getByRole("button", { name: /switch to dark mode/i }));
    expect(mockToggleTheme).toHaveBeenCalled();
  });
});
