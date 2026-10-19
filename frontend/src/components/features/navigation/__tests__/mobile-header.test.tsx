import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { SidebarProvider } from "@/components/ui/sidebar";
import { MobileHeader } from "../mobile-header";

function renderHeader() {
  return render(
    <SidebarProvider>
      <MobileHeader />
    </SidebarProvider>,
  );
}

describe("MobileHeader", () => {
  it("menu button has an accessible name and a 44px coarse-pointer target", () => {
    renderHeader();
    const button = screen.getByRole("button", { name: "Open navigation menu" });
    expect(button).toHaveClass("coarse:size-11");
  });

  it("renders the brand mark and a Fraunces wordmark", () => {
    renderHeader();
    expect(screen.getByRole("img", { name: "Home Hub" })).toBeInTheDocument();
    expect(screen.getByText("Home Hub")).toHaveClass("font-heading", "text-lg");
  });
});
