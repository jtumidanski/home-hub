import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { Settings } from "lucide-react";
import { SidebarProvider } from "@/components/ui/sidebar";
import { NavGroup } from "../nav-group";

const group = {
  key: "home",
  label: "Home",
  items: [
    { to: "/app/tasks", icon: Settings, label: "Tasks" },
    { to: "/app/reminders", icon: Settings, label: "Reminders" },
  ],
};

function renderGroup(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <SidebarProvider>
        <NavGroup group={group} isOpen onToggle={() => {}} />
      </SidebarProvider>
    </MemoryRouter>,
  );
}

describe("NavGroup", () => {
  it("marks only the active item with aria-current=page", () => {
    renderGroup("/app/tasks");
    expect(screen.getByRole("link", { name: "Tasks" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: "Reminders" })).not.toHaveAttribute("aria-current");
  });

  it("group trigger is a 44px target on coarse pointers", () => {
    renderGroup("/app/tasks");
    expect(screen.getByRole("button", { name: "Home" })).toHaveClass("coarse:min-h-11");
  });
});
