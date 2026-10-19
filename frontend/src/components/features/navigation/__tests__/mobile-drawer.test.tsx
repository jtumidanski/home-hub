import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";

vi.mock("@/components/providers/auth-provider", () => ({
  useAuth: () => ({ user: null }),
}));
vi.mock("@/lib/hooks/api/use-auth", () => ({ useLogout: () => ({ mutate: vi.fn() }) }));
vi.mock("@/components/features/navigation/dashboards-nav-group", () => ({
  DashboardsNavGroup: () => <div>Dashboards</div>,
}));
vi.mock("@/components/features/households/mobile-household-selector", () => ({
  MobileHouseholdSelector: () => <div>Household selector</div>,
}));

import { MobileDrawer } from "../mobile-drawer";

function renderDrawer() {
  return render(
    <MemoryRouter>
      <SidebarProvider>
        <SidebarTrigger aria-label="Open navigation menu" />
        <MobileDrawer />
      </SidebarProvider>
    </MemoryRouter>,
  );
}

describe("MobileDrawer", () => {
  it("is not in the DOM (so not in the Tab order) while closed", () => {
    renderDrawer();
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(screen.queryByRole("link", { name: "Tasks" })).toBeNull();
  });

  it("opens as a labelled dialog and returns focus to the trigger on Escape", async () => {
    const user = userEvent.setup();
    renderDrawer();
    const trigger = screen.getByRole("button", { name: "Open navigation menu" });
    await user.click(trigger);
    expect(await screen.findByRole("dialog", { name: "Navigation" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Tasks" })).toBeInTheDocument();
    await user.keyboard("{Escape}");
    await vi.waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
    expect(trigger).toHaveFocus();
  });

  it("close button dismisses the drawer", async () => {
    const user = userEvent.setup();
    renderDrawer();
    await user.click(screen.getByRole("button", { name: "Open navigation menu" }));
    await user.click(await screen.findByRole("button", { name: "Close navigation menu" }));
    await vi.waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });
});
