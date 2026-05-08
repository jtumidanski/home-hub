import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";

const mockUseAuth = vi.fn();
const mockLogoutMutate = vi.fn();
const mockNavigate = vi.fn();

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => mockUseAuth(),
}));

vi.mock("@/lib/hooks/api/use-auth", () => ({
  useLogout: () => ({ mutate: mockLogoutMutate }),
}));

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual<typeof import("react-router-dom")>("react-router-dom");
  return { ...actual, useNavigate: () => mockNavigate };
});

import { UserMenu } from "../user-menu";

function renderUserMenu(props: { onAction?: () => void } = {}) {
  return render(
    <MemoryRouter>
      <UserMenu {...props} />
    </MemoryRouter>,
  );
}

describe("UserMenu", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAuth.mockReturnValue({
      user: {
        id: "test-user-id",
        attributes: {
          displayName: "Test User",
          email: "test@example.com",
          avatarUrl: "",
          providerAvatarUrl: "",
        },
      },
    });
  });

  it("renders nothing when there is no user", () => {
    mockUseAuth.mockReturnValue({ user: null });
    const { container } = renderUserMenu();
    expect(container).toBeEmptyDOMElement();
  });

  it("renders only Settings and Sign Out items in the pop-over, in that order", async () => {
    const user = userEvent.setup();
    renderUserMenu();
    await user.click(screen.getByRole("button", { name: /test user/i }));
    const items = await screen.findAllByRole("menuitem");
    expect(items).toHaveLength(2);
    expect(items[0]).toHaveTextContent(/settings/i);
    expect(items[1]).toHaveTextContent(/sign out/i);
  });

  it("does not render any theme toggle item", async () => {
    const user = userEvent.setup();
    renderUserMenu();
    await user.click(screen.getByRole("button", { name: /test user/i }));
    await screen.findByRole("menuitem", { name: /sign out/i });
    expect(screen.queryByRole("menuitem", { name: /dark mode/i })).toBeNull();
    expect(screen.queryByRole("menuitem", { name: /light mode/i })).toBeNull();
  });

  it("Settings item navigates to /app/settings and calls onAction", async () => {
    const user = userEvent.setup();
    const onAction = vi.fn();
    renderUserMenu({ onAction });
    await user.click(screen.getByRole("button", { name: /test user/i }));
    const settingsItem = await screen.findByRole("menuitem", { name: /settings/i });
    await user.click(settingsItem);
    expect(mockNavigate).toHaveBeenCalledTimes(1);
    expect(mockNavigate).toHaveBeenCalledWith("/app/settings");
    expect(onAction).toHaveBeenCalledTimes(1);
  });

  it("Sign Out item triggers logout and calls onAction", async () => {
    const user = userEvent.setup();
    const onAction = vi.fn();
    renderUserMenu({ onAction });
    await user.click(screen.getByRole("button", { name: /test user/i }));
    const signOutItem = await screen.findByRole("menuitem", { name: /sign out/i });
    await user.click(signOutItem);
    expect(mockLogoutMutate).toHaveBeenCalledTimes(1);
    expect(onAction).toHaveBeenCalledTimes(1);
  });
});
