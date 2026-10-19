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

const mockMutate = vi.fn();
vi.mock("@/lib/hooks/api/use-auth", () => ({
  useUpdateMe: () => ({ mutate: mockMutate, isPending: false }),
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
        attributes: { displayName: "Jane Doe", email: "jane@example.com", avatarUrl: "", providerAvatarUrl: "" },
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
    render(<SettingsPage />);
    expect(screen.getByRole("status", { name: "Loading" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Settings" })).toBeInTheDocument();
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
    expect(screen.getByText("Owner")).toBeInTheDocument();
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

  it("names avatar tiles by position and exposes pressed state", async () => {
    const user = userEvent.setup();
    render(<SettingsPage />);
    const tile = screen.getByRole("button", { name: "Choose avatar 1" });
    expect(tile).toHaveAttribute("aria-pressed", "false");
    expect(tile.querySelector("img")).toHaveAttribute("alt", "");
    expect(screen.getByRole("button", { name: "Adventurer" })).toHaveAttribute("aria-pressed", "true");
    await user.click(screen.getByRole("button", { name: "Robots" }));
    expect(screen.getByRole("button", { name: "Robots" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "Adventurer" })).toHaveAttribute("aria-pressed", "false");
  });

  it("marks the selected avatar tile as pressed", () => {
    mockUseAuth.mockReturnValue({
      user: {
        id: "u-1",
        type: "users",
        attributes: { displayName: "Jane Doe", email: "jane@example.com", avatarUrl: "dicebear:adventurer:seed2", providerAvatarUrl: "" },
      },
      appContext: { attributes: { resolvedRole: "owner" }, relationships: {} },
      isLoading: false,
    });
    render(<SettingsPage />);
    expect(screen.getByRole("button", { name: "Choose avatar 3" })).toHaveAttribute("aria-pressed", "true");
  });

  it("shows the current theme text", () => {
    render(<SettingsPage />);
    expect(screen.getByText("Current theme: Light")).toBeInTheDocument();
  });
});
