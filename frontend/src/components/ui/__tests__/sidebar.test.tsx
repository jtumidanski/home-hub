import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import {
  SidebarProvider,
  SidebarMenuButton,
  SidebarTrigger,
  useSidebar,
} from "@/components/ui/sidebar";

function OpenState() {
  const { openMobile } = useSidebar();
  return <span data-testid="open">{openMobile ? "open" : "closed"}</span>;
}

describe("Sidebar primitives", () => {
  it("SidebarTrigger toggles the shared mobile open state", async () => {
    const user = userEvent.setup();
    render(
      <SidebarProvider>
        <SidebarTrigger />
        <OpenState />
      </SidebarProvider>,
    );

    expect(screen.getByTestId("open")).toHaveTextContent("closed");
    await user.click(screen.getByRole("button", { name: /toggle navigation/i }));
    expect(screen.getByTestId("open")).toHaveTextContent("open");
    await user.click(screen.getByRole("button", { name: /toggle navigation/i }));
    expect(screen.getByTestId("open")).toHaveTextContent("closed");
  });

  it("SidebarMenuButton asChild forwards active styling onto the child element", () => {
    render(
      <SidebarProvider>
        <SidebarMenuButton asChild isActive className="extra-class">
          <a href="/x">Item</a>
        </SidebarMenuButton>
      </SidebarProvider>,
    );

    const link = screen.getByRole("link", { name: "Item" });
    expect(link).toHaveAttribute("data-slot", "sidebar-menu-button");
    expect(link).toHaveAttribute("data-active", "true");
    expect(link.className).toContain("extra-class");
  });

  it("SidebarMenuButton asChild omits data-active when inactive", () => {
    render(
      <SidebarProvider>
        <SidebarMenuButton asChild>
          <a href="/y">Idle</a>
        </SidebarMenuButton>
      </SidebarProvider>,
    );

    expect(screen.getByRole("link", { name: "Idle" })).not.toHaveAttribute("data-active");
  });
});
