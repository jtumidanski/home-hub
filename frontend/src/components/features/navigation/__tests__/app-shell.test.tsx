import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";

const mockUseAuth = vi.fn();
const mockToggleTheme = vi.fn();
const mockLogoutMutate = vi.fn();

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => mockUseAuth(),
}));

vi.mock("@/lib/hooks/use-theme-toggle", () => ({
  useThemeToggle: () => ({ theme: "light", toggleTheme: mockToggleTheme }),
}));

vi.mock("@/lib/hooks/api/use-auth", () => ({
  useLogout: () => ({ mutate: mockLogoutMutate }),
}));

vi.mock("@/components/features/households/household-switcher", () => ({
  HouseholdSwitcher: () => <div data-testid="household-switcher">Switcher</div>,
}));

vi.mock("@/components/features/navigation/mobile-header", () => ({
  MobileHeader: ({ onMenuOpen }: { onMenuOpen: () => void }) => (
    <button data-testid="mobile-header" onClick={onMenuOpen}>Menu</button>
  ),
}));

vi.mock("@/components/features/navigation/mobile-drawer", () => ({
  MobileDrawer: ({ open }: { open: boolean }) =>
    open ? <div data-testid="mobile-drawer">Drawer</div> : null,
}));

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return { ...actual, Outlet: () => <div data-testid="outlet">Page Content</div> };
});

import { AppShell } from "../app-shell";

function renderWithRouter() {
  return render(
    <MemoryRouter initialEntries={["/app"]}>
      <AppShell />
    </MemoryRouter>,
  );
}

describe("AppShell", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAuth.mockReturnValue({
      user: { attributes: { displayName: "Test User", email: "test@example.com" } },
    });
  });

  it("renders sidebar with navigation links", () => {
    renderWithRouter();
    expect(screen.getByText("Home Hub")).toBeInTheDocument();
    expect(screen.getByText("Dashboard")).toBeInTheDocument();
    expect(screen.getByText("Tasks")).toBeInTheDocument();
    expect(screen.getByText("Reminders")).toBeInTheDocument();
    expect(screen.getByText("Households")).toBeInTheDocument();
    expect(screen.getByText("Settings")).toBeInTheDocument();
  });

  it("renders household switcher", () => {
    renderWithRouter();
    expect(screen.getByTestId("household-switcher")).toBeInTheDocument();
  });

  it("renders user info in sidebar", () => {
    renderWithRouter();
    expect(screen.getByText("Test User")).toBeInTheDocument();
    expect(screen.getByText("test@example.com")).toBeInTheDocument();
  });

  it("renders outlet for page content", () => {
    renderWithRouter();
    expect(screen.getByTestId("outlet")).toBeInTheDocument();
  });

  it("calls toggleTheme when theme button is clicked", async () => {
    const user = userEvent.setup();
    renderWithRouter();
    await user.click(screen.getByRole("button", { name: /dark mode/i }));
    expect(mockToggleTheme).toHaveBeenCalledTimes(1);
  });

  it("calls logout when sign out button is clicked", async () => {
    const user = userEvent.setup();
    renderWithRouter();
    await user.click(screen.getByRole("button", { name: /sign out/i }));
    expect(mockLogoutMutate).toHaveBeenCalledTimes(1);
  });

  it("does not render user info when user is null", () => {
    mockUseAuth.mockReturnValue({ user: null });
    renderWithRouter();
    expect(screen.queryByText("Test User")).not.toBeInTheDocument();
  });
});
